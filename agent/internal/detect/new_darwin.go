//go:build darwin

package detect

import "github.com/brianpowell/on-air/internal/detect/platform/darwin"

// New returns the platform detector for this OS.
func New() Detector {
	return darwin.New()
}
