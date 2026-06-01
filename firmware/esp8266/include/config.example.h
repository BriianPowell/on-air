#pragma once

#include "protocol.h"

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
#define MQTT_TOPIC_FILTER onair::kTopicFilter

// --- WS2812 strip ---
// NodeMCU: D4 = GPIO2 (default board in platformio.ini)
#define LED_PIN 2
#define LED_COUNT 10
#define LED_BRIGHTNESS 80

// One person on-air: full strip uses this color (solid or snake when camera on).
#define SIGN_SOLO_R 255
#define SIGN_SOLO_G 0
#define SIGN_SOLO_B 0

// Snake animation speed (one step per zone per interval).
#define SNAKE_STEP_MS 120

struct ZoneConfig {
	const char *topicSuffix;
	uint16_t ledFirst;
	uint16_t ledCount;
	onair::Color color;
};

static const ZoneConfig kZones[] = {
	{"brian-mac", 0, 5, {0, 255, 255}},    // cyan — left half
	{"lauren-win", 5, 5, {255, 55, 0}},    // orange — right half
};

static const size_t kZoneCount = sizeof(kZones) / sizeof(kZones[0]);
