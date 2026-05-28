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

1. Single poll (useful for debugging):

```bash
go run ./cmd/on-air-agent --once
```

1. Run with MQTT (requires a broker such as Mosquitto):

```bash
go run ./cmd/on-air-agent
```

## Config

See `config.example.yaml`. Key fields:

- `device` — identifier used in MQTT topic and JSON payload
- `mqtt.broker` — e.g. `tcp://192.168.1.10:1883`
- `mqtt.topic` — retained status topic the ESP8266 subscribes to
- `on_debounce` / `off_debounce` — avoid flicker when muting/unmuting

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
- `mic_active` or `camera_active` changes while hardware is still active (mid-call LED color updates)
- Not during the off-debounce wind-down when both are idle but `on_air` is still true internally — the final `on_air=false` publish covers that

## LED color matrix

Drive colors from `mic_active` and `camera_active` on the ESP8266 — not `on_air` alone.

| `camera_active` | `mic_active` | Meaning | Suggested RGB | Through white sign |
|:-:|:-:|-|-|-|
| true | true | Live video call, mic capturing | `(255, 0, 0)` | Classic broadcast red |
| true | false | Video on, mic not capturing | `(255, 140, 0)` | Warm amber — often a muted video call |
| false | true | Audio-only call | `(0, 180, 60)` | Soft green |
| false | false | Off / idle | `(0, 0, 0)` | LEDs off (or `(30, 20, 15)` for a faint warm glow) |

### Color notes

These values assume WS2812-style RGB LEDs behind a white diffuser (acrylic, 3D-printed panel, or vinyl). White material shifts colors toward pastels — higher saturation in config compensates for that.

- **Red `(255, 0, 0)`** — the unmistakable "ON AIR" look; reads clearly through white
- **Amber `(255, 140, 0)`** — distinct from red, still reads as "busy"; good for camera-on/mic-muted
- **Green `(0, 180, 60)`** — avoids confusion with red; signals "on a call" without video
- **Off** — prefer fully off over dim white; a faint glow can look like a fourth state

Tune brightness (e.g. cap at 30–40% / `(80, 0, 0)` instead of 255) if the sign is too bright at night.

### Mic muted during a video call

`mic_active` reflects **hardware capture**, not meeting-app mute state. Many apps release the mic when muted, so a muted video call often reports `mic_active=false` and maps to **amber**, not red. That is expected with the current detector.

## Platform support

- **macOS** — CoreAudio mic detection + CoreMediaIO camera detection
- **Windows** — stub (not implemented yet)

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

Revive runs via `go tool revive` (pinned in `go.mod`) — no separate `go install` needed.

Run manually against the whole repo:

```bash
pre-commit run --all-files
```
