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
	"fmt"
	"os"

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
