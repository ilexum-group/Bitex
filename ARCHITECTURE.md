# Bitex Architecture

## Overview

Bitex is a standalone forensic engine designed for disk and block-device analysis with The Sleuth Kit (TSK) integration. It follows a minimal, explicit architecture focused on read-only disk access and forensically sound evidence extraction.

## Directory Structure

### `cmd/bitex/`
Command-line interface and application entry point. Handles flag parsing, input validation, and orchestrates the disk analysis workflow.

### `internal/disk/`
Disk imaging and block device access module. Provides read-only operations for disk images and physical block devices, ensuring no modifications to source evidence.

### `internal/tsk/`
The Sleuth Kit integration layer. Wraps TSK command-line tools (mmls, fsstat, fls, istat) for partition analysis, filesystem metadata extraction, and file system artifact collection.

### `internal/models/`
Data structures for disk analysis results, partition information, and filesystem metadata.

### `internal/logger/`
Logging system for tracking disk operations and analysis steps.

### `internal/sender/`
HTTP transmission module for sending analysis results to Processor backend with Bearer token authentication.

### `internal/utils/`
Shared utility functions for file operations, path handling, and common forensic tasks.

## Data Flow

1. CLI parses disk image path and configuration flags
2. Disk module opens source in read-only mode
3. TSK wrapper executes analysis tools (mmls, fsstat, fls)
4. Results are collected and structured into JSON format
5. Sender transmits results to Processor (if configured)
6. Output is written to stdout for standalone usage
7. All operations logged for audit trail
