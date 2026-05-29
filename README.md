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
- `mqtt.broker` — `tcp://` for local LAN; `ssl://host:443` for TLS (e.g. over work VPN when 8883 is blocked)
- `mqtt.topic` — defaults to `on-air/{device}` if omitted
- `mqtt.tls` — optional `ca_file` / `insecure_skip_verify` overrides
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
| **Windows** | WASAPI session enumeration (`go-wca`) | Privacy registry timestamps (best-effort) | Implemented — needs real-device validation |

Target apps: Slack huddles, Zoom, and Microsoft Teams on macOS; Zoom and Teams on Windows.

`mic_active` and `camera_active` reflect **hardware capture**, not in-app mute/camera buttons. Some apps keep the mic stream open when muted, so the sign may stay green during a muted audio call. Waiting rooms often open the camera for preview but defer mic capture until you are admitted.

Grant the agent **Microphone** and **Camera** privacy permissions on macOS (System Settings → Privacy & Security). On Windows, mic/camera access is per-app; the agent reads system APIs as your logged-in user — run it under your account (Task Scheduler logon trigger), not as a service.

### Windows camera note

Camera detection reads `HKCU\...\CapabilityAccessManager\ConsentStore\webcam` and treats an app as active when `LastUsedTimeStart > LastUsedTimeStop`. This mirrors how Windows records webcam use but may lag or miss edge cases. Validate on Lauren's machine and report mismatches.

## Run at login

Config is always local — the build pipelines only produce the binary. You still create `~/.config/on-air/config.yaml` (or the Windows equivalent) on each laptop and set up autostart once.

### Get the binary

**Option A — CI build (no Go required on the laptop)**

1. Merge to `main` (or open **Actions** → **Build macOS** / **Build Windows** → **Run workflow**).
2. Open the completed workflow run → **Artifacts** → download `on-air-agent-macos` or `on-air-agent-windows`.
3. Unzip and pick the right binary:

| Your machine | Binary from artifact |
|--------------|----------------------|
| Apple Silicon Mac | `on-air-agent-darwin-arm64` |
| Intel Mac | `on-air-agent-darwin-amd64` |
| Windows (64-bit) | `on-air-agent-windows-amd64.exe` |

With [GitHub CLI](https://cli.github.com/) (after a build on `main`):

```bash
# macOS — use run id from: gh run list --workflow=build-macos.yml
gh run download RUN_ID -n on-air-agent-macos -D /tmp/on-air
chmod +x /tmp/on-air/on-air-agent-darwin-arm64
sudo install -m 755 /tmp/on-air/on-air-agent-darwin-arm64 /usr/local/bin/on-air-agent
```

```powershell
# Windows — use run id from: gh run list --workflow=build-windows.yml
gh run download RUN_ID -n on-air-agent-windows -D $env:TEMP\on-air
```

Artifacts expire after 90 days (GitHub default). For tagged releases with permanent downloads, we can add a release workflow later.

**Option B — build locally** (see [Build](#build) below).

Pull request CI (`.github/workflows/ci.yml`) only validates code — it does not publish installable binaries.

Unsigned CI binaries may be blocked by Gatekeeper — run `xattr -dr com.apple.quarantine` on the downloaded binary, or build locally with `go build`.

### macOS (launchd)

1. Install the binary (from CI or local build):

```bash
# CI artifact already on disk as on-air-agent-darwin-arm64, or:
go build -o on-air-agent ./cmd/on-air-agent
sudo install -m 755 on-air-agent /usr/local/bin/on-air-agent   # or the darwin-* name
mkdir -p ~/.config/on-air
cp config.example.yaml ~/.config/on-air/config.yaml   # edit device, mqtt, etc.
```

1. Copy and customize the example plist (`USERNAME`, paths):

```bash
cp deploy/launchd/com.github.brianpowell.on-air-agent.plist.example \
   ~/Library/LaunchAgents/com.github.brianpowell.on-air-agent.plist
# edit ProgramArguments to match your config path
launchctl load ~/Library/LaunchAgents/com.github.brianpowell.on-air-agent.plist
```

1. Check status:

```bash
launchctl list | grep on-air
tail -f /tmp/on-air-agent.log
```

Unload: `launchctl unload ~/Library/LaunchAgents/com.github.brianpowell.on-air-agent.plist`

### Windows (Task Scheduler)

1. Install the binary and config:

```powershell
mkdir "$env:USERPROFILE\.config\on-air"
copy config.example.yaml "$env:USERPROFILE\.config\on-air\config.yaml"
# edit config.yaml, then copy on-air-agent-windows-amd64.exe to e.g. C:\Program Files\on-air\on-air-agent.exe
```

Or build locally:

```bash
GOOS=windows GOARCH=amd64 go build -o on-air-agent.exe ./cmd/on-air-agent
```

1. Register the scheduled task (edit `deploy/windows/on-air-agent-task.xml.example` first — replace `DOMAIN\USERNAME` and paths):

```powershell
schtasks /Create /TN "on-air-agent" /XML deploy\windows\on-air-agent-task.xml.example /F
schtasks /Run /TN "on-air-agent"
```

Remove: `schtasks /Delete /TN "on-air-agent" /F`

## Build

```bash
go build -o on-air-agent ./cmd/on-air-agent
```

Cross-compile for Windows from macOS:

```bash
GOOS=windows GOARCH=amd64 go build -o on-air-agent.exe ./cmd/on-air-agent
```

### CI

Pull requests run `.github/workflows/ci.yml`, mirroring pre-commit:

| Step | Command |
|------|---------|
| Format | `gofmt -l .` (must be clean) |
| Modules | `go mod tidy` (must not change `go.mod` / `go.sum`) |
| Vet | `go vet ./...` |
| Lint | `go tool revive -config=revive.toml ./...` |
| Test | `go test ./...` |
| Build | `go build ./...` on Ubuntu, macOS, and Windows |

### Release builds

Platform artifact workflows run on push to `main` and manual dispatch (`.github/workflows/build-macos.yml`, `build-windows.yml`):

| Artifact | Platform |
|----------|----------|
| `on-air-agent-darwin-arm64` | Apple Silicon Mac |
| `on-air-agent-darwin-amd64` | Intel Mac |
| `on-air-agent-windows-amd64.exe` | Windows |

Download from the workflow run → **Artifacts** section on GitHub.

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
