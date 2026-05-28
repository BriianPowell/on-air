package mqtt

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStatusMessageJSON(t *testing.T) {
	ts := time.Date(2026, 5, 28, 14, 32, 0, 0, time.UTC)
	msg := StatusMessage{
		Device:       "brian-mac",
		OnAir:        true,
		MicActive:    true,
		CameraActive: false,
		Timestamp:    ts,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded["device"] != "brian-mac" {
		t.Fatalf("device: got %v", decoded["device"])
	}
	if decoded["on_air"] != true {
		t.Fatalf("on_air: got %v", decoded["on_air"])
	}
	if decoded["mic_active"] != true {
		t.Fatalf("mic_active: got %v", decoded["mic_active"])
	}
	if decoded["camera_active"] != false {
		t.Fatalf("camera_active: got %v", decoded["camera_active"])
	}
	if _, ok := decoded["source"]; ok {
		t.Fatal("source field should not be present")
	}
}
