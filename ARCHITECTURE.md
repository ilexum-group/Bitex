# Bitex Architecture

## Overview

Bitex is a forensic disk analysis tool implementing clean architecture with dependency injection, OS abstraction, and comprehensive custody chain tracking. The architecture prioritizes testability, maintainability, and forensic integrity.

## Core Principles

1. **Dependency Injection** - All components receive dependencies via constructors
2. **OS Abstraction** - Platform-specific operations isolated behind interfaces
3. **Read-Only Access** - No file content reading, only metadata extraction
4. **Custody Chain** - Complete audit trail of all operations
5. **Testability** - External test packages with black-box testing

## Directory Structure

```
bitex/
├── cmd/bitex/              # Application entry point
│   └── main.go            # CLI initialization and orchestration
│
├── internal/              # Private application code
│   ├── acquisition/       # Disk acquisition orchestration
│   │   └── acquisition.go # TSK analysis workflow coordinator
│   ├── config/           # Configuration management
│   │   └── config.go     # CLI flag parsing and validation
│   ├── logger/           # RFC 5424 compliant logging
│   │   └── logger.go     # Structured logging implementation
│   ├── os/               # OS abstraction layer
│   │   ├── os.go         # Interface definition
│   │   ├── default.go    # Base implementation with logging
│   │   ├── windows.go    # Windows-specific operations
│   │   ├── linux.go      # Linux-specific operations
│   │   └── darwin.go     # macOS-specific operations
│   ├── sender/           # HTTP transmission
│   │   └── sender.go     # POST request handling
│   ├── tsk/              # The Sleuth Kit integration
│   │   └── tsk.go        # TSK command execution and parsing
│   └── utils/            # Shared utilities
│       └── utils.go      # UUID generation, helpers
│
├── pkg/models/           # Public data structures
│   ├── models.go         # TSK analysis structures
│   ├── custody_chain.go  # Custody chain definition
│   └── custody_helpers.go # Custody chain operations
│
└── tests/                # Unit tests (external packages)
    ├── config_test.go           # Config parsing tests
    ├── custody_helpers_test.go  # Custody chain tests
    ├── os_test.go               # OS abstraction tests
    ├── utils_test.go            # Utility function tests
    └── logger_test.go           # Logging tests
```

## Data Flow

```
1. CLI Flags → Config Parsing → Validation
   ↓
2. OS Abstraction Initialization (Windows/Linux/Darwin)
   ↓
3. Logger Initialization (RFC 5424 compliant)
   ↓
4. Custody Chain Creation
   ↓
5. TSK Analysis (via acquisition layer)
   │  ├── Tool Version Detection
   │  ├── Partition Analysis (mmls)
   │  ├── Filesystem Analysis (fsstat)
   │  ├── File Listing (fls)
   │  └── Inode Analysis (istat)
   ↓
6. Custody Chain Finalization (with hashes)
   ↓
7. HTTP POST to server (authenticated)
```

## Component Details

### cmd/bitex/main.go

**Purpose:** Application entry point and initialization orchestrator

**Responsibilities:**
- Parse and validate CLI flags
- Initialize OS abstraction for current platform
- Set up RFC 5424 logger
- Create custody chain
- Orchestrate TSK analysis via acquisition layer
- Transmit results to server via authenticated POST
- Exit code management (all exits use code 1)

**Dependencies:**
- `internal/config` - Configuration management
- `internal/os` - OS abstraction
- `internal/logger` - Logging
- `internal/acquisition` - TSK orchestration
- `internal/sender` - HTTP transmission
- `pkg/models` - Data structures

### internal/config/config.go

**Purpose:** CLI configuration management

**Responsibilities:**
- Parse command-line flags
- Validate required parameters
- Validate flag combinations (e.g., --token requires --server)
- Provide Config struct to application

**Data Structure:**
```go
type Config struct {
    DiskPath  string  // Path to disk or image
    CaseID    string  // Optional case identifier
    ServerURL string  // Optional server endpoint
    Token     string  // Optional auth token
}
```

### internal/os/

**Purpose:** Platform-specific operations abstraction

**Interface:**
```go
type OS interface {
    Hostname() (string, error)
    Getenv(key string) string
    Getwd() (string, error)
    Stat(name string) (os.FileInfo, error)
    GetProcessID() int
    SetLogger(logger LoggerFunc)
}
```

**Implementations:**
- `DefaultImpl` - Base implementation with logging
- `Windows` - Windows-specific platform detection
- `Linux` - Linux-specific platform detection
- `Darwin` - macOS-specific platform detection

**Logger Integration:**
All OS operations are logged via callback function for complete custody chain tracking.

### internal/logger/logger.go

**Purpose:** RFC 5424 compliant structured logging

**Log Levels:**
- DEBUG (7) - Detailed debugging
- INFO (6) - Informational messages
- WARNING (4) - Warning conditions
- ERROR (3) - Error conditions
- CRITICAL (2) - Critical conditions

**Features:**
- Structured log entries with timestamps
- Hostname and process ID tracking
- Severity level enforcement
- Thread-safe concurrent logging

### internal/acquisition/acquisition.go

**Purpose:** TSK analysis workflow orchestration

**Responsibilities:**
- Coordinate TSK tool execution
- Manage custody chain during analysis
- Handle errors and log all operations
- Aggregate results into TSKAnalysis structure

**Workflow:**
1. Detect TSK tool versions
2. Run mmls for partition analysis
3. For each partition: run fsstat for filesystem info
4. Run fls for file listing (including deleted files)
5. Aggregate results with complete custody chain

### internal/tsk/tsk.go

**Purpose:** The Sleuth Kit integration

**TSK Commands:**
- `mmls` - Partition table extraction
- `fsstat` - Filesystem metadata
- `fls` - File and directory listing
- `istat` - Inode information

**Features:**
- Command version detection
- Exit code handling
- Output parsing into structured data
- Offset detection for partition analysis
- Deleted file detection in listings

### pkg/models/

**Purpose:** Shared data structures

**Key Structures:**
- `TSKAnalysis` - Complete analysis results
- `PartitionEntry` - Partition table entry
- `FilesystemInfo` - Filesystem metadata
- `FileEntry` - File listing entry
- `CustodyChainEntry` - Forensic custody chain
- `LogEntry` - Structured log entry
- `CommandExecution` - Command execution record

### pkg/models/custody_helpers.go

**Purpose:** Custody chain operations

**Functions:**
- `NewCustodyChainEntry()` - Initialize chain
- `LogCommand()` - Log TSK command execution
- `LogInfo/Warning/Error()` - Structured logging
- `FinalizeFromReader()` - Compute hashes and close chain

**Hash Algorithms:**
- MD5 (128-bit) - Legacy compatibility
- SHA1 (160-bit) - Legacy compatibility
- SHA256 (256-bit) - Primary verification

## Testing Architecture

### Test Organization

All tests are in `tests/` directory using external test packages:
- `config_test` - Configuration parsing and validation
- `models_test` - Custody chain operations
- `os_test` - OS abstraction layer
- `utils_test` - Utility functions
- `logger_test` - Logging system

### Testing Patterns

**Table-Driven Tests:**
```go
tests := []struct {
    name     string
    input    string
    expected string
    wantErr  bool
}{
    // Test cases...
}
```

**Concurrent Testing:**
```go
var wg sync.WaitGroup
for i := 0; i < numGoroutines; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // Test concurrent operations
    }()
}
wg.Wait()
```

**Benchmark Tests:**
```go
func BenchmarkGenerateRandomID(b *testing.B) {
    for i := 0; i < b.N; i++ {
        utils.GenerateRandomID()
    }
}
```

### Test Coverage

- **Config Tests:** Flag parsing, validation rules, error cases
- **Custody Chain Tests:** Creation, logging, command tracking, hash generation
- **OS Tests:** Platform detection, OS operations, logger integration
- **Utils Tests:** UUID generation, uniqueness, thread-safety
- **Logger Tests:** Log levels, concurrent logging, special characters

Run tests: `go test ./tests/... -v`

## Design Patterns

### Dependency Injection
Components receive dependencies via constructor parameters rather than creating them internally. This enables testing with mock dependencies.

### Interface Abstraction
OS operations are abstracted behind interfaces, allowing platform-specific implementations and test doubles.

### Command Pattern
TSK operations are logged as CommandExecution records with complete context (args, timestamps, exit codes).

### Builder Pattern
CustodyChainEntry is built incrementally through method calls, then finalized with hashes.

## Key Decisions

1. **No File Content Reading** - Only metadata extraction for forensic safety
2. **Exit Code Standardization** - All exits use code 1 for consistency
3. **External Test Packages** - Black-box testing approach with `*_test` packages
4. **Platform Detection at Runtime** - OS abstraction selects implementation based on runtime.GOOS
5. **Custody Chain Embedded in Output** - Complete audit trail travels with analysis results
6. **Server-Only Mode** - All results must be transmitted to remote server

## Security Considerations

- **Read-Only Disk Access** - No write operations to evidence
- **No File Content Reading** - Only metadata extraction
- **Complete Audit Trail** - Every operation logged in custody chain
- **Hash Verification** - MD5, SHA1, SHA256 for integrity
- **Authenticated Transmission** - Bearer token for server mode
- **Input Validation** - All CLI flags validated before execution
