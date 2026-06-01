#pragma once

#include <Arduino.h>

// WS2812 sign rendering from zone MQTT state (see config.h for layout/colors).
void signBegin();
void signHandleMqtt(char *topic, byte *payload, unsigned int length);
void signTick();
