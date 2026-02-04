# Bitex Deployment

## Requirements

- Go 1.25 or higher
- Make utility (optional, for build automation)
- The Sleuth Kit (TSK) installed and available in PATH

## Building

```bash
# Clone repository
git clone https://github.com/yourusername/bitex.git
cd bitex

# Build for current platform
make build

# Build for all platforms
make build-all

# Manual build (no Make required)
go build -o build/bitex ./cmd/bitex
```

## Testing

Verify the build with tests:

```bash
# Run all tests
go test ./tests/... -v

# Run with coverage
go test ./tests/... -cover

# Run specific test suites
go test ./tests/config_test.go -v
go test ./tests/custody_helpers_test.go -v
go test ./tests/os_test.go -v
```

## Execution

### Standalone Analysis (JSON output)
```bash
# Output analysis to stdout
./build/bitex -disk /path/to/disk.img

# With case ID
./build/bitex -disk /dev/sda -case-id "CASE-2025-001"
```

### Server Integration
```bash
# Analyze and transmit to server
./build/bitex -disk /path/to/disk.img \
  -server https://server.com/api/analysis \
  -token your-auth-token

# With case ID
./build/bitex -disk /path/to/disk.img \
  -case-id "CASE-2025-001" \
  -server https://server.com/api/analysis \
  -token your-auth-token
```

### Key Flags
- `--disk PATH` - Path to disk image or block device (required)
- `--case-id ID` - Case identifier for correlation (required)
- `--server URL` - Server endpoint for results transmission (required)
- `--token TOKEN` - Authentication token for server communication (required)
- `-h, --help` - Show help
- `-v, --version` - Show version

## TSK Dependencies

Bitex requires The Sleuth Kit binaries in system PATH:
- `mmls` - Partition table analysis
- `fsstat` - Filesystem metadata
- `fls` - File listing with deleted files
- `istat` - Inode information

### Installation
- Linux: `sudo apt-get install sleuthkit`
- macOS: `brew install sleuthkit`
- Windows: Download from sleuthkit.org

## Deployment Checklist

1. ✓ Build the binary for target platform
2. ✓ Run test suite to verify functionality
3. ✓ Install TSK on target system
4. ✓ Verify TSK commands in PATH (`which mmls fsstat fls istat`)
5. ✓ Test with sample disk image
6. ✓ Configure server endpoint (if using server mode)
7. ✓ Set authentication token (if using server mode)
8. ✓ Verify read-only disk access permissions
