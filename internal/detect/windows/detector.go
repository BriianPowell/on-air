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
	micActive, err := micInUse()
	if err != nil {
		return detect.Status{}, fmt.Errorf("mic detection: %w", err)
	}

	cameraActive, err := cameraInUse()
	if err != nil {
		return detect.Status{}, fmt.Errorf("camera detection: %w", err)
	}

	return detect.Status{
		MicActive:    micActive,
		CameraActive: cameraActive,
	}, nil
}
