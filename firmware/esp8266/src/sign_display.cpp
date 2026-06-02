#include "sign_display.h"

#include <Adafruit_NeoPixel.h>
#include <ArduinoJson.h>

#include "config.h"
#include "protocol.h"

namespace {

struct ZoneState {
	bool known;
	bool micActive;
	bool cameraActive;
};

Adafruit_NeoPixel strip(LED_COUNT, LED_PIN, NEO_GRB + NEO_KHZ800);
ZoneState zones[kZoneCount];
uint8_t flashFrame = 0;
unsigned long lastAnimMs = 0;

constexpr onair::Color kSignSoloColor = {SIGN_SOLO_R, SIGN_SOLO_G, SIGN_SOLO_B};

uint32_t toPixelColor(const onair::Color &color) {
	const uint16_t scale = LED_BRIGHTNESS;
	const uint8_t r = (uint16_t)color.r * scale / 255;
	const uint8_t g = (uint16_t)color.g * scale / 255;
	const uint8_t b = (uint16_t)color.b * scale / 255;
	return strip.Color(r, g, b);
}

int zoneIndexForTopic(const char *topic) {
	const char *suffixes[kZoneCount];
	for (size_t i = 0; i < kZoneCount; i++) {
		suffixes[i] = kZones[i].topicSuffix;
	}
	return onair::zoneIndexForTopic(topic, suffixes, kZoneCount);
}

bool parseBoolField(JsonDocument &doc, const char *key) {
	if (doc[key].isNull()) {
		return false;
	}
	if (doc[key].is<bool>()) {
		return doc[key].as<bool>();
	}
	if (doc[key].is<int>() || doc[key].is<long>()) {
		return doc[key].as<int>() != 0;
	}
	if (doc[key].is<const char *>()) {
		const char *text = doc[key].as<const char *>();
		return text != nullptr &&
			(strcmp(text, "true") == 0 || strcmp(text, "1") == 0);
	}

	return false;
}

bool zoneOnAir(size_t zoneIndex) {
	return zones[zoneIndex].known &&
		(zones[zoneIndex].micActive || zones[zoneIndex].cameraActive);
}

size_t countOnAir() {
	size_t count = 0;
	for (size_t i = 0; i < kZoneCount; i++) {
		if (zoneOnAir(i)) {
			count++;
		}
	}
	return count;
}

bool anyCameraOnAir() {
	for (size_t i = 0; i < kZoneCount; i++) {
		if (zoneOnAir(i) && zones[i].cameraActive) {
			return true;
		}
	}
	return false;
}

void clearStrip() {
	for (uint16_t pixel = 0; pixel < LED_COUNT; pixel++) {
		strip.setPixelColor(pixel, 0);
	}
}

void renderSolid(uint16_t first, uint16_t count, const onair::Color &color) {
	const uint32_t pixelColor = toPixelColor(color);
	for (uint16_t i = 0; i < count; i++) {
		strip.setPixelColor(first + i, pixelColor);
	}
}

void renderFlash(uint16_t first, uint16_t count, const onair::Color &color, uint8_t frame) {
	if (count == 0 || (frame & 1) == 0) {
		return;
	}
	renderSolid(first, count, color);
}

void renderSegment(uint16_t first, uint16_t count, const onair::Color &color, bool flash) {
	if (flash) {
		renderFlash(first, count, color, flashFrame);
	} else {
		renderSolid(first, count, color);
	}
}

void logRenderState(size_t onAirCount) {
	Serial.printf("render on_air=%u flash=%d", static_cast<unsigned>(onAirCount), anyCameraOnAir());
	for (size_t i = 0; i < kZoneCount; i++) {
		Serial.printf(
			" %s=%d cam=%d",
			kZones[i].topicSuffix,
			zoneOnAir(i) ? 1 : 0,
			zones[i].cameraActive ? 1 : 0);
	}
	Serial.println();
}

void render(bool logState) {
	const size_t onAirCount = countOnAir();
	clearStrip();

	if (onAirCount == 1) {
		for (size_t i = 0; i < kZoneCount; i++) {
			if (!zoneOnAir(i)) {
				continue;
			}
			renderSegment(0, LED_COUNT, kSignSoloColor, zones[i].cameraActive);
			break;
		}
	} else if (onAirCount >= 2) {
		for (size_t i = 0; i < kZoneCount; i++) {
			if (!zoneOnAir(i)) {
				continue;
			}
			const ZoneConfig &zone = kZones[i];
			renderSegment(zone.ledFirst, zone.ledCount, zone.color, zones[i].cameraActive);
		}
	}

	strip.show();

	if (logState) {
		logRenderState(onAirCount);
	}
}

}  // namespace

void signBegin() {
	for (size_t i = 0; i < kZoneCount; i++) {
		zones[i] = {false, false, false};
	}

	strip.begin();
	strip.clear();
	strip.show();
}

void signHandleMqtt(char *topic, byte *payload, unsigned int length) {
	const int zoneIndex = zoneIndexForTopic(topic);
	if (zoneIndex < 0) {
		return;
	}

	JsonDocument doc;
	const DeserializationError err = deserializeJson(doc, payload, length);
	if (err) {
		Serial.printf("JSON parse error on %s: %s\n", topic, err.c_str());
		return;
	}

	const bool micActive = parseBoolField(doc, onair::kJsonMic);
	const bool cameraActive = parseBoolField(doc, onair::kJsonCamera);
	const bool cameraTurnedOn = cameraActive && !zones[zoneIndex].cameraActive;

	zones[zoneIndex].known = true;
	zones[zoneIndex].micActive = micActive;
	zones[zoneIndex].cameraActive = cameraActive;

	if (cameraTurnedOn) {
		flashFrame = 1;
		lastAnimMs = millis();
	}

	Serial.printf(
		"%s mic=%d camera=%d\n",
		topic,
		zones[zoneIndex].micActive,
		zones[zoneIndex].cameraActive);

	render(true);
}

void signTick() {
	if (!anyCameraOnAir()) {
		return;
	}

	const unsigned long now = millis();
	if (now - lastAnimMs < CAMERA_FLASH_MS) {
		return;
	}

	lastAnimMs = now;
	flashFrame++;
	render(false);
}
