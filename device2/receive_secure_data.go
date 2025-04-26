package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"os"
)

// Decrypt data using AES-256
func decryptAES(key []byte, encrypted string) (string, error) {
	ciphertext, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func main() {
	key := []byte("0123456789abcdef0123456789abcdef") // 32-byte AES key

	// Read encrypted message
	encrypted, err := os.ReadFile("secure_message.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Decrypt message
	decrypted, err := decryptAES(key, string(encrypted))
	if err != nil {
		fmt.Println("Decryption error:", err)
		return
	}

	fmt.Println("Decrypted message:", decrypted)
}
