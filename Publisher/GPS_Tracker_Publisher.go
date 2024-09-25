package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// type Data = GPSDataParams

// var files = GPSLogFiles

type Data = ThermostatDataParams

var files = ThermostatLogFiles

// type Data = WeatherDataParams
// var files = WeatherLogFiles

// MQTT settings
const (
	mqttBroker = "tcp://192.168.1.103:31883" // Replace with your MQTT broker address
	topic      = "test/GPS_Tracker"          // Replace with your topic
	clientID   = "go_mqtt_client"
	timeFormat = "15:04:05.000" // Time format for timestamps
)

func main() {
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

	messageCount := 0
	var scanners []*bufio.Scanner

	// Loop through the files to create scanners
	for _, fileName := range files {
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("Failed to open file: %v", err) // Stop program if file fails to open
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanners = append(scanners, scanner)
	}

	// Send initial message with id: 0 and the current time
	initialMessage := Data{
		CurrentTime: time.Now().Format(timeFormat),
	}
	sendMessage(client, initialMessage)

	// Loop through the scanners to process each line
	for _, scanner := range scanners {
		for scanner.Scan() {
			line := scanner.Text()

			// Parse the JSON line into the Data struct
			var data Data
			err := json.Unmarshal([]byte(line), &data)
			if err != nil {
				log.Fatalf("Error parsing line: %v", err)
				continue
			}

			data.CurrentTime = time.Now().Format(timeFormat)
			sendMessage(client, data)
			messageCount++
		}

		// Handle any scanning errors for the current file
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading file: %v", err)
		}
	}

	// Send final message with id: 99 and the current time
	finalMessage := Data{
		CurrentTime: time.Now().Format(timeFormat),
	}
	sendMessage(client, finalMessage)

	// Disconnect from the broker after all files are processed
	client.Disconnect(250)
	fmt.Println("Done sending data from all files.")

	// Parse the initial timestamp
	initialTime, err := time.Parse(timeFormat, initialMessage.CurrentTime)
	if err != nil {
		log.Fatalf("Error parsing initial timestamp: %v", err)
	}

	// Parse the final timestamp
	finalTime, err := time.Parse(timeFormat, finalMessage.CurrentTime)
	if err != nil {
		log.Fatalf("Error parsing final timestamp: %v", err)
	}

	// Calculate the duration
	duration := finalTime.Sub(initialTime)
	log.Printf("Total time taken from first to last message: %v", duration)
	log.Printf("Total real messages sent: %d", messageCount)

}

// Helper function to send a message to the MQTT broker
func sendMessage(client MQTT.Client, data Data) {
	// Convert Data struct to JSON
	msg, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	// Publish the message to the MQTT broker
	token := client.Publish(topic, 0, false, msg)
	token.Wait()
}
