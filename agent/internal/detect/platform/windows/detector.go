//go:build windows

package windows

import (
	"fmt"

	"github.com/brianpowell/on-air/internal/detect/types"
)

type Detector struct{}

func New() *Detector {
	return &Detector{}
}

func (d *Detector) Poll() (types.Status, error) {
	micActive, err := micInUse()
	if err != nil {
		return types.Status{}, fmt.Errorf("mic detection: %w", err)
	}

	cameraActive, err := cameraInUse()
	if err != nil {
		return types.Status{}, fmt.Errorf("camera detection: %w", err)
	}

	return types.Status{
		MicActive:    micActive,
		CameraActive: cameraActive,
	}, nil
}
