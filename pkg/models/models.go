// Package models provides Bitex models for TSK analysis results (migrated from Tracium).
package models

// TSKAnalysis represents The Sleuth Kit analysis results for a disk.
// (copied from Tracium for Bitex independence).
type TSKAnalysis struct {
	DiskPath        string              `json:"disk_path"`
	FilesystemStats *TSKFilesystemStats `json:"filesystem_stats,omitempty"`
	FileListing     []TSKFileEntry      `json:"file_listing,omitempty"`
	ToolVersions    map[string]string   `json:"tool_versions"`
	CaseID          string              `json:"case_id"`       // Case identifier for correlation
	CustodyChain    *CustodyChainEntry  `json:"custody_chain"` // Custody Chain - Complete digital evidence custody tracking
}

// TSKFilesystemStats contains filesystem-level statistics from TSK analysis.
type TSKFilesystemStats struct {
	FilesystemType string `json:"filesystem_type"`
	BlockSize      int    `json:"block_size"`
	BlockCount     int64  `json:"block_count"`
	FreeBlocks     int64  `json:"free_blocks"`
	InodeCount     int64  `json:"inode_count"`
	FreeInodes     int64  `json:"free_inodes"`
	LastMountTime  int64  `json:"last_mount_time,omitempty"`
	LastWriteTime  int64  `json:"last_write_time,omitempty"`
	LastCheckTime  int64  `json:"last_check_time,omitempty"`
}

// TSKFileEntry represents a single file entry from TSK analysis.
type TSKFileEntry struct {
	Path         string `json:"path"`
	Inode        uint64 `json:"inode"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	ModifiedTime int64  `json:"modified_time"`
	AccessedTime int64  `json:"accessed_time"`
	CreatedTime  int64  `json:"created_time"`
	DeletionTime int64  `json:"deletion_time,omitempty"`
	Permissions  string `json:"permissions"`
	UID          int    `json:"uid"`
	GID          int    `json:"gid"`
	Deleted      bool   `json:"deleted"`
}
