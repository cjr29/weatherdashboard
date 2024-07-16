/******************************************************************
 *
 * Dashboard Menu handler functions
 *
 ******************************************************************/

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//*************************************************
// Menu action handlers
//*************************************************

var hideWidgetHandler = func(value bool) {
	hideflag = value
}

var showHumidityHandler = func(value bool) {
}

// editSpecificSensorHandler - Used by dashboard widgets to edit a tapped widget
func editSpecificSensorHandler(key string) {
	ewwFlag = true // Prevent more than one edit window at a time
	resetHiLoFlag := false
	// Load widget using selected sensor
	s := activeSensors[key]
	// Save values in case edit is canceled by user
	sav_Station := s.Station
	sav_Name := s.Name
	sav_Location := s.Location
	sav_Hide := s.Hide
	sav_HasHumidity := s.HasHumidity
	// Load form fields
	s_Station_widget := widget.NewEntry()
	s_Station_widget.SetText(s.Station)
	s_Station_widget.SetPlaceHolder("Home")
	s_Name_widget := widget.NewEntry()
	s_Name_widget.SetText(s.Name)
	s_Name_widget.SetPlaceHolder("Name")
	s_Location_widget := widget.NewEntry()
	s_Location_widget.SetText(s.Location)
	s_Location_widget.SetPlaceHolder("Location")
	s_Hide_widget := widget.NewCheck("Check to hide sensor on weather dashboard", hideWidgetHandler)
	s_Hide_widget.SetChecked(s.Hide)
	s_HasHumidity_widget := widget.NewCheck("Check if sensor also provides humidity", showHumidityHandler)
	s_HasHumidity_widget.SetChecked(s.HasHumidity)
	s_ResetHiLo_widget := widget.NewCheck("Reset Hi/Lo", func(value bool) {
		if value {
			resetHiLoFlag = true
		}
	})
	s_ResetHiLo_widget.SetChecked(false)
	s_Model_widget := widget.NewLabel("")
	s_Model_widget.SetText(s.Model)
	s_Id_widget := widget.NewLabel("")
	s_Id_widget.SetText(strconv.Itoa(s.Id))
	s_Channel_widget := widget.NewLabel("")
	s_Channel_widget.SetText(s.Channel)
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	s_LastEdit_widget := widget.NewLabel("")
	s_LastEdit_widget.SetText(st)
	SetStatus(fmt.Sprintf("Last edit set to %s", st))

	// Pop up a window to let user edit record, then save the record
	editSensorWindow := a.NewWindow("Edit active sensor properties.")
	editSensorWindow.SetOnClosed(func() {
		// Reload and display the updated widgets in dashboard
		ewwFlag = false
		reloadDashboard()
	})
	editSensorContainer := container.NewVBox(
		widget.NewLabel("Update Home, Name, Location, and Visibility, then press Submit to save."),
		s_Station_widget,
		s_Name_widget,
		s_Location_widget,
		s_Hide_widget,
		s_HasHumidity_widget,
		s_ResetHiLo_widget,
		s_Model_widget,
		s_Id_widget,
		s_Channel_widget,
		s_LastEdit_widget,
		widget.NewButton("Submit", func() {
			// Save updated record back to activeSensors
			// First be sure that there is a widget record. User may have cleared the Hide flag, expecting a new sensor to display
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
			resetHiLoFlag = false
			activeSensors[key] = s
			editSensorWindow.Close()
		}),
		widget.NewButton("Cancel", func() {
			// Restore orginal values
			s_Station_widget.SetText(sav_Station)
			s_Name_widget.SetText(sav_Name)
			s_Location_widget.SetText(sav_Location)
			s_Hide_widget.SetChecked(sav_Hide)
			s_HasHumidity_widget.SetChecked(sav_HasHumidity)
			resetHiLoFlag = false
			editSensorWindow.Close()
		}),
	)
	editSensorWindow.SetContent(editSensorContainer)

	editSensorWindow.Show()
}

/******************************************************************
 *
 * exitHandler - callback used by main window when it is closed
 *
 ******************************************************************/
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
		dataWindow = a.NewWindow("Station Data Live Feed")
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
