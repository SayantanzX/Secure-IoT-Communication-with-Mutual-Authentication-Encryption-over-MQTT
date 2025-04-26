package main

import (
    "log"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "myproject/shared" // Ensure this package contains EncryptAES and DecryptAES functions
)

const broker = "tcp://localhost:1883"
const publishTopic = "device1/data"
const subscribeTopic = "device2/response"

var encryptionKey = []byte("0123456789abcdef0123456789abcdef") // 32-byte key

func responseHandler(client mqtt.Client, msg mqtt.Message) {
    encryptedMessage := string(msg.Payload())

    decryptedMessage, err := shared.DecryptAES(encryptionKey, encryptedMessage)
    if err != nil {
        log.Println("❌ Decryption failed:", err)
        return
    }

    log.Println("📩 Response from Device 2:", decryptedMessage)
}

func main() {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(broker)
    opts.SetClientID("Device1-Publisher")

    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Fatal(token.Error())
    }
    log.Println("📡 Connected to MQTT Broker")

    // Subscribe to receive responses from Device 2
    client.Subscribe(subscribeTopic, 0, responseHandler)

    // Encrypt & send message
    message := "Hello Device 2, can you hear me?"
    encryptedMessage, err := shared.EncryptAES(encryptionKey, message)
    if err != nil {
        log.Println("❌ Encryption failed:", err)
        return
    }

    token := client.Publish(publishTopic, 0, false, encryptedMessage)
    token.Wait()
    log.Println("📤 Sent message to Device 2:", encryptedMessage)

    // Keep running to receive responses
    select {}
}

