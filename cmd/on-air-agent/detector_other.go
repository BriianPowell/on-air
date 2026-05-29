//go:build !darwin && !windows

package main

import (
	"fmt"

	"github.com/brianpowell/on-air/internal/detect"
)

type unsupportedDetector struct{}

func (unsupportedDetector) Poll() (detect.Status, error) {
	return detect.Status{}, fmt.Errorf("platform not supported")
}

func newDetector() detect.Detector {
	return unsupportedDetector{}
}
