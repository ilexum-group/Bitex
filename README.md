# Bitex

## Description

Bitex is a forensic disk analysis tool that performs metadata extraction using The Sleuth Kit (TSK). It provides strict read-only access to disks and disk images, ensuring forensically sound evidence extraction with comprehensive custody chain tracking. All analysis results are transmitted to a remote server for centralized forensic management.

## Purpose

Bitex analyzes disk images and block devices, extracting partition tables, filesystem metadata, and file system artifacts without reading file contents. It integrates with TSK tools (mmls, fsstat, fls, istat) to provide comprehensive disk-level forensic analysis with complete audit trails.

## Problem It Solves

Forensic disk analysis requires safe, read-only access to storage devices without risking data modification. Bitex provides a CLI tool that performs TSK-based metadata extraction with complete custody chain logging, transmitting structured JSON results to centralized analysis servers for forensic investigations.

## Key Features

- **Read-only metadata extraction** - No file content reading, only TSK metadata
- **Custody chain tracking** - Complete audit trail of all operations
- **Cross-platform support** - Windows, Linux, and macOS
- **RFC 5424 logging** - Compliant forensic logging
- **Server transmission** - All results sent to remote server

### Example CLI Usage

```bash
# Analyze disk and send to server
bitex --disk /path/to/disk.img --case-id "CASE-2025-001" \
  --server https://server.com/api/analysis \
  --token your-auth-token

# Analyze block device
bitex --disk /dev/sda \
  --case-id "CASE-2025-001" \
  --server https://server.com/api/analysis \
  --token your-auth-token
```

This command performs forensic analysis and transmits the results via authenticated POST request to the specified server.

### Command-Line Flags

All runtime configuration is passed via CLI flags:

- `--disk PATH` (required): Path to disk image or block device (read-only)
- `--case-id ID` (required): Case identifier for correlation
- `--server URL` (required): Remote server endpoint to send analysis results
- `--token TOKEN` (required): Authentication token for server communication
- `-h, --help`: Show help message
- `-v, --version`: Show version information

**Environment variables** are not used for runtime configuration. The only external dependency is **The Sleuth Kit (TSK)** which must be available in the system PATH.

## Server Integration

Bitex transmits all analysis results via authenticated POST request to the configured server. The server receives:

- Complete TSK analysis results (partition tables, filesystem metadata, file listings)
- Embedded custody chain with all commands executed
- Hash verification (MD5, SHA1, SHA256)
- Complete audit trail of all operations

## Core Responsibilities

- **Metadata extraction** - TSK integration for disk analysis (no file content reading)
- **Filesystem analysis** - Partition tables (mmls), filesystem info (fsstat), file listings (fls)
- **Custody chain** - Complete tracking of all operations with timestamps
- **JSON output** - Structured results with embedded custody chain
- **Server transmission** - Authenticated POST to remote server

## Digital Evidence Custody Chain

Bitex implements RFC 5424 compliant forensic custody chain tracking:

### Custody Chain Features

**Cryptographic Integrity:**
- MD5 (128-bit) - Legacy compatibility
- SHA1 (160-bit) - Legacy compatibility  
- SHA256 (256-bit) - Primary integrity verification
- Hashes calculated for complete analysis package

**Command Logging:**
- Every TSK command (mmls, fsstat, fls, istat) tracked
- Command arguments, start/end timestamps, exit codes
- Working directory and target resources logged
- Error messages captured for failed commands

**OS Operations Logging:**
- All system operations tracked automatically
- File access, environment variables, hostname queries
- Process IDs and execution context
- Timestamps and exit codes for all operations

**Audit Trail:**
- Comprehensive log entries (INFO, WARNING, ERROR, DEBUG, CRITICAL)
- Structured metadata for each log entry
- Timestamps for all operations
- Complete chain from start to finish

### Components

**Custody Chain:**
- Automatic creation and tracking
- Command execution logging
- Structured logging with severity levels
- Hash generation (MD5, SHA1, SHA256)

**TSK Analysis:**
- Automatic tool version detection
- Partition table analysis (mmls)
- Filesystem metadata extraction (fsstat)
- File and directory listings (fls)
- Deleted file detection
- Inode information (istat)

## Project Structure

```
bitex/
├── cmd/bitex/              # CLI entry point and initialization
├── internal/
│   ├── acquisition/        # Disk acquisition orchestration
│   ├── config/            # Configuration parsing and validation
│   ├── logger/            # RFC 5424 compliant logging
│   ├── os/                # OS abstraction layer (Windows/Linux/Darwin)
│   ├── sender/            # HTTP transmission to server
│   ├── tsk/               # The Sleuth Kit integration
│   └── utils/             # Shared utilities (ID generation)
├── pkg/models/            # Data structures and custody chain
└── tests/                 # Unit tests for all packages
```

## Testing

Run tests with:

```bash
# Run all tests
go test ./tests/... -v

# Run specific test files
go test ./tests/config_test.go -v
go test ./tests/custody_helpers_test.go -v
go test ./tests/os_test.go -v
go test ./tests/utils_test.go -v
go test ./tests/logger_test.go -v
```

## Dependencies

- **Go 1.25+** - Programming language
- **The Sleuth Kit** - Forensic analysis tools (mmls, fsstat, fls, istat)
- **github.com/google/uuid** - UUID generation for custody chain

For detailed architecture information, see [ARCHITECTURE.md](ARCHITECTURE.md).  
For deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md)