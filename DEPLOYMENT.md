# Bitex Deployment

## Requirements

- Go 1.25 or higher
- Make utility
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
```

## Execution

### Standalone Analysis (JSON output)
```bash
# Output analysis to stdout
./build/bitex --disk /path/to/disk.img
```

### Send to Processor
```bash
# Analyze and transmit to Processor
./build/bitex --disk /path/to/disk.img \
  --server-url http://localhost:8080/api/v1/bitex/evidence \
  --agent-token your-auth-token
```

### Key Flags
- `--disk PATH` - Path to disk image or block device (required)
- `--server-url URL` - Processor endpoint for results transmission
- `--agent-token TOKEN` - Bearer token for authentication
- `--case-id ID` - Case identifier for correlation

## TSK Dependencies

Bitex requires The Sleuth Kit binaries in system PATH:
- `mmls` - Partition table analysis
- `fsstat` - Filesystem metadata
- `fls` - File listing
- `istat` - Inode information

### Installation
- Linux: `sudo apt-get install sleuthkit`
- macOS: `brew install sleuthkit`
- Windows: Download from sleuthkit.org
