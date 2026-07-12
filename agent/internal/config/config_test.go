package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAppliesDefaultsFromPartialYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`device: office-mac
mqtt:
  broker: tcp://192.168.1.5:1883
allowed_networks:
  - 192.168.1.0/24
  - 10.9.0.12
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Device != "office-mac" {
		t.Fatalf("device: got %q", cfg.Device)
	}
	if cfg.MQTT.Broker != "tcp://192.168.1.5:1883" {
		t.Fatalf("broker: got %q", cfg.MQTT.Broker)
	}
	if cfg.MQTT.Topic != "on-air/office-mac" {
		t.Fatalf("topic: got %q", cfg.MQTT.Topic)
	}
	if cfg.MQTT.ClientID != "on-air-agent-office-mac" {
		t.Fatalf("client_id: got %q", cfg.MQTT.ClientID)
	}
	if cfg.PollInterval != time.Second {
		t.Fatalf("poll_interval: got %v", cfg.PollInterval)
	}
	if cfg.OnDebounce != 3*time.Second {
		t.Fatalf("on_debounce: got %v", cfg.OnDebounce)
	}
	if cfg.OffDebounce != 10*time.Second {
		t.Fatalf("off_debounce: got %v", cfg.OffDebounce)
	}
	if len(cfg.AllowedNetworks) != 2 {
		t.Fatalf("allowed_networks: got %v", cfg.AllowedNetworks)
	}
}

func TestLoadRequiresDevice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`mqtt:
  topic: on-air/test
  client_id: test
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing device")
	}
}

func TestDefaultFillsMQTTFromDevice(t *testing.T) {
	cfg := Default("kitchen")

	if cfg.MQTT.Topic != "on-air/kitchen" {
		t.Fatalf("topic: got %q", cfg.MQTT.Topic)
	}
	if cfg.MQTT.ClientID != "on-air-agent-kitchen" {
		t.Fatalf("client_id: got %q", cfg.MQTT.ClientID)
	}
}

func TestLoadRejectsInvalidAllowedNetwork(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`device: office-mac
mqtt:
  topic: on-air/test
  client_id: test
allowed_networks:
  - home
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid allowed network")
	}
}
