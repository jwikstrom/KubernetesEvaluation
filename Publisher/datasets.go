package main

import "fmt"

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
	DeviceTitle        string  `json:"device title"`
	CurrentTemperature float64 `json:"current temperature"`
	AC_State           bool    `json:"AC_state"`
	StateOfThermostat  string  `json:"state of thermostat"`
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

const fileSets = 16

func generateFilePaths(basePath, fileType string) [][]string {
	var filePaths [][]string
	for i := 0; i <= fileSets; i++ {
		var copyPath string
		if i == 0 {
			copyPath = basePath
		} else if i == 1 {
			copyPath = fmt.Sprintf("%s - kopia", basePath)
		} else {
			copyPath = fmt.Sprintf("%s - kopia (%d)", basePath, i)
		}
		filePaths = append(filePaths, []string{
			fmt.Sprintf("%s/%s_1.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_2.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_3.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_4.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_5.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_6.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_7.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_8.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_9.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_10.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_11.log", copyPath, fileType),
			fmt.Sprintf("%s/%s_12.log", copyPath, fileType),
		})
	}
	return filePaths
}

// File lists for each dataset
var (
	GPSLogFiles        = generateFilePaths("../Testdata/Files", "IoT_normal_GPS_Tracker")
	WeatherLogFiles    = generateFilePaths("../Testdata/Files", "IoT_normal_Weather")
	ThermostatLogFiles = generateFilePaths("../Testdata/Files", "IoT_normal_Thermostat")
)
