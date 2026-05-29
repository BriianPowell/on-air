//go:build darwin

package main

import (
	"github.com/brianpowell/on-air/internal/detect"
	"github.com/brianpowell/on-air/internal/detect/darwin"
)

func newDetector() detect.Detector {
	return darwin.New()
}
