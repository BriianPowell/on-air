package types

// Status is a snapshot of local media device usage.
type Status struct {
	MicActive    bool
	CameraActive bool
}

func (s Status) OnAir() bool {
	return s.MicActive || s.CameraActive
}

// Detector reports whether the mic or camera is in use.
type Detector interface {
	Poll() (Status, error)
}
