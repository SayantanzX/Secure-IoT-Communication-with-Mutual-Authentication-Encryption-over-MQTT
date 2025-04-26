#include <Arduino.h>
#include "keys.h"  // Contains: my_private_key[] and its length
#include "mbedtls/pk.h"
#include "mbedtls/sha256.h"
#include "mbedtls/ctr_drbg.h"
#include "mbedtls/entropy.h"

#define RED_LED_PIN 2  // Change this to your actual GPIO pin for the red LED

void blinkSuccessLED() {
  for (int i = 0; i < 3; i++) {
    digitalWrite(RED_LED_PIN, HIGH);
    delay(200);
    digitalWrite(RED_LED_PIN, LOW);
    delay(200);
  }
}

void setup() {
  Serial.begin(115200);
  delay(1000);
  Serial.println("🚀 ESP32 A Sender - ECDSA Signing Test");

  pinMode(RED_LED_PIN, OUTPUT);
  digitalWrite(RED_LED_PIN, LOW); // LED off initially

  // 1. Setup mbedTLS contexts
  mbedtls_pk_context pk;
  mbedtls_entropy_context entropy;
  mbedtls_ctr_drbg_context ctr_drbg;

  mbedtls_pk_init(&pk);
  mbedtls_entropy_init(&entropy);
  mbedtls_ctr_drbg_init(&ctr_drbg);

  const char *pers = "ecdsa_sign";
  if (mbedtls_ctr_drbg_seed(&ctr_drbg, mbedtls_entropy_func, &entropy,
                             (const unsigned char *)pers, strlen(pers)) != 0) {
    Serial.println("❌ Failed to initialize random number generator");
    digitalWrite(RED_LED_PIN, HIGH); // Solid ON for failure
    return;
  }

  // 2. Load Private Key (DER format)
  if (mbedtls_pk_parse_key(&pk, my_private_key, my_private_key_len, NULL, 0,
                           mbedtls_ctr_drbg_random, &ctr_drbg) != 0) {
    Serial.println("❌ Failed to parse private key");
    digitalWrite(RED_LED_PIN, HIGH); // Solid ON for failure
    return;
  }

  // 3. Hash Message
  const char *message = "Hello from ESP32 A";
  uint8_t hash[32];

  mbedtls_sha256_context sha_ctx;
  mbedtls_sha256_init(&sha_ctx);
  mbedtls_sha256_starts(&sha_ctx, 0);
  mbedtls_sha256_update(&sha_ctx, (const uint8_t *)message, strlen(message));
  mbedtls_sha256_finish(&sha_ctx, hash);
  mbedtls_sha256_free(&sha_ctx);

  Serial.println("📦 Message Hash:");
  for (size_t i = 0; i < sizeof(hash); i++) {
    Serial.printf("%02X", hash[i]);
  }
  Serial.println();

  // 4. Sign Hash
  uint8_t signature[MBEDTLS_ECDSA_MAX_LEN];
  size_t sig_len = 0;

  if (mbedtls_pk_sign(&pk, MBEDTLS_MD_SHA256,
                      hash, sizeof(hash),
                      signature, sizeof(signature),
                      &sig_len,
                      mbedtls_ctr_drbg_random, &ctr_drbg) == 0) {
    Serial.println("✅ Signature:");
    for (size_t i = 0; i < sig_len; i++) {
      Serial.printf("%02X", signature[i]);
    }
    Serial.println();

    blinkSuccessLED();  // Blink red LED for success
  } else {
    Serial.println("❌ Signing failed");
    digitalWrite(RED_LED_PIN, HIGH); // Solid ON for failure
  }

  // 5. Cleanup
  mbedtls_pk_free(&pk);
  mbedtls_entropy_free(&entropy);
  mbedtls_ctr_drbg_free(&ctr_drbg);
}

void loop() {
  // Nothing here for now
}
