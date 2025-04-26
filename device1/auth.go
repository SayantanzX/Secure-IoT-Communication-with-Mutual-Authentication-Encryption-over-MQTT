package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	
)

func main() {
	// Load Device 2's public key
	pubKeyData, err := ioutil.ReadFile("device2/keys/public_key.pem")
	if err != nil {
		log.Fatalf("Error reading public key: %v", err)
	}

	block, _ := pem.Decode(pubKeyData)
	if block == nil {
		log.Fatalf("Failed to parse PEM block")
	}

	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
    	log.Fatalf("Error parsing public key: %v", err)
	}

	pubKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
    	log.Fatalf("Public key is not of type RSA")
	}


	// Load the challenge message
	challenge, err := ioutil.ReadFile("challenge.txt")
	if err != nil {
		log.Fatalf("Error reading challenge message: %v", err)
	}

	// Compute SHA-256 hash of the challenge
	hash := sha256.Sum256(challenge)

	// Load the received signature
	sig, err := ioutil.ReadFile("challenge.sig")
	if err != nil {
		log.Fatalf("Error reading signature: %v", err)
	}

	// Verify the signature
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sig)
	if err != nil {
		log.Println("Signature verification failed! 🚨")
	} else {
		fmt.Println("Signature verified successfully! ✅ Device 2 is authenticated.")
	}
}

