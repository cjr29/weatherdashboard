package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	// YYYY-MM-DD: 2022-03-23
	YYYYMMDD = "2006-01-02"
	// 24h hh:mm:ss: 14:23:20
	HHMMSS24h = "15:04:05"
	// 12h hh:mm:ss: 2:23:20 PM
)

var (
	a                   fyne.App
	Client              mqtt.Client
	status              string
	incoming            WeatherDataRaw
	outgoing            WeatherData
	Console             = container.NewVBox()
	ConsoleScroller     = container.NewVScroll(Console)
	WeatherDataDisp     = container.NewVBox()
	WeatherScroller     = container.NewVScroll(WeatherDataDisp)
	SensorDisplay       = container.NewVBox()
	SensorScroller      = container.NewVScroll(SensorDisplay)
	TopicDisplay        = container.NewVBox()
	TopicScroller       = container.NewVScroll(TopicDisplay)
	EditSensorContainer *fyne.Container
	statusContainer     *fyne.Container
	buttonContainer     *fyne.Container
	dataWindow          fyne.Window
	sensorWindow        fyne.Window
	selectSensorWindow  fyne.Window
	editSensorWindow    fyne.Window
	topicWindow         fyne.Window
	swflag              bool          = false // Sensor window flag. If true, window has been initilized.
	ddflag              bool          = false // Data display flag. If true, window has been initialized.
	tflag               bool          = false // Topic display flag. If true, window has been initialized
	cancelEdit          bool          = false // Set to true to cancel current edit
	s_Station_widget    *widget.Entry = widget.NewEntry()
	s_Name_widget       *widget.Entry = widget.NewEntry()
	s_Location_widget   *widget.Entry = widget.NewEntry()
	s_Model_widget      *widget.Label = widget.NewLabel("")
	s_Id_widget         *widget.Label = widget.NewLabel("")
	s_Channel_widget    *widget.Label = widget.NewLabel("")
	s_LastEdit_widget   *widget.Label = widget.NewLabel("")
	selectedValue       string        = ""
	logdata_flg         bool          = false
	selections          []string
	sav_Station         string // to restore in case of edit cancel
	sav_Name            string
	sav_Location        string
)

var selectionHandler = func(value string) {
	selectedValue = value
	// Get key for selected sensor - Key: field of value
	key := strings.Split(selectedValue, "Key: ")[1] // Found at end of the displayed string after "Key: ..."
	// Load widgets using selected sensor
	s := activeSensors[key]
	sav_Station = s.Station
	sav_Name = s.Name
	sav_Location = s.Location
	s_Station_widget.SetText(s.Station)
	s_Station_widget.SetPlaceHolder("Home")
	s_Name_widget.SetText(s.Name)
	s_Name_widget.SetPlaceHolder("Name")
	s_Location_widget.SetText(s.Location)
	s_Location_widget.SetPlaceHolder("Location")
	s_Model_widget.SetText(s.Model)
	s_Id_widget.SetText(strconv.Itoa(s.Id))
	s_Channel_widget.SetText(s.Channel)
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	s_LastEdit_widget.SetText(st)
	SetStatus(fmt.Sprintf("Last edit set to %s", st))
	selectSensorWindow.Close()
	// Pop up a window to let user edit record, then save the record
	editSensorWindow = a.NewWindow("Edit active sensor properties.")
	editSensorWindow.SetContent(EditSensorContainer)
	editSensorWindow.SetOnClosed(func() {
		if cancelEdit {
			cancelEdit = false // reset flag for next time
			// Restore orginal values
			s_Station_widget.SetText(sav_Station)
			s_Name_widget.SetText(sav_Name)
			s_Location_widget.SetText(sav_Location)
			editSensorWindow.Close()
			return
		}
		// Save updated record back to activeSensors
		s.Station = s_Station_widget.Text
		s.Name = s_Name_widget.Text
		s.Location = s_Location_widget.Text
		s.LastEdit = st
		activeSensors[key] = s
		editSensorWindow.Close()
	})
	editSensorWindow.Show()
}

/**********************************************************************************
 *	Program Control
 **********************************************************************************/

func main() {
	//**********************************
	// Set up Fyne window before trying to write to Status line!!!
	//**********************************
	a = app.NewWithID("github.com/cjr29/weatherdashboard")
	w := a.NewWindow("Weather Dashboard")
	w.Resize(fyne.NewSize(640, 460))
	w.SetMaster()
	os.Setenv("FYNE_THEME", "light")

	//**********************************
	// Menus - Doesn't work in macOS
	/* 	newSensorItem := fyne.NewMenuItem("New", func() { log.Println("New Sensor Menu Item") })
	   	addSensorItem := fyne.NewMenuItem("Add Sensor", func() { log.Println("Add Sensor Menu Item") })
	   	removeSensorItem := fyne.NewMenuItem("Remove Sensor", func() { log.Println("Remove Sensor Menu Item") })
	   	availableSensorsItem := fyne.NewMenuItem("Available Sensors", func() { log.Println("Available Sensors Menu Item") })
	   	sensorMenu := fyne.NewMenu("Sensors", newSensorItem, addSensorItem, removeSensorItem, availableSensorsItem)

	   	internetSpeedItem := fyne.NewMenuItem("Internet Speed", func() { log.Println("Internet Speed by Station") })
	   	stationStatusItem := fyne.NewMenuItem("Station Status", func() { log.Println("Station Status") })
	   	statusMenu := fyne.NewMenu("Status", internetSpeedItem, stationStatusItem)

	   	menu := fyne.NewMainMenu(sensorMenu, statusMenu)

	   	w.SetMainMenu(menu)
	   	menu.Refresh() */

	//*************************************************
	// Buttons & Containers

	exitButton := widget.NewButton("Exit", func() {
		// Close data files
		for _, d := range dataFiles {
			d.file.Sync()
			d.file.Close()
		}

		// Output current configuration for later reload
		writeConfig()

		// Now, exit program
		os.Exit(0)
	})

	// Display a box to check to turn data logging on and off
	toggleDataLogging := widget.NewCheck("Log Data", func(b bool) {
		if b {
			logdata_flg = true
			SetStatus("Data logging turned on")
		} else {
			logdata_flg = false
			SetStatus("Data logging turned off")
		}
	})

	displaySensors := widget.NewButton("Active Sensors", func() {
		// Get displayable list of sensors
		if !swflag {
			sensorWindow = a.NewWindow("Active Sensors")
			DisplaySensors(activeSensors)
			sensorWindow.SetContent(SensorScroller)
			sensorWindow.SetOnClosed(func() {
				swflag = false
			})
			swflag = true
			sensorWindow.Show()
		} else {
			DisplaySensors(activeSensors)
			sensorWindow.Show()
			swflag = true
		}
	})

	dataDisplay := widget.NewButton("Data", func() {
		if !ddflag {
			dataWindow = a.NewWindow("Weather Data From Sensors")
			dataWindow.SetContent(WeatherScroller)
			dataWindow.SetOnClosed(func() {
				ddflag = false
			})
			ddflag = true
			dataWindow.Show()
		} else {
			dataWindow.Show()
			ddflag = true
		}
	})

	// editSensor
	//
	EditSensorContainer = container.NewVBox(
		widget.NewLabel("Update Home, Name, and Location, then press Submit to save."),
		s_Station_widget,
		s_Name_widget,
		s_Location_widget,
		s_Model_widget,
		s_Id_widget,
		s_Channel_widget,
		s_LastEdit_widget,
		widget.NewButton("Submit", func() {
			cancelEdit = false
			editSensorWindow.Close()
		}),
		widget.NewButton("Cancel", func() {
			cancelEdit = true
			editSensorWindow.Close()
		}),
	)

	editSensorButton := widget.NewButton("Edit Sensor", func() {
		selectSensorWindow = a.NewWindow("Select a Sensor to edit from the Active Sensors list")
		vlist := buildSensorList(activeSensors) // Get list of visible sensors
		pickSensor := widget.NewSelect(vlist, selectionHandler)
		/* 		pickSensor := widget.NewSelect(vlist, func(value string) {
			selectedValue = value
			// Get key for selected sensor - Key: field of value
			key := strings.Split(selectedValue, "Key: ")[1] // Found at end of the displayed string after "Key: ..."
			// Load widgets using selected sensor
			s := activeSensors[key]
			s_Station_widget.SetText(s.Station)
			s_Station_widget.SetPlaceHolder("Home")
			s_Name_widget.SetText(s.Name)
			s_Name_widget.SetPlaceHolder("Name")
			s_Location_widget.SetText(s.Location)
			s_Location_widget.SetPlaceHolder("Location")
			s_Model_widget.SetText(s.Model)
			s_Id_widget.SetText(strconv.Itoa(s.Id))
			s_Channel_widget.SetText(s.Channel)
			t := time.Now().Local()
			st := t.Format(YYYYMMDD + " " + HHMMSS24h)
			s_LastEdit_widget.SetText(st)
			SetStatus(fmt.Sprintf("Last edit set to %s", st))
			selectSensorWindow.Close()
			// Pop up a window to let user edit record, then save the record
			editSensorWindow = a.NewWindow("Edit active sensor properties.")
			editSensorWindow.SetContent(EditSensorContainer)
			editSensorWindow.SetOnClosed(func() {
				// Save updated record back to activeSensors
				s.Station = s_Station_widget.Text
				s.Name = s_Name_widget.Text
				s.Location = s_Location_widget.Text
				s.LastEdit = st
				activeSensors[key] = s
				editSensorWindow.Close()
			})
			editSensorWindow.Show()
		}) */
		// pickSensorContainer := container.NewVBox(
		// 	widget.NewLabel("Pick a sensor to edit. Use the toggle to scroll through active sensors to find sensor to edit. ***************************"+
		// 		"******************************************"),
		// 	pickSensor,
		// )

		header := widget.NewLabel("Pick a sensor to edit. Use the toggle to scroll through active sensors to find sensor to edit. ***************************" +
			"******************************************")
		pickSensorContainer := container.NewVBox(header, pickSensor)
		selectSensorWindow.SetContent(pickSensorContainer)
		selectSensorWindow.Show()
	})

	addActiveSensorsButton := widget.NewButton("Add Sensor(s)", func() {
		addSensorWindow := a.NewWindow("Select Sensors to add to active list")
		vlist := buildSensorList(availableSensors) // Get list of visible sensors
		pickSensors := widget.NewCheckGroup(vlist, func(choices []string) {
			selections = choices
		})
		addSensorContainer := container.NewVBox(
			widget.NewLabel("Select all sensors to be added to the active sensors list"),
			pickSensors,
			widget.NewButton("Submit", func() {
				// SetStatus("Submit pressed to add selected sensors to active list.")
				// Process selected sensors and add to the active sensors map
				for i := 0; i < len(selections); i++ {
					// Get key for selected sensor - Key: field of value
					key := strings.Split(selections[i], "Key: ")[1] // Found at end of the displayed string after "Key: ..."
					// Load widgets using selected sensor
					if checkSensor(key, availableSensors) {
						// Add sensors to activeSensors map - TBD
						activeSensors[key] = availableSensors[key]
						SetStatus(fmt.Sprintf("Added sensor to active sensors: %s", key))
					}
				}
				addSensorWindow.Close()
			}),
		)
		addSensorWindow.SetContent(addSensorContainer)
		addSensorWindow.Show()
	})

	removeActiveSensorsButton := widget.NewButton("Remove Sensor(s)", func() {
		removeSensorWindow := a.NewWindow("Select Sensors to remove from active list")
		vlist := buildSensorList(activeSensors) // Get list of active sensors
		pickSensors := widget.NewCheckGroup(vlist, func(choices []string) {
			selections = choices
		})
		removeSensorContainer := container.NewVBox(
			widget.NewLabel("Select all sensors to be removed from the active sensors list"),
			pickSensors,
			widget.NewButton("Submit", func() {
				// Process selected sensors and add to the active sensors map
				for i := 0; i < len(selections); i++ {
					// Get key for selected sensor - Key: field of value
					key := strings.Split(selections[i], "Key: ")[1] // Found at end of the displayed string after "Key: ..."
					// Delete sensor using selected key
					if checkSensor(key, activeSensors) {
						// Add sensors to activeSensors map - TBD
						delete(activeSensors, key)
						SetStatus(fmt.Sprintf("Removed sensor from active sensors: %s", key))
					}
				}
				removeSensorWindow.Close()
			}),
		)
		removeSensorWindow.SetContent(removeSensorContainer)
		removeSensorWindow.Show()

	})

	listTopicsButton := widget.NewButton("Subscribed Topics", func() {
		DisplayTopics(messages)
		if !tflag {
			topicWindow = a.NewWindow("Subscribed Topics")
			// Get displayable list of subscribed topics
			topicWindow.SetContent(TopicScroller)
			topicWindow.SetOnClosed(func() {
				tflag = false
			})
			tflag = true
			topicWindow.Show()
		} else {
			topicWindow.Show()
			tflag = true
		}
	})

	addTopicButton := widget.NewButton("Add Topic", func() {
		addTopicWindow := a.NewWindow("New Topic")
		inputT := widget.NewEntry()
		inputS := widget.NewEntry()
		inputT.SetPlaceHolder("Topic")
		inputS.SetPlaceHolder("Station")
		addTopicContainer := container.NewVBox(
			widget.NewLabel("Enter the full topic and its station name to which you want to subscribe."),
			inputT,
			inputS,
			widget.NewButton("Submit", func() {
				SetStatus(fmt.Sprintf("Added Topic: %s, Station: %s", inputT.Text, inputS.Text))
				// Add input text to messages[]
				var m Message
				m.Topic = inputT.Text
				m.Station = inputS.Text
				key := rand.Int()
				messages[key] = m
				Client.Subscribe(m.Topic, 0, messageHandler)
				SetStatus(fmt.Sprintf("Subscribed to Topic: %s", m.Topic))
				addTopicWindow.Close()
			}),
		)
		addTopicWindow.SetContent(addTopicContainer)
		addTopicWindow.Show()
	})

	delTopicButton := widget.NewButton("Delete Topic", func() {
		delTopicWindow := a.NewWindow("Delete Topic")
		var choices []string
		tlist := buildMessageList(messages)
		for _, m := range tlist {
			choices = append(choices, m.Display)
		}
		pickTopics := widget.NewCheckGroup(choices, func(c []string) {
			selections = c
		})
		delTopicContainer := container.NewVBox(
			widget.NewLabel("Select the topic you want to remove from subscribing."),
			pickTopics,
			widget.NewButton("Submit", func() {
				for i := 0; i < len(selections); i++ {
					j, err := strconv.ParseInt(strings.Split(selections[i], ":")[0], 10, 32) // Index of choice at head of string "0: ", "1: ", ...
					check(err)
					k := int(j) // k is index into the tlist array of []ChoicesIntKey where .Key is the Message key
					// Verify message is in map before rying to delete
					if checkMessage(tlist[k].Key, messages) {
						// Delete using key
						key := tlist[k].Key
						unsubscribe(Client, messages[key])
						delete(messages, key)
					}
				}
				delTopicWindow.Close()
			}),
		)
		delTopicWindow.SetContent(delTopicContainer)
		delTopicWindow.Show()
	})

	ConsoleScroller.SetMinSize(fyne.NewSize(640, 400))
	WeatherScroller.SetMinSize(fyne.NewSize(700, 500))
	SensorScroller.SetMinSize(fyne.NewSize(550, 500))
	TopicScroller.SetMinSize(fyne.NewSize(300, 200))

	buttonContainer = container.NewHBox(
		displaySensors,
		dataDisplay,
		addActiveSensorsButton,
		editSensorButton,
		removeActiveSensorsButton,
		listTopicsButton,
		addTopicButton,
		delTopicButton,
		exitButton,
		toggleDataLogging,
	)

	statusContainer = container.NewVBox(
		ConsoleScroller,
	)

	mainContainer := container.NewVBox(
		buttonContainer,
		widget.NewLabel("Dashboard Status Scrolling Window"),
		statusContainer,
	)

	w.SetContent(mainContainer)

	//**********************************
	// Read configuration file
	//**********************************
	readConfig()

	//**********************************
	// Set configuration for MQTT, read from config.ini file in local directory
	//**********************************
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", brokers[0].Path, brokers[0].Port))
	opts.SetClientID(clientID)
	opts.SetUsername(brokers[0].Uid)
	opts.SetPassword(brokers[0].Pwd)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	//**********************************
	// Initialize MQTT client
	//**********************************
	Client = mqtt.NewClient(opts)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		SetStatus("Error connecting with broker. Closing program.")
		log.Println("Error connecting with broker. Closing program.")
		panic(token.Error())
	}
	SetStatus(fmt.Sprintf("Client connected to broker %s", brokers[0].Path+":"+strconv.Itoa(brokers[0].Port)))

	//**********************************
	// Turn over control to the GUI
	//**********************************

	w.ShowAndRun()
}

func check(e error) {
	if e != nil {
		log.Println(e)
		panic(e)
	}
}

func SetStatus(s string) {
	status = s
	ConsoleWrite(status)
}

// ConsoleWrite - call this function to write a string to the scrolling console status window
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

// DisplayData - Call this function to display a weather data string in the weather display scrolling window
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

// DisplayTopics - Call this function to display a list of topics in a scrolling window
func DisplayTopics(t map[int]Message) {
	TopicDisplay.RemoveAll()
	header := fmt.Sprintf("Number of Subscribed Topics = %d", len(t))
	TopicDisplay.Add(&canvas.Text{
		Text:      header,
		Color:     color.Black,
		TextSize:  12,
		TextStyle: fyne.TextStyle{Monospace: true},
	})
	for m := range t {
		msg := t[m]
		text := msg.Topic
		TopicDisplay.Add(&canvas.Text{
			Text:      text,
			Color:     color.Black,
			TextSize:  12,
			TextStyle: fyne.TextStyle{Monospace: true},
		})
	}

	if len(TopicDisplay.Objects) > 100 {
		TopicDisplay.Remove(TopicDisplay.Objects[0])
	}
	delta := (TopicDisplay.Size().Height - TopicScroller.Size().Height) - TopicScroller.Offset.Y

	if delta < 100 {
		TopicScroller.ScrollToBottom()
	}
	TopicDisplay.Refresh()
}

// DisplaySensors - First, erase previously displayed list
func DisplaySensors(m map[string]Sensor) {
	SensorDisplay.RemoveAll()
	header := fmt.Sprintf("Number of Sensors = %d", len(m))
	SensorDisplay.Add(&canvas.Text{
		Text:      header,
		Color:     color.Black,
		TextSize:  12,
		TextStyle: fyne.TextStyle{Monospace: true},
	})
	for s := range m {
		sens := m[s]
		text := sens.FormatSensor(1)
		SensorDisplay.Add(&canvas.Text{
			Text:      text,
			Color:     color.Black,
			TextSize:  12,
			TextStyle: fyne.TextStyle{Monospace: true},
		})
	}
	if len(SensorDisplay.Objects) > 100 {
		SensorDisplay.Remove(SensorDisplay.Objects[0])
	}
	delta := (SensorDisplay.Size().Height - SensorScroller.Size().Height) - SensorScroller.Offset.Y

	if delta < 100 {
		SensorScroller.ScrollToBottom()
	}
	SensorDisplay.Refresh()
}

func UpdateAll() {
	statusContainer.Refresh()
}
