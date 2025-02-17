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
	CurrentTime        string  `json:"current_time"`
	ID                 int     `json:"id"`
	DeviceTitle        string  `json:"device_title"`
	CurrentTemperature float64 `json:"current_temperature"`
	AC_State           bool    `json:"AC_state"`
	StateOfThermostat  string  `json:"state_of_thermostat"`
}

// MQTT settings
const (
	defaultMqttBroker = "tcp://192.168.1.103:31883" // Replace with your MQTT broker address
	clientID          = "go_mqtt_subscriber"
	timeFormat        = "01-02 15:04:05.0000Z07:00" // Time format for timestamps with month, date, and timezone
)

// Map each topic to its corresponding message handler
var defaultTopics = []string{
	//"run/GPS_Tracker",
	"run/Thermo_Sensor",
	//"run/Weather_Station",
}

var (
	totalTempChan       = make(chan float64, 100)
	messageCounterChan  = make(chan int, 100)
	tempVarianceSumChan = make(chan float64, 100)
	minTemperatureChan  = make(chan float64, 100)
	maxTemperatureChan  = make(chan float64, 100)
)

const defaultTopicSets = 16
const expectedMessages = 1500741

func main() {

	// Initialize channels with initial values
	totalTempChan <- 0
	messageCounterChan <- 0
	tempVarianceSumChan <- 0
	minTemperatureChan <- 9999
	maxTemperatureChan <- -9999

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

	// Set up callback to handle connection loss
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		fmt.Printf("Connection lost: %v\n", err)
	})

	// Connect to the MQTT broker
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to the MQTT broker: %v", token.Error())
	}
	defer client.Disconnect(500)

	// Subscribe to all topics
	for i := 1; i <= topicSets; i++ {
		for _, topic := range topics {
			topicWithSuffix := fmt.Sprintf("%s_%d", topic, i)
			go subscribeToTopic(client, topicWithSuffix)
			//sleep 100ms so that the subscribers don't all subscribe at the same time
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Keep the main function alive indefinitely to keep receiving messages
	select {} // Block forever
}

func subscribeToTopic(client MQTT.Client, topic string) {
	var initialSubTime, finalSubTime time.Time
	var initialPubTime, finalPubTime time.Time
	var receivedTime, publishedTime time.Time
	var messageCount int
	var transmissionTimes []time.Duration // Track transmission times for this topic

	// Create a message handler for the subscription
	messageHandler := func(client MQTT.Client, msg MQTT.Message) {

		// Parse the received message into the Message struct
		var receivedMsg Message
		err := json.Unmarshal(msg.Payload(), &receivedMsg)
		if err != nil {
			log.Printf("[%s] %s: Error unmarshaling message: %v\n", time.Now().Format(timeFormat), topic, err)
			return
		}

		// Log the message data for debugging
		//log.Print(len(transmissionTimes), string(msg.Payload()))
		//fmt.Printf("[%s] %s:\t %v %s\n", receivedTime.Format(timeFormat), topic, len(transmissionTimes), string(msg.Payload()))

		if receivedMsg.ID == 0 {
			// Initial message (id: 0)
			messageCount = 0
			transmissionTimes = []time.Duration{} // Empty the transmissionTimes slice

			fmt.Printf("[%s] %s:\tReceived initial message\n", time.Now().Format(timeFormat), topic)
		} else if messageCount >= expectedMessages {
			fmt.Printf("[%s] %s:\t %v %s\n", receivedTime.Format(timeFormat), topic, messageCount, string(msg.Payload()))
			// Final message (id: 99)
			finalSubTime = receivedTime
			finalPubTime = publishedTime

			subDuration := finalSubTime.Sub(initialSubTime)
			pubDuration := finalPubTime.Sub(initialPubTime)
			totalDuration := finalSubTime.Sub(initialPubTime)

			//log.Printf("[%s] %s:\tTransmission times: %v\n", finalSubTime.Format(timeFormat), topic, transmissionTimes)

			fmt.Printf("[%s] %s:\tTotal received: %d|%d. || SubDuration: %v, PubDuration: %v, TotalDuration: %v, AvgTransmissionTime: %v \n",
				finalSubTime.Format(timeFormat), topic,
				len(transmissionTimes), messageCount, subDuration, pubDuration, totalDuration, getAverageTransmissionTime(transmissionTimes))
			sendAck(client, topic)

		} else {
			// Regular message
			receivedTime = getcurrentTimeFormatted()
			if messageCount%(150000) == 0 {
				fmt.Printf("[%s] %s:\t %v %s\n", receivedTime.Format(timeFormat), topic, messageCount, string(msg.Payload()))
			}

			// Parse the publish time from the message
			publishedTime, err = time.Parse(timeFormat, receivedMsg.CurrentTime)
			if err != nil {
				log.Printf("[%s] %s: Error parsing message timestamp: %v\n", time.Now().Format(timeFormat), topic, err)
				return
			}

			messageCount++
			if messageCount == 1 {
				// First received message
				initialSubTime = receivedTime
			}
			if receivedMsg.ID == 1 {
				// First published message
				initialPubTime = publishedTime
			}

			transmissionTime := receivedTime.Sub(publishedTime)

			// Store the transmission time
			transmissionTimes = append(transmissionTimes, transmissionTime)
			// // log receivedtime and publishedtime
			// if len(transmissionTimes) < 5 {
			// 	log.Printf("[%s] %s:\tP:%s || S:%s\n", receivedTime.Format(timeFormat), topic, publishedTime, receivedTime)
			// }
			// Calculate transmission time (time difference between sent and received)
			// Handle regular message

			processRegularMessage(msg)
		}
	}

	// Subscribe to the topic
	if token := client.Subscribe(topic, 0, messageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("%s: Error subscribing to topic: %v", topic, token.Error())
	}

	fmt.Printf("%s:\tSubscribed to topic, waiting for messages...\n", topic)
}

func getcurrentTimeFormatted() time.Time {
	var currentTime, err = time.Parse(timeFormat, time.Now().Format(timeFormat))
	if err != nil {
		log.Printf("[%s]: Error formatting time now: %v\n", time.Now().Format(timeFormat), err)
	}
	return currentTime
}

func getAverageTransmissionTime(transmissionTimes []time.Duration) time.Duration {
	if len(transmissionTimes) == 0 {
		log.Println("No transmission times to calculate average")
	}

	var total time.Duration
	for _, t := range transmissionTimes {
		total += t
	}

	average := total / time.Duration(len(transmissionTimes))
	return average
}

func processRegularMessage(msg MQTT.Message) {
	// Parse the message
	var receivedMsg Message
	err := json.Unmarshal(msg.Payload(), &receivedMsg)
	if err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	// Validate the message data with more conditions
	if !validateData(receivedMsg.CurrentTemperature, receivedMsg.AC_State, receivedMsg.StateOfThermostat) {
		log.Printf("Validation failed for message ID: %d", receivedMsg.ID)
		return
	}

	// Data transformation: Convert temperature to Fahrenheit
	temperatureFahrenheit := transformTemperatureToFahrenheit(receivedMsg.CurrentTemperature)

	// Perform data aggregation and advanced statistical calculations
	updateStatistics(receivedMsg.CurrentTemperature)

	// Perform more statistical calculations (e.g., variance, min/max)
	calculateVariance(receivedMsg.CurrentTemperature)
	updateMinMax(receivedMsg.CurrentTemperature)

	// Log the transformed and aggregated data, including advanced statistics
	log.Printf("Processed message ID: %d, Device: %s, Temperature: %.2f°C (%.2f°F), Avg Temp: %.2f°C, Variance: %.4f, AC State: %t, Thermostat State: %s",
		receivedMsg.ID, receivedMsg.DeviceTitle, receivedMsg.CurrentTemperature, temperatureFahrenheit, getAverageTemp(), getTemperatureVariance(), receivedMsg.AC_State, receivedMsg.StateOfThermostat)
}

func validateData(temperature float64, acState bool, thermostatState string) bool {
	// Extended temperature range validation
	if temperature < -50 || temperature > 60 {
		log.Printf("Invalid temperature: %.2f", temperature)
		return false
	}

	// Check consistency of AC state and thermostat state
	if acState && thermostatState != "cooling" {
		log.Printf("AC is on but thermostat state is not 'cooling': AC_state=%t, thermostat_state=%s", acState, thermostatState)
		return false
	}
	if !acState && thermostatState != "heating" && thermostatState != "off" {
		log.Printf("AC is off but thermostat state is not valid: AC_state=%t, thermostat_state=%s", acState, thermostatState)
		return false
	}

	// Validate that thermostat state is in allowed values
	validStates := []string{"off", "cooling", "heating"}
	for _, state := range validStates {
		if thermostatState == state {
			return true
		}
	}

	log.Printf("Invalid thermostat state: %s", thermostatState)
	return false
}

func transformTemperatureToFahrenheit(tempCelsius float64) float64 {
	// Transform temperature from Celsius to Fahrenheit
	return (tempCelsius * 9 / 5) + 32
}

func updateStatistics(temp float64) {
	totalTemp := <-totalTempChan
	messageCounter := <-messageCounterChan

	totalTemp += temp
	messageCounter++

	totalTempChan <- totalTemp
	messageCounterChan <- messageCounter
}

func getAverageTemp() float64 {
	totalTemp := <-totalTempChan
	messageCounter := <-messageCounterChan

	if messageCounter == 0 {
		totalTempChan <- totalTemp
		messageCounterChan <- messageCounter
		return 0
	}

	average := totalTemp / float64(messageCounter)
	totalTempChan <- totalTemp
	messageCounterChan <- messageCounter
	return average
}

func calculateVariance(temp float64) {
	tempVarianceSum := <-tempVarianceSumChan

	avgTemp := getAverageTemp()
	variance := (temp - avgTemp) * (temp - avgTemp)
	tempVarianceSum += variance

	tempVarianceSumChan <- tempVarianceSum
}

func getTemperatureVariance() float64 {
	tempVarianceSum := <-tempVarianceSumChan
	messageCounter := <-messageCounterChan

	if messageCounter == 0 {
		tempVarianceSumChan <- tempVarianceSum
		messageCounterChan <- messageCounter
		return 0
	}

	variance := tempVarianceSum / float64(messageCounter)
	tempVarianceSumChan <- tempVarianceSum
	messageCounterChan <- messageCounter
	return variance
}

func updateMinMax(temp float64) {
	minTemperature := <-minTemperatureChan
	maxTemperature := <-maxTemperatureChan

	if temp < minTemperature {
		minTemperature = temp
	}
	if temp > maxTemperature {
		maxTemperature = temp
	}

	minTemperatureChan <- minTemperature
	maxTemperatureChan <- maxTemperature
}

func sendAck(client MQTT.Client, topic string) {
	ackTopic := fmt.Sprintf("%s/ack", topic) // The ack topic is based on the main topic
	ackMessage := "ack"

	// Publish acknowledgment message to the MQTT broker
	token := client.Publish(ackTopic, 0, false, ackMessage)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Error publishing ack message to %s: %v", ackTopic, token.Error())
	} else {
		fmt.Printf("[%s] %s:\tAcknowledgment sent to %s\n", time.Now().Format(timeFormat), topic, ackTopic)
	}
}
