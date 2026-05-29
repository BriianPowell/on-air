#pragma once

// Copy to config.h and fill in your values:
//   cp include/config.example.h include/config.h

// --- Wi-Fi ---
#define WIFI_SSID "your-ssid"
#define WIFI_PASS "your-wifi-password"

// --- MQTT (local broker on home LAN) ---
#define MQTT_HOST "10.0.2.11"
#define MQTT_PORT 1883
#define MQTT_USER "homeassistant"
#define MQTT_PASS "your-mqtt-password"
#define MQTT_TOPIC_FILTER "on-air/#"

// --- WS2812 strip ---
// D4 on NodeMCU / Wemos D1 mini = GPIO2
#define LED_PIN 2
#define LED_COUNT 2
#define LED_BRIGHTNESS 80

// Map MQTT topic suffixes (after "on-air/") to LED ranges on the strip.
// Example: on-air/brian-mac -> pixels 0..0, on-air/lauren-win -> pixels 1..1
struct ZoneConfig {
	const char *topicSuffix;
	uint16_t ledFirst;
	uint16_t ledCount;
};

static const ZoneConfig kZones[] = {
	{"brian-mac", 0, 1},
	{"lauren-win", 1, 1},
};

static const size_t kZoneCount = sizeof(kZones) / sizeof(kZones[0]);
