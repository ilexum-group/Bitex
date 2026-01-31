
# Bitex Architecture

Bitex is a low-level forensic engine designed for disk and block-device analysis and integration with The Sleuth Kit (TSK). Bitex is a fully independent CLI tool and can be executed standalone. It does not require Tracium or any other orchestrator to run, but can be integrated into workflows as needed.


## Key Components
- **internal/disk/**: Disk imaging, block device access, read-only operations
- **internal/tsk/**: TSK integration and wrappers
- **internal/utils/**: Shared utilities

## Principles
- Strict read-only access to all disk and image sources
- Forensically sound evidence extraction
- Minimal, explicit, and maintainable design
- No direct reporting or orchestration logic



## Example Workflow
1. Bitex is executed directly from the command line with a disk image path:
	```bash
	# Output JSON to stdout
	bitex --disk /path/to/disk.img

	# Send results to a remote server
	bitex --disk /path/to/disk.img --server-url https://api.example.com/v1/data --agent-token your-auth-token
	```
2. Bitex performs read-only imaging or analysis, optionally using TSK.
3. Bitex outputs results as JSON to stdout and, if configured, sends results to a remote server via HTTP POST.

---

For implementation details, see the README.md and internal/ modules.
