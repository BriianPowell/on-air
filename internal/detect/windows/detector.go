//go:build windows

package windows

import (
	"fmt"

	"github.com/brianpowell/on-air/internal/detect"
)

type Detector struct{}

func New() *Detector {
	return &Detector{}
}

func (d *Detector) Poll() (detect.Status, error) {
	// Planned: WASAPI session enumeration for mic, Windows camera APIs for video.
	// Target apps: Zoom, Microsoft Teams (app-agnostic hardware detection).
	return detect.Status{}, fmt.Errorf("windows detection is not implemented yet")
}
