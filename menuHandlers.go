/******************************************************************
 *
 * Dashboard Menu handler functions
 *
 ******************************************************************/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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
	sav_HasHumidity = s.HasHumidity
	s_Station_widget.SetText(s.Station)
	s_Station_widget.SetPlaceHolder("Home")
	s_Name_widget.SetText(s.Name)
	s_Name_widget.SetPlaceHolder("Name")
	s_Location_widget.SetText(s.Location)
	s_Location_widget.SetPlaceHolder("Location")
	s_Hide_widget.SetChecked(s.Hide)
	s_HasHumidity_widget.SetChecked(s.HasHumidity)
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
			s_HasHumidity_widget.SetChecked(sav_HasHumidity)
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
		s.HasHumidity = s_HasHumidity_widget.Checked
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

var showHumidityHandler = func(value bool) {
}

var resetHiLoHandler = func(value bool) {
	if value {
		resetHiLoFlag = true
		return
	}
	resetHiLoFlag = false
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

// EDIT SPECIFIC SENSOR (in widget)
var editSpecificSensorHandler = func(key string) {
	s := activeSensors[key]
	sav_Station = s.Station
	sav_Name = s.Name
	sav_Location = s.Location
	sav_Hide = s.Hide
	sav_HasHumidity = s.HasHumidity
	s_Station_widget.SetText(s.Station)
	s_Station_widget.SetPlaceHolder("Home")
	s_Name_widget.SetText(s.Name)
	s_Name_widget.SetPlaceHolder("Name")
	s_Location_widget.SetText(s.Location)
	s_Location_widget.SetPlaceHolder("Location")
	s_Hide_widget.SetChecked(s.Hide)
	s_HasHumidity_widget.SetChecked(s.HasHumidity)
	s_ResetHiLo_widget.SetChecked(false)
	s_Model_widget.SetText(s.Model)
	s_Id_widget.SetText(strconv.Itoa(s.Id))
	s_Channel_widget.SetText(s.Channel)
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	s_LastEdit_widget.SetText(st)
	SetStatus(fmt.Sprintf("Last edit set to %s", st))
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
			s_HasHumidity_widget.SetChecked(sav_HasHumidity)
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
		s.HasHumidity = s_HasHumidity_widget.Checked
		s.LastEdit = st
		if resetHiLoFlag {
			s.HighTemp = s.Temp
			s.LowTemp = s.Temp
			s.HighHumidity = s.Humidity
			s.LowHumidity = s.Humidity
		}
		activeSensors[key] = s
		reloadDashboard()
		editSensorWindow.Close()
	})
	editSensorWindow.Show()
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

var newSensorDisplayListHandler = func() {
	sensorSelectWindow = a.NewWindow("Scrolling list of sensors")
	sensorSelectWindow.Resize(fyne.NewSize(sensorDisplayWidgetSizeX+20, sensorDisplayWidgetSizeY*10))
	FillSensorSelectionContainer()
	sensorSelectWindow.SetContent(SensorSelectScroller)
	sensorSelectWindow.Show()
}
