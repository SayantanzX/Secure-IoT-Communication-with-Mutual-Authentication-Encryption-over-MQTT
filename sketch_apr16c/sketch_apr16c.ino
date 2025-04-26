#include <WiFi.h>
#include <PubSubClient.h>
#include "mbedtls/sha256.h"
#include "mbedtls/pk.h"
#include "mbedtls/ctr_drbg.h"
#include "mbedtls/entropy.h"
#include "Keys.h"

#define GREEN_LED 18
#define BLUE_LED 19
#define BUZZER 23

const char* ssid = "Moto G24 Power";      // 🔁 Replace with your WiFi SSID
const char* password = "00000000";  // 🔁 Replace with your WiFi password
const char* mqtt_server = "192.168.227.89";    // 🔁 Replace with your MQTT broker IP

WiFiClient espClient;
PubSubClient client(espClient);

void setup_wifi() {
  Serial.println("[WiFi] Connecting to WiFi...");
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\n[WiFi] Connected!");
}

void messageReceived(char* topic, byte* payload, unsigned int length) {
  Serial.println("[MQTT] Message received");

  if (length < 32) {
    Serial.println("[MQTT] Invalid payload length.");
    return;
  }

  uint8_t hash[32];
  memcpy(hash, payload, 32);
  uint8_t* signature = payload + 32;
  size_t sig_len = length - 32;

  mbedtls_pk_context pk;
  mbedtls_pk_init(&pk);

  Serial.println("[Crypto] Parsing public key...");
  if (mbedtls_pk_parse_public_key(&pk, public_key_der, public_key_der_len) != 0) {
    Serial.println("[Crypto] Public key parse failed!");
    return;
  }

  Serial.println("[Crypto] Verifying signature...");
  int ret = mbedtls_pk_verify(&pk, MBEDTLS_MD_SHA256, hash, sizeof(hash), signature, sig_len);
  if (ret == 0) {
    Serial.println("[Crypto] Signature VERIFIED ✅");
    digitalWrite(GREEN_LED, HIGH);
    digitalWrite(BLUE_LED, LOW);
    digitalWrite(BUZZER, LOW);
  } else {
    Serial.println("[Crypto] Signature verification FAILED ❌");
    digitalWrite(GREEN_LED, LOW);
    digitalWrite(BLUE_LED, HIGH);
    digitalWrite(BUZZER, HIGH);
  }

  mbedtls_pk_free(&pk);
}

void reconnect() {
  while (!client.connected()) {
    Serial.print("[MQTT] Attempting MQTT connection...");
    if (client.connect("ESP32_B")) {
      Serial.println("connected");
      client.subscribe("iot/data");
    } else {
      Serial.print(" failed, rc=");
      Serial.print(client.state());
      Serial.println(" retrying in 5 seconds...");
      delay(5000);
    }
  }
}

void setup() {
  Serial.begin(115200);
  delay(2000); // Let Serial settle
  Serial.println("\n[System] Booting...");

  pinMode(GREEN_LED, OUTPUT);
  pinMode(BLUE_LED, OUTPUT);
  pinMode(BUZZER, OUTPUT);
  digitalWrite(GREEN_LED, LOW);
  digitalWrite(BLUE_LED, LOW);
  digitalWrite(BUZZER, LOW);

  setup_wifi();
  client.setServer(mqtt_server, 1883);
  client.setCallback(messageReceived);

  Serial.println("[System] Setup complete.");
}

void loop() {
  if (!client.connected()) {
    reconnect();
  }
  client.loop();
}
