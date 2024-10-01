package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// General structure for messages with an ID and timestamp
type Message struct {
	ID          int    `json:"id"`
	CurrentTime string `json:"current_time"`
}

// MQTT settings
const (
	defaultMqttBroker = "tcp://192.168.1.103:31883" // Replace with your MQTT broker address
	clientID          = "go_mqtt_subscriber"
	timeFormat        = "15:04:05.000" // Time format for timestamps
)

// Map each topic to its corresponding message handler
var defaultTopics = []string{
	"run/GPS_Tracker",
	"run/Thermo_Sensor",
	"run/Weather_Station",
}

const defaultTopicSets = 1

func main() {

	// Read the broker address from the environment variable
	mqttBroker := os.Getenv("MQTT_BROKER")
	if mqttBroker == "" {
		log.Printf("MQTT_BROKER environment variable not set")
		mqttBroker = defaultMqttBroker
	}

	// Read the topics from the environment variable and split them into a slice
	topicsEnv := os.Getenv("MQTT_TOPICS")
	topics := strings.Split(topicsEnv, ",") // Split by commas to get individual topics
	if topicsEnv == "" {
		log.Printf("MQTT_TOPICS environment variable not set")
		topics = defaultTopics
	}

	// Read the amount of topic sets from the environment variable
	topicSetsStr := os.Getenv("MQTT_TOPIC_SETS")
	topicSets, err := strconv.Atoi(topicSetsStr)
	if err != nil {
		log.Printf("MQTT_TOPIC_SETS environment variable not set or invalid, using default value")
		topicSets = defaultTopicSets
	}

	// Create a new MQTT client
	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqttBroker)
	opts.SetClientID(clientID)

	// Set up callback to handle connection loss
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		fmt.Printf("Connection lost: %v\n", err)
	})

	// Connect to the MQTT broker
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to the MQTT broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// Subscribe to all topics
	for i := 1; i <= topicSets; i++ {
		for _, topic := range topics {
			topicWithSuffix := fmt.Sprintf("%s_%d", topic, i)
			go subscribeToTopic(client, topicWithSuffix)
			//sleep 100ms so that the subscribers don't all subscribe at the same time
			time.Sleep(00 * time.Millisecond)
		}
	}

	// Keep the main function alive indefinitely to keep receiving messages
	select {} // Block forever
}

func subscribeToTopic(client MQTT.Client, topic string) {
	var initialTime, finalTime time.Time
	var messageCount int

	// Create a message handler for the subscription
	messageHandler := func(client MQTT.Client, msg MQTT.Message) {
		messageCount++

		// Parse the received message into the Message struct
		var receivedMsg Message
		err := json.Unmarshal(msg.Payload(), &receivedMsg)
		if err != nil {
			log.Printf("[%s] %s: Error unmarshaling message: %v\n", time.Now().Format(timeFormat), topic, err)
			return
		}

		// Log the message data for debugging
		//log.Print(messageCount, string(msg.Payload()))

		// Check for initial message (id: 0)
		if receivedMsg.ID == 0 {
			messageCount = 0
			initialTime = time.Now()
			fmt.Printf("[%s] %s:\tReceived initial message\n", initialTime.Format(timeFormat), topic)
		}

		// Check for final message (id: 99)
		if receivedMsg.ID == 99 {
			finalTime = time.Now()
			duration := finalTime.Sub(initialTime)
			fmt.Printf("[%s] %s:\tFinal message received. Total messages received: %d. Duration: %v\n", finalTime.Format(timeFormat), topic, messageCount-1, duration)
		}
	}

	// Subscribe to the topic
	if token := client.Subscribe(topic, 0, messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("%s: Error subscribing to topic: %v", topic, token.Error())
	}

	fmt.Printf("%s:\tSubscribed to topic, waiting for messages...\n", topic)
}
