# ESP8266 sign firmware

WS2812 LED sign that subscribes to `on-air/#` on your **local** MQTT broker and sets each person's zone from `mic_active` / `camera_active`.

## Hardware

| Part | Default |
|------|---------|
| Board | NodeMCU v1.0 (ESP-12E, CP2102) — `nodemcuv2` in PlatformIO |
| LEDs | WS2812 / NeoPixel strip |
| Data pin | GPIO2 (`D4` on NodeMCU) |
| Power | 5 V supply sized for your LED count |

Wire data to `D4`, share ground with the ESP, and power the strip separately if you have more than a few pixels.

Using a Wemos D1 mini instead? Build with `pio run -e d1_mini -t upload` (same `D4` / GPIO2 pin).

## Setup

1. Install [PlatformIO](https://platformio.org/) (VS Code extension or CLI).

2. Copy config:

    ```bash
    cd firmware/esp8266
    cp include/config.example.h include/config.h
    ```

3. Edit `include/config.h`:

    - Wi-Fi credentials
    - MQTT broker IP (local LAN — same broker your ESP8266 can reach without TLS)
    - MQTT username/password
    - `kZones` — topic suffix → LED range mapping
    - `LED_COUNT` — total pixels on the strip

4. Build and upload:

    ```bash
    pio run -t upload
    pio device monitor
    ```

## Zone mapping

Agents publish to `on-air/brian-mac`, `on-air/lauren-win`, etc. Map each suffix to a contiguous run of pixels:

```cpp
static const ZoneConfig kZones[] = {
    {"brian-mac", 0, 1},   // pixel 0
    {"lauren-win", 1, 1},  // pixel 1
};
```

Use `ledCount > 1` for a multi-pixel zone (same color across the range).

## MQTT payload

See [`protocol/README.md`](../../protocol/README.md) for the shared topic, payload, and color contract.

Agents publish retained JSON:

```json
{
  "mic_active": true,
  "camera_active": false
}
```

## Colors

See [`protocol/README.md`](../../protocol/README.md#led-colors).

## Notes

- The ESP uses **plain MQTT on your LAN** (`tcp://host:1883`). Laptops use TLS on 443 over VPN; the sign at home does not need that.
- On boot, retained messages replay and the sign restores the last state.
- `include/config.h` is gitignored — do not commit credentials.
