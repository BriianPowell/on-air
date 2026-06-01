#pragma once

#include <cstdint>
#include <cstring>

// Shared MQTT + LED contract for the Go agent and ESP8266 sign.
// Keep in sync with protocol/protocol.go.

namespace onair {

constexpr char kTopicPrefix[] = "on-air/";
constexpr char kTopicFilter[] = "on-air/#";
constexpr char kJsonMic[] = "mic_active";
constexpr char kJsonCamera[] = "camera_active";
constexpr char kSignClientID[] = "on-air-sign";

struct Color {
	uint8_t r;
	uint8_t g;
	uint8_t b;
};

inline int zoneIndexForTopic(const char *topic, const char *const *suffixes, size_t count) {
	if (strncmp(topic, kTopicPrefix, strlen(kTopicPrefix)) != 0) {
		return -1;
	}

	const char *suffix = topic + strlen(kTopicPrefix);
	for (size_t i = 0; i < count; i++) {
		if (strcmp(suffix, suffixes[i]) == 0) {
			return static_cast<int>(i);
		}
	}
	return -1;
}

}  // namespace onair
