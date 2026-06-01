#include <Arduino.h>

#include "connectivity.h"
#include "sign_display.h"

void setup() {
	Serial.begin(115200);
	delay(100);
	Serial.println();
	Serial.println("on-air sign starting");

	signBegin();
	connectWiFi();
	connectMQTT();
}

void loop() {
	networkEnsure();
	networkLoop();
	signTick();
}
