package main

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	challenge     string
	device2PubKey *rsa.PublicKey
)

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func generateChallenge() string {
	nonce := fmt.Sprintf("%d-%x", time.Now().UnixNano(), sha256.Sum256([]byte("salt")))
	hash := sha256.Sum256([]byte(nonce))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func verifySignature(signature string) bool {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		log.Printf("⚠️ Signature decode error: %v", err)
		return false
	}

	hashed := sha256.Sum256([]byte(challenge))
	err = rsa.VerifyPKCS1v15(device2PubKey, crypto.SHA256, hashed[:], sig)
	if err != nil {
		log.Printf("⚠️ Verification failed: %v", err)
		return false
	}
	return true
}

func main() {
	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Load public key
	var err error
	device2PubKey, err = loadPublicKey("../device2/keys/public_key.pem")
	if err != nil {
		log.Fatalf("🔴 Failed to load public key: %v", err)
	}

	// 2. Configure MQTT with enhanced options
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID(fmt.Sprintf("device1-verifier-%d", time.Now().UnixNano())).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetKeepAlive(30 * time.Second).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Printf("⚠️ Connection lost: %v", err)
		})

	client := mqtt.NewClient(opts)
	
	// 3. Connect with timeout
	if token := client.Connect(); !token.WaitTimeout(10*time.Second) {
		log.Fatal("🔴 Connection timeout")
	} else if token.Error() != nil {
		log.Fatalf("🔴 Connection failed: %v", token.Error())
	}
	defer client.Disconnect(250)

	// 4. Generate and send challenge
	challenge = generateChallenge()
	log.Printf("📤 Challenge: %s", challenge)
	if token := client.Publish("iot/auth/challenge", 1, false, challenge); token.Wait() && token.Error() != nil {
		log.Printf("⚠️ Publish failed: %v", token.Error())
		return
	}

	// 5. Handle responses with timeout
	responseChan := make(chan string, 1)
	if token := client.Subscribe("iot/auth/response", 1, func(_ mqtt.Client, msg mqtt.Message) {
		responseChan <- string(msg.Payload())
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("🔴 Subscribe failed: %v", token.Error())
	}

	select {
	case signature := <-responseChan:
		if verifySignature(signature) {
			log.Println("✅ Authentication successful!")
		} else {
			log.Println("❌ Authentication failed!")
		}
	case <-time.After(30 * time.Second):
		log.Println("⚠️ Timeout waiting for response")
	case <-ctx.Done():
		log.Println("🛑 Shutdown requested")
	}
}
