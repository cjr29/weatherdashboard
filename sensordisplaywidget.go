/******************************************************************
 *
 * SensorDisplayWidget - Generates and manages the custom widget
 *           used by the GUI dashboard to display sensor info
 *           and selection lists
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
	SensorSelectDisp     = container.NewVBox()
	SensorSelectScroller = container.NewVScroll(SensorSelectDisp)

	sensorDisplayWidgetBackgroundColor = color.RGBA{R: 214, G: 240, B: 246, A: 255}
	sensorDisplayWidgetForegroundColor = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // Black
	sensorDisplayWidgetFrameColor      = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // Black
)

type SensorDisplayWidget struct {
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
	renderer          *SensorDisplayWidgetRenderer
}

type SensorDisplayWidgetRenderer struct {
	widget       *SensorDisplayWidget
	frame        *canvas.Rectangle
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

// FillSensorSelectionContainer - Call this function to put weather data string in the weather display scrolling window
func FillSensorSelectionContainer() {
	for key, sens := range activeSensors {
		// temp := widget.NewLabel("")
		// temp.TextStyle = fyne.TextStyle{Monospace: true}
		temp := SensorDisplayWidget{}
		temp.Init(sens)
		a := *activeSensors[key]
		temp.station = a.Station
		temp.name = a.Name
		temp.model = a.Model
		temp.id = a.Id
		temp.channel = a.Channel
		temp.dateAdded = a.DateAdded
		temp.latestUpdate = a.LastEdit
		temp.hasHumidity = a.HasHumidity
		SensorSelectDisp.Add(&temp)
	}

	SensorSelectScroller.ScrollToBottom()

	SensorSelectDisp.Refresh()
}

/******************************************
 * Renderer Methods
 ******************************************/

func newSensorDisplayWidgetRenderer(sdw *SensorDisplayWidget) fyne.WidgetRenderer {
	r := SensorDisplayWidgetRenderer{}
	sdw.renderer = &r

	frame := &canvas.Rectangle{
		FillColor:   sensorDisplayWidgetBackgroundColor,
		StrokeColor: sensorDisplayWidgetFrameColor,
		StrokeWidth: sensorDisplayStrokeWidth,
	}
	frame.SetMinSize(fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY))
	frame.Resize(fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY))
	frame.CornerRadius = sensorDisplayCornerRadius

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
	r.station = st
	r.name = sn
	r.model = mo
	r.id = id
	r.channel = ch
	r.latestUpdate = latestUpdate
	r.dateAdded = dateAdded
	r.hasHumidity = hi
	r.objects = append(r.objects, frame, st, sn, mo, id, ch, hi, latestUpdate, dateAdded)

	r.widget.ExtendBaseWidget(sdw)

	return &r
}

func (r *SensorDisplayWidgetRenderer) Destroy() {

}

func (r *SensorDisplayWidgetRenderer) Layout(size fyne.Size) {
	r.frame.Move(fyne.NewPos(0, 0))
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

func (r *SensorDisplayWidgetRenderer) MinSize() fyne.Size {
	return fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY)
}

func (r SensorDisplayWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *SensorDisplayWidgetRenderer) Refresh() {
	r.frame.Resize(fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY)) // This is critical or frame won't appear
	r.frame.Show()
	r.name.Text = r.widget.name
	r.station.Text = r.widget.station
	r.model.Text = r.widget.model
	r.id.Text = strconv.Itoa(r.widget.id)
	r.channel.Text = r.widget.channel
	r.latestUpdate.Text = r.widget.latestUpdate
	r.dateAdded.Text = r.widget.dateAdded
}

/************************************
 * SensorDisplayWidget Methods
 ************************************/

func (sdw *SensorDisplayWidget) CreateRenderer() fyne.WidgetRenderer {
	r := newSensorDisplayWidgetRenderer(sdw)
	return r
}

func (sdw *SensorDisplayWidget) Hide() {

}

func (sdw *SensorDisplayWidget) Refresh() {
	if sdw == nil {
		return
	}
	sdw.BaseWidget.Refresh()
	if sdw.renderer != nil {
		sdw.renderer.Refresh()
	}
}

/* func (sdw *SensorDisplayWidget) SetSensorName(name string) {
	sdw.sensorName = name
}

func (sdw *SensorDisplayWidget) SetTemp(temp float64) {
	sdw.temp = temp
}

func (sdw *SensorDisplayWidget) SetHumidity(humidity float64) {
	sdw.humidity = humidity
}

func (sdw *SensorDisplayWidget) SetHighTemp(temp float64) {
	sdw.highTemp = temp
}

func (sdw *SensorDisplayWidget) SetLowTemp(temp float64) {
	sdw.lowTemp = temp
}

func (sdw *SensorDisplayWidget) SetHighHumidity(humidity float64) {
	sdw.highHumidity = humidity
}

func (sdw *SensorDisplayWidget) SetLowHumidity(humidity float64) {
	sdw.lowHumidity = humidity
}

func (sdw *SensorDisplayWidget) SetLatestUpdate(latest string) {
	sdw.latestUpdate = latest
} */

func (sdw *SensorDisplayWidget) UpdateDate() {
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	sdw.latestUpdate = st
}

// Initialize fields of a SensorDisplayWidget using data from Sensor
func (sdw *SensorDisplayWidget) Init(s *Sensor) {
	sdw.sensorKey = s.Key
	sdw.name = s.Name
	sdw.station = s.Station
	sdw.model = s.Model
	sdw.id = s.Id
	sdw.dateAdded = s.DateAdded
	sdw.latestUpdate = s.LastEdit
	sdw.hasHumidity = s.HasHumidity
}

/* func (sdw *SensorDisplayWidget) Tapped(*fyne.PointEvent) {
	fmt.Println("Widget tapped")
	// Call handler to edit the widget elements
	editSpecificSensorHandler(sdw.sensorKey)
}

func (mc *weatherWidget) TappedSecondary(*fyne.PointEvent) {} */
