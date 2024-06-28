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
	"log"
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
		var m Message
		fmt.Println("Enter full path to weather broker: ")
		fmt.Scanln(&b.Path)
		b.Port = 1883
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", b.Path, b.Port))
		fmt.Println("Enter user Id: ")
		fmt.Scanln(&b.Uid)
		opts.SetUsername(b.Uid)
		fmt.Printf("Enter password to weather broker %s: \n", b.Path)
		fmt.Scanln(&b.Pwd)
		opts.SetPassword(b.Pwd)
		opts.SetClientID(clientID)
		// Copy properties into the brokers array
		brokers = append(brokers, b)
		fmt.Println("Enter a topic to subscribe to:")
		fmt.Scanln(&m.Topic)
		fmt.Println("Enter station name:")
		fmt.Scanln(&m.Station)
		// Add to messages
		key := rand.Int()
		messages[key] = m
	}

	// Disable data logging
	logdata_flg = false

	//**********************************
	// Open data output files, one for each message subscription
	//**********************************
	for _, m := range messages {
		fp := "./WeatherData-" + m.Station + ".txt"
		dfile := new(DataFile)
		dfile.path = fp
		//dfile.file, err = os.Create(dfile.path)
		dfile.file, err = os.OpenFile(fp, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			SetStatus(fmt.Sprintf("Unable to create/open output file. %s", err))
			log.Println(fmt.Sprintf("Unable to create/open output file. %s", err))
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
	c := Configuration{
		Brokers:       brokers,
		Messages:      messages,
		ActiveSensors: activeSensors,
	}
	data, _ := json.MarshalIndent(c, "", "    ")
	//fmt.Println(string(data))
	err := os.WriteFile("config.json", data, 0644)
	return err
}

// jsonInput() - Read configuration information from a .json file
func jsonInput() (e error) {
	// Declare a configuration structure composed of the structures and maps needed
	c := Configuration{
		Brokers:       brokers,
		Messages:      messages,
		ActiveSensors: activeSensors,
	}

	inidata, err := os.ReadFile("config.json")
	if err != nil {
		log.Println("Unable to open config.json file.")
		return err
	}
	//fmt.Print(string(inidata))

	err = json.Unmarshal(inidata, &c)
	if err != nil {
		log.Printf("Unable to unmarshal JSON due to %s", err)
		return err
	}

	brokers = nil

	// Copy the configuration information into the data structures.
	brokers = append(brokers, c.Brokers...)
	for key, value := range c.Messages {
		messages[key] = value
	}
	for key, value := range c.ActiveSensors {
		activeSensors[key] = value
	}

	/* for _, b := range brokers {
		fmt.Printf("Broker path: %s, port: %d\n", b.Path, b.Port)
	}
	for _, m := range messages {
		fmt.Printf("Subscribed to: %s\n", m.Topic)
	}
	for _, s := range activeSensors {
		fmt.Printf("Active sensor: %s\n", s.Key)
	} */

	return nil
}
