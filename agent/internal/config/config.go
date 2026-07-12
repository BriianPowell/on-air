package config

import (
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/brianpowell/on-air/protocol"
	"gopkg.in/yaml.v3"
)

type MQTTLTLS struct {
	CAFile             string `yaml:"ca_file"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

type MQTT struct {
	Broker   string   `yaml:"broker"`
	Topic    string   `yaml:"topic"`
	ClientID string   `yaml:"client_id"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	TLS      MQTTLTLS `yaml:"tls"`
}

// MQTTTLSEnabled reports whether the broker URL uses TLS.
func (m MQTT) MQTTTLSEnabled() bool {
	u, err := url.Parse(m.Broker)
	if err != nil {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "ssl", "tls", "mqtts", "mqtt+ssl", "tcps":
		return true
	default:
		return false
	}
}

type Config struct {
	Device          string        `yaml:"device"`
	MQTT            MQTT          `yaml:"mqtt"`
	AllowedNetworks []string      `yaml:"allowed_networks"`
	PollInterval    time.Duration `yaml:"poll_interval"`
	OnDebounce      time.Duration `yaml:"on_debounce"`
	OffDebounce     time.Duration `yaml:"off_debounce"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	cfg.ApplyDefaults()

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Default(device string) Config {
	cfg := Config{Device: device}
	cfg.ApplyDefaults()
	return cfg
}

func (c *Config) ApplyDefaults() {
	if c.PollInterval == 0 {
		c.PollInterval = time.Second
	}
	if c.OnDebounce == 0 {
		c.OnDebounce = 3 * time.Second
	}
	if c.OffDebounce == 0 {
		c.OffDebounce = 10 * time.Second
	}
	if c.MQTT.Broker == "" {
		c.MQTT.Broker = "tcp://localhost:1883"
	}
	if c.MQTT.Topic == "" && c.Device != "" {
		c.MQTT.Topic = protocol.TopicFor(c.Device)
	}
	if c.MQTT.ClientID == "" && c.Device != "" {
		c.MQTT.ClientID = protocol.ClientIDFor(c.Device)
	}
}

func (c Config) validate() error {
	if c.Device == "" {
		return fmt.Errorf("device is required")
	}
	if c.MQTT.Topic == "" {
		return fmt.Errorf("mqtt.topic is required")
	}
	if c.MQTT.ClientID == "" {
		return fmt.Errorf("mqtt.client_id is required")
	}
	for _, network := range c.AllowedNetworks {
		if _, err := parseNetwork(network); err != nil {
			return fmt.Errorf("allowed_networks: %q is not an IP address or CIDR: %w", network, err)
		}
	}
	return nil
}

func parseNetwork(value string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(value)
	if err == nil {
		return prefix.Masked(), nil
	}

	addr, addrErr := netip.ParseAddr(value)
	if addrErr != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}
