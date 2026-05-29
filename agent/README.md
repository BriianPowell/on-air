# on-air agent

Go service that detects mic/camera usage and publishes status to MQTT.

```bash
cp config.example.yaml config.yaml   # edit device, mqtt broker, etc.
go run ./cmd/on-air-agent --dry-run
```

See the [root README](../README.md) for full setup, deployment, and multi-user configuration.
