// The Sleuth Kit (TSK) based disk analysis logic migrated from Tracium
package tsk

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ilexum-group/bitex/internal/logger"
	"github.com/ilexum-group/bitex/internal/models"
)

// AnalyzeDisk performs metadata-only analysis of a disk using TSK tools
func AnalyzeDisk(diskPath string) (*models.TSKAnalysis, error) {
	logger.Info("Starting disk analysis", map[string]string{"diskPath": diskPath})

	analysis := &models.TSKAnalysis{
		DiskPath:          diskPath,
		AnalysisTimestamp: time.Now().Unix(),
		ToolVersions:      make(map[string]string),
		CommandLogs:       []models.TSKCommandLog{},
		Errors:            []string{},
	}

	if err := getToolVersions(analysis); err != nil {
		analysis.Errors = append(analysis.Errors, fmt.Sprintf("Version check failed: %v", err))
		logger.Error("Tool version check failed", map[string]string{"error": err.Error()})
	}

	offset, err := detectOffset(diskPath, analysis)
	if err != nil {
		analysis.Errors = append(analysis.Errors, fmt.Sprintf("mmls failed: %v", err))
		logger.Error("Failed to detect partition offset", map[string]string{"error": err.Error()})
		return analysis, nil
	}
	logger.Info("Detected partition offset", map[string]string{"offset": strconv.FormatUint(offset, 10)})

	if fsStats, err := runFsstat(diskPath, offset, analysis); err != nil {
		analysis.Errors = append(analysis.Errors, fmt.Sprintf("fsstat failed: %v", err))
		logger.Error("Filesystem stats analysis failed", map[string]string{"error": err.Error()})
	} else {
		analysis.FilesystemStats = fsStats
		logger.Info("Filesystem stats collected", map[string]string{"filesystemType": fsStats.FilesystemType})
	}

	if fileListing, err := runFls(diskPath, offset, analysis); err != nil {
		analysis.Errors = append(analysis.Errors, fmt.Sprintf("fls failed: %v", err))
		logger.Error("File listing failed", map[string]string{"error": err.Error()})
	} else {
		analysis.FileListing = fileListing
		logger.Info("File listing completed", map[string]string{"fileCount": strconv.Itoa(len(fileListing))})
	}

	if err := runIstat(diskPath, offset, analysis, analysis.FileListing); err != nil {
		analysis.Errors = append(analysis.Errors, fmt.Sprintf("istat failed: %v", err))
		logger.Error("Inode metadata analysis failed", map[string]string{"error": err.Error()})
	} else {
		logger.Info("Deletion times updated for deleted files", map[string]string{})
	}

	logger.Info("Disk analysis completed", map[string]string{"diskPath": diskPath})
	return analysis, nil
}

// -------------------
// Offset detection
// -------------------

func detectOffset(diskPath string, analysis *models.TSKAnalysis) (uint64, error) {
	output, err := runCommandWithTimeout("mmls", []string{diskPath}, analysis, 30*time.Second)
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Example:
		// 002:  000:000   0000000032   0060620799   0060620768   Win95 FAT32
		if len(line) == 0 || strings.HasPrefix(line, "Slot") || strings.HasPrefix(line, "Meta") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		start, err := strconv.ParseUint(fields[2], 10, 64)
		if err == nil && start > 0 {
			return start, nil
		}
	}

	return 0, fmt.Errorf("no valid partition offset found")
}

// -------------------
// Tool versions
// -------------------

func getToolVersions(analysis *models.TSKAnalysis) error {
	tools := []string{"mmls", "fsstat", "fls", "istat"}
	for _, tool := range tools {
		version, err := runCommandWithTimeout(tool, []string{"-V"}, analysis, 10*time.Second)
		if err != nil {
			return fmt.Errorf("failed to get %s version: %w", tool, err)
		}
		lines := strings.Split(strings.TrimSpace(version), "\n")
		if len(lines) > 0 {
			analysis.ToolVersions[tool] = strings.TrimSpace(lines[0])
		}
	}
	return nil
}

// -------------------
// fsstat
// -------------------

func runFsstat(diskPath string, offset uint64, analysis *models.TSKAnalysis) (*models.TSKFilesystemStats, error) {
	args := []string{"-o", strconv.FormatUint(offset, 10), diskPath}
	output, err := runCommandWithTimeout("fsstat", args, analysis, 30*time.Second)
	if err != nil {
		return nil, err
	}

	stats := &models.TSKFilesystemStats{}
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		clusterSize := 0

		switch {
		case strings.Contains(line, "File System Type:"):
			stats.FilesystemType = strings.TrimSpace(strings.Split(line, ":")[1])

		case strings.Contains(line, "Block Size:"):
			if v, err := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1])); err == nil {
				stats.BlockSize = v
			}

		case strings.Contains(line, "Block Count:"):
			if v, err := strconv.ParseInt(strings.TrimSpace(strings.Split(line, ":")[1]), 10, 64); err == nil {
				stats.BlockCount = v
			}

		case strings.Contains(line, "Free Blocks:"):
			if v, err := strconv.ParseInt(strings.TrimSpace(strings.Split(line, ":")[1]), 10, 64); err == nil {
				stats.FreeBlocks = v
			}

		case strings.Contains(line, "Inode Count:"):
			if v, err := strconv.ParseInt(strings.TrimSpace(strings.Split(line, ":")[1]), 10, 64); err == nil {
				stats.InodeCount = v
			}

		case strings.Contains(line, "Free Inodes:"):
			if v, err := strconv.ParseInt(strings.TrimSpace(strings.Split(line, ":")[1]), 10, 64); err == nil {
				stats.FreeInodes = v
			}

		// FAT32-specific parsing
		case strings.HasPrefix(line, "Sector Size:"):
			if v, err := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1])); err == nil {
				stats.BlockSize = v
			}

		case strings.HasPrefix(line, "Cluster Size:"):
			if v, err := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1])); err == nil {
				clusterSize = v
			}
			// opcional: usar clusterSize como referencia de bloque lógico
			if stats.BlockSize == 0 {
				stats.BlockSize = clusterSize
			}

		case strings.HasPrefix(line, "Total Cluster Range:"):
			parts := strings.Split(strings.TrimSpace(strings.Split(line, ":")[1]), "-")
			if len(parts) == 2 {
				start, err1 := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
				end, err2 := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err1 == nil && err2 == nil {
					stats.BlockCount = end - start + 1
				}
			}

		case strings.HasPrefix(line, "Free Sector Count"):
			if v, err := strconv.ParseInt(strings.TrimSpace(strings.Split(line, ":")[1]), 10, 32); err == nil {
				// convertir a bloques usando cluster size / sector size
				if stats.BlockSize > 0 {
					stats.FreeBlocks = v / int64(clusterSize/stats.BlockSize)
				} else {
					stats.FreeBlocks = v
				}
			}
		}
	}

	return stats, nil
}

// -------------------
// fls
// -------------------

func runFls(diskPath string, offset uint64, analysis *models.TSKAnalysis) ([]models.TSKFileEntry, error) {
	args := []string{
		"-o", strconv.FormatUint(offset, 10),
		"-r", "-m", "/", "-d", "-p",
		diskPath,
	}

	output, err := runCommandWithTimeout("fls", args, analysis, 30*time.Second)
	if err != nil {
		return nil, err
	}

	var files []models.TSKFileEntry
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		isDeleted := strings.HasPrefix(line, "*") || strings.Contains(line, "(deleted)")
		if isDeleted {
			line = strings.TrimPrefix(line, "*")
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		fsID := parts[0]

		var inode uint64
		if fsID == "0" {
			inode, err = strconv.ParseUint(parts[2], 10, 64)
		} else {
			inode, err = strconv.ParseUint(parts[0], 10, 64)
		}
		if err != nil {
			continue
		}

		size, _ := strconv.ParseInt(parts[6], 10, 64) // este es el tamaño que a veces sale 0
		modified, _ := strconv.ParseInt(parts[7], 10, 64)
		accessed, _ := strconv.ParseInt(parts[8], 10, 64)
		created, _ := strconv.ParseInt(parts[10], 10, 64)
		uid, _ := strconv.Atoi(parts[4])
		gid, _ := strconv.Atoi(parts[5])

		files = append(files, models.TSKFileEntry{
			Path:         parts[1],
			Inode:        inode,
			Permissions:  parts[3],
			Deleted:      isDeleted,
			Size:         size,
			ModifiedTime: modified,
			AccessedTime: accessed,
			CreatedTime:  created,
			UID:          uid,
			GID:          gid,
		})
	}

	return files, nil
}

// -------------------
// istat
// -------------------

func runIstat(
	diskPath string,
	offset uint64,
	analysis *models.TSKAnalysis,
	fileListing []models.TSKFileEntry,
) error {

	inodeSet := make(map[uint64]bool)
	for _, f := range fileListing {
		if f.Deleted {
			inodeSet[f.Inode] = true
		}
	}

	for inode := range inodeSet {
		args := []string{
			"-o", strconv.FormatUint(offset, 10),
			diskPath,
			strconv.FormatUint(inode, 10),
		}

		output, err := runCommandWithTimeout("istat", args, analysis, 30*time.Second)
		if err != nil {
			continue
		}

		var deletionTime int64
		scanner := bufio.NewScanner(strings.NewReader(output))

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// Deleted Time
			if strings.Contains(line, "Deleted Time:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					if v, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
						deletionTime = v
					}
				}
			}
		}

		// Set DeletionTime in fileListing
		for i := range fileListing {
			if fileListing[i].Inode == inode && fileListing[i].Deleted {
				fileListing[i].DeletionTime = deletionTime
				break
			}
		}
	}

	return nil
}

// -------------------
// Command runner
// -------------------

func runCommandWithTimeout(command string, args []string, analysis *models.TSKAnalysis, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.Output()

	exitCode := 0
	var errorMsg string

	// si el contexto expiró (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		errorMsg = fmt.Sprintf("command timed out after %s", timeout)
		// registrar log aunque haya timeout
		logEntry := models.TSKCommandLog{
			Command:    command,
			Arguments:  args,
			Timestamp:  time.Now().Unix(),
			ExitCode:   0,
			OutputSize: len(output),
			Error:      errorMsg,
		}
		analysis.CommandLogs = append(analysis.CommandLogs, logEntry)

		// devolvemos output parcial si lo hay, con warning
		if len(output) > 0 {
			return string(output), nil
		}
		return "", fmt.Errorf(errorMsg)
	}

	// manejar otros errores normales
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
			errorMsg = exitError.Error()
		} else {
			exitCode = -1
			errorMsg = err.Error()
		}
	}

	// registrar log
	logEntry := models.TSKCommandLog{
		Command:    command,
		Arguments:  args,
		Timestamp:  time.Now().Unix(),
		ExitCode:   exitCode,
		OutputSize: len(output),
		Error:      errorMsg,
	}
	analysis.CommandLogs = append(analysis.CommandLogs, logEntry)

	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}
