// Package protocol defines the MQTT contract shared by the Go agent and ESP8266 sign.
package protocol

import "strings"

const (
	TopicPrefix = "on-air"
	TopicFilter = "on-air/#"

	FieldMicActive    = "mic_active"
	FieldCameraActive = "camera_active"

	ClientIDPrefix = "on-air-agent-"
	SignClientID   = "on-air-sign"
)

// Status is the retained JSON payload on each person's topic.
type Status struct {
	MicActive    bool `json:"mic_active"`
	CameraActive bool `json:"camera_active"`
}

// Active reports whether mic or camera hardware is in use.
func (s Status) Active() bool {
	return s.MicActive || s.CameraActive
}

// OnAir is a synonym for Active used by the agent debounce logic.
func (s Status) OnAir() bool {
	return s.Active()
}

// TopicFor returns the MQTT topic for a device id (e.g. on-air/brian-mac).
func TopicFor(device string) string {
	return TopicPrefix + "/" + device
}

// TopicSuffix extracts the device id from a full topic, or "" if not an on-air topic.
func TopicSuffix(topic string) string {
	prefix := TopicPrefix + "/"
	if !strings.HasPrefix(topic, prefix) {
		return ""
	}
	return strings.TrimPrefix(topic, prefix)
}

// ClientIDFor returns the default MQTT client id for an agent device.
func ClientIDFor(device string) string {
	return ClientIDPrefix + device
}
