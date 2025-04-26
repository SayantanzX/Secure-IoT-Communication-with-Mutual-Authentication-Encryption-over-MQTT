package main

import (
	"bytes" // Added this import
	"context"
	"crypto"
	"crypto/rand"
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

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if key, ok := priv.(*rsa.PrivateKey); ok {
			return key, nil
		}
		return nil, fmt.Errorf("not an RSA private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func signChallenge(privateKey *rsa.PrivateKey, challenge string) (string, error) {
	// Trim whitespace from challenge
	challenge = string(bytes.TrimSpace([]byte(challenge)))
	
	hash := sha256.Sum256([]byte(challenge))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("signing failed: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func main() {
	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Load private key
	privateKey, err := loadPrivateKey("keys/private_key.pem")
	if err != nil {
		log.Fatalf("🔴 Failed to load private key: %v", err)
	}

	// 2. Configure MQTT client
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID(fmt.Sprintf("device2-responder-%d", time.Now().UnixNano())).
		SetCleanSession(false).
		SetResumeSubs(true).
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

	// 4. Handle challenges
	if token := client.Subscribe("iot/auth/challenge", 1, func(_ mqtt.Client, msg mqtt.Message) {
		challenge := string(msg.Payload())
		log.Printf("📥 Received challenge: %s", challenge)

		signature, err := signChallenge(privateKey, challenge)
		if err != nil {
			log.Printf("⚠️ Signing error: %v", err)
			return
		}

		log.Printf("📤 Sending signature (%d bytes)", len(signature))
		if token := client.Publish("iot/auth/response", 1, false, signature); token.Wait() && token.Error() != nil {
			log.Printf("⚠️ Publish failed: %v", token.Error())
		}
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("🔴 Subscribe failed: %v", token.Error())
	}

	log.Println("🔄 Ready for challenges (Ctrl+C to exit)")
	<-ctx.Done()
	log.Println("🛑 Shutting down gracefully...")
}
