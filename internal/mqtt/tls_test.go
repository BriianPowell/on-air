package mqtt

import (
	"testing"

	"github.com/brianpowell/on-air/internal/config"
)

func TestBrokerOptionsPlainTCP(t *testing.T) {
	broker, tlsConfig, err := brokerOptions(config.MQTT{
		Broker: "tcp://localhost:1883",
	})
	if err != nil {
		t.Fatal(err)
	}
	if broker != "tcp://localhost:1883" {
		t.Fatalf("broker: got %q", broker)
	}
	if tlsConfig != nil {
		t.Fatal("expected no TLS config")
	}
}

func TestBrokerOptionsSSLScheme(t *testing.T) {
	broker, tlsConfig, err := brokerOptions(config.MQTT{
		Broker: "ssl://mqtt.example.com:443",
	})
	if err != nil {
		t.Fatal(err)
	}
	if broker != "tls://mqtt.example.com:443" {
		t.Fatalf("broker: got %q", broker)
	}
	if tlsConfig == nil {
		t.Fatal("expected TLS config")
	}
	if tlsConfig.ServerName != "mqtt.example.com" {
		t.Fatalf("ServerName: got %q", tlsConfig.ServerName)
	}
}
