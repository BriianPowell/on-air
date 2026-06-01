package protocol

import (
	"encoding/json"
	"testing"
)

func TestStatusJSON(t *testing.T) {
	payload, err := json.Marshal(Status{MicActive: true, CameraActive: false})
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != `{"mic_active":true,"camera_active":false}` {
		t.Fatalf("payload: got %s", payload)
	}
}

func TestTopicFor(t *testing.T) {
	if got := TopicFor("brian-mac"); got != "on-air/brian-mac" {
		t.Fatalf("TopicFor: got %q", got)
	}
}

func TestTopicSuffix(t *testing.T) {
	if got := TopicSuffix("on-air/brian-mac"); got != "brian-mac" {
		t.Fatalf("TopicSuffix: got %q", got)
	}
	if got := TopicSuffix("other/topic"); got != "" {
		t.Fatalf("TopicSuffix: got %q want empty", got)
	}
}
