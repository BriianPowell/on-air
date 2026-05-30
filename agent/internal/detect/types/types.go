package types

import "github.com/brianpowell/on-air/protocol"

// Status is the mic/camera snapshot shared with the MQTT payload.
type Status = protocol.Status

// Detector reports whether the mic or camera is in use.
type Detector interface {
	Poll() (Status, error)
}
