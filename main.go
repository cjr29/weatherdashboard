package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/ini.v1"
)

var (
	uid      string
	pwd      string
	broker   string
	port     int    = 1883
	clientID string = "weatherdashboard"
	topic1   string = "home/weather/sensors"
	topic2   string = "bus/weather/sensors"
	opts            = mqtt.NewClientOptions()
	datafile *os.File
	// bufdatafile *bufio.Writer
	nbytes           int
	err              error
	status           string
	incoming         WeatherDataRaw
	outgoing         WeatherData
	availableSensors = make(map[string]Sensor) // Available sensors table, no dups allowed
	activeSensors    = make(map[string]Sensor) // Active sensors table
	Console          = container.NewVBox()
	ConsoleScroller  = container.NewVScroll(Console)
	WeatherDataDisp  = container.NewVBox()
	WeatherScroller  = container.NewVScroll(WeatherDataDisp)
	statusContainer  *fyne.Container
	buttonContainer  *fyne.Container
)

/**********************************************************************************
 *	MQTT Message Handling
 **********************************************************************************/

var messageHandler1 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	err := json.Unmarshal(msg.Payload(), &incoming)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON due to %s", err)
	}
	outgoing.CopyWDRtoWD(incoming)
	outgoing.Home = "home"
	writeWeatherData(outgoing)
	DisplayData(fmt.Sprintf("station: %s, time: %s, model: %s, id: %d, channel: %s, temp: %.1f, humidity: %.1f",
		outgoing.Home, outgoing.Time, outgoing.Model, outgoing.Id, outgoing.Channel, outgoing.Temperature_F, outgoing.Humidity))
	// Add sensor to avalableSensors table(map)
	skey := outgoing.BuildSensorKey()
	if _, ok := availableSensors[skey]; !ok {
		// Sensor not in map. Add it.
		sens := outgoing.GetSensorFromData() // Create Sensor record
		availableSensors[skey] = sens        // Add it to the available sensors
		// log.Println(sens.FormatSensor(1))    // Log comma-separated line
	}

}

var messageHandler2 mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	err := json.Unmarshal(msg.Payload(), &incoming)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON due to %s", err)
	}
	outgoing.CopyWDRtoWD(incoming)
	outgoing.Home = "bus"
	writeWeatherData(outgoing)
	DisplayData(fmt.Sprintf("station: %s, time: %s, model: %s, id: %d, channel: %s, temp: %.1f, humidity: %.1f",
		outgoing.Home, outgoing.Time, outgoing.Model, outgoing.Id, outgoing.Channel, outgoing.Temperature_F, outgoing.Humidity))
	// Add sensor to avalableSensors table(map)
	skey := outgoing.BuildSensorKey()
	if _, ok := availableSensors[skey]; !ok {
		// Sensor not in map. Add it.
		sens := outgoing.GetSensorFromData() // Create Sensor record
		availableSensors[skey] = sens        // Add it to the available sensors
		// log.Println(sens.FormatSensor(1))    // Log comma-separated line
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
	go sub(client)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

func sub(client mqtt.Client) {
	log.Printf("Subscribing to topic ==> %s\n", topic1)
	// SetStatus(fmt.Sprintf("Subscribing to topic ==> %s", topic1))
	client.Subscribe(topic1, 0, messageHandler1)
	log.Printf("Subscribed to topic %s\n", topic1)
	// SetStatus(fmt.Sprintf("Subscribed to topic %s", topic1))
	log.Printf("Subscribing to topic ==> %s\n", topic2)
	// SetStatus(fmt.Sprintf("Subscribing to topic ==> %s", topic2))
	client.Subscribe(topic2, 0, messageHandler2)
	log.Printf("Subscribed to topic %s\n", topic2)
	// SetStatus(fmt.Sprintf("Subscribed to topic %s", topic2))
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

// checkSensor - Check if sensor is in active sensor table
func checkSensor(key string, m map[string]Sensor) bool {
	if _, ok := m[key]; ok {
		return true
	}
	return false
}

func writeWeatherData(wd WeatherData) {
	nbytes, err = datafile.WriteString(fmt.Sprintf("station: %s, time: %s, model: %s, id: %d, channel: %s, temp: %.1f, humidity: %.1f\n",
		wd.Home, wd.Time, wd.Model, wd.Id, wd.Channel, wd.Temperature_F, wd.Humidity))
	check(err)
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

/**********************************************************************************
 *	Program Control
 **********************************************************************************/

// Exit program successfully if Ctrl-C pressed
func handleInterrupt(c chan os.Signal) {
	for {
		<-c
		log.Println("User halted program. Normal exit.")
		log.Println("Size of available sensors map = ", len(availableSensors))
		log.Println(buildSensorList(availableSensors))
		os.Exit(0)
	}
}

func config() {
	inidata, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Unable to read configuration file: %v", err)
		SetStatus(fmt.Sprintf("Unable to read configuration file: %v", err))
		os.Exit(1)
	}
	section := inidata.Section("broker")

	broker = section.Key("host").String()
	uid = section.Key("username").String()
	pwd = section.Key("password").String()
}

func main() {
	// Read configuration file
	config()
	datafile, err = os.Create("./WeatherData.txt")
	if err != nil {
		log.Fatal("Unable to create/open output file.\n", err)
		SetStatus(fmt.Sprintf("Unable to create/open output file.\n", err))
		panic(err.Error)
	}
	defer datafile.Close()
	datafile.Sync()
	// bufdatafile = bufio.NewWriter(datafile)
	// defer bufdatafile.Flush()

	// Exit program on Ctrl-C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go handleInterrupt(c)
	os.Setenv("FYNE_THEME", "light")

	// fmt.Printf("Enter full path to weather broker %s: \n")
	// fmt.Scanln(&broker)
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(clientID)
	opts.SetUsername(uid)
	// fmt.Printf("Enter password to weather broker %s: \n", broker)
	// fmt.Scanln(&pwd)
	opts.SetPassword(pwd)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Error connecting. Closing program.")
		panic(token.Error())
	}

	log.Println("Client connected to broker")
	// SetStatus(fmt.Sprintln("Client connected to broker"))

	// Set up Fyne window before trying to write to Status line!!!
	a := app.NewWithID("github.com/cjr29/weatherdashboard")
	w := a.NewWindow("Weather Dashboard")
	w.Resize(fyne.NewSize(640, 460))

	// Menus
	newSensorItem := fyne.NewMenuItem("New", func() { log.Println("New Sensor Menu Item") })
	addSensorItem := fyne.NewMenuItem("Add Sensor", func() { log.Println("Add Sensor Menu Item") })
	removeSensorItem := fyne.NewMenuItem("Remove Sensor", func() { log.Println("Remove Sensor Menu Item") })
	availableSensorsItem := fyne.NewMenuItem("Available Sensors", func() { log.Println("Available Sensors Menu Item") })
	sensorMenu := fyne.NewMenu("Sensors", newSensorItem, addSensorItem, removeSensorItem, availableSensorsItem)

	internetSpeedItem := fyne.NewMenuItem("Internet Speed", func() { log.Println("Internet Speed by Station") })
	stationStatusItem := fyne.NewMenuItem("Station Status", func() { log.Println("Station Status") })
	statusMenu := fyne.NewMenu("Status", internetSpeedItem, stationStatusItem)

	menu := fyne.NewMainMenu(sensorMenu, statusMenu)

	w.SetMainMenu(menu)
	menu.Refresh()

	//*************************************************
	// Buttons & Containers
	exitButton := widget.NewButton("Exit", func() {
		SetStatus("Exiting dashboard")
		log.Println("User halted program. Normal exit.")
		os.Exit(0)
	})

	displaySensors := widget.NewButton("Sensors", func() {
		SetStatus("*****************************")
		SetStatus(fmt.Sprintf("Number of Available Sensors = %d", len(availableSensors)))
		for s := range availableSensors {
			sens := availableSensors[s]
			line := sens.FormatSensor(1)
			SetStatus(line)
		}
		SetStatus("*****************************")
	})

	dataDisplay := widget.NewButton("Data", func() {
		dataWindow := a.NewWindow("Weather Data From Sensors")
		dataWindow.SetMainMenu(menu)
		dataWindow.SetContent(WeatherScroller)
		dataWindow.Show()
	})

	ConsoleScroller.SetMinSize(fyne.NewSize(640, 400))
	WeatherScroller.SetMinSize(fyne.NewSize(700, 500))

	buttonContainer = container.NewHBox(
		displaySensors,
		dataDisplay,
		exitButton,
	)

	statusContainer = container.NewVBox(
		ConsoleScroller,
	)

	mainContainer := container.NewVBox(
		buttonContainer,
		statusContainer,
	)

	w.SetContent(mainContainer)

	UpdateAll()

	w.ShowAndRun()
}

func SetStatus(s string) {
	status = s
	ConsoleWrite(status)
}

func ConsoleWrite(text string) {
	Console.Add(&canvas.Text{
		Text:      text,
		Color:     color.Black,
		TextSize:  12,
		TextStyle: fyne.TextStyle{Monospace: true},
	})

	if len(Console.Objects) > 100 {
		Console.Remove(Console.Objects[0])
	}
	delta := (Console.Size().Height - ConsoleScroller.Size().Height) - ConsoleScroller.Offset.Y

	if delta < 100 {
		ConsoleScroller.ScrollToBottom()
	}
	Console.Refresh()
}

func DisplayData(text string) {
	WeatherDataDisp.Add(&canvas.Text{
		Text:      text,
		Color:     color.Black,
		TextSize:  12,
		TextStyle: fyne.TextStyle{Monospace: true},
	})

	if len(WeatherDataDisp.Objects) > 100 {
		WeatherDataDisp.Remove(WeatherDataDisp.Objects[0])
	}
	delta := (WeatherDataDisp.Size().Height - WeatherScroller.Size().Height) - WeatherScroller.Offset.Y

	if delta < 100 {
		WeatherScroller.ScrollToBottom()
	}
	WeatherDataDisp.Refresh()
}

func UpdateAll() {
	statusContainer.Refresh()
}
