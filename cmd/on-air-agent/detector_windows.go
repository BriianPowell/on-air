//go:build windows

package main

import (
	"github.com/brianpowell/on-air/internal/detect"
	"github.com/brianpowell/on-air/internal/detect/windows"
)

func newDetector() detect.Detector {
	return windows.New()
}
