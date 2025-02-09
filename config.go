/****************************************************************
 *
 * Program configuration functions
 *
 * config.go - Reads the config.json file and sets up the connection options, generates
 *			message handlers for each subscribed topic, and opens the data output file
 *			for writing if requested by the data logging flag in the dashboard.
 *			When program exits, rewrites config.json with the current configuration.
 *
******************************************************************/

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

var clientID string = "weatherdashboard"

func readConfig() {
	/*********************
	* TO DO - Add code to check for config.yaml configuration file and restore from that instead of config.ini
	**********************/

	// Read config from the config.json file
	err := jsonInput()
	if err != nil {
		// Can't open json file, ask user for broker info
		SetStatus("Unable to find or open the config.json file. Asking user for broker info.")
		var b Broker
		var m Subscription
		fmt.Println("Enter full path to weather broker: ")
		fmt.Scanln(&b.Path)
		b.Port = 1883
		// opts.AddBroker(fmt.Sprintf("tcp://%s:%d", b.Path, b.Port))
		fmt.Println("Enter user Id: ")
		fmt.Scanln(&b.Uid)
		// opts.SetUsername(b.Uid)
		fmt.Printf("Enter password to weather broker %s: \n", b.Path)
		fmt.Scanln(&b.Pwd)
		// opts.SetPassword(b.Pwd)
		// opts.SetClientID(clientID)
		// Copy properties into the brokers array
		bkey := rand.Int()
		brokers[bkey] = b
		fmt.Println("Enter a topic to subscribe to:")
		fmt.Scanln(&m.Topic)
		fmt.Println("Enter station name:")
		fmt.Scanln(&m.Station)
		// Add to subscriptions
		skey := rand.Int()
		subscriptions[skey] = &m
	}

	// Disable data logging
	logdata_flg = false

	//**********************************
	// Open data output files, one for each subscription
	//**********************************
	for _, m := range subscriptions {
		fp := "./WeatherData-" + m.Station + ".txt"
		dfile := new(DataFile)
		dfile.path = fp
		dfile.file, err = os.OpenFile(fp, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			SetStatus(fmt.Sprintf("Unable to create/open output file. %s", err))
			fmt.Printf("Unable to create/open output file. %s", err)
			panic(err.Error)
		}
		dataFiles[m.Station] = *dfile // Add the new DataFile object to the array of data files
		SetStatus(fmt.Sprintf("Opened data file %s", fp))
	}
}

func writeConfig() {
	err := jsonOutput()
	check(err)
}

// jsonOutput() - Write out all configuration options to a .json file
func jsonOutput() (e error) {

	// Declare a configuration structure composed of the structures and maps needed
	var as = make(map[string]Sensor)
	var subs = make(map[int]Subscription)

	// retrieve all current active sensors, making sure to dereference the address first!
	for key := range activeSensors {
		a := *activeSensors[key]
		as[key] = a
	}

	for key := range subscriptions {
		s := *subscriptions[key]
		subs[key] = s
	}

	c := Configuration{
		Brokers:       brokers,
		Subscriptions: subs,
		ActiveSensors: as,
	}

	data, _ := json.MarshalIndent(c, "", "    ")
	err := os.WriteFile("config.json", data, 0644)
	return err
}

// jsonInput() - Read configuration information from a .json file
func jsonInput() (e error) {
	// Declare a configuration structure composed of the structures and maps needed
	var as map[string]Sensor
	var subs map[int]Subscription

	c := Configuration{
		Brokers:       brokers,
		Subscriptions: subs,
		ActiveSensors: as,
	}

	inidata, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Unable to open config.json file.")
		return err
	}

	err = json.Unmarshal(inidata, &c)
	if err != nil {
		fmt.Printf("Unable to unmarshal JSON due to %s", err)
		return err
	}

	//brokers = nil // Erase current list
	// Copy the configuration information into the data structures.
	//brokers = append(brokers, c.Brokers...)

	// Load the input brokers
	for key, value := range c.Brokers {
		brokers[key] = value
	}

	// Load the input subscriptions
	for key, value := range c.Subscriptions {
		subscriptions[key] = &value
	}

	// Load the input sensors, being sure to store the address of the sensor, not the sensor
	for key, value := range c.ActiveSensors {
		activeSensors[key] = &value
	}

	return nil
}
