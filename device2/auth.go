package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	// Load private key
	privateKeyPath := "device2/keys/private_key.pem"
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		fmt.Println("Error reading private key:", err)
		return
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		fmt.Println("Failed to parse PEM block")
		return
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("Error parsing private key:", err)
		return
	}

	// Read challenge file
	challengePath := "/home/sayan-bhattacharya/Proj/challenge.txt"
	challengeBytes, err := os.ReadFile(challengePath)
	if err != nil {
		fmt.Println("Error reading challenge file:", err)
		return
	}

	// Hash the challenge
	hashed := sha256.Sum256(challengeBytes)

	// Sign the challenge
	signature, err := rsa.SignPKCS1v15(nil, privateKey.(*rsa.PrivateKey), crypto.SHA256, hashed[:])
	if err != nil {
		fmt.Println("Error signing challenge:", err)
		return
	}

	// Write signature to file
	signaturePath := "/home/sayan-bhattacharya/Proj/challenge.sig"
	err = os.WriteFile(signaturePath, signature, 0644)
	if err != nil {
		fmt.Println("Error saving signature:", err)
		return
	}

	fmt.Println("Challenge signed successfully. Signature saved to:", signaturePath)
}

