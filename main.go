package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	pwd      string
	broker   string = ""
	port     int    = 1883
	clientID string = "weatherdashboard"
	uid      string = "weather"
	topic1   string = "home/weather/sensors"
	topic2   string = "bus/weather/sensors"
	opts            = mqtt.NewClientOptions()
	datafile *os.File
	// bufdatafile *bufio.Writer
	nbytes   int
	err      error
	incoming WeatherDataRaw
	outgoing WeatherData
)

type WeatherDataRaw struct {
	Time          string        `json:"time"`          //"2024-06-11 10:33:52"
	Model         string        `json:"model"`         //"Acurite-5n1"
	Message_type  int           `json:"message_type"`  //56
	Id            int           `json:"id"`            //1997
	Channel       CustomChannel `json:"channel"`       //"A" or 1
	Sequence_num  int           `json:"sequence_num"`  //0
	Battery_ok    int           `json:"battery_ok"`    //1
	Wind_avg_mi_h float64       `json:"wind_avg_mi_h"` //4.73634
	Temperature_F float64       `json:"temperature_F"` //69.4
	Humidity      float64       `json:"humidity"`      // Can appear as integer or a decimal value
	Mic           string        `json:"mic"`           //"CHECKSUM"
}

type CustomChannel struct {
	Channel string
}

func (cc *CustomChannel) channel() string {
	return cc.Channel
}

type WeatherData struct {
	Time          string  `json:"time"`          //"2024-06-11 10:33:52"
	Model         string  `json:"model"`         //"Acurite-5n1"
	Message_type  int     `json:"message_type"`  //56
	Id            int     `json:"id"`            //1997
	Channel       string  `json:"channel"`       //"A" or 1
	Sequence_num  int     `json:"sequence_num"`  //0
	Battery_ok    int     `json:"battery_ok"`    //1
	Wind_avg_mi_h float64 `json:"wind_avg_mi_h"` //4.73634
	Temperature_F float64 `json:"temperature_F"` //69.4
	Humidity      float64 `json:"humidity"`      // Can appear as integer or a decimal value
	Mic           string  `json:"mic"`           //"CHECKSUM"
}

var messageHandler1 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	// Sometimes, JSON for channel returns an integer instead of a letter. Check and convert to string.
	err := json.Unmarshal(msg.Payload(), &incoming)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON due to %s", err)
	}
	copyWDRtoWD()
	printWeatherData(outgoing, "home")
}

var messageHandler2 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	err := json.Unmarshal(msg.Payload(), &incoming)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON due to %s", err)
	}
	copyWDRtoWD()
	printWeatherData(outgoing, "bus")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
	go sub(client)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

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
			log.Println("Can't unmarshal the channel field")
		}
		t.Channel = strconv.Itoa(channelInt)
		return nil
	}

	// Set the fields to the new struct,
	t.Channel = channel

	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func printWeatherData(wd WeatherData, home string) {
	nbytes, err = datafile.WriteString(fmt.Sprintf("station: %s, time: %s, model: %s, id: %d, channel: %s, temp: %.1f, humidity: %.1f\n",
		home, wd.Time, wd.Model, wd.Id, wd.Channel, wd.Temperature_F, wd.Humidity))
	check(err)
}

func sub(client mqtt.Client) {
	log.Printf("Subscribing to topic ==> %s\n", topic1)
	client.Subscribe(topic1, 0, messageHandler1)
	log.Printf("Subscribed to topic %s\n", topic1)
	log.Printf("Subscribing to topic ==> %s\n", topic2)
	client.Subscribe(topic2, 0, messageHandler2)
	log.Printf("Subscribed to topic %s\n", topic2)
}

func copyWDRtoWD() {
	outgoing.Time = incoming.Time
	outgoing.Model = incoming.Model
	outgoing.Message_type = incoming.Message_type
	outgoing.Id = incoming.Id
	outgoing.Channel = incoming.Channel.channel()
	outgoing.Sequence_num = incoming.Sequence_num
	outgoing.Battery_ok = incoming.Battery_ok
	outgoing.Wind_avg_mi_h = incoming.Wind_avg_mi_h
	outgoing.Temperature_F = incoming.Temperature_F
	outgoing.Humidity = incoming.Humidity
	outgoing.Mic = incoming.Mic
}

func main() {
	datafile, err = os.Create("./WeatherData.txt")
	if err != nil {
		log.Fatal("Unable to create/open output file.\n", err)
		panic(err.Error)
	}
	defer datafile.Close()
	datafile.Sync()
	// bufdatafile = bufio.NewWriter(datafile)
	// defer bufdatafile.Flush()

	fmt.Printf("Enter full path to weather broker %s: \n")
	fmt.Scanln(&broker)
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(clientID)
	opts.SetUsername(uid)
	fmt.Printf("Enter password to weather broker %s: \n", broker)
	fmt.Scanln(&pwd)
	opts.SetPassword(pwd)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Error connecting. Closing program.")
		panic(token.Error())
	}

	log.Println("Client connected to broker")

	// Loop forever while goroutine handles subscribed messages
	for {
	}
}
