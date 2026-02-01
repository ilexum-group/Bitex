// Package disk provides imaging and VM-based analysis logic migrated from Tracium.
package disk

import (
	"fmt"

	"github.com/ilexum-group/bitex/internal/models"
	"github.com/ilexum-group/bitex/internal/tsk"
)

// AnalyzeDisk performs metadata-only forensic analysis of a disk.
func AnalyzeDisk(diskPath string) (*models.TSKAnalysis, error) {
	analysis, err := tsk.AnalyzeDisk(diskPath)
	if err != nil {
		return nil, fmt.Errorf("TSK analysis failed: %w", err)
	}
	return analysis, nil
}
