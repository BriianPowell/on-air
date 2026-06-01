#include "connectivity.h"

#include <ESP8266WiFi.h>
#include <PubSubClient.h>

#include "config.h"
#include "protocol.h"
#include "sign_display.h"

namespace {

WiFiClient wifiClient;
PubSubClient mqtt(wifiClient);

void mqttCallback(char *topic, byte *payload, unsigned int length) {
	signHandleMqtt(topic, payload, length);
}

}  // namespace

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

void networkLoop() {
	mqtt.loop();
}

void networkEnsure() {
	if (WiFi.status() != WL_CONNECTED) {
		connectWiFi();
	}
	if (!mqtt.connected()) {
		connectMQTT();
	}
}
