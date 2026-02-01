# Bitex

## Description

Bitex is a disk acquisition and forensic analysis tool that performs disk imaging and analysis using The Sleuth Kit (TSK). It provides strict read-only access to disks and disk images, ensuring forensically sound evidence extraction.

## Purpose

Bitex acquires and analyzes disk images and block devices, extracting partition tables, filesystem metadata, and file system artifacts. It integrates with TSK tools (mmls, fsstat, fls, istat) to provide comprehensive disk-level forensic analysis.

## Problem It Solves

Forensic disk analysis requires safe, read-only access to storage devices without risking data modification. Bitex provides a standalone CLI tool that performs disk acquisition and TSK-based analysis, outputting structured JSON results that can be transmitted to centralized analysis servers (Processor) or used independently for forensic investigations.


### Example CLI Usage

```bash
# Output JSON to stdout
bitex --disk /path/to/disk.img

# Send results to Processor
bitex --disk /path/to/disk.img \
	--server-url http://localhost:8080/api/v1/bitex/evidence \
	--agent-token your-auth-token
```

This command outputs a JSON report to stdout and, if --server-url is provided, POSTs the result to the specified Processor endpoint using the given bearer token.

### Flags (Configuration Model)
All runtime configuration is passed via CLI flags.

- `--disk` (required): Path to disk image or block device (read-only)
- `--server-url` (optional): Processor endpoint to send analysis results as JSON
- `--agent-token` (optional): Bearer token for authentication when sending results
- `--case-id` (optional): Case identifier for correlation

**Environment variables** are not used for runtime configuration. The only relevant external dependency is **TSK** (mmls, fsstat, fls, istat), which must be available on the system PATH.

## Integration with Processor

Bitex sends evidence to Processor via:

- `POST /api/v1/bitex/evidence`

Processor stores the evidence and exposes it to Processor-UI. Bitex does not perform reporting or workflow orchestration.



## Responsibilities

- Disk imaging (TSK integration)
- Filesystem analysis (mmls, fsstat, fls, istat)
- JSON output with TSK results
- Optional transmission to Processor (POST /api/v1/bitex/evidence)

## Digital Evidence Custody Chain

Bitex implements a comprehensive digital evidence custody chain for disk analysis:

### Custody Chain Features

**Standardized Hash Algorithms:**
- MD5 (128-bit) - Legacy compatibility
- SHA1 (160-bit) - Legacy compatibility
- SHA256 (256-bit) - Primary integrity verification
- Hashes calculated for complete analysis package

**TSK Command Logging:**
- Every TSK command (mmls, fsstat, fls, istat) logged
- Command arguments, timestamps, exit codes tracked
- Output sizes and error messages captured
- Converted to standardized CommandExecution format

**Custody Transfer Tracking:**
- Initial disk analysis by agent
- Transmission to Processor for review
- Verification hashes at each transfer point

**Timeline Generation:**
- File created, modified, accessed, deleted timestamps
- Extracted from TSK file entries
- Formatted for TimeAnalysis integration
- Includes deleted files with deletion timestamps

**Processor Integration:**
- TSKAnalysis with embedded custody chain
- Automatic timeline extraction from file listing
- TimeAnalysis and Report reference tracking

### Usage Example

```go
// Create custody chain for disk analysis
chain, _ := models.NewCustodyChainEntry(caseID, version)

// Convert TSK command logs to custody chain
commands := models.ConvertTSKCommandLogs(tskAnalysis.CommandLogs)
for _, cmd := range commands {
    chain.AddCommandExecution(cmd)
}

// Finalize with analysis data
analysisJSON, _ := json.Marshal(tskAnalysis)
chain.Finalize(analysisJSON, len(tskAnalysis.FileListing))

// Generate timeline for Processor
timeline := models.GenerateTimelineFromTSK(tskAnalysis)

// Mark transmission
chain.MarkTransmitted(processorURL, response)
```

## Responsibilities
- Disk and block-device forensic analysis
- Integration with The Sleuth Kit (TSK)
- Strict read-only access to disks or disk images


## Structure
- `cmd/bitex/` - CLI entry point
- `internal/disk/` - Disk imaging, block device access
- `internal/tsk/` - TSK integration
- `internal/utils/` - Shared utilities


For more information, see the ARCHITECTURE.md file.
