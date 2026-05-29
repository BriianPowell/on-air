# on-air

Status agent that detects mic/camera usage on your work laptop and publishes an on-air signal for an ESP8266 sign.

## Quick start

1. Install Go with [goenv](https://github.com/go-nv/goenv):

```bash
goenv install 1.26.3   # reads .go-version in this directory
goenv local 1.26.3     # already set if .go-version exists
go version             # should report go1.26.3
```

1. Copy the example config and edit it:

```bash
cp config.example.yaml config.yaml
```

1. Test detection without MQTT:

```bash
go run ./cmd/on-air-agent --dry-run
```

1. Single poll:

```bash
go run ./cmd/on-air-agent --once
```

1. Run with MQTT (requires a broker such as Mosquitto):

```bash
go run ./cmd/on-air-agent
```

## Config

See `config.example.yaml`. Key fields:

- `device` — unique ID for this laptop/person (used in topic and JSON)
- `mqtt.broker` — e.g. `tcp://homeassistant.local:1883`
- `mqtt.topic` — defaults to `on-air/{device}` if omitted
- `on_debounce` / `off_debounce` — avoid flicker when hardware state flaps

## Multi-user sign

One agent per laptop. Each person gets their own MQTT topic; the ESP8266 maps topic suffixes to LED zones in firmware.

```
brian-mac (agent) ──► on-air/brian-mac ──┐
                                         ├──► Mosquitto (HA) ──► ESP8266 sign
lauren-win (agent) ──► on-air/lauren-win ──┘         └──► HA dashboard (optional)
```

| Laptop | `device` | `mqtt.topic` |
|--------|----------|--------------|
| Brian MacBook | `brian-mac` | `on-air/brian-mac` |
| Lauren Windows | `lauren-win` | `on-air/lauren-win` |

Each agent publishes independently with the retain flag, so the sign always knows everyone's last state — even if one laptop is offline.

### Home Assistant

Create one MQTT sensor per person (Developer Tools → MQTT to sniff topics first):

```yaml
mqtt:
  sensor:
    - name: "Brian on air"
      state_topic: "on-air/brian-mac"
      value_template: "{{ value_json.on_air }}"
      json_attributes_topic: "on-air/brian-mac"
      json_attributes_template: "{{ value_json | tojson }}"

    - name: "Lauren on air"
      state_topic: "on-air/lauren-win"
      value_template: "{{ value_json.on_air }}"
      json_attributes_topic: "on-air/lauren-win"
      json_attributes_template: "{{ value_json | tojson }}"
```

Attributes include `mic_active` and `camera_active` for automations or the dashboard.

## MQTT payload

```json
{
  "device": "brian-mac",
  "on_air": true,
  "mic_active": true,
  "camera_active": false,
  "ts": "2026-05-28T14:32:00Z"
}
```

Messages are published with QoS 1 and the retain flag so new subscribers get the latest state immediately.

The agent publishes when:

- The debounced `on_air` state changes (after `on_debounce` / `off_debounce`)
- `mic_active` or `camera_active` changes while still considered on-air internally

## LED color matrix

Drive colors from `mic_active` and `camera_active` per person — not `on_air` alone.

| `camera_active` | `mic_active` | Color | Suggested RGB |
|:-:|:-:|-|-|
| true | true | Red | `(255, 0, 0)` |
| true | false | Amber | `(255, 140, 0)` |
| false | true | Green | `(0, 180, 60)` |
| false | false | Off | `(0, 0, 0)` |

These values assume WS2812-style RGB LEDs behind a white diffuser. Tune brightness down (e.g. cap red at `(80, 0, 0)`) if the sign is too bright at night.

## Detection

Detection is **OS-level and app-agnostic** — we watch whether the mic or camera hardware is in use, not whether a specific app is in the foreground. Zoom, Teams, Slack huddles, and similar apps work without per-app integration.

| Platform | Mic | Camera | Status |
|----------|-----|--------|--------|
| **macOS** | CoreAudio process input streams, device fallback | CoreMediaIO | Working |
| **Windows** | WASAPI (planned) | Windows camera APIs (planned) | Not implemented |

Target apps: Slack huddles, Zoom, and Microsoft Teams on macOS; Zoom and Teams on Windows.

`mic_active` and `camera_active` reflect **hardware capture**, not in-app mute/camera buttons. Some apps keep the mic stream open when muted, so the sign may stay green during a muted audio call. Waiting rooms often open the camera for preview but defer mic capture until you are admitted.

Grant the agent **Microphone** and **Camera** privacy permissions on macOS (System Settings → Privacy & Security).

## Build

```bash
go build -o on-air-agent ./cmd/on-air-agent
```

Cross-compile for Windows from macOS:

```bash
GOOS=windows GOARCH=amd64 go build -o on-air-agent.exe ./cmd/on-air-agent
```

## Pre-commit

Uses [pre-commit-golang](https://github.com/TekWizely/pre-commit-golang) for Go checks on each commit. Revive uses a local hook with `go tool revive` (pinned via `tool` in `go.mod`).

| Hook | What it does |
|------|--------------|
| `go-fmt-repo` | Format code (`-w` auto-fixes) |
| `go-revive` | Style lint via [revive](https://github.com/mgechev/revive) (`revive.toml`) |
| `go-mod-tidy-repo` | Keep `go.mod` / `go.sum` in sync |
| `go-vet-repo-mod` | `go vet ./...` |
| `go-test-repo-mod` | `go test ./...` |
| `go-build-repo-mod` | `go build ./...` |

Setup:

```bash
brew install pre-commit   # or: pip install pre-commit
go mod download            # installs pinned tools (see `tool` in go.mod)
pre-commit install
```

Run manually against the whole repo:

```bash
pre-commit run --all-files
```
