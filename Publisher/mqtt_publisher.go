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

// MQTT settings
const (
	mqttBroker = "tcp://192.168.1.103:31883" // Replace with your MQTT broker address
	timeFormat = "01-02 15:04:05.0000Z07:00" // Time format for timestamps with month, date, and timezone
)

var (
	durations      = make(map[string]time.Duration)
	totalDurations = make(map[string]time.Duration)
	mutex          sync.Mutex
)

const numRuns = 1  //2 ~ 1 min || med 1ms delay => ca 7,5 min
const datasets = 8 // 1 dataset = 3 topics

const initialSleep = 1000 * time.Millisecond
const realtimedelay = 50000 * time.Nanosecond

const runTime = 180 * time.Second

func main() {
	var wg sync.WaitGroup

	if datasets > fileSets {
		log.Fatalf("Number of datasets exceeds the number of file sets.")
	}

	// Start separate goroutines for each dataset

	wg.Add(1 * datasets)
	for i := 0; i < datasets; i++ {
		var istr = fmt.Sprintf("%d", i+1)
		// go func() {
		// 	defer wg.Done()
		// 	processDataset("GPS_"+istr, GPSLogFiles[i], "run/GPS_Tracker_"+istr, func() any { return new(GPSDataParams) })
		// }()

		go func() {
			defer wg.Done()
			processDataset("Thermo_"+istr, ThermostatLogFiles[i], "run/Thermo_Sensor_"+istr, func() any { return new(ThermostatDataParams) })
		}()

		// go func() {
		// 	defer wg.Done()
		// 	processDataset("Weather_"+istr, WeatherLogFiles[i], "run/Weather_Station_"+istr, func() any { return new(WeatherDataParams) })
		// }()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("All datasets processed.")

	for topic, duration := range durations {
		fmt.Printf("Topic: %s, Duration: %v\n", topic, duration)
	}
}

// Function to process a dataset with specific log files, topic, and data struct type
func processDataset(datasetName string, files []string, topic string, dataConstructor func() any) {
	var initialTimeClean time.Time
	var finalTimeSentClean time.Time
	// Create a new MQTT client for each dataset
	clientID := fmt.Sprintf("go_mqtt_client_%s", datasetName)
	opts := MQTT.NewClientOptions()
	opts.SetOrderMatters(false)
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
	defer client.Disconnect(500)

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
	initialTimeFormatted := getCurrentTime(initialMessage)
	fmt.Printf("[%s] %s:\tInitial message sent, waiting 1 second then continuing transmit\n", initialTimeFormatted.Format(timeFormat), topic)

	messageCount := 0
	filesCount := len(files)
	runs := 0
	// Wait for 1 second before starting the transmission
	time.Sleep(initialSleep)
	// Loop through the scanners to process each line
	startTime := time.Now()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			fmt.Printf("[%s] %s:\tProgress: %d messages sent\n", time.Now().Format(timeFormat), topic, messageCount)
		}
	}()

	for i, scanner := range scanners {
		if i%filesCount == 0 && i != 0 {
			runs++
			fmt.Printf("[%s] %s:\tRun:%v\n", time.Now().Format(timeFormat), topic, runs)
		}
		for scanner.Scan() {
			if time.Since(startTime) >= runTime {
				break
			}
			line := scanner.Text()

			// Parse the JSON line into the dataset-specific struct
			data := dataConstructor()
			err := json.Unmarshal([]byte(line), data)
			if err != nil {
				log.Printf("[%s] %s: Error parsing line: %v", time.Now().Format(timeFormat), topic, err)
				continue
			}
			setCurrentTime(data)

			if messageCount == 0 {
				setID(data, 1)
				initialTimeFormatted = getCurrentTime(data)
				initialTimeClean = time.Now()
			}
			// Set the current timestamp for the message

			sendMessage(client, topic, data)
			messageCount++
			time.Sleep(realtimedelay)
		}

		// Handle any scanning errors for the current file
		if err := scanner.Err(); err != nil {
			log.Fatalf("[%s] %s: Error reading file: %v", time.Now().Format(timeFormat), topic, err)
		}
		if time.Since(startTime) >= runTime {
			break
		}
	}

	// Send final message with id: 99 and the current time
	finalMessage := dataConstructor()
	setCurrentTime(finalMessage)
	finalTimeSentClean = time.Now()
	setID(finalMessage, 99)

	sendMessage(client, topic, finalMessage)

	finalTime := getCurrentTime(finalMessage)
	duration := finalTime.Sub(initialTimeFormatted)
	fmt.Printf("[%s] %s:\tFinal message sent. Total messages sent: %d. Duration: %v\n", finalTime.Format(timeFormat), topic, messageCount, duration)

	mutex.Lock()
	durations[topic] = duration
	totalDurations[topic] = finalTimeSentClean.Sub(initialTimeClean)
	mutex.Unlock()
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
	token := client.Publish(topic, 0, false, msg)
	token.Wait()
}
