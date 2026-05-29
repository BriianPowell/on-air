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

- `device` — unique ID for this laptop/person (used in topic and JSON)
- `mqtt.broker` — e.g. `tcp://homeassistant.local:1883`
- `mqtt.topic` — defaults to `on-air/{device}` if omitted
- `on_debounce` / `off_debounce` — avoid flicker when muting/unmuting

## Multi-user sign

One agent per laptop. Each person gets their own MQTT topic and LED zone on the same physical sign.

```
brian-mac (agent) ──► on-air/brian-mac ──┐
                                         ├──► Mosquitto (HA) ──► ESP8266 sign
jane-win  (agent) ──► on-air/jane-win  ──┘         └──► HA dashboard (optional)
```

| Laptop | `device` | `mqtt.topic` |
|--------|----------|--------------|
| Brian MacBook | `brian-mac` | `on-air/brian-mac` |
| Jane Windows | `jane-win` | `on-air/jane-win` |

Each agent publishes **independently** with the retain flag, so the sign always knows everyone's last state — even if one laptop is offline.

The ESP8266 subscribes to `on-air/#`, uses the **topic suffix** (e.g. `brian-mac`) to identify who updated, and maps that to a LED zone in firmware.

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

    - name: "Jane on air"
      state_topic: "on-air/jane-win"
      value_template: "{{ value_json.on_air }}"
      json_attributes_topic: "on-air/jane-win"
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
- `mic_active` or `camera_active` changes while hardware is still active (mid-call LED color updates)
- Not during the off-debounce wind-down when both are idle but `on_air` is still true internally — the final `on_air=false` publish covers that

## LED color matrix

Drive colors from `mic_active` and `camera_active` per person — not `on_air` alone. The ESP applies the matrix independently to each zone (mapped from the MQTT topic).

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

### Zoom waiting room

Zoom can show the mic as **unmuted in the UI** while you are still in the waiting room, but macOS may not open the mic for capture until you are admitted to the meeting. If the OS has not started an input stream, `mic_active` stays false even though Zoom's button looks "on". Camera preview in the waiting room *does* typically open the camera — which matches `camera_active=true` with `mic_active=false`.

The macOS detector checks CoreAudio **process input streams** first (which app is actually capturing), then falls back to device state.

## Platform support

Detection is **OS-level and app-agnostic** — we watch whether the mic or camera hardware is actively in use, not whether a specific app is in the foreground. That means Zoom, Teams, Slack huddles, FaceTime, etc. all work without per-app integration, as long as the app actually opens capture at the OS level.

### Target applications

| Platform | Apps | Typical signals |
|----------|------|-----------------|
| **macOS** | Slack huddles, Zoom, Microsoft Teams | Huddles: mic only. Zoom/Teams calls: mic + camera |
| **Windows** | Zoom, Microsoft Teams | Mic + camera during calls |

### Implementation status

| Platform | Mic | Camera | Status |
|----------|-----|--------|--------|
| **macOS** | CoreAudio process input streams, device fallback | CoreMediaIO `DeviceIsRunningSomewhere` | Working |
| **Windows** | WASAPI active capture sessions (planned) | Media/device enumeration (planned) | Not implemented |

### macOS

Uses CoreAudio **process input streams** first (`kAudioProcessPropertyIsRunningInput`), then falls back to input device state. Camera uses CoreMediaIO. No Slack/Zoom/Teams-specific code — if the app opens hardware, we detect it.

### Windows

Stub only today. Planned approach mirrors macOS: app-agnostic hardware detection via WASAPI (mic) and Windows camera APIs, not per-app hooks. Same JSON payload and agent logic on both platforms.

### Known limitations (all apps)

These apply to every target app on both platforms:

- **App mute ≠ hardware idle** — many apps release the mic when you mute in-meeting, so `mic_active` may be false during a muted call (sign shows amber, not red).
- **Pre-join / waiting room** — apps often open the camera for preview but defer mic capture until you are admitted (e.g. Zoom waiting room: `camera_active=true`, `mic_active=false`).
- **Slack huddles** — audio-only; expect `mic_active=true`, `camera_active=false` (green on the sign).
- **Bluetooth / virtual devices** — some headsets and virtual devices report state inconsistently at the OS level.

Grant the agent **Microphone** and **Camera** privacy permissions on macOS (System Settings → Privacy & Security). Windows will need equivalent permissions once detection is implemented.

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
