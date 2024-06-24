package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

/**********************************************************************************
 *	MQTT Message Handling
 **********************************************************************************/

// UnmarshalJSON custom method for handling different types
// of the amount field.
func (t *CustomChannel) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// Handle Channel which can be "A" or 1, string or int
	var channel string // try to unmarshal to string

	if err := json.Unmarshal(data, &channel); err != nil {
		// Try to convert int to string before failing
		var channelInt int
		if err := json.Unmarshal(data, &channelInt); err != nil {
			// log.Println("Can't unmarshal the channel field")
			SetStatus("Can't unmarshal the channel field")
		}
		t.Channel = strconv.Itoa(channelInt)
		return nil
	}

	// Set the fields to the new struct,
	t.Channel = channel

	return nil
}

// checkSensor - Check if sensor is in active sensor table
func checkSensor(key string, m map[string]Sensor) bool {
	if _, ok := m[key]; ok {
		return true
	}
	return false
}

// Create sensor list
func buildSensorList(m map[string]Sensor) []string {
	var list []string
	for s := range m {
		sens := m[s]
		list = append(list, sens.FormatSensor(1))
	}
	return list
}

// writeWeatherData - Output weather record to appropriate file based on the station (home)
func writeWeatherData(wd WeatherData) {
	datafile := DataFiles[wd.Home].file
	nbytes, err = datafile.WriteString(fmt.Sprintf("station: %s, sensor: %s, location: %s, temp: %.1f, humidity: %.1f, time: %s, model: %s, id: %d, channel: %s\n",
		wd.Home, wd.SensorName, wd.SensorLocation, wd.Temperature_F, wd.Humidity, wd.Time, wd.Model, wd.Id, wd.Channel))
	check(err)
}
