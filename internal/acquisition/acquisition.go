// Package acquisition handles disk analysis acquisition and chain of custody for Bitex.
// This package orchestrates the complete forensic acquisition process, including:
//   - Disk validation and accessibility checks
//   - TSK-based metadata extraction
//   - Custody chain construction and logging
//   - Analysis finalization with integrity verification
package acquisition

import (
	"fmt"
	"time"

	internalos "github.com/ilexum-group/bitex/internal/os"
	"github.com/ilexum-group/bitex/internal/tsk"
	"github.com/ilexum-group/bitex/internal/utils"
	"github.com/ilexum-group/bitex/pkg/models"
)

// ============================================================================
// Acquirer Structure
// ============================================================================

// Acquirer manages the complete disk analysis acquisition process with custody chain tracking.
// It coordinates between the OS abstraction layer, TSK analysis tools, and custody chain logging.
type Acquirer struct {
	custodyChain *models.CustodyChainEntry // Custody chain for forensic logging
	diskPath     string                    // Path to disk image or block device
	tskAnalyzer  *tsk.Analyzer             // TSK analyzer instance
	osImpl       internalos.OS             // Operating system abstraction
}

// NewAcquirer creates a new Acquirer instance with all required dependencies.
// Parameters:
//   - osImpl: Operating system abstraction for file operations
//   - diskPath: Path to the disk image or block device to analyze
//   - custodyChainEntry: Initialized custody chain entry for logging
//   - tskAnalyzer: TSK analyzer instance for disk analysis
//
// Returns a configured Acquirer ready to perform disk acquisition.
func NewAcquirer(osImpl internalos.OS, diskPath string, custodyChainEntry *models.CustodyChainEntry, tskAnalyzer *tsk.Analyzer) *Acquirer {
	return &Acquirer{
		osImpl:       osImpl,
		diskPath:     diskPath,
		custodyChain: custodyChainEntry,
		tskAnalyzer:  tskAnalyzer,
	}
}

// ============================================================================
// Public Methods
// ============================================================================

// AcquireDisk performs the complete disk acquisition process with custody chain logging.
// This method orchestrates:
//  1. Disk validation and accessibility verification
//  2. TSK-based metadata analysis
//  3. Tool version logging
//  4. Statistical analysis and logging
//
// All operations are logged to the custody chain for forensic integrity.
//
// Returns:
//   - TSKAnalysis structure containing all collected metadata
//   - Error if critical failure occurs (disk inaccessible, TSK analysis fails)
func (a *Acquirer) AcquireDisk() (*models.TSKAnalysis, error) {
	a.custodyChain.LogInfo("DiskAcquisitionStart", fmt.Sprintf("Starting acquisition: %s", a.diskPath))

	// Step 1: Validate disk accessibility
	if err := a.validateDisk(); err != nil {
		return nil, err
	}

	// Step 2: Perform TSK analysis
	analysis, err := a.performTSKAnalysis()
	if err != nil {
		return nil, err
	}

	// Step 3: Log tool versions
	a.logToolVersions(analysis)

	// Step 4: Log analysis statistics
	a.logAnalysisStatistics(analysis)

	a.custodyChain.LogInfo("DiskAcquisitionComplete", fmt.Sprintf("Successfully acquired: %s", a.diskPath))

	return analysis, nil
}

// GetAnalysisWithCustody finalizes the custody chain and attaches it to the analysis.
// This method:
//   - Sets the end timestamp and calculates total duration
//   - Calculates total size from file listing
//   - Computes item count
//   - Attaches the custody chain to the analysis
//
// Parameters:
//   - analysis: TSKAnalysis structure to finalize
//
// Returns the analysis with attached custody chain.
func (a *Acquirer) GetAnalysisWithCustody(analysis *models.TSKAnalysis) *models.TSKAnalysis {
	a.custodyChain.LogInfo("CustodyChainFinalization", "Finalizing custody chain metadata")

	// Finalize timestamps
	a.custodyChain.EndTimestamp = time.Now().UTC()
	a.custodyChain.Duration = a.custodyChain.EndTimestamp.Sub(a.custodyChain.StartTimestamp).String()

	// Calculate totals from file listing
	totalSize := a.calculateTotalSize(analysis.FileListing)
	a.custodyChain.TotalSizeBytes = totalSize
	a.custodyChain.ItemCount = len(analysis.FileListing)

	a.custodyChain.LogInfo("CustodyChainComplete", fmt.Sprintf(
		"Custody chain finalized - Items: %d, Size: %d bytes, Duration: %s",
		a.custodyChain.ItemCount,
		a.custodyChain.TotalSizeBytes,
		a.custodyChain.Duration,
	))

	// Attach custody chain to analysis
	analysis.CustodyChain = a.custodyChain

	return analysis
}

// ============================================================================
// Private Helper Methods
// ============================================================================

// validateDisk verifies that the target disk exists and is accessible.
// Logs the validation command and results to the custody chain.
//
// Returns error if disk is not accessible.
func (a *Acquirer) validateDisk() error {
	a.custodyChain.LogInfo("DiskValidation", fmt.Sprintf("Validating disk: %s", a.diskPath))

	startTime := time.Now()
	fileInfo, err := a.osImpl.Stat(a.diskPath)
	endTime := time.Now()

	// Log the validation command
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	a.custodyChain.LogCommand(
		utils.GenerateRandomID(),
		"os.Stat",
		[]string{a.diskPath},
		startTime,
		endTime,
		exitCode,
		err,
		"",
		a.diskPath,
	)

	if err != nil {
		a.custodyChain.LogError("DiskValidation", fmt.Sprintf("Disk not accessible: %s", a.diskPath), err)
		return fmt.Errorf("disk not found or inaccessible: %s", a.diskPath)
	}

	a.custodyChain.LogInfo("DiskValidated", fmt.Sprintf("Disk accessible - Size: %d bytes", fileInfo.Size()))
	return nil
}

// performTSKAnalysis executes the TSK analysis and logs the operation.
//
// Returns:
//   - TSKAnalysis structure with collected metadata
//   - Error if TSK analysis fails
func (a *Acquirer) performTSKAnalysis() (*models.TSKAnalysis, error) {
	a.custodyChain.LogInfo("TSKAnalysisStart", fmt.Sprintf("Starting TSK analysis: %s", a.diskPath))

	startTime := time.Now()
	analysis, err := a.tskAnalyzer.AnalyzeDisk(a.diskPath)
	endTime := time.Now()

	// Log the analysis operation
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	a.custodyChain.LogCommand(
		utils.GenerateRandomID(),
		"tsk.AnalyzeDisk",
		[]string{a.diskPath},
		startTime,
		endTime,
		exitCode,
		err,
		"",
		a.diskPath,
	)

	if err != nil {
		a.custodyChain.LogError("TSKAnalysisFailed", fmt.Sprintf("TSK analysis error: %s", a.diskPath), err)
		return nil, fmt.Errorf("TSK analysis failed: %w", err)
	}

	a.custodyChain.LogInfo("TSKAnalysisComplete", "TSK metadata extraction completed")
	return analysis, nil
}

// logToolVersions logs all TSK tool versions used in the analysis.
// This is critical for forensic documentation and reproducibility.
//
// Parameters:
//   - analysis: Analysis containing tool version information
func (a *Acquirer) logToolVersions(analysis *models.TSKAnalysis) {
	if len(analysis.ToolVersions) == 0 {
		a.custodyChain.LogWarning("ToolVersions", "No tool versions recorded")
		return
	}

	for tool, version := range analysis.ToolVersions {
		a.custodyChain.LogInfo("ToolVersion", fmt.Sprintf("%s: %s", tool, version))
	}
}

// logAnalysisStatistics calculates and logs comprehensive analysis statistics.
// Includes file counts, deletion statistics, and filesystem information.
//
// Parameters:
//   - analysis: Analysis containing file listing and filesystem stats
func (a *Acquirer) logAnalysisStatistics(analysis *models.TSKAnalysis) {
	// Calculate file statistics
	fileCount := len(analysis.FileListing)
	deletedCount := a.countDeletedFiles(analysis.FileListing)

	a.custodyChain.LogInfo("FileStatistics", fmt.Sprintf(
		"Total files: %d, Deleted files: %d",
		fileCount,
		deletedCount,
	))

	// Log filesystem statistics if available
	if analysis.FilesystemStats != nil {
		a.custodyChain.LogInfo("FilesystemInfo", fmt.Sprintf(
			"Type: %s, Block Size: %d, Block Count: %d",
			analysis.FilesystemStats.FilesystemType,
			analysis.FilesystemStats.BlockSize,
			analysis.FilesystemStats.BlockCount,
		))
	} else {
		a.custodyChain.LogWarning("FilesystemInfo", "No filesystem statistics available")
	}
}

// calculateTotalSize computes the total size of all files in the listing.
//
// Parameters:
//   - files: Array of file entries
//
// Returns total size in bytes.
func (a *Acquirer) calculateTotalSize(files []models.TSKFileEntry) int64 {
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}
	return totalSize
}

// countDeletedFiles counts the number of deleted files in the listing.
//
// Parameters:
//   - files: Array of file entries
//
// Returns count of deleted files.
func (a *Acquirer) countDeletedFiles(files []models.TSKFileEntry) int {
	count := 0
	for _, file := range files {
		if file.Deleted {
			count++
		}
	}
	return count
}
