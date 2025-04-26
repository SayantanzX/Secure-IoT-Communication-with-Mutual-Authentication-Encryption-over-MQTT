package main

import (
    "log"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "myproject/shared" // Import your encryption functions
)

const broker = "tcp://localhost:1883"
const subscribeTopic = "device1/data"
const publishTopic = "device2/response"

var encryptionKey = []byte("0123456789abcdef0123456789abcdef") // 32-byte key

func messageHandler(client mqtt.Client, msg mqtt.Message) {
    encryptedMessage := string(msg.Payload())

    decryptedMessage, err := shared.DecryptAES(encryptionKey, encryptedMessage)
    if err != nil {
        log.Println("❌ Decryption failed:", err)
        return
    }

    log.Println("📩 Received from Device 1:", decryptedMessage)

    // Prepare response
    responseMessage := "Hello Device 1, received your message at " + time.Now().Format(time.RFC1123)
    encryptedResponse, err := shared.EncryptAES(encryptionKey, responseMessage)
    if err != nil {
        log.Println("❌ Encryption failed:", err)
        return
    }

    // Publish response
    token := client.Publish(publishTopic, 0, false, encryptedResponse)
    token.Wait()
    log.Println("📤 Sent response to Device 1:", encryptedResponse)
}

func main() {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(broker)
    opts.SetClientID("Device2-Subscriber")

    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Fatal(token.Error())
    }
    log.Println("📡 Connected to MQTT Broker")

    client.Subscribe(subscribeTopic, 0, messageHandler)

    // Keep running
    select {}
}

