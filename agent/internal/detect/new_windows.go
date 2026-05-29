//go:build windows

package detect

import "github.com/brianpowell/on-air/internal/detect/platform/windows"

// New returns the platform detector for this OS.
func New() Detector {
	return windows.New()
}
