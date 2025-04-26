import paho.mqtt.client as mqtt
import json
import time
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.hazmat.primitives import hashes

# Broker details
broker = "localhost"  # Change to your broker IP if running on a different machine
port = 1883
topic = "device/challenge"

# The client ID for this device
device_id = "Device 2"

# Initialize an elliptic curve private key (ECDSA)
private_key = ec.generate_private_key(ec.SECP256R1())

def on_connect(client, userdata, flags, rc):
    print(f"✅ Connected to MQTT Broker with result code: {rc}")
    # Subscribe to the topic where Device 1 sends the challenge
    client.subscribe(topic)

def on_message(client, userdata, msg):
    print(f"📩 Raw Payload from Device 1: {msg.payload.decode()}")
    try:
        # Parse the message from Device 1
        challenge_data = json.loads(msg.payload.decode())
        print(f"📥 Received Challenge: {challenge_data['challenge']}")
        
        # Sign the challenge
        signature = sign_message(private_key, challenge_data['challenge'])
        
        # Send back the response with the signature
        response = {
            "device_id": device_id,
            "challenge": challenge_data['challenge'],
            "signature": signature.hex()  # Convert to hex for easier transmission
        }
        
        print(f"🔏 Signed Challenge: {signature.hex()}")
        print(f"📤 Sent Response with Signature: {response}")
        
        # Publish the response to the same topic
        client.publish(topic, json.dumps(response))

    except json.JSONDecodeError:
        print("⚠️ Error decoding JSON from the message")

def sign_message(private_key, message):
    signature = private_key.sign(
        message.encode(),  # Encode string to bytes
        ec.ECDSA(hashes.SHA256())
    )
    return signature

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
    
    # Keep the script running
    client.loop_forever()

if __name__ == "__main__":
    main()

