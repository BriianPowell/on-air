package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/brianpowell/on-air/internal/config"
	"github.com/brianpowell/on-air/internal/detect"
)

type StatusMessage struct {
	Device       string    `json:"device"`
	OnAir        bool      `json:"on_air"`
	MicActive    bool      `json:"mic_active"`
	CameraActive bool      `json:"camera_active"`
	Timestamp    time.Time `json:"ts"`
}

type Publisher struct {
	client pahomqtt.Client
	topic  string
	device string
}

func New(cfg config.MQTT, device string) (*Publisher, error) {
	opts := pahomqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}

	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return nil, fmt.Errorf("mqtt connect timed out")
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt connect: %w", err)
	}

	return &Publisher{
		client: client,
		topic:  cfg.Topic,
		device: device,
	}, nil
}

func (p *Publisher) Publish(onAir bool, status detect.Status) error {
	msg := StatusMessage{
		Device:       p.device,
		OnAir:        onAir,
		MicActive:    status.MicActive,
		CameraActive: status.CameraActive,
		Timestamp:    time.Now().UTC(),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal status: %w", err)
	}

	token := p.client.Publish(p.topic, 1, true, payload)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("mqtt publish timed out")
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("mqtt publish: %w", err)
	}

	log.Printf("published on_air=%t mic=%t camera=%t topic=%s", onAir, status.MicActive, status.CameraActive, p.topic)
	return nil
}

func (p *Publisher) Close() {
	if p.client != nil && p.client.IsConnected() {
		p.client.Disconnect(250)
	}
}
