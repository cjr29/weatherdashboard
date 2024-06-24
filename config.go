/*************************
 *
 * config.go - Reads the config.ini file and sets up the connection options, generates
 *			message handlers for each subscribed topic, and opens the data output file
 *			for writing if requested in config.ini
 *
 *			Eventually, the config.ini file will be supplemented by a GUI menu to allow the
 *			user to configure the program dynamically and then preserve the configuration upon exit.
 *
 */
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/ini.v1"
)

var (
	uid      string
	pwd      string
	broker   string
	port     int    = 1883
	clientID string = "weatherdashboard"
	opts            = mqtt.NewClientOptions()
)

/**********************************************************************************
 *	MQTT Message Handling
 **********************************************************************************/

var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	err := json.Unmarshal(msg.Payload(), &incoming)
	if err != nil {
		SetStatus(fmt.Sprintf("Unable to unmarshal JSON due to %s", err))
	}
	outgoing.CopyWDRtoWD(incoming)
	outgoing.Station = strings.Split(msg.Topic(), "/")[0] // station, or home, is the first segment of the msg.Topic
	skey := outgoing.BuildSensorKey()
	// Add sensor to visibleSensors table(map)
	if _, ok := visibleSensors[skey]; !ok {
		// Sensor not in map. Add it.
		sens := outgoing.GetSensorFromData() // Create Sensor record
		visibleSensors[skey] = sens          // Add it to the visible sensors
		SetStatus(fmt.Sprintf("Added sensor to visible sensors: %s", skey))
	}
	// If sensor is active, write to output file
	if checkSensor(skey, activeSensors) {
		s := activeSensors[skey]
		outgoing.Station = s.Station
		outgoing.SensorName = s.Name
		outgoing.SensorLocation = s.Location
		writeWeatherData(outgoing)
		DisplayData(fmt.Sprintf("station: %s, sensor: %s, location: %s, temp: %.1f, humidity: %.1f, time: %s, model: %s, id: %d, channel: %s",
			outgoing.Station, outgoing.SensorName, outgoing.SensorLocation, outgoing.Temperature_F, outgoing.Humidity, outgoing.Time, outgoing.Model, outgoing.Id, outgoing.Channel))
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	go sub(client)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	SetStatus("Connection to broker lost")
}

func sub(client mqtt.Client) {
	for _, m := range messages {
		//SetStatus(fmt.Sprintf("Subscribing to topic ==> %s", m.Topic))
		client.Subscribe(m.Topic, 0, messageHandler)
		SetStatus(fmt.Sprintf("Subscribed to topic %s", m.Topic))
	}
}

func config() {
	/*********************
	* TO DO - Add code to check for config.yaml configuration file and restore from tht instead of config.ini
	**********************/

	inidata, err := ini.Load("config.ini")
	if err != nil {
		SetStatus(fmt.Sprintf("Unable to read configuration file: %v", err))
		os.Exit(1)
	}

	// Retrieve broker info from "broker" section of config.ini file
	section := inidata.Section("broker")
	broker = section.Key("host").String()
	uid = section.Key("username").String()
	pwd = section.Key("password").String()

	// First broker - initialize from the config.ini file
	// FUTURE ENHANCEMENT: Support multiple brokers, data structures already in place
	b := brokers[0]
	b.Path = broker
	b.Uid = uid
	b.Pwd = pwd
	brokers[0] = b // Replace Broker with updated data

	//**********************************
	// Open data output files, one for each message subscription
	//**********************************
	for _, m := range messages {
		fp := "./WeatherData-" + m.Station + ".txt"
		dfile := new(DataFile)
		dfile.path = fp
		dfile.file, err = os.Create(dfile.path)
		if err != nil {
			SetStatus(fmt.Sprintf("Unable to create/open output file. %s", err))
			panic(err.Error)
		}
		dataFiles[m.Station] = *dfile // Add the new DataFile object to the array of data files
		SetStatus(fmt.Sprintf("Opened data file %s", fp))
	}

	/*********************
	* TO DO - Add code to output configuration to a config.yaml file for restoration on next run
	**********************/
}
