/******************************************************************
 *
 * chooseSensors - Generates a list of sensors from map supplied in argument
 *      and pops up a selection window allowing user to pick one or more
 *      sensors from the list. Once user presses Submit, updates an
 *      array of sensor keys, resultKeys, that can be used for processing.
 *   title is the heading to be displayed at top of window frame
 *   sensors is the sensor map to use for the selections
 *   action is the process to be performed on the selected sensors
 *
 * Usage:
 *      chooseSensors(title string, sensors map[string]*Sensor, action Action)
 *
 ******************************************************************/

package main

import (
	"fmt"
	"image/color"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	sensorDisplayWidgetSizeX   float32 = 800
	sensorDisplayWidgetSizeY   float32 = 40
	sensorDisplayWidgetPadding float32 = 0 // separation between widgets
	sensorDisplayCornerRadius  float32 = 5
	sensorDisplayStrokeWidth   float32 = 1
)

type Action int

const (
	ListActive Action = iota + 1 // List active sensors with no check box
	ListAvail                    // List available sensors with no check box
	Edit                         // Edit selected sensors
	Add                          // Add selected sensors
	Remove                       // Remove selected sensors
)

var (
	resultKeys                         []string // storage for sensor keys selected by user
	sensorDisplayWidgetBackgroundColor          = color.RGBA{R: 214, G: 240, B: 246, A: 255}
	sensorDisplayWidgetForegroundColor          = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // Black
	sensorDisplayWidgetFrameColor               = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // Black
	showCheckBoxesFlag                 bool     = false
)

type sensorDisplayWidget struct {
	widget.BaseWidget // Inherit from BaseWidget
	sensorKey         string
	station           string
	name              string
	model             string
	id                int
	channel           string
	dateAdded         string
	latestUpdate      string
	hasHumidity       bool
	check             bool
	renderer          *sensorDisplayWidgetRenderer
	sync.Mutex
}

type sensorDisplayWidgetRenderer struct {
	widget       *sensorDisplayWidget
	frame        *canvas.Rectangle
	checkbox     *widget.Check
	station      *canvas.Text
	name         *canvas.Text
	model        *canvas.Text
	id           *canvas.Text
	channel      *canvas.Text
	dateAdded    *canvas.Text
	latestUpdate *canvas.Text
	hasHumidity  *canvas.Text
	objects      []fyne.CanvasObject
}

// chooseSensors
//
//	if showChecks, display check boxes
//	Use Action type to determine which activity to initiate after selections are made
func chooseSensors(title string, sensors map[string]*Sensor, action Action) {
	var sensorSelectDisp = container.NewVBox()
	var sensorSelectScroller = container.NewVScroll(sensorSelectDisp)
	sensorSelectScroller.SetMinSize(fyne.NewSize(600, 50))

	// Only display one instance of a window at a time
	if listActiveSensorsFlag && (action == ListActive) {
		return
	}
	if listAvailableSensorsFlag && (action == ListAvail) {
		return
	}
	if editActiveSensorsFlag && (action == Edit) {
		return
	}
	if addActiveSensorsFlag && (action == Add) {
		return
	}
	if removeActiveSensorsFlag && (action == Remove) {
		return
	}

	// Depending on action, turn check boxes on or off by setting the global flag
	switch action {
	case ListActive:
		listActiveSensorsFlag = true
		showCheckBoxesFlag = false // Set the global flag for widgets to see
	case ListAvail:
		listAvailableSensorsFlag = true
		showCheckBoxesFlag = false // Set the global flag for widgets to see
	case Edit:
		editActiveSensorsFlag = true
		showCheckBoxesFlag = true
	case Add:
		addActiveSensorsFlag = true
		showCheckBoxesFlag = true
	case Remove:
		removeActiveSensorsFlag = true
		showCheckBoxesFlag = true
	default:
		showCheckBoxesFlag = false
	}

	// Prepare the containers and widgets to display the selection list
	sensorSelectWindow := a.NewWindow(title)
	sensorSelectWindow.SetOnClosed(func() {
		clearSelectedFlag(action)
	})

	sensorSelectWindow.Resize(fyne.NewSize(sensorDisplayWidgetSizeX+20, sensorDisplayWidgetSizeY*10))

	// Make a slice with capacity of all sensors passed in the argument
	resultKeys = make([]string, 0, len(sensors))
	var choices []*sensorDisplayWidget
	for key, sens := range sensors {
		temp := sensorDisplayWidget{}
		temp.Lock()
		temp.init(sens)
		a := sensors[key]
		temp.station = a.Station
		temp.name = a.Name
		temp.model = a.Model
		temp.id = a.Id
		temp.channel = a.Channel
		temp.dateAdded = a.DateAdded
		temp.latestUpdate = a.LastEdit
		temp.hasHumidity = a.HasHumidity
		sensorSelectDisp.Add(&temp)
		choices = append(choices, &temp)
		temp.Unlock()
	}
	sensorSelectScroller.ScrollToBottom()
	sensorSelectDisp.Refresh()

	var buttonContainer *fyne.Container
	// Show Submit only if a choice is required
	if showCheckBoxesFlag {
		buttonContainer = container.NewHBox(
			widget.NewButton("Submit", func() {
				for _, value := range choices {
					value.Lock()
					if value.check {
						resultKeys = append(resultKeys, value.sensorKey)
					}
					value.Unlock()
				}
				clearSelectedFlag(action)
				sensorSelectWindow.Close()
			}),
			widget.NewButton("Cancel", func() {
				clearSelectedFlag(action)
				sensorSelectWindow.Close()
			}),
		)
	} else {
		buttonContainer = container.NewHBox(
			widget.NewButton("Cancel", func() {
				sensorSelectWindow.Close()
			}),
		)
	}
	overallContainer := container.NewBorder(
		buttonContainer,      // top
		nil,                  // bottom
		nil,                  // left
		nil,                  // right
		sensorSelectScroller, // middle
	)
	sensorSelectWindow.SetContent(overallContainer)
	sensorSelectWindow.Show()
}

// clearSelectedFlag(action)
//
//	Helper to clear specific window flag
func clearSelectedFlag(action Action) {
	switch action {
	case ListActive:
		listActiveSensorsFlag = false
		return
	case ListAvail:
		listAvailableSensorsFlag = false
		return
	case Edit:
		editSensors()
		editActiveSensorsFlag = false
	case Add:
		addSensors()
		addActiveSensorsFlag = false
	case Remove:
		removeSensors()
		removeActiveSensorsFlag = false
	default:
		return
	}
}

/**********************************************************
*
* editSensors()
*
**********************************************************/
func editSensors() {
	// Loop over selected sensors to edit
	if len(resultKeys) == 0 {
		return
	}
	for _, key := range resultKeys {

		// Load widgets using selected sensor
		resetHiLoFlag := false
		activeSensorsMutex.Lock()
		s := activeSensors[key]
		activeSensorsMutex.Unlock()
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
				s.HighTemp = s.Temp
				s.LowTemp = s.Temp
				s.HighHumidity = s.Humidity
				s.LowHumidity = s.Humidity
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
		// Pop up a window to let user edit record, then save the record
		editSensorWindow := a.NewWindow("Edit active sensor properties.")
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
				SetStatus(fmt.Sprintf("Saving updated sensor record for  %s", s_Name_widget.Text))
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
				// If dashboard is visible, reload it since we changed a sensor in a widget
				if dashFlag {
					reloadDashboard()
				}
				editSensorWindow.Close()
			}),
			widget.NewButton("Cancel", func() {
				// Restore orginal values
				s_Station_widget.SetText(sav_Station)
				s_Name_widget.SetText(sav_Name)
				s_Location_widget.SetText(sav_Location)
				s_Hide_widget.SetChecked(sav_Hide)
				s_HasHumidity_widget.SetChecked(sav_HasHumidity)
				editSensorWindow.Close()
			}),
		)
		editSensorWindow.SetContent(editSensorContainer)

		editSensorWindow.Show()
	}
}

/**********************************************************
*
* addSensors()
*
**********************************************************/
func addSensors() {
	for _, key := range resultKeys {
		if checkSensor(key, availableSensors) && !availableSensors[key].Hide {
			availableSensorsMutex.Lock()
			s := *availableSensors[key]
			availableSensorsMutex.Unlock()
			activeSensorsMutex.Lock()
			activeSensors[key] = &s
			activeSensorsMutex.Unlock()
			reloadDashboard()
			SetStatus(fmt.Sprintf("Added sensor to active sensors: %s", key))
		}
	}
}

/**********************************************************
*
* removeSensors()
*
**********************************************************/
func removeSensors() {
	for _, key := range resultKeys {
		if checkSensor(key, activeSensors) { // Be sure the sensor is there
			activeSensorsMutex.Lock()
			delete(activeSensors, key)
			activeSensorsMutex.Unlock()
			reloadDashboard()
			SetStatus(fmt.Sprintf("Removed sensor from active sensors: %s", key))
		}
	}
}

/******************************************
 * Renderer Methods
 ******************************************/

func newSensorDisplayWidgetRenderer(sdw *sensorDisplayWidget) fyne.WidgetRenderer {
	r := sensorDisplayWidgetRenderer{}
	sdw.renderer = &r

	frame := &canvas.Rectangle{
		FillColor:   sensorDisplayWidgetBackgroundColor,
		StrokeColor: sensorDisplayWidgetFrameColor,
		StrokeWidth: sensorDisplayStrokeWidth,
	}
	frame.SetMinSize(fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY))
	frame.Resize(fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY))
	frame.CornerRadius = sensorDisplayCornerRadius

	/******************************
	* Add a Check Box widget
	******************************/
	check := widget.NewCheck("", func(b bool) {
		r.widget.check = b
		// selectedItems[r.widget.sensorKey] = b
	})

	st := canvas.NewText(sdw.station, sensorDisplayWidgetForegroundColor)
	st.TextSize = 14
	st.TextStyle = fyne.TextStyle{Bold: true}

	sn := canvas.NewText(sdw.name, sensorDisplayWidgetForegroundColor)
	sn.TextSize = 14

	ch := canvas.NewText("CH: "+sdw.channel, sensorDisplayWidgetForegroundColor)
	ch.TextSize = 14

	mo := canvas.NewText(sdw.model, sensorDisplayWidgetForegroundColor)
	mo.TextSize = 14

	id := canvas.NewText("ID: "+strconv.Itoa(sdw.id), sensorDisplayWidgetForegroundColor)
	id.TextSize = 14

	hi := canvas.NewText(fmt.Sprintf("humidity: %v", sdw.hasHumidity), sensorDisplayWidgetForegroundColor)
	hi.TextSize = 11

	latestUpdate := canvas.NewText("Latest update:   "+sdw.latestUpdate, sensorDisplayWidgetForegroundColor)
	latestUpdate.TextSize = 11

	dateAdded := canvas.NewText("Date added:   "+sdw.dateAdded, sensorDisplayWidgetForegroundColor)
	dateAdded.TextSize = 11

	r.widget = sdw
	r.frame = frame
	r.checkbox = check
	r.station = st
	r.name = sn
	r.model = mo
	r.id = id
	r.channel = ch
	r.latestUpdate = latestUpdate
	r.dateAdded = dateAdded
	r.hasHumidity = hi
	if showCheckBoxesFlag {
		r.objects = append(r.objects, frame, check, st, sn, mo, id, ch, hi, latestUpdate, dateAdded)
	} else {
		r.objects = append(r.objects, frame, st, sn, mo, id, ch, hi, latestUpdate, dateAdded)
	}

	r.widget.ExtendBaseWidget(sdw)

	return &r
}

func (r *sensorDisplayWidgetRenderer) Destroy() {
}

func (r *sensorDisplayWidgetRenderer) Layout(size fyne.Size) {
	r.frame.Move(fyne.NewPos(0, 0))
	r.checkbox.Resize(fyne.NewSize(20, 20))
	r.checkbox.Move(fyne.NewPos(5, 10))
	ypos := ((sensorDisplayWidgetSizeY - r.name.TextSize) / 2) + 5
	r.station.Move(fyne.NewPos(40, ypos))
	r.hasHumidity.Move(fyne.NewPos(40, ypos-15))
	ypos = ((sensorDisplayWidgetSizeY - r.name.TextSize) / 2) - 5
	r.name.Move(fyne.NewPos(120, ypos))
	xpos := r.name.Size().Width + 275
	r.model.Move(fyne.NewPos(xpos, ypos))
	xpos = xpos + r.model.Size().Width + 180
	r.id.Move(fyne.NewPos(xpos, ypos))
	xpos = xpos + r.id.Size().Width + 80
	r.channel.Move(fyne.NewPos(xpos, ypos))
	r.dateAdded.Move(fyne.NewPos((sensorDisplayWidgetSizeX-r.latestUpdate.MinSize().Width)-5, (sensorDisplayWidgetSizeY-r.dateAdded.TextSize)*0.25))
	r.latestUpdate.Move(fyne.NewPos((sensorDisplayWidgetSizeX-r.latestUpdate.MinSize().Width)-5, (sensorDisplayWidgetSizeY-r.latestUpdate.TextSize)*0.75))
}

func (r *sensorDisplayWidgetRenderer) MinSize() fyne.Size {
	return fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY)
}

func (r sensorDisplayWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *sensorDisplayWidgetRenderer) Refresh() {
	r.frame.Resize(fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY)) // This is critical or frame won't appear
	r.frame.Show()
	r.checkbox.Resize(fyne.NewSize(20, 20))
	r.checkbox.Move(fyne.NewPos(5, 10))
	r.checkbox.Show()
	r.name.Text = r.widget.name
	r.station.Text = r.widget.station
	r.model.Text = r.widget.model
	r.id.Text = strconv.Itoa(r.widget.id)
	r.channel.Text = r.widget.channel
	r.latestUpdate.Text = r.widget.latestUpdate
	r.dateAdded.Text = r.widget.dateAdded
}

/************************************
 * sensorDisplayWidget Methods
 ************************************/

func (sdw *sensorDisplayWidget) CreateRenderer() fyne.WidgetRenderer {
	r := newSensorDisplayWidgetRenderer(sdw)
	return r
}

func (sdw *sensorDisplayWidget) Hide() {
}

func (sdw *sensorDisplayWidget) Checked() bool {
	return sdw.check
}

func (sdw *sensorDisplayWidget) Refresh() {
	if sdw == nil {
		return
	}
	sdw.BaseWidget.Refresh()
	if sdw.renderer != nil {
		sdw.renderer.Refresh()
	}
}

func (sdw *sensorDisplayWidget) UpdateDate() {
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	sdw.latestUpdate = st
}

// Initialize fields of a sensorDisplayWidget using data from Sensor
func (sdw *sensorDisplayWidget) init(s *Sensor) {
	sdw.sensorKey = s.Key
	sdw.name = s.Name
	sdw.station = s.Station
	sdw.model = s.Model
	sdw.id = s.Id
	sdw.dateAdded = s.DateAdded
	sdw.latestUpdate = s.LastEdit
	sdw.hasHumidity = s.HasHumidity
	sdw.check = false
}
