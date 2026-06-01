#include <Arduino.h>
#include <Adafruit_NeoPixel.h>
#include <ArduinoJson.h>
#include <ESP8266WiFi.h>
#include <PubSubClient.h>

#include "config.h"
#include "protocol.h"

namespace {

struct ZoneState {
	bool known;
	bool micActive;
	bool cameraActive;
};

WiFiClient wifiClient;
PubSubClient mqtt(wifiClient);
Adafruit_NeoPixel strip(LED_COUNT, LED_PIN, NEO_GRB + NEO_KHZ800);

ZoneState zones[kZoneCount];

uint32_t scaleColor(uint8_t r, uint8_t g, uint8_t b) {
	const uint16_t scale = LED_BRIGHTNESS;
	r = (uint16_t)r * scale / 255;
	g = (uint16_t)g * scale / 255;
	b = (uint16_t)b * scale / 255;
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
	if (!doc[key].is<bool>()) {
		return false;
	}
	return doc[key].as<bool>();
}

bool zoneOnAir(size_t zoneIndex) {
	if (!zones[zoneIndex].known) {
		return false;
	}
	return zones[zoneIndex].micActive || zones[zoneIndex].cameraActive;
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

int zoneIndexForPixel(uint16_t pixel) {
	for (size_t i = 0; i < kZoneCount; i++) {
		const ZoneConfig &zone = kZones[i];
		if (pixel >= zone.ledFirst && pixel < zone.ledFirst + zone.ledCount) {
			return static_cast<int>(i);
		}
	}
	return -1;
}

onair::Color soloColorForActiveZone() {
	for (size_t i = 0; i < kZoneCount; i++) {
		if (!zoneOnAir(i)) {
			continue;
		}
		if (zones[i].cameraActive) {
			return {SIGN_SOLO_CAMERA_R, SIGN_SOLO_CAMERA_G, SIGN_SOLO_CAMERA_B};
		}
		return {SIGN_SOLO_MIC_ONLY_R, SIGN_SOLO_MIC_ONLY_G, SIGN_SOLO_MIC_ONLY_B};
	}
	return {};
}

onair::Color colorForPixel(uint16_t pixel, size_t onAirCount) {
	if (onAirCount == 0) {
		return {};
	}

	if (onAirCount == 1) {
		return soloColorForActiveZone();
	}

	const int zoneIndex = zoneIndexForPixel(pixel);
	if (zoneIndex < 0 || !zoneOnAir(static_cast<size_t>(zoneIndex))) {
		return {};
	}

	const ZoneConfig &zone = kZones[zoneIndex];
	return {zone.whenBothR, zone.whenBothG, zone.whenBothB};
}

void renderAll() {
	const size_t onAirCount = countOnAir();

	for (uint16_t pixel = 0; pixel < LED_COUNT; pixel++) {
		const onair::Color color = colorForPixel(pixel, onAirCount);
		strip.setPixelColor(pixel, scaleColor(color.r, color.g, color.b));
	}

	strip.show();

	Serial.printf("render on_air=%u", static_cast<unsigned>(onAirCount));
	for (size_t i = 0; i < kZoneCount; i++) {
		Serial.printf(
			" %s=%d",
			kZones[i].topicSuffix,
			zoneOnAir(i) ? 1 : 0);
	}
	Serial.println();
}

void mqttCallback(char *topic, byte *payload, unsigned int length) {
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

	zones[zoneIndex].known = true;
	zones[zoneIndex].micActive = parseBoolField(doc, onair::kJsonMic);
	zones[zoneIndex].cameraActive = parseBoolField(doc, onair::kJsonCamera);

	Serial.printf(
		"%s mic=%d camera=%d\n",
		topic,
		zones[zoneIndex].micActive,
		zones[zoneIndex].cameraActive);

	renderAll();
}

void connectWiFi() {
	WiFi.mode(WIFI_STA);
	WiFi.begin(WIFI_SSID, WIFI_PASS);
	Serial.printf("Wi-Fi connecting to %s", WIFI_SSID);

	while (WiFi.status() != WL_CONNECTED) {
		delay(500);
		Serial.print('.');
	}
	Serial.printf("\nWi-Fi connected: %s\n", WiFi.localIP().toString().c_str());
}

void connectMQTT() {
	mqtt.setServer(MQTT_HOST, MQTT_PORT);
	mqtt.setCallback(mqttCallback);
	mqtt.setBufferSize(512);

	while (!mqtt.connected()) {
		Serial.printf("MQTT connecting to %s:%d...", MQTT_HOST, MQTT_PORT);
		const bool ok = mqtt.connect(onair::kSignClientID, MQTT_USER, MQTT_PASS);
		if (ok) {
			Serial.println(" connected");
			mqtt.subscribe(MQTT_TOPIC_FILTER);
			Serial.printf("Subscribed to %s\n", MQTT_TOPIC_FILTER);
			return;
		}

		Serial.printf(" failed (rc=%d), retrying...\n", mqtt.state());
		delay(3000);
	}
}

}  // namespace

void setup() {
	Serial.begin(115200);
	delay(100);
	Serial.println();
	Serial.println("on-air sign starting");

	for (size_t i = 0; i < kZoneCount; i++) {
		zones[i] = {false, false, false};
	}

	strip.begin();
	strip.clear();
	strip.show();

	connectWiFi();
	connectMQTT();
}

void loop() {
	if (WiFi.status() != WL_CONNECTED) {
		connectWiFi();
	}

	if (!mqtt.connected()) {
		connectMQTT();
	}

	mqtt.loop();
}
