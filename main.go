package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
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
const widgetPadding float32 = 5 // separation between widgets
const numColumns = 5

var (
	a               fyne.App
	Client          mqtt.Client
	status          string
	Console         = container.NewVBox()
	ConsoleScroller = container.NewVScroll(Console)
	WeatherDataDisp = container.NewVBox()
	WeatherScroller = container.NewVScroll(WeatherDataDisp)
	SensorDisplay   = container.NewVBox()
	SensorScroller  = container.NewVScroll(SensorDisplay)
	SensorDisplay2  = container.NewVBox()
	SensorScroller2 = container.NewVScroll(SensorDisplay2)
	TopicDisplay    = container.NewVBox()
	TopicScroller   = container.NewVScroll(TopicDisplay)

	EditSensorContainer *fyne.Container
	statusContainer     *fyne.Container
	//buttonContainer     *fyne.Container
	dashboardContainer *fyne.Container
	dataWindow         fyne.Window
	sensorWindow       fyne.Window
	sensorWindow2      fyne.Window
	selectSensorWindow fyne.Window
	editSensorWindow   fyne.Window
	topicWindow        fyne.Window
	dashboardWindow    fyne.Window
	dashFlag           bool          = false // Dashboard window flag. If true, window has been initialized.
	swflag             bool          = false // Sensor window flag. If true, window has been initilized.
	swflag2            bool          = false // Sensor window2 flag. If true, window has been initilized.
	ddflag             bool          = false // Data display flag. If true, window has been initialized.
	tflag              bool          = false // Topic display flag. If true, window has been initialized
	hideflag           bool          = false // Used by hideWidgetHandler DO NOT DELETE!
	cancelEdit         bool          = false // Set to true to cancel current edit
	s_Station_widget   *widget.Entry = widget.NewEntry()
	s_Name_widget      *widget.Entry = widget.NewEntry()
	s_Location_widget  *widget.Entry = widget.NewEntry()
	s_Model_widget     *widget.Label = widget.NewLabel("")
	s_Id_widget        *widget.Label = widget.NewLabel("")
	s_Channel_widget   *widget.Label = widget.NewLabel("")
	s_LastEdit_widget  *widget.Label = widget.NewLabel("")
	s_Hide_widget      *widget.Check = widget.NewCheck("Check to hide sensor on weather dashboard", hideWidgetHandler)
	selectedValue      string        = ""
	logdata_flg        bool          = false
	selections         []string
	sav_Station        string // to restore in case of edit cancel
	sav_Name           string
	sav_Location       string
	sav_Hide           bool
)

//*************************************************
// Menu action handlers
//*************************************************

// HANDLER FOR EDIT ACTIVE SENSOR
var selectionHandler = func(value string) {
	selectedValue = value
	// Get key for selected sensor - Key: field of value
	key := strings.Split(selectedValue, "Key: ")[1] // Found at end of the displayed string after "Key: ..."
	// Load widgets using selected sensor
	s := activeSensors[key]
	sav_Station = s.Station
	sav_Name = s.Name
	sav_Location = s.Location
	sav_Hide = s.Hide
	s_Station_widget.SetText(s.Station)
	s_Station_widget.SetPlaceHolder("Home")
	s_Name_widget.SetText(s.Name)
	s_Name_widget.SetPlaceHolder("Name")
	s_Location_widget.SetText(s.Location)
	s_Location_widget.SetPlaceHolder("Location")
	s_Hide_widget.SetChecked(s.Hide)
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
			s_Hide_widget.SetChecked(sav_Hide)
			editSensorWindow.Close()
			return
		}
		// Save updated record back to activeSensors
		// First be sure that there is a widget record. User may have cleared the Hide flag, expecting a new sensor to display
		if !checkWeatherWidget(key) {
			// Sensor doesn't exist, regenerate the weather widgets
			generateWeatherWidgets()
			dashboardContainer.Refresh()
		}
		SetStatus(fmt.Sprintf("Saving updated sensor record: %s", s_Name_widget.Text+":"+key))
		s.Station = s_Station_widget.Text
		s.Name = s_Name_widget.Text
		s.Location = s_Location_widget.Text
		s.Hide = s_Hide_widget.Checked
		s.LastEdit = st
		activeSensors[key] = s
		reloadDashboard()
		editSensorWindow.Close()
	})
	editSensorWindow.Show()
}

var hideWidgetHandler = func(value bool) {
	hideflag = value
}

// LIST ACTIVE SENSORS
var listSensorsHandler = func() {
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
}

// LIST AVAILABLE SENSORS
var listAvailableSensorsHandler = func() {
	// Get displayable list of sensors
	if !swflag2 {
		sensorWindow2 = a.NewWindow("Available Sensors")
		DisplaySensors2(availableSensors)
		sensorWindow2.SetContent(SensorScroller2)
		sensorWindow2.SetOnClosed(func() {
			swflag2 = false
		})
		swflag2 = true
		sensorWindow2.Show()
	} else {
		DisplaySensors2(availableSensors)
		sensorWindow2.Show()
		swflag2 = true
	}
}

// ADD ACTIVE SENSOR
var addSensorHandler = func() {
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
				if checkSensor(key, availableSensors) && !availableSensors[key].Hide {
					// Add sensors to activeSensors map - TBD
					s := *availableSensors[key]
					activeSensors[key] = &s
					reloadDashboard()
					SetStatus(fmt.Sprintf("Added sensor to active sensors: %s", key))
				}
			}
			addSensorWindow.Close()
		}),
	)
	addSensorWindow.SetContent(addSensorContainer)
	addSensorWindow.Show()
}

// EDIT ACTIVE SENSOR
var editSensorHandler = func() {
	selectSensorWindow = a.NewWindow("Select a Sensor to edit from the Active Sensors list")
	vlist := buildSensorList(activeSensors) // Get list of visible sensors
	pickSensor := widget.NewSelect(vlist, selectionHandler)
	header := widget.NewLabel("Pick a sensor to edit. Use the toggle to scroll through active sensors to find sensor to edit. ***************************" +
		"******************************************")
	pickSensorContainer := container.NewVBox(header, pickSensor)
	selectSensorWindow.SetContent(pickSensorContainer)
	selectSensorWindow.Show()
}

// REMOVE ACTIVE SENSOR
var removeSensorHandler = func() {
	removeSensorWindow := a.NewWindow("Select Sensors to remove from active list")
	vlist := buildSensorList(activeSensors) // Get list of active sensors
	pickSensors := widget.NewCheckGroup(vlist, func(choices []string) {
		selections = choices
	})
	removeSensorScroller := container.NewVScroll(
		pickSensors,
	)
	removeSensorContainer := container.NewVBox(
		widget.NewLabel("Select all sensors to be removed from the active sensors list"),
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
					// if checkWeatherWidget(key) {
					// 	delete(weatherWidgets, key)
					// 	generateWeatherWidgets()
					// }
					reloadDashboard()
				}
			}
			removeSensorWindow.Close()
		}),
		removeSensorScroller,
	)
	removeSensorWindow.SetContent(removeSensorContainer)
	removeSensorWindow.Show()
}

// LIST TOPIC
var listTopicsHandler = func() {
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
}

// ADD TOPIC
var addTopicHandler = func() {
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
}

// REMOVE TOPIC
var removeTopicHandler = func() {
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
}

var exitHandler = func() {
	// Close data files
	for _, d := range dataFiles {
		d.file.Sync()
		d.file.Close()
	}

	// Output current configuration for later reload
	writeConfig()

	// Now, exit program
	os.Exit(0)
}

var scrollDataHandler = func() {
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
}

// Opens a new window that displays a dashboard of climate data from selected sensors
var dashboardHandler = func() {
	// generateWeatherWidgets()
	// for _, ww := range weatherWidgets {
	// 	fmt.Printf("dashboardHandler:WeatherWidget: Key = %s, Name = %s\n", ww.sensorKey, ww.sensorName)

	// }
	// Loop here to add all sensors to display in dashboard
	// keys := sortWeatherWidgets() // Display in columns sorted by Station and Name
	// for _, k := range keys {
	// 	weatherWidgets[k].Refresh()
	// 	dashboardContainer.Add(weatherWidgets[k])
	// 	// Start background widget update routine
	// 	go weatherWidgets[k].goHandler(k)
	// }
	// dashboardWindow.SetContent(dashboardContainer)
	// numRows := float32(math.Round(float64(len(weatherWidgets)) / float64(numColumns)))
	// dashboardWindow.Resize(fyne.NewSize((widgetSizeX+widgetPadding)*float32(numColumns), (widgetSizeY+widgetPadding)*numRows))

	reloadDashboard()
	dashboardWindow.Show()
}

var dataLoggingOnHandler = func() {
	logdata_flg = true
	SetStatus("Data logging turned on")
}

var dataLoggingOffHandler = func() {
	logdata_flg = false
	SetStatus("Data logging turned off")
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
	// Menus
	listActiveSensorsItem := fyne.NewMenuItem("List Active Sensors", listSensorsHandler)
	listAvailableSensorsItem := fyne.NewMenuItem("List Available Sensors", listAvailableSensorsHandler)
	addActiveSensorItem := fyne.NewMenuItem("Add Active Sensor", addSensorHandler)
	removeActiveSensorItem := fyne.NewMenuItem("Remove Active Sensor", removeSensorHandler)
	editActiveSensorItem := fyne.NewMenuItem("Edit Active Sensor", editSensorHandler)
	sensorMenu := fyne.NewMenu("Sensors",
		listActiveSensorsItem,
		listAvailableSensorsItem,
		addActiveSensorItem,
		editActiveSensorItem,
		removeActiveSensorItem)

	listTopicsItem := fyne.NewMenuItem("List", listTopicsHandler)
	addTopicItem := fyne.NewMenuItem("New", addTopicHandler)
	removeTopicItem := fyne.NewMenuItem("Remove", removeTopicHandler)
	topicMenu := fyne.NewMenu("Topics", listTopicsItem, addTopicItem, removeTopicItem)

	dataDisplayItem := fyne.NewMenuItem("Weather Data Scroller", scrollDataHandler)
	dashboardItem := fyne.NewMenuItem("Dashboard Widgets", dashboardHandler)
	dataMenuSeparator := fyne.NewMenuItemSeparator()
	toggleDataLoggingOnItem := fyne.NewMenuItem("Data Logging On", dataLoggingOnHandler)
	toggleDataLoggingOffItem := fyne.NewMenuItem("DataLogging Off", dataLoggingOffHandler)
	weatherMenu := fyne.NewMenu("Display Data",
		dataDisplayItem,
		dashboardItem,
		dataMenuSeparator,
		toggleDataLoggingOnItem,
		toggleDataLoggingOffItem,
	)

	menu := fyne.NewMainMenu(weatherMenu, sensorMenu, topicMenu)

	w.SetMainMenu(menu)
	menu.Refresh()

	//*************************************************
	// Containers
	//*************************************************

	// editSensor
	//
	EditSensorContainer = container.NewVBox(
		widget.NewLabel("Update Home, Name, Location, and Visibility, then press Submit to save."),
		s_Station_widget,
		s_Name_widget,
		s_Location_widget,
		s_Hide_widget,
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

	ConsoleScroller.SetMinSize(fyne.NewSize(640, 400))
	WeatherScroller.SetMinSize(fyne.NewSize(700, 500))
	SensorScroller.SetMinSize(fyne.NewSize(550, 500))
	SensorScroller2.SetMinSize(fyne.NewSize(550, 500))
	TopicScroller.SetMinSize(fyne.NewSize(300, 200))

	statusContainer = container.NewVBox(
		ConsoleScroller,
	)

	mainContainer := container.NewVBox(
		//buttonContainer,
		widget.NewLabel("Dashboard Status Scrolling Window"),
		statusContainer,
	)

	// Put main container in the primary window
	w.SetContent(mainContainer)

	//**********************************
	// Read configuration file
	//**********************************
	readConfig()

	// Build the widgets used by the dashboard before activating the GUI
	generateWeatherWidgets()

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
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	SetStatus(fmt.Sprintf("%s : Client connected to broker %s", st, brokers[0].Path+":"+strconv.Itoa(brokers[0].Port)))

	//**********************************
	// Turn over control to the GUI
	//**********************************
	w.SetOnClosed(exitHandler)

	w.ShowAndRun()

	//*************************************************
	// Program blocked until GUI closes
	//*************************************************
}

//*************************************************
// Support functions
//*************************************************

func check(e error) {
	if e != nil {
		log.Println(e)
		panic(e)
	}
}

// SetStatus - publishes message on scrolling GUI status console
func SetStatus(s string) {
	status = s
	ConsoleWrite(status)
}

// generateWeatherWidgets - reads active sensor table and creates widgets for each sensor
//
//	stores pointers to the widgets in the weatherWidgets array
func generateWeatherWidgets() {
	keys := sortActiveSensors()
	for _, s := range keys {
		// Only generate widget if sensor is not hidden
		if !activeSensors[s].Hide {
			newWidget := new(weatherWidget)
			sens := activeSensors[s]
			newWidget.Init(sens)
			weatherWidgets[s] = newWidget
		}
	}
}

// checkWeatherWidget - Check if widget is available
func checkWeatherWidget(key string) bool {
	if _, ok := weatherWidgets[key]; ok {
		return true
	}
	return false
}

// removeAllWeatherWidgets - Loops through map and removes all widgets
//
//	Allows the map to remain available to other processes
func removeAllWeatherWidgets() {
	for w := range weatherWidgets {
		delete(weatherWidgets, w)
	}
}

// reloadDashboard() - Regenerate and reload the dashboard container
func reloadDashboard() {

	// Empty dashboard container
	// Create new window if one does't already exist
	if !dashFlag {
		dashboardWindow = a.NewWindow("Weather Dashboard")
		dashFlag = true
	}
	dashboardWindow.SetOnClosed(func() {
		dashFlag = false
	})
	dashboardContainer = container.NewGridWithColumns(numColumns)

	// Delete existing weatherWidgets
	removeAllWeatherWidgets()

	// Regenerate weatherWidgets
	generateWeatherWidgets()

	// Load the container with the new widgets
	keys := sortWeatherWidgets() // Display in columns sorted by Station and Name
	for _, k := range keys {
		weatherWidgets[k].Refresh()
		dashboardContainer.Add(weatherWidgets[k])
		// Start background widget update routine
		go weatherWidgets[k].goHandler(k)
	}
	dashboardWindow.SetContent(dashboardContainer)
	numRows := float32(math.Round(float64(len(weatherWidgets)) / float64(numColumns)))
	dashboardWindow.Resize(fyne.NewSize((widgetSizeX+widgetPadding)*float32(numColumns), (widgetSizeY+widgetPadding)*numRows))

	// Show the reloaded container in window
	dashboardContainer.Refresh()
	dashboardWindow.Show()
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
func DisplaySensors(m map[string]*Sensor) {
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

// DisplayAvailableSensors - First, erase previously displayed list
func DisplaySensors2(m map[string]*Sensor) {
	SensorDisplay2.RemoveAll()
	header := fmt.Sprintf("Number of Sensors = %d", len(m))
	SensorDisplay2.Add(&canvas.Text{
		Text:      header,
		Color:     color.Black,
		TextSize:  12,
		TextStyle: fyne.TextStyle{Monospace: true},
	})
	for s := range m {
		sens := m[s]
		text := sens.FormatSensor(1)
		SensorDisplay2.Add(&canvas.Text{
			Text:      text,
			Color:     color.Black,
			TextSize:  12,
			TextStyle: fyne.TextStyle{Monospace: true},
		})
	}
	if len(SensorDisplay2.Objects) > 100 {
		SensorDisplay2.Remove(SensorDisplay2.Objects[0])
	}
	delta := (SensorDisplay2.Size().Height - SensorScroller2.Size().Height) - SensorScroller2.Offset.Y

	if delta < 100 {
		SensorScroller2.ScrollToBottom()
	}
	SensorDisplay2.Refresh()
}

// UpdateAll - updates selected containers
func UpdateAll() {
	if dashboardContainer != nil {
		dashboardContainer.Refresh()
	}
}
