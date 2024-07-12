/******************************************************************
 *
 * chooseSensors - Generates a list of sensors from map supplied in argument
 *      and pops up a selection window allowing user to pick one or more
 *      sensors from the list. Once user presses Submit, updates an
 *      array of sensor keys, resultKeys, that can be used for processing.
 *   title is the heading to be displayed at top of window frame
 *   showChecks set to true to show selection check boxes
 *   singleCheck set to true to allow only one box to be checked
 *
 * Usage:
 *      chooseSensors(title string, sensors map[string]*Sensor, showChecks bool, singleCheck bool)
 *
 ******************************************************************/

package main

import (
	"fmt"
	"image/color"
	"strconv"
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

var (
	sensorSelectDisp                   = container.NewVBox()
	sensorSelectScroller               = container.NewVScroll(sensorSelectDisp)
	submitButton                       = widget.NewButton("Submit", getSelectedItems)
	cancelButton                       = widget.NewButton("Cancel", cancelSelection)
	selectedItems                      map[string]bool
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
	renderer          *sensorDisplayWidgetRenderer
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
//	if singleCheck, allow only one box to be checked
func chooseSensors(title string, sensors map[string]*Sensor, showChecks bool, singleCheck bool) {
	// Check to be sure window isn't already displayed
	if (sensorSelectWindow != nil) && sensorSelectWindow.Content().Visible() {
		return
	}
	showCheckBoxesFlag = showChecks // Set the global flag for widgets to see
	// Make a slice with capacity of all sensors passed in the argument
	resultKeys = make([]string, 0, len(sensors))

	// Prepare the containers and widgets to display the selection list
	sensorSelectWindow = a.NewWindow(title)
	sensorSelectWindow.SetOnClosed(processSelection)
	sensorSelectWindow.Resize(fyne.NewSize(sensorDisplayWidgetSizeX+20, sensorDisplayWidgetSizeY*10))
	fillSensorSelectionContainer(sensors)
	var buttonContainer *fyne.Container
	// Show Submit only if a choice is required
	if showCheckBoxesFlag {
		buttonContainer = container.NewHBox(
			submitButton,
			cancelButton,
		)
	} else {
		buttonContainer = container.NewHBox(
			cancelButton,
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

// var processSelection = func() {
func processSelection() {
	fmt.Println("processSelections")
	// Loop over selected sensors to edit
	if len(resultKeys) == 0 {
		fmt.Println("No selection")
		return
	}
	for _, key := range resultKeys {
		// Open an edit window for each selected sensor

		// Pause until edit finished for previous sensor
		// for inprocessFlag {
		// 	// set timer for 5 secs
		// 	time.Sleep(5 * time.Second)
		// }

		fmt.Println("editSensorHandler: sensor ", key)
		// Get key for selected sensor - Key: field of value
		// key := strings.Split(selectedValue, "Key: ")[1] // Found at end of the displayed string after "Key: ..."
		// Load widgets using selected sensor
		s := activeSensors[key]
		sav_Station := s.Station
		sav_Name := s.Name
		sav_Location := s.Location
		sav_Hide := s.Hide
		sav_HasHumidity := s.HasHumidity
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
		s_ResetHiLo_widget := widget.NewCheck("Reset Hi/Lo", resetHiLoHandler)
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
		//selectSensorWindow.Close()
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
				SetStatus(fmt.Sprintf("Saving updated sensor record: %s", s_Name_widget.Text+":"+key))
				s.Station = s_Station_widget.Text
				s.Name = s_Name_widget.Text
				s.Location = s_Location_widget.Text
				s.Hide = s_Hide_widget.Checked
				s.HasHumidity = s_HasHumidity_widget.Checked
				s.LastEdit = st
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
				editSensorWindow.Close()
			}),
		)
		editSensorWindow.SetContent(editSensorContainer)
		fmt.Println("Showing edit window for sensor ", s.Station)
		editSensorWindow.Show()
	}
}

// fillSensorSelectionContainer - Call this function to put weather data string in the weather display scrolling window
func fillSensorSelectionContainer(sensors map[string]*Sensor) {
	// Clear selected items map
	selectedItems = make(map[string]bool)
	sensorSelectDisp.Objects = nil
	for key, sens := range sensors {
		temp := sensorDisplayWidget{}
		temp.init(sens)
		a := *activeSensors[key]
		temp.station = a.Station
		temp.name = a.Name
		temp.model = a.Model
		temp.id = a.Id
		temp.channel = a.Channel
		temp.dateAdded = a.DateAdded
		temp.latestUpdate = a.LastEdit
		temp.hasHumidity = a.HasHumidity
		sensorSelectDisp.Add(&temp)
	}

	sensorSelectScroller.ScrollToBottom()

	sensorSelectDisp.Refresh()
}

// Ranges through the sensorSelectDisp container, checking each item for a checked box
func getSelectedItems() {
	for key, value := range selectedItems {
		if value {
			resultKeys = append(resultKeys, key)
		}
	}

	// TESTING - remove for production
	for _, value := range resultKeys {
		fmt.Printf("Sensor %s selected\n", value)
	}
	// TESTING

	sensorSelectWindow.Close()
	sensorSelectWindow = nil
}

func cancelSelection() {
	sensorSelectWindow.Close()
	sensorSelectWindow = nil
}

// function to check if a bool value present in the map
// func checkForBool(b bool, m map[string]bool) bool {

// 	//traverse through the map
// 	for _, value := range m {
// 		//check if present value is equals to b
// 		if value == b {
// 			//if same return true
// 			return true
// 		}
// 	}

// 	//if value not found return false
// 	return false
// }

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
		selectedItems[r.widget.sensorKey] = b
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
}
