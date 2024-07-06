/******************************************************************
 *
 * MQTT message handling functions - used to process incoming data
 *
 ******************************************************************/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var opts = mqtt.NewClientOptions()

/**********************************************************************************
 *	MQTT Message Handling
 **********************************************************************************/

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var incoming WeatherDataRaw
	var outgoing WeatherData

	err := json.Unmarshal(msg.Payload(), &incoming)
	if err != nil {
		fmt.Println("messageHandler: Unable to unmarshal JSON due to %s", err)
		SetStatus(fmt.Sprintf("messageHandler: Unable to unmarshal JSON due to %s", err))
	}
	outgoing.CopyWDRtoWD(incoming)
	outgoing.Station = strings.Split(msg.Topic(), "/")[0] // station, or home, is the first segment of the msg.Topic
	skey := outgoing.BuildSensorKey()
	// Add sensor to availableSensors table(map) if not already there AND if not already in activeSensors
	if !checkSensor(skey, activeSensors) {
		// Sensor not in active sensors map
		if _, ok := availableSensors[skey]; !ok {
			// Sensor not in available sensors map. Add it.
			sens := outgoing.GetSensorFromData() // Create Sensor record
			sens.init(skey)                      // Initialize sensor record
			availableSensors[skey] = &sens       // Add it to the visible sensors
			SetStatus(fmt.Sprintf("Added sensor to visible sensors: %s", skey))
		}
	} else {
		// Sensor is active, write record to output file
		s := *activeSensors[skey]
		outgoing.Station = s.Station
		outgoing.SensorName = s.Name
		outgoing.SensorLocation = s.Location
		// Update Sensor's WeatherWidget if not hidden and widget exists
		if checkWeatherWidget(skey) && !s.Hide {
			nd := newData{skey, outgoing.Temperature_F, outgoing.Humidity, outgoing.Time}
			// Use a go routine to prevent blocking of this event handler
			// Each incoming data record gets its own goroutine
			go notifyWidget(nd)

			if dashboardContainer != nil {
				dashboardContainer.Refresh()
			}
		}
		// Log data if data logging is turned on
		if logdata_flg {
			writeWeatherData(outgoing)
		}
		// Always write record to the data display scrolling console
		DisplayData(fmt.Sprintf("station: %s, sensor: %s, location: %s, temp: %.1f, humidity: %.1f, time: %s, model: %s, id: %d, channel: %s",
			outgoing.Station, outgoing.SensorName, outgoing.SensorLocation, outgoing.Temperature_F, outgoing.Humidity, outgoing.Time, outgoing.Model, outgoing.Id, outgoing.Channel))
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	go sub(client)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	SetStatus(fmt.Sprintf("%s : Connection to broker lost", st))
}

// Send channel message to goroutine to update widget. Runs once and quits.
func notifyWidget(nd newData) {

	key := nd.key
	temp := nd.temp
	humidity := nd.humidity
	date := nd.date

	// Sometimes, the temp is blank from the sensor, so check if 0.0, don't update
	// Simple test of value changed from last time. If it didn't, do nothing
	if (temp != activeSensors[key].Temp) && (temp != 0) {
		weatherWidgets[key].temp = temp
		activeSensors[key].Temp = temp
	}
	if (humidity != activeSensors[key].Humidity) && (humidity != 0) {
		weatherWidgets[key].humidity = humidity
		activeSensors[key].Humidity = humidity
	}
	weatherWidgets[key].latestUpdate = date
	activeSensors[key].DataDate = date

	//Calculate & update highs and lows
	if (temp != 0) && (temp > weatherWidgets[key].highTemp) {
		weatherWidgets[key].highTemp = temp
		activeSensors[key].HighTemp = temp
	}
	if (temp != 0) && (temp < weatherWidgets[key].lowTemp) {
		weatherWidgets[key].lowTemp = temp
		activeSensors[key].LowTemp = temp
	}
	if (humidity != 0) && (humidity > (weatherWidgets[key].highHumidity)) {
		weatherWidgets[key].highHumidity = humidity
		activeSensors[key].HighHumidity = humidity
	}
	if (humidity != 0) && (humidity < weatherWidgets[key].lowHumidity) {
		weatherWidgets[key].lowHumidity = humidity
		activeSensors[key].LowHumidity = humidity
	}
	// Send channel signal to background processor
	weatherWidgets[key].channel <- key
}

func sub(client mqtt.Client) {
	for _, m := range messages {
		//SetStatus(fmt.Sprintf("Subscribing to topic ==> %s", m.Topic))
		client.Subscribe(m.Topic, 0, messageHandler)
		SetStatus(fmt.Sprintf("Subscribed to topic %s", m.Topic))
	}
}

func unsubscribe(client mqtt.Client, msg Message) {
	//SetStatus(fmt.Sprintf("Unsubscribing from topic ==> %s", m.Topic))
	client.Unsubscribe(msg.Topic)
	SetStatus(fmt.Sprintf("Unsubscribed from topic %s", msg.Topic))
}

// UnmarshalJSON custom method for handling different types
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
func checkSensor(key string, s map[string]*Sensor) bool {
	if _, ok := s[key]; ok {
		return true
	}
	return false
}

// checkMessage - Check if message is in message table
func checkMessage(key int, m map[int]Message) bool {
	if _, ok := m[key]; ok {
		return true
	}
	return false
}

// Create sensor list
func buildSensorList(m map[string]*Sensor) []string {
	var list []string
	for s := range m {
		sens := *m[s]
		list = append(list, sens.FormatSensor(1))
	}
	return list
}

// Create message (topic) list
func buildMessageList(m map[int]Message) []ChoicesIntKey {
	var list []ChoicesIntKey
	i := 0
	for key, message := range m {
		var c ChoicesIntKey
		c.Display = strconv.Itoa(i) + ": " + message.FormatMessage(1)
		c.Key = key
		list = append(list, c)
		i++
	}
	return list
}

// writeWeatherData - Output weather record to appropriate file based on the station (home)
func writeWeatherData(wd WeatherData) {
	datafile := dataFiles[wd.Station].file
	_, err := datafile.WriteString(fmt.Sprintf("station: %s, sensor: %s, location: %s, temp: %.1f, humidity: %.1f, time: %s, model: %s, id: %d, channel: %s\n",
		wd.Station, wd.SensorName, wd.SensorLocation, wd.Temperature_F, wd.Humidity, wd.Time, wd.Model, wd.Id, wd.Channel))
	check(err)
}
