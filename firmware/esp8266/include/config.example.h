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

// Full-strip colors when exactly one person is on-air.
#define SIGN_SOLO_CAMERA_R 255
#define SIGN_SOLO_CAMERA_G 0
#define SIGN_SOLO_CAMERA_B 0

// Mic-only solo: lower G = more orange, higher G = more yellow (try 40–80).
#define SIGN_SOLO_MIC_ONLY_R 255
#define SIGN_SOLO_MIC_ONLY_G 55
#define SIGN_SOLO_MIC_ONLY_B 0

// Map MQTT topic suffixes (after onair::kTopicPrefix) to LED ranges on the strip.
// whenBoth* is used when two or more people are on-air at the same time.
struct ZoneConfig {
	const char *topicSuffix;
	uint16_t ledFirst;
	uint16_t ledCount;
	uint8_t whenBothR;
	uint8_t whenBothG;
	uint8_t whenBothB;
};

static const ZoneConfig kZones[] = {
	{"brian-mac", 0, 5, 0, 0, 255},    // blue — left half
	{"lauren-win", 5, 5, 0, 255, 0},   // green — right half
};

static const size_t kZoneCount = sizeof(kZones) / sizeof(kZones[0]);
