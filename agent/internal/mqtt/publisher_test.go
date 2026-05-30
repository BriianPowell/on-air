package mqtt

import (
	"encoding/json"
	"testing"

	"github.com/brianpowell/on-air/protocol"
)

func TestPublishPayloadShape(t *testing.T) {
	payload, err := json.Marshal(protocol.Status{MicActive: true, CameraActive: false})
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != `{"mic_active":true,"camera_active":false}` {
		t.Fatalf("payload: got %s", payload)
	}
}
