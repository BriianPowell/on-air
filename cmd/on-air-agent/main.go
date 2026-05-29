package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brianpowell/on-air/internal/agent"
	"github.com/brianpowell/on-air/internal/config"
	"github.com/brianpowell/on-air/internal/detect"
	onairmqtt "github.com/brianpowell/on-air/internal/mqtt"
)

type stdoutPublisher struct {
	device string
}

func (p stdoutPublisher) Publish(onAir bool, status detect.Status) error {
	fmt.Printf(
		"[%s] on_air=%t mic=%t camera=%t\n",
		time.Now().Format(time.RFC3339),
		onAir,
		status.MicActive,
		status.CameraActive,
	)
	return nil
}

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	dryRun := flag.Bool("dry-run", false, "print status to stdout instead of publishing to MQTT")
	once := flag.Bool("once", false, "poll once and exit")
	flag.Parse()

	cfg, err := loadConfig(*configPath, *once || *dryRun)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	detector := newDetector()

	if *once {
		status, err := detector.Poll()
		if err != nil {
			log.Fatalf("poll: %v", err)
		}
		fmt.Printf("mic=%t camera=%t on_air=%t\n",
			status.MicActive,
			status.CameraActive,
			status.OnAir(),
		)
		return
	}

	var publisher agent.Publisher
	if *dryRun {
		publisher = stdoutPublisher{device: cfg.Device}
	} else {
		pub, err := onairmqtt.New(cfg.MQTT, cfg.Device)
		if err != nil {
			log.Fatalf("mqtt: %v", err)
		}
		defer pub.Close()
		publisher = pub
	}

	a := agent.New(cfg, detector, publisher)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("on-air agent started for device %q (dry_run=%t)", cfg.Device, *dryRun)
	if err := a.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("agent stopped: %v", err)
	}
}

func loadConfig(path string, allowDefault bool) (config.Config, error) {
	if _, err := os.Stat(path); err == nil {
		return config.Load(path)
	}
	if allowDefault {
		host, _ := os.Hostname()
		if host == "" {
			host = "local"
		}
		return config.Default(host), nil
	}
	return config.Config{}, fmt.Errorf("read config: open %s: file not found", path)
}
