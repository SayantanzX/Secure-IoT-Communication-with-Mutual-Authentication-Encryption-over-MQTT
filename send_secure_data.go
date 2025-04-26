package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Encrypt data using AES-256
func encryptAES(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return hex.EncodeToString(ciphertext), nil
}

func main() {
	key := []byte("0123456789abcdef0123456789abcdef") // 32-byte AES key
	message := "Hello Device 2, this is a secure message!"

	// Encrypt message
	encrypted, err := encryptAES(key, message)
	if err != nil {
		fmt.Println("Encryption error:", err)
		return
	}

	// Save encrypted message to a file
	err = os.WriteFile("secure_message.txt", []byte(encrypted), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("Encrypted message saved to secure_message.txt")

	// Send file to Device 2 via SCP (manually run this after)
	fmt.Println("Run this command to send the file:")
	fmt.Printf("scp secure_message.txt sayan-bhattacharya@192.168.0.185:/home/sayan-bhattacharya/Proj/\n")
}
