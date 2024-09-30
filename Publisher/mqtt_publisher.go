package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// GPSDataParams represents the parameters for GPS data
type GPSDataParams struct {
	CurrentTime string  `json:"current_time"`
	ID          int     `json:"id"`
	Timestamp   string  `json:"timestamp"`
	Lat         float64 `json:"lat"`
	Long        float64 `json:"long"`
}

// ThermostatDataParams represents the parameters for thermostat data
type ThermostatDataParams struct {
	CurrentTime        string  `json:"current_time"`
	ID                 int     `json:"id"`
	DeviceTitle        string  `json:"device_title"`
	CurrentTemperature float64 `json:"current_temperature"`
	AC_State           bool    `json:"AC_state"`
	StateOfThermostat  string  `json:"state_of_thermostat"`
}

// WeatherDataParams represents the parameters for weather data
type WeatherDataParams struct {
	CurrentTime string  `json:"current_time"`
	ID          int     `json:"id"`
	Timestamp   string  `json:"timestamp"`
	Temperature float64 `json:"temperature"`
	Pressure    float64 `json:"pressure"`
	Humidity    float64 `json:"humidity"`
}

// File lists for each dataset
var (
	GPSLogFiles = []string{
		"../../Testdata/IoT_normal_GPS_Tracker_1.log",
		"../../Testdata/IoT_normal_GPS_Tracker_2.log",
		"../../Testdata/IoT_normal_GPS_Tracker_3.log"}
	ThermostatLogFiles = []string{
		"../../Testdata/IoT_normal_Thermostat_1.log",
		"../../Testdata/IoT_normal_Thermostat_2.log",
		"../../Testdata/IoT_normal_Thermostat_3.log"}
	WeatherLogFiles = []string{
		"../../Testdata/IoT_normal_Weather_1.log",
		"../../Testdata/IoT_normal_Weather_2.log",
		"../../Testdata/IoT_normal_Weather_3.log"}
)

// MQTT settings
const (
	mqttBroker = "tcp://192.168.1.103:31883" // Replace with your MQTT broker address
	timeFormat = "15:04:05.000"              // Time format for timestamps
)
const numRuns = 1

func main() {
	var wg sync.WaitGroup

	// Start separate goroutines for each dataset
	wg.Add(3)

	go func() {
		defer wg.Done()
		processDataset("GPS Tracker", GPSLogFiles, "test/GPS_Tracker", func() any { return new(GPSDataParams) })
	}()

	go func() {
		defer wg.Done()
		processDataset("Thermostat Sensor", ThermostatLogFiles, "test/Thermostat_Sensor", func() any { return new(ThermostatDataParams) })
	}()

	go func() {
		defer wg.Done()
		processDataset("Weather Station", WeatherLogFiles, "test/Weather_Station", func() any { return new(WeatherDataParams) })
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("All datasets processed.")
}

// Function to process a dataset with specific log files, topic, and data struct type
func processDataset(datasetName string, files []string, topic string, dataConstructor func() any) {
	// Create a new MQTT client for each dataset
	clientID := fmt.Sprintf("go_mqtt_client_%s", datasetName)
	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqttBroker)
	opts.SetClientID(clientID)

	// Set up callback to handle connection loss
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		fmt.Printf("[%s] %s: Connection lost: %v\n", time.Now().Format(timeFormat), topic, err)
	})

	// Connect to the MQTT broker
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("[%s] %s: Error connecting to the MQTT broker: %v", time.Now().Format(timeFormat), topic, token.Error())
	}
	defer client.Disconnect(250)

	var scanners []*bufio.Scanner

	// Loop through the files to create scanners
	for run := 0; run < numRuns; run++ {
		for _, fileName := range files {
			file, err := os.Open(fileName)
			if err != nil {
				log.Fatalf("[%s] %s: Failed to open file: %v", time.Now().Format(timeFormat), topic, err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			scanners = append(scanners, scanner)
		}
	}

	// Send initial message with id: 0 and the current time
	initialMessage := dataConstructor()
	setCurrentTime(initialMessage)
	sendMessage(client, topic, initialMessage)
	initialTime := getCurrentTime(initialMessage)
	fmt.Printf("[%s] %s:\tInitial message sent\n", initialTime.Format(timeFormat), topic)

	messageCount := 0
	filesCount := len(files)
	runs := 0
	// Loop through the scanners to process each line
	for i, scanner := range scanners {
		if i%filesCount == 0 {
			runs++
			fmt.Printf("[%s] %s:\tRun:%v\n", initialTime.Format(timeFormat), topic, runs)
		}
		for scanner.Scan() {
			line := scanner.Text()

			// Parse the JSON line into the dataset-specific struct
			data := dataConstructor()
			err := json.Unmarshal([]byte(line), data)
			if err != nil {
				log.Printf("[%s] %s: Error parsing line: %v", time.Now().Format(timeFormat), topic, err)
				continue
			}

			// Set the current timestamp for the message
			setCurrentTime(data)
			sendMessage(client, topic, data)
			messageCount++
		}

		// Handle any scanning errors for the current file
		if err := scanner.Err(); err != nil {
			log.Fatalf("[%s] %s: Error reading file: %v", time.Now().Format(timeFormat), topic, err)
		}
	}

	// Send final message with id: 99 and the current time
	finalMessage := dataConstructor()
	setCurrentTime(finalMessage)
	setID(finalMessage, 99)
	sendMessage(client, topic, finalMessage)

	finalTime := getCurrentTime(finalMessage)
	duration := finalTime.Sub(initialTime)
	fmt.Printf("[%s] %s:\tFinal message sent. Total messages sent: %d. Duration: %v\n", finalTime.Format(timeFormat), topic, messageCount, duration)
}

// Helper function to set the current time in the dataset
func setCurrentTime(data any) {
	switch v := data.(type) {
	case *GPSDataParams:
		v.CurrentTime = time.Now().Format(timeFormat)
	case *ThermostatDataParams:
		v.CurrentTime = time.Now().Format(timeFormat)
	case *WeatherDataParams:
		v.CurrentTime = time.Now().Format(timeFormat)
	}
}

// Helper function to set the ID in the dataset
func setID(data any, id int) {
	switch v := data.(type) {
	case *GPSDataParams:
		v.ID = id
	case *ThermostatDataParams:
		v.ID = id
	case *WeatherDataParams:
		v.ID = id
	}
}

// Helper function to get the current time from the dataset
func getCurrentTime(data any) time.Time {
	switch v := data.(type) {
	case *GPSDataParams:
		t, _ := time.Parse(timeFormat, v.CurrentTime)
		return t
	case *ThermostatDataParams:
		t, _ := time.Parse(timeFormat, v.CurrentTime)
		return t
	case *WeatherDataParams:
		t, _ := time.Parse(timeFormat, v.CurrentTime)
		return t
	}
	return time.Time{}
}

// Helper function to send a message to the MQTT broker
func sendMessage(client MQTT.Client, topic string, data any) {
	// Convert Data struct to JSON
	msg, err := json.Marshal(data)
	if err != nil {
		log.Printf("[%s] %s: Error marshaling message: %v", time.Now().Format(timeFormat), topic, err)
		return
	}

	// Publish the message to the MQTT broker
	client.Publish(topic, 0, false, msg)
	//token.Wait()
}
