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
	return detect.Status{}, fmt.Errorf("windows detection is not implemented yet")
}
