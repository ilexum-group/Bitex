
# Bitex - Disk and Forensic Engine

Bitex is a standalone forensic engine for disk and block-device analysis and integration with The Sleuth Kit (TSK). It provides strict read-only access to disks and disk images for forensically sound evidence extraction.

## Usage

Bitex is executed independently as a CLI tool. It does not require Tracium or any other orchestrator to run.


### Example CLI Usage

```bash
# Output JSON to stdout
bitex --disk /path/to/disk.img

# Send results to a remote server
bitex --disk /path/to/disk.img --server-url https://api.example.com/v1/data --agent-token your-auth-token
```

This command will output a JSON report of the forensic analysis to stdout and, if --server-url is provided, POST the results to the specified URL using the given bearer token.

### Flags
- `--disk` (required): Path to disk image or block device (read-only)
- `--server-url` (optional): URL to send analysis results as JSON
- `--agent-token` (optional): Bearer token for authentication when sending results

## Integration

Bitex can be invoked by other tools (e.g., Tracium) via CLI, API, or IPC. It does not perform reporting or workflow orchestration.



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
