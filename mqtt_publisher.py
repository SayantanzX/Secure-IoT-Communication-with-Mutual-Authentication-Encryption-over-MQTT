import paho.mqtt.client as mqtt
import json
import time
import random

# Broker details
broker = "localhost"
port = 1883
topic = "device/challenge"

# The client ID for this device
device_id = "Device 1"

def on_connect(client, userdata, flags, rc):
    print(f"✅ Connected to MQTT Broker with result code: {rc}")
    
def on_message(client, userdata, msg):
    print(f"📩 Raw Payload from Device 2: {msg.payload.decode()}")
    try:
        response = json.loads(msg.payload.decode())  # Decode the JSON message
        print(f"📥 Received Response: {response}")
    except json.JSONDecodeError:
        print("⚠️ Error decoding JSON from the message")

def main():
    # Create a new MQTT client instance
    client = mqtt.Client()

    # Set up callback functions
    client.on_connect = on_connect
    client.on_message = on_message

    # Connect to the MQTT broker
    print("🔌 Connecting to MQTT broker...")
    client.connect(broker, port, 60)

    # Start the network loop
    client.loop_start()

    # Wait for the connection to establish before processing
    time.sleep(2)

    # Create a random challenge
    challenge = f"Challenge from Device 1: {random.randint(1000000000, 9999999999)}"
    
    # Send challenge to Device 2
    challenge_message = {
        "device_id": device_id,
        "challenge": challenge
    }
    print(f"📤 Sending Challenge: {challenge}")
    
    client.publish(topic, json.dumps(challenge_message))  # Send as JSON string
    
    # Wait for the response
    time.sleep(5)

    # Stop the loop and disconnect
    client.loop_stop()
    client.disconnect()

if __name__ == "__main__":
    main()

