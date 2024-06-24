package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	//nbytes              int
	err                 error
	status              string
	incoming            WeatherDataRaw
	outgoing            WeatherData
	visibleSensors      = make(map[string]Sensor) // Visible sensors table, no dups allowed
	activeSensors       = make(map[string]Sensor) // Active sensors table
	Console             = container.NewVBox()
	ConsoleScroller     = container.NewVScroll(Console)
	WeatherDataDisp     = container.NewVBox()
	WeatherScroller     = container.NewVScroll(WeatherDataDisp)
	SensorDisplay       = container.NewVBox()
	SensorScroller      = container.NewVScroll(SensorDisplay)
	EditSensorContainer *fyne.Container
	statusContainer     *fyne.Container
	buttonContainer     *fyne.Container
	dataWindow          fyne.Window
	sensorWindow        fyne.Window
	editSensorWindow    fyne.Window
	swflag              bool          = false // Sensor window flag. If true, window has been initilized.
	ddflag              bool          = false // Data display flag. If true, window has been initialized.
	s_Station_widget    *widget.Entry = widget.NewEntry()
	s_Name_widget       *widget.Entry = widget.NewEntry()
	s_Location_widget   *widget.Entry = widget.NewEntry()
	s_Model_widget      *widget.Label = widget.NewLabel("")
	s_Id_widget         *widget.Label = widget.NewLabel("")
	s_Channel_widget    *widget.Label = widget.NewLabel("")
	s_LastEdit_widget   *widget.Label = widget.NewLabel("")
	selectedValue       string        = ""
)

/**********************************************************************************
 *	Program Control
 **********************************************************************************/

func main() {
	//**********************************
	// Set up Fyne window before trying to write to Status line!!!
	//**********************************
	a := app.NewWithID("github.com/cjr29/weatherdashboard")
	w := a.NewWindow("Weather Dashboard")
	w.Resize(fyne.NewSize(640, 460))
	w.SetMaster()
	os.Setenv("FYNE_THEME", "light")

	//**********************************
	// Menus
	newSensorItem := fyne.NewMenuItem("New", func() { log.Println("New Sensor Menu Item") })
	addSensorItem := fyne.NewMenuItem("Add Sensor", func() { log.Println("Add Sensor Menu Item") })
	removeSensorItem := fyne.NewMenuItem("Remove Sensor", func() { log.Println("Remove Sensor Menu Item") })
	visibleSensorsItem := fyne.NewMenuItem("Available Sensors", func() { log.Println("Available Sensors Menu Item") })
	sensorMenu := fyne.NewMenu("Sensors", newSensorItem, addSensorItem, removeSensorItem, visibleSensorsItem)

	internetSpeedItem := fyne.NewMenuItem("Internet Speed", func() { log.Println("Internet Speed by Station") })
	stationStatusItem := fyne.NewMenuItem("Station Status", func() { log.Println("Station Status") })
	statusMenu := fyne.NewMenu("Status", internetSpeedItem, stationStatusItem)

	menu := fyne.NewMainMenu(sensorMenu, statusMenu)

	w.SetMainMenu(menu)
	menu.Refresh()

	//*************************************************
	// Buttons & Containers

	exitButton := widget.NewButton("Exit", func() {
		for _, d := range dataFiles {
			d.file.Sync()
			d.file.Close()
		}
		os.Exit(0)
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
			editSensorWindow.Close()
		}),
	)

	editSensorButton := widget.NewButton("Edit Sensor", func() {
		selectSensorWindow := a.NewWindow("Select a Sensor to edit")
		vlist := buildSensorList(activeSensors) // Get list of visible sensors
		pickSensor := widget.NewSelect(vlist, func(value string) {
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
			s_LastEdit_widget.SetText(s.LastEdit)
			selectSensorWindow.Close()
			// Pop up a window to let user edit record, then save the record
			editSensorWindow = a.NewWindow("Edit active sensor properties.")
			editSensorWindow.SetContent(EditSensorContainer)
			editSensorWindow.SetOnClosed(func() {
				// Save updated record back to activeSensors
				s.Station = s_Station_widget.Text
				s.Name = s_Name_widget.Text
				s.Location = s_Location_widget.Text
				activeSensors[key] = s
				editSensorWindow.Close()
			})
			editSensorWindow.Show()
		})
		pickSensorContainer := container.NewVBox(
			widget.NewLabel("Pick a sensor to edit. Use the toggle to scroll through active sensors to find sensor to edit. ***************************"+
				"******************************************"),
			pickSensor,
		)
		selectSensorWindow.SetContent(pickSensorContainer)
		selectSensorWindow.Show()
	})

	var selections []string
	addActiveSensorsButton := widget.NewButton("Add Sensor(s)", func() {
		addSensorWindow := a.NewWindow("Select Sensors to add to active list")
		vlist := buildSensorList(visibleSensors) // Get list of visible sensors
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
					if checkSensor(key, visibleSensors) {
						// Add sensors to activeSensors map - TBD
						activeSensors[key] = visibleSensors[key]
					}
				}
				addSensorWindow.Close()
			}),
		)
		addSensorWindow.SetContent(addSensorContainer)
		addSensorWindow.Show()
	})

	ConsoleScroller.SetMinSize(fyne.NewSize(640, 400))
	WeatherScroller.SetMinSize(fyne.NewSize(700, 500))
	SensorScroller.SetMinSize(fyne.NewSize(550, 500))

	buttonContainer = container.NewHBox(
		displaySensors,
		dataDisplay,
		addActiveSensorsButton,
		editSensorButton,
		exitButton,
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
	config()

	//**********************************
	// Set configuration for MQTT, read from config.ini file in local directory
	//**********************************
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(clientID)
	opts.SetUsername(uid)
	opts.SetPassword(pwd)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	//**********************************
	// Initialize MQTT client
	//**********************************
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		SetStatus("Error connecting with broker. Closing program.")
		panic(token.Error())
	}
	SetStatus(fmt.Sprintf("Client connected to broker %s", broker))

	//**********************************
	// Turn over control to the GUI
	//**********************************

	w.ShowAndRun()
}

func check(e error) {
	if e != nil {
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
