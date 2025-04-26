package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Generate a unique device ID using HMAC-SHA256
func generateDeviceID(secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte("UniqueIoTIdentifier"))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	secret := "SuperSecureSecret"
	deviceID := generateDeviceID(secret)
	fmt.Println("Generated Unique Device ID:", deviceID)
}
