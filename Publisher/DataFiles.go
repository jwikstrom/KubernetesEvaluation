package main

// DataParams struct represents the parameters in the log file + the current time
type GPSDataParams struct {
	CurrentTime string  `json:"current_time"`
	ID          int     `json:"id"`
	Timestamp   string  `json:"timestamp"`
	Lat         float64 `json:"lat"`
	Long        float64 `json:"long"`
}

// LogFiles contains the list of log file paths
var GPSLogFiles = []string{
	"../../Testdata/IoT_normal_GPS_Tracker_1.log",
	"../../Testdata/IoT_normal_GPS_Tracker_2.log",
	"../../Testdata/IoT_normal_GPS_Tracker_3.log",
}

// DataParams struct represents the parameters in the log file + the current time
type ThermostatDataParams struct {
	CurrentTime        string  `json:"current_time"`
	ID                 int     `json:"id"`
	DeviceTitle        string  `json:"device title"`
	CurrentTemperature float64 `json:"current temperature"`
	AC_State           bool    `json:"AC_state"`
	StateOfThermostat  string  `json:"state of thermostat"`
}

// LogFiles contains the list of log file paths
var ThermostatLogFiles = []string{
	"../../Testdata/IoT_normal_Thermostat_1.log",
	"../../Testdata/IoT_normal_Thermostat_2.log",
	"../../Testdata/IoT_normal_Thermostat_3.log",
}

// DataParams struct represents the parameters in the log file + the current time
type WeatherDataParams struct {
	CurrentTime string  `json:"current_time"`
	ID          int     `json:"id"`
	Timestamp   string  `json:"timestamp"`
	Temperature float64 `json:"temperature"`
	Pressure    float64 `json:"pressure"`
	Humidity    float64 `json:"humidity"`
}

// LogFiles contains the list of log file paths
var WeatherLogFiles = []string{
	"../../Testdata/IoT_normal_Weather_1.log",
	"../../Testdata/IoT_normal_Weather_2.log",
	"../../Testdata/IoT_normal_Weather_3.log",
}
