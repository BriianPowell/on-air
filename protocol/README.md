# on-air protocol

Shared contract between the **Go agent** (publisher) and **ESP8266 sign** (subscriber).

## Topics

| Item | Value |
|------|-------|
| Per-person topic | `on-air/{device}` |
| Sign subscription | `on-air/#` |
| Agent client id | `on-air-agent-{device}` |
| Sign client id | `on-air-sign` |

The topic suffix (`brian-mac`, `lauren-win`, …) must match `device` in agent config and `topicSuffix` in firmware `kZones`.

## Payload

Retained JSON, QoS 1:

```json
{
  "mic_active": true,
  "camera_active": false
}
```

Field names are defined in `protocol.go` / `protocol.h` (`mic_active`, `camera_active`).

The agent debounces internally; subscribers only see mic/camera hardware state.

## LED colors

Applied by firmware from `mic_active` + `camera_active`:

| Camera | Mic | Color | RGB |
|:------:|:---:|-------|-----|
| on | on | Red | `(255, 0, 0)` |
| on | off | Amber | `(255, 140, 0)` |
| off | on | Green | `(0, 180, 60)` |
| off | off | Off | `(0, 0, 0)` |

Firmware scales RGB by `LED_BRIGHTNESS` before sending to the strip.

## Source files

| File | Used by |
|------|---------|
| `protocol.go` | Go agent |
| `protocol.h` | ESP8266 firmware |
