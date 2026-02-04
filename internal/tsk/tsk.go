// Package tsk provides The Sleuth Kit (TSK) based disk analysis logic.
// This package encapsulates all TSK tool interactions (mmls, fsstat, fls, istat) for forensic disk analysis.
package tsk

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	internalos "github.com/ilexum-group/bitex/internal/os"
	"github.com/ilexum-group/bitex/internal/utils"
	"github.com/ilexum-group/bitex/pkg/models"
)

// ============================================================================
// TSKAnalyzer Structure
// ============================================================================

// Analyzer encapsulates TSK analysis functionality with custody chain logging.
type Analyzer struct {
	custodyChainEntry *models.CustodyChainEntry
	osImpl            *internalos.OS
}

// NewTSKAnalyzer creates a new Analyzer instance.
// Parameters:
//   - custodyChainEntry: Custody chain for logging analysis operations
//   - osImpl: Operating system abstraction (currently unused, reserved for future use)
func NewTSKAnalyzer(custodyChainEntry *models.CustodyChainEntry, osImpl *internalos.OS) *Analyzer {
	return &Analyzer{
		custodyChainEntry: custodyChainEntry,
		osImpl:            osImpl,
	}
}

// ============================================================================
// Main Analysis Function
// ============================================================================

// AnalyzeDisk performs comprehensive metadata-only disk analysis using TSK tools.
// This function orchestrates the complete analysis pipeline:
//  1. Check TSK tool versions
//  2. Detect all partitions using mmls
//  3. For each partition:
//     - Extract filesystem statistics using fsstat
//     - List all files (including deleted) using fls
//     - Get deletion times for deleted files using istat
//
// Parameters:
//   - diskPath: Path to the disk image or block device to analyze
//
// Returns:
//   - TSKAnalysis structure containing all collected metadata
//   - Error if critical failure occurs (non-critical errors are logged but don't stop analysis)
func (t *Analyzer) AnalyzeDisk(diskPath string) (*models.TSKAnalysis, error) {
	t.custodyChainEntry.LogInfo("DiskAnalysisStart", fmt.Sprintf("Starting TSK analysis: %s", diskPath))

	analysis := &models.TSKAnalysis{
		DiskPath:     diskPath,
		ToolVersions: make(map[string]string),
		Partitions:   []models.PartitionAnalysis{},
	}

	// Step 1: Get tool versions
	if err := t.getToolVersions(analysis); err != nil {
		t.custodyChainEntry.LogError("ToolVersionCheck", "Failed to get TSK tool versions", err)
	}

	// Step 2: Detect all partitions
	partitions, err := t.detectPartitions(diskPath)
	if err != nil {
		t.custodyChainEntry.LogError("PartitionDetection", "Failed to detect partitions", err)
		return analysis, nil
	}
	t.custodyChainEntry.LogInfo("PartitionsDetected", fmt.Sprintf("Found %d partitions", len(partitions)))

	// Step 3-5: Analyze each partition
	for _, partition := range partitions {
		t.custodyChainEntry.LogInfo("PartitionAnalysisStart", fmt.Sprintf("Analyzing partition %d (offset: %d)", partition.PartitionNumber, partition.StartSector))

		partAnalysis := models.PartitionAnalysis{
			PartitionNumber: partition.PartitionNumber,
			StartSector:     partition.StartSector,
			EndSector:       partition.EndSector,
			Length:          partition.Length,
			Description:     partition.Description,
		}

		// Get filesystem statistics
		if fsStats, err := t.runFsstat(diskPath, partition.StartSector, analysis); err != nil {
			t.custodyChainEntry.LogError("FilesystemStats", fmt.Sprintf("Failed for partition %d", partition.PartitionNumber), err)
		} else {
			partAnalysis.FilesystemStats = fsStats
			t.custodyChainEntry.LogInfo("FilesystemStats", fmt.Sprintf("Partition %d - Type: %s", partition.PartitionNumber, fsStats.FilesystemType))
		}

		// List all files
		if fileListing, err := t.runFls(diskPath, partition.StartSector, analysis, partition.PartitionNumber); err != nil {
			t.custodyChainEntry.LogError("FileListing", fmt.Sprintf("Failed for partition %d", partition.PartitionNumber), err)
		} else {
			partAnalysis.FileListing = fileListing
			t.custodyChainEntry.LogInfo("FileListingComplete", fmt.Sprintf("Partition %d - Found %d files", partition.PartitionNumber, len(fileListing)))
		}

		// Get deletion times for deleted files
		t.runIstat(diskPath, partition.StartSector, analysis, partAnalysis.FileListing)
		t.custodyChainEntry.LogInfo("DeletionTimes", fmt.Sprintf("Partition %d - Updated deletion times", partition.PartitionNumber))

		analysis.Partitions = append(analysis.Partitions, partAnalysis)
	}

	// Populate legacy fields with first partition for backward compatibility
	if len(analysis.Partitions) > 0 {
		analysis.FilesystemStats = analysis.Partitions[0].FilesystemStats
		analysis.FileListing = analysis.Partitions[0].FileListing
	}

	t.custodyChainEntry.LogInfo("DiskAnalysisComplete", fmt.Sprintf("Completed TSK analysis: %s", diskPath))
	return analysis, nil
}

// ============================================================================
// TSK Tool Version Detection
// ============================================================================

// getToolVersions queries all TSK tools for their version information.
// This ensures the custody chain documents which tool versions were used.
//
// Parameters:
//   - analysis: Analysis structure where tool versions are stored
//
// Returns error if any tool version cannot be retrieved.
func (t *Analyzer) getToolVersions(analysis *models.TSKAnalysis) error {
	tools := []string{"mmls", "fsstat", "fls", "istat"}

	for _, tool := range tools {
		version, err := t.runCommandWithTimeout(tool, []string{"-V"}, 10*time.Second)
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

// ============================================================================
// Partition Detection (mmls)
// ============================================================================
// detectPartitions uses mmls to detect all partitions on the disk.
// Skips unallocated space and meta entries.
//
// Parameters:
//   - diskPath: Path to disk image
//
// Returns:
//   - Array of PartitionInfo structures for all valid partitions
//   - Error if mmls fails
func (t *Analyzer) detectPartitions(diskPath string) ([]models.PartitionInfo, error) {
	output, err := t.runCommandWithTimeout("mmls", []string{diskPath}, 30*time.Second)
	if err != nil {
		return nil, err
	}

	var partitions []models.PartitionInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	partitionNumber := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip header lines, empty lines, meta entries, and unallocated space
		if len(line) == 0 || strings.HasPrefix(line, "Slot") ||
			strings.Contains(line, "Meta") || strings.Contains(line, "-------") {
			continue
		}

		// Parse partition table line: "002:  000:000   0000002048   0001126399   0001124352   NTFS / exFAT (0x07)"
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Parse start sector (field 2), end sector (field 3), length (field 4)
		start, err1 := strconv.ParseUint(fields[2], 10, 64)
		end, err2 := strconv.ParseUint(fields[3], 10, 64)
		length, err3 := strconv.ParseUint(fields[4], 10, 64)

		if err1 == nil && err2 == nil && err3 == nil && start > 0 {
			// Get description (rest of the fields)
			description := strings.Join(fields[5:], " ")

			partitions = append(partitions, models.PartitionInfo{
				PartitionNumber: partitionNumber,
				StartSector:     start,
				EndSector:       end,
				Length:          length,
				Description:     description,
			})
			partitionNumber++
		}
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("no valid partitions found in mmls output")
	}

	return partitions, nil
}

// ============================================================================
// Filesystem Statistics (fsstat)
// ============================================================================

// runFsstat executes fsstat to collect filesystem metadata.
// Supports both ext4/NTFS and FAT32 filesystem types.
//
// Parameters:
//   - diskPath: Path to disk image
//   - offset: Partition offset in sectors
//   - analysis: Analysis structure for command logging
//
// Returns:
//   - FilesystemStats structure with collected metadata
//   - Error if fsstat execution fails
func (t *Analyzer) runFsstat(diskPath string, offset uint64, _ *models.TSKAnalysis) (*models.TSKFilesystemStats, error) {
	args := []string{"-o", strconv.FormatUint(offset, 10), diskPath}
	output, err := t.runCommandWithTimeout("fsstat", args, 30*time.Second)
	if err != nil {
		return nil, err
	}

	stats := &models.TSKFilesystemStats{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	clusterSize := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		// Common filesystem fields
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

		// FAT32-specific fields
		case strings.HasPrefix(line, "Sector Size:"):
			if v, err := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1])); err == nil {
				stats.BlockSize = v
			}

		case strings.HasPrefix(line, "Cluster Size:"):
			if v, err := strconv.Atoi(strings.TrimSpace(strings.Split(line, ":")[1])); err == nil {
				clusterSize = v
				// Use cluster size as block size if not set
				if stats.BlockSize == 0 {
					stats.BlockSize = clusterSize
				}
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
				// Convert sectors to blocks using cluster/sector size ratio
				if stats.BlockSize > 0 && clusterSize > 0 {
					stats.FreeBlocks = v / int64(clusterSize/stats.BlockSize)
				} else {
					stats.FreeBlocks = v
				}
			}
		}
	}

	return stats, nil
}

// ============================================================================
// File Listing (fls)
// ============================================================================

// runFls executes fls to list all files (including deleted) in the filesystem.
// Uses machine-readable output format (-m flag) for reliable parsing.
//
// Parameters:
//   - diskPath: Path to disk image
//   - offset: Partition offset in sectors
//   - analysis: Analysis structure for command logging
//   - partitionNumber: Number of the partition being analyzed
//
// Returns:
//   - Array of TSKFileEntry structures
//   - Error if fls execution fails
func (t *Analyzer) runFls(diskPath string, offset uint64, _ *models.TSKAnalysis, partitionNumber int) ([]models.TSKFileEntry, error) {
	args := []string{
		"-o", strconv.FormatUint(offset, 10),
		"-r",      // Recursive
		"-m", "/", // Machine-readable format
		"-d", // Include deleted files
		"-p", // Display full path
		diskPath,
	}

	output, err := t.runCommandWithTimeout("fls", args, 30*time.Second)
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

		// Check if file is deleted (marked with * or contains "(deleted)")
		isDeleted := strings.HasPrefix(line, "*") || strings.Contains(line, "(deleted)")
		if isDeleted {
			line = strings.TrimPrefix(line, "*")
		}

		// Parse pipe-separated fields from machine-readable output
		parts := strings.Split(line, "|")
		if len(parts) < 11 {
			continue
		}

		// Parse inode number (handle special case where fsID is 0)
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

		// Parse file metadata
		size, _ := strconv.ParseInt(parts[6], 10, 64)
		modified, _ := strconv.ParseInt(parts[7], 10, 64)
		accessed, _ := strconv.ParseInt(parts[8], 10, 64)
		created, _ := strconv.ParseInt(parts[10], 10, 64)
		uid, _ := strconv.Atoi(parts[4])
		gid, _ := strconv.Atoi(parts[5])

		files = append(files, models.TSKFileEntry{
			Path:            parts[1],
			Inode:           inode,
			Permissions:     parts[3],
			Deleted:         isDeleted,
			Size:            size,
			ModifiedTime:    modified,
			AccessedTime:    accessed,
			CreatedTime:     created,
			UID:             uid,
			GID:             gid,
			PartitionNumber: partitionNumber,
		})
	}

	return files, nil
}

// ============================================================================
// Inode Statistics (istat) - Deletion Time Extraction
// ============================================================================

// runIstat executes istat for deleted files to extract deletion timestamps.
// This provides forensic evidence of when files were deleted.
//
// Parameters:
//   - diskPath: Path to disk image
//   - offset: Partition offset in sectors
//   - analysis: Analysis structure for command logging
//   - fileListing: Array of file entries to update with deletion times
//
// Note: Modifies fileListing in-place by setting DeletionTime for deleted files.
func (t *Analyzer) runIstat(
	diskPath string,
	offset uint64,
	_ *models.TSKAnalysis,
	fileListing []models.TSKFileEntry,
) {
	// Build set of unique inodes for deleted files
	inodeSet := make(map[uint64]bool)
	for _, f := range fileListing {
		if f.Deleted {
			inodeSet[f.Inode] = true
		}
	}

	// Query istat for each deleted inode
	for inode := range inodeSet {
		args := []string{
			"-o", strconv.FormatUint(offset, 10),
			diskPath,
			strconv.FormatUint(inode, 10),
		}

		output, err := t.runCommandWithTimeout("istat", args, 30*time.Second)
		if err != nil {
			continue
		}

		// Parse deletion time from istat output
		var deletionTime int64
		scanner := bufio.NewScanner(strings.NewReader(output))

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if strings.Contains(line, "Deleted Time:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					if v, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
						deletionTime = v
					}
				}
			}
		}

		// Update deletion time in file listing
		for i := range fileListing {
			if fileListing[i].Inode == inode && fileListing[i].Deleted {
				fileListing[i].DeletionTime = deletionTime
				break
			}
		}
	}
}

// ============================================================================
// Command Execution with Timeout
// ============================================================================

// runCommandWithTimeout executes a TSK command with timeout protection.
// All command executions are logged to the analysis command history for custody chain.
//
// Parameters:
//   - command: Name of the TSK tool to execute (mmls, fsstat, fls, istat)
//   - args: Command-line arguments
//   - timeout: Maximum execution time before cancellation
//
// Returns:
//   - Command output as string
//   - Error if command fails or times out
func (t *Analyzer) runCommandWithTimeout(command string, args []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	timeCommandStart := time.Now()
	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.Output()

	exitCode := 0

	// Handle timeout
	if ctx.Err() == context.DeadlineExceeded {
		timeoutMsg := fmt.Sprintf("command timed out after %s", timeout)
		t.custodyChainEntry.LogWarning("CommandTimeout", fmt.Sprintf("%s timeout: %s", command, timeout))

		// Return partial output if available
		if len(output) > 0 {
			return string(output), nil
		}
		return "", fmt.Errorf("%s", timeoutMsg)
	}

	// Handle other errors
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// Log command execution to custody chain
	t.custodyChainEntry.LogCommand(
		utils.GenerateRandomID(),
		command,
		args,
		timeCommandStart, // Approximate start time
		time.Now(),       // End time
		exitCode,
		err,
		"", // Working directory (optional)
		"", // Target resource (optional)
	)

	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}
