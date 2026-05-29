#include <Arduino.h>
#include <Adafruit_NeoPixel.h>
#include <ArduinoJson.h>
#include <ESP8266WiFi.h>
#include <PubSubClient.h>

#include "config.h"

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

void colorForStatus(bool mic, bool camera, uint8_t &r, uint8_t &g, uint8_t &b) {
	if (camera && mic) {
		r = 255;
		g = 0;
		b = 0;
	} else if (camera) {
		r = 255;
		g = 140;
		b = 0;
	} else if (mic) {
		r = 0;
		g = 180;
		b = 60;
	} else {
		r = 0;
		g = 0;
		b = 0;
	}
}

int zoneIndexForTopic(const char *topic) {
	const char *prefix = "on-air/";
	if (strncmp(topic, prefix, strlen(prefix)) != 0) {
		return -1;
	}

	const char *suffix = topic + strlen(prefix);
	for (size_t i = 0; i < kZoneCount; i++) {
		if (strcmp(suffix, kZones[i].topicSuffix) == 0) {
			return static_cast<int>(i);
		}
	}
	return -1;
}

bool parseBoolField(JsonDocument &doc, const char *key) {
	if (!doc[key].is<bool>()) {
		return false;
	}
	return doc[key].as<bool>();
}

void applyZoneColor(size_t zoneIndex) {
	const ZoneConfig &zone = kZones[zoneIndex];
	uint8_t r = 0;
	uint8_t g = 0;
	uint8_t b = 0;

	if (zones[zoneIndex].known) {
		colorForStatus(zones[zoneIndex].micActive, zones[zoneIndex].cameraActive, r, g, b);
	}

	const uint32_t color = scaleColor(r, g, b);
	for (uint16_t i = 0; i < zone.ledCount; i++) {
		const uint16_t pixel = zone.ledFirst + i;
		if (pixel < LED_COUNT) {
			strip.setPixelColor(pixel, color);
		}
	}
}

void renderAll() {
	for (size_t i = 0; i < kZoneCount; i++) {
		applyZoneColor(i);
	}
	strip.show();
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
	zones[zoneIndex].micActive = parseBoolField(doc, "mic_active");
	zones[zoneIndex].cameraActive = parseBoolField(doc, "camera_active");

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
		const bool ok = mqtt.connect("on-air-sign", MQTT_USER, MQTT_PASS);
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
