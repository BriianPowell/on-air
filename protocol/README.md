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

**On-air** (for sign display) = `mic_active || camera_active`.

## Sign display (firmware)

LED colors are **not** part of the MQTT payload — the ESP8266 derives on-air state from the fields above and applies display rules in `firmware/esp8266/src/main.cpp` and `include/config.h`.

Default layout (`LED_COUNT 10`):

| Pixels | Person | Topic suffix |
|--------|--------|--------------|
| 0–4 | Brian | `brian-mac` |
| 5–9 | Lauren | `lauren-win` |

| On-air count | Camera | Display |
|:------------:|:------:|---------|
| 0 | — | Off |
| 1 | off | Full strip **red** (solid) |
| 1 | on | Full strip **red snake** |
| 2+ | off (per zone) | Zone color, solid (Brian **cyan**, Lauren **orange**) |
| 2+ | on (per zone) | **Snake** on that zone only |

Snake = one pixel stepping through the zone. Tune `SNAKE_STEP_MS` and colors in `firmware/esp8266/include/config.h`.

Tune colors and pixel ranges in `firmware/esp8266/include/config.h`. See [`firmware/esp8266/README.md`](../firmware/esp8266/README.md).

Firmware scales RGB by `LED_BRIGHTNESS` before sending to the strip.

## Source files

| File | Used by |
|------|---------|
| `protocol.go` | Go agent |
| `protocol.h` | ESP8266 firmware |
