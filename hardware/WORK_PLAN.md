# On-Air Sign — Hardware Work Plan

Build plan for the ESP8266 + addressable LED sign that pairs with the Go agent and MQTT firmware in this repo.

## Inventory: what you already have


| Item                        | Works for this project?        | Notes                                                                                                                                                                                                                                      |
| --------------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **NodeMCU CP2102 ESP-12E**  | **Yes**                        | Default board in `platformio.ini` (`nodemcuv2`). Use **D4 (GPIO2)** for LED data. Flash via micro-USB; CP2102 driver is standard on macOS.                                                                                                 |
| **100–240 V AC LED strips** | **No** (for per-person status) | Mains strips are **not individually addressable**. They are on/off (or dimmable) only — you cannot show red / amber / green per zone or per person. They also run at **dangerous mains voltage**, not the **5 V DC** the firmware expects. |


**Bottom line:** Your NodeMCU is the right brain. You need a **low-voltage addressable strip** (WS2812B or similar) for the status sign. Keep the AC strips for another project (under-cabinet lighting, etc.) or as a separate non-MQTT ambient light.

---

## Recommended LED strip (addressable + diffusion)

The firmware uses **Adafruit NeoPixel** with **WS2812B timing** (`NEO_GRB + NEO_KHZ800`). Buy strips that identify as **WS2812B** (or “NeoPixel-compatible”).

### Pick one approach


| Approach                                                     | Best for                       | Diffusion                                | Cost |
| ------------------------------------------------------------ | ------------------------------ | ---------------------------------------- | ---- |
| **A. WS2812B neon flex** (“360°” silicone tube)              | Clean glow, minimal build work | Excellent — light wraps in the tube      | $$   |
| **B. WS2812B 5050 strip + aluminum channel + opal diffuser** | Rectangular sign, custom width | Very good — frosted cover spreads pixels | $    |
| **C. WS2812B COB strip + diffuser panel**                    | Very even “light bar” look     | Excellent — COB is already smooth        | $$   |


**Recommendation for a household “on air” sign:** **Option B** — 30 LEDs/m WS2812B in a ~½″ aluminum channel with an opal polycarbonate diffuser. Cut to length, mount behind white acrylic or vinyl lettering. Easy to expand zones (`ledCount > 1` in firmware).

### Specs to order


| Spec      | Suggested value                              | Why                                                      |
| --------- | -------------------------------------------- | -------------------------------------------------------- |
| Protocol  | **WS2812B** (5 V)                            | Matches existing firmware; no code changes               |
| Density   | **30 LEDs/m** (or 60/m for shorter segments) | 30/m hides better behind diffusion; fewer pixels to wire |
| Length    | **0.5–1 m** total (cut to zones)             | 2 people × 3–6 LEDs each is plenty for a name badge      |
| IP rating | IP30 indoor is fine                          | Sign is indoors                                          |
| Wire      | Pre-soldered **5 V, GND, DIN** at one end    | Easier first build                                       |


### Optional upgrade (longer cable runs)

If the strip is **far from the NodeMCU** or you have **>30 pixels**, consider **WS2815** (12 V, backup data line) — **requires firmware/library changes**. Stick with **WS2812B** for v1.

### Parts to buy (minimal BOM)


| Qty | Part                                                                   | Est.    |
| --- | ---------------------------------------------------------------------- | ------- |
| 1   | WS2812B strip, 30 LEDs/m, 5 V, cut length                              | $8–15   |
| 1   | 5 V power supply, **2 A** (barrel or USB)                              | $6–10   |
| 1   | 1000 µF 6.3 V+ electrolytic capacitor                                  | $0.50   |
| 1   | 330 Ω resistor                                                         | $0.10   |
| 1   | 74AHCT125 (or pre-made 3.3 V→5 V level shifter module)                 | $1–3    |
| —   | 22 AWG hookup wire (red/black/green)                                   | on hand |
| —   | *(Option B)* Aluminum channel + opal diffuser, length matched to strip | $10–20  |
| —   | *(Option B)* White acrylic / PVC face for the sign                     | $5–15   |


You likely **do not** need a logic level shifter for bench testing with 2–8 pixels and a short data wire; add it for the final mounted sign.

---

## Circuit summary (NodeMCU)

```
5 V PSU (+) ──┬──► Strip 5V
              └──► NodeMCU 5V pin (small builds) OR separate USB for NodeMCU

5 V PSU (−) ──┬──► Strip GND
              └──► NodeMCU GND

NodeMCU D4 (GPIO2) ──[330Ω]──► Level shifter IN ──► Strip DIN
                              (optional for bench; recommended for final)
Strip 5V/GND ──[1000µF cap]── at strip input
```


| NodeMCU label | GPIO  | Connect to                                 |
| ------------- | ----- | ------------------------------------------ |
| **D4**        | GPIO2 | WS2812 data (via 330 Ω)                    |
| **G**         | GND   | Strip GND, PSU −                           |
| **VIN / 5V**  | —     | 5 V supply (or power NodeMCU via USB only) |


**PlatformIO note:** Default env is `nodemcuv2` (NodeMCU). Wemos D1 mini users: `pio run -e d1_mini -t upload` — same **D4 / GPIO2** pin.

---

## Work plan

### Phase 0 — Confirm software path (no new hardware)

**Goal:** Verify agent → MQTT → sign logic before buying or wiring.

- MQTT broker reachable on home LAN (Home Assistant / Mosquitto)
- Copy and edit `agent/config.example.yaml` → `config.yaml` per laptop
- Run agent dry-run / once: `go run ./cmd --dry-run`
- Copy `firmware/esp8266/include/config.example.h` → `config.h`
- Set Wi-Fi, MQTT host, `kZones`, and `LED_COUNT` in `config.h`
- *(Optional)* Temporarily test firmware with **no strip** connected — confirm Wi-Fi + MQTT in serial monitor (115200 baud)

**Exit criteria:** Agent publishes retained JSON to `on-air/{device}`; you understand zone → color mapping.

---

### Phase 1 — Bench prototype (NodeMCU + new WS2812B strip)

**Goal:** End-to-end proof with 2–8 pixels on the desk.

- Cut strip to **at least** `LED_COUNT` pixels (default 2; leave spare for zone expansion)
- Wire: D4 → DIN, common GND, 5 V to strip (NodeMCU on USB or shared 5 V)
- Add 1000 µF cap across strip 5 V/GND at the input end
- Add 330 Ω on data line
- Flash firmware: `cd firmware/esp8266 && pio run -t upload && pio device monitor`
- Publish test MQTT payload (or run agent on one laptop) and confirm:
  - Camera + mic → red
  - Camera only → amber
  - Mic only → green
  - Neither → off
- Tune `LED_BRIGHTNESS` in `config.h` (start at 80)

**Exit criteria:** Both zones show correct colors from real or test MQTT messages; no ESP resets when LEDs change.

**If pixels flicker or wrong colors:** Add 74AHCT125 level shifter; shorten data wire; check common ground.

---

### Phase 2 — Sign mechanical design

**Goal:** Layout zones, diffusion, and enclosure before permanent wiring.

- Decide sign size and labels (e.g. two names side by side)
- Map each person to pixel range in `kZones` (increase `ledCount` if each zone is a bar)
- Update `LED_COUNT` to total pixels
- Choose diffusion path (neon flex **or** channel + opal cover **or** acrylic panel)
- Sketch pixel positions behind the face — aim for **≥15 mm** between pixel center and diffuser for even glow
- Plan wire route: strip → cap/resistor/shifter → NodeMCU (mount MCU where USB access is easy)

**Exit criteria:** Cut list for strip, channel, and face material; firmware config matches pixel map.

---

### Phase 3 — Final assembly

**Goal:** Permanent wiring and mounted sign.

- Mount strip in channel / behind diffuser
- Power strip from **5 V PSU** directly; power NodeMCU via **USB** (recommended) or shared 5 V
- Tie **all grounds** together (PSU −, strip GND, NodeMCU GND)
- Install level shifter on data line
- Secure cap at strip input; strain-relief data and power wires
- Attach face (white acrylic / vinyl labels)
- Re-flash firmware if `LED_COUNT` or zones changed
- Full test with both agents online

**Exit criteria:** Sign mounted; both users’ states visible; survives Wi-Fi reconnect and power cycle (retained MQTT restores state).

---

### Phase 4 — Production tuning

**Goal:** Live-in polish.

- Adjust `LED_BRIGHTNESS` for room lighting and diffuser
- Confirm agent debounce feels right (`on_debounce` / `off_debounce` in agent config)
- Optional: Home Assistant dashboard / automations (see root `README.md`)
- Document your pixel map and PSU choice in a personal note or PR to this repo

**Exit criteria:** Sign is trustworthy day-to-day; no annoying flicker or false “on air.”

---

## What to do with the AC strips

These **cannot** replace WS2812B for this project without **rewriting firmware** and adding **mains-safe switching** (SSR/relay), and you would still only get **one color / on-off per switched segment** — not red/amber/green per pixel.

Reasonable alternatives:

- Use elsewhere in the house as static lighting
- Future **separate** project: one SSR + “sign is lit when anyone is on air” (binary), losing per-person colors

---

## Risk checklist


| Risk                          | Mitigation                                                                        |
| ----------------------------- | --------------------------------------------------------------------------------- |
| ESP resets when LEDs turn on  | Larger PSU; power strip directly, not through NodeMCU 5 V pin                     |
| Wrong colors / random flashes | Level shifter; cap at strip; shorter DIN wire                                     |
| Mains shock from AC strip     | Do **not** connect AC strip data/power to NodeMCU; use WS2812B only for this sign |
| GPIO2 boot issues             | Use **D4 only**; avoid pulling GPIO2 low at boot                                  |
| Zone mismatch                 | Topic suffix in agent `device` must match `kZones` entry exactly                  |


---

## Timeline (rough)


| Phase                       | Effort                                |
| --------------------------- | ------------------------------------- |
| Phase 0 — Software          | 1–2 hours                             |
| Phase 1 — Bench prototype   | 2–3 hours (+ shipping wait for strip) |
| Phase 2 — Mechanical design | 2–4 hours                             |
| Phase 3 — Final assembly    | 3–6 hours                             |
| Phase 4 — Tuning            | 30 min ongoing                        |


---

## References in this repo

- Firmware README: `[firmware/esp8266/README.md](../firmware/esp8266/README.md)`
- Config template: `[firmware/esp8266/include/config.example.h](../firmware/esp8266/include/config.example.h)`
- Agent config: `[agent/config.example.yaml](../agent/config.example.yaml)`
- Architecture + colors: `[README.md](../README.md)`

