//go:build !darwin && !windows

package detect

import (
	"fmt"

	"github.com/brianpowell/on-air/internal/detect/types"
)

type unsupportedDetector struct{}

func (unsupportedDetector) Poll() (types.Status, error) {
	return types.Status{}, fmt.Errorf("platform not supported")
}

// New returns the platform detector for this OS.
func New() Detector {
	return unsupportedDetector{}
}
