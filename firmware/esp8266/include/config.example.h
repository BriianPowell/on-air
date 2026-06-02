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
#define LED_COUNT 56
#define LED_BRIGHTNESS 80

// One person on-air: full strip uses this color (solid or flash when camera on).
#define SIGN_SOLO_R 255
#define SIGN_SOLO_G 0
#define SIGN_SOLO_B 0

// Camera-on animation: zone flashes on/off (ms per half-cycle).
#define CAMERA_FLASH_MS 400

struct ZoneConfig {
	const char *topicSuffix;
	uint16_t ledFirst;
	uint16_t ledCount;
	onair::Color color;
};

static const ZoneConfig kZones[] = {
	{"brian-mac", 0, 28, {0, 255, 255}},    // cyan — left half
	{"lauren-win", 28, 28, {255, 55, 0}},   // orange — right half
};

static const size_t kZoneCount = sizeof(kZones) / sizeof(kZones[0]);
