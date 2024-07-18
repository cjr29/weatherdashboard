/******************************************************************
 *
 * WeatherWidget - Generates and manages the custom widget used by
 *                 the GUI dashboard
 *
 ******************************************************************/
/*
*    The container object to display the data from a given sensor
 */

package main

import (
	"image/color"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	widgetBackgroundColor = color.RGBA{R: 214, G: 240, B: 246, A: 255}
	widgetFrameColor      = color.Black
)

// Goroutine to run for each weather widget to watch for channel messages
var wwHandler func(key string) = func(key string) {
	// Loop forever
	var c = weatherWidgets[key].channel
	for {
		// Kill yourself if widget has been removed
		if weatherWidgets[key] == nil {
			return
		}
		key := <-c
		// If weatherWidget is missing, return because no longer need this handler
		if !checkWeatherWidget(key) {
			return
		}

		weatherWidgets[key].Refresh()
	}
}

/******************************************
 * Renderer Methods
 ******************************************/

func newWeatherWidgetRenderer(ww *weatherWidget) fyne.WidgetRenderer {
	r := weatherWidgetRenderer{}
	ww.renderer = &r

	frame := &canvas.Rectangle{
		FillColor:   widgetBackgroundColor,
		StrokeColor: widgetFrameColor,
		StrokeWidth: strokeWidth,
	}
	frame.SetMinSize(fyne.NewSize(widgetSizeX, widgetSizeY))
	frame.Resize(fyne.NewSize(widgetSizeX, widgetSizeY))
	frame.CornerRadius = cornerRadius

	header := canvas.NewText(ww.sensorName, color.Black)
	header.TextSize = 18

	st := canvas.NewText(ww.sensorStation, color.Black)
	st.TextSize = 10
	st.TextStyle = fyne.TextStyle{Bold: true}

	tw := canvas.NewText(strconv.FormatFloat(ww.temp, 'f', 1, 64)+"º", color.Black)
	tw.TextSize = 40
	tw.TextStyle = fyne.TextStyle{Bold: true}
	tw2 := canvas.NewText("fahrenheit", color.Black)
	tw2.TextSize = 10
	tw2.TextStyle = fyne.TextStyle{Bold: false}
	xpos := ((widgetSizeX / 2) - (tw2.MinSize().Width)/2)
	tw2.Move(fyne.NewPos(xpos, 70))

	hw := canvas.NewText(strconv.FormatFloat(ww.humidity, 'f', 1, 64)+"%", color.Black)
	hw.TextSize = 20
	hw.TextStyle = fyne.TextStyle{Italic: true}
	xpos = ((widgetSizeX / 2) - (hw.MinSize().Width)/2)
	hw.Move(fyne.NewPos(xpos, 85))
	hw2 := canvas.NewText("Humidity", color.Black)
	hw2.TextSize = 10
	hw2.TextStyle = fyne.TextStyle{Italic: true}
	xpos = ((widgetSizeX / 2) - (hw2.MinSize().Width)/2)
	hw2.Move(fyne.NewPos(xpos, 110))

	htw := canvas.NewText("Hi "+strconv.FormatFloat(ww.highTemp, 'f', 1, 64), color.RGBA{R: 247, G: 19, B: 2, A: 255})
	htw.TextSize = 10

	ltw := canvas.NewText("Lo "+strconv.FormatFloat(ww.lowTemp, 'f', 1, 64), color.RGBA{R: 11, G: 11, B: 243, A: 255})
	ltw.TextSize = 10

	hhw := canvas.NewText("Hi "+strconv.FormatFloat(ww.highHumidity, 'f', 1, 64)+"%", color.RGBA{R: 247, G: 19, B: 2, A: 255})
	hhw.TextSize = 10

	lhw := canvas.NewText("Lo "+strconv.FormatFloat(ww.lowHumidity, 'f', 1, 64)+"%", color.RGBA{R: 11, G: 11, B: 243, A: 255})
	lhw.TextSize = 10

	latestUpdate := canvas.NewText("Latest update:   "+ww.latestUpdate, color.Black)
	latestUpdate.TextSize = 12

	r.widget = ww
	r.frame = frame
	r.sensorName = header
	r.station = st
	r.temp = tw
	r.humidity = hw
	r.humidity2 = hw2
	r.highTemp = htw
	r.lowTemp = ltw
	r.highHumidity = hhw
	r.lowHumidity = lhw
	r.latestUpdate = latestUpdate
	r.objects = append(r.objects, frame, header, st, tw, tw2, hw, hw2, htw, ltw, hhw, lhw, latestUpdate)

	r.widget.ExtendBaseWidget(ww)

	return &r
}

func (r *weatherWidgetRenderer) Destroy() {

}

func (r *weatherWidgetRenderer) Layout(size fyne.Size) {
	r.frame.Move(fyne.NewPos(0, 0))
	// Calculate start of sensor name in widget
	xpos := ((widgetSizeX / 2) - (r.sensorName.MinSize().Width)/2)
	r.sensorName.Move(fyne.NewPos(xpos, 0))
	// r.station.Move(fyne.NewPos(4, 5))
	xpos = widgetSizeX - widgetPadding - r.station.MinSize().Width
	r.station.Move(fyne.NewPos(xpos, 65))
	xpos = ((widgetSizeX / 2) - (r.temp.MinSize().Width)/2)
	r.temp.Move(fyne.NewPos(xpos, 25))
	xpos = ((widgetSizeX / 2) - (r.humidity.MinSize().Width)/2)
	r.humidity.Move(fyne.NewPos(xpos, 85))
	xpos = ((widgetSizeX / 2) - (r.humidity2.MinSize().Width)/2)
	r.humidity2.Move(fyne.NewPos(xpos, 110))
	r.highTemp.Move(fyne.NewPos(4, 40))
	r.lowTemp.Move(fyne.NewPos(4, 55))
	r.highHumidity.Move(fyne.NewPos(4, 85))
	r.lowHumidity.Move(fyne.NewPos(4, 100))
	xpos = ((widgetSizeX / 2) - (r.latestUpdate.MinSize().Width)/2)
	r.latestUpdate.Move(fyne.NewPos(xpos, 130))
	if !r.widget.hasHumidity {
		r.humidity.Hide()
		r.highHumidity.Hide()
		r.lowHumidity.Hide()
		r.humidity2.Hide()
	}
}

func (r *weatherWidgetRenderer) MinSize() fyne.Size {
	return fyne.NewSize(widgetSizeX, widgetSizeY)
}

func (r *weatherWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *weatherWidgetRenderer) Refresh() {
	r.frame.Resize(fyne.NewSize(widgetSizeX, widgetSizeY)) // This is critical or frame won't appear
	r.frame.Show()
	r.sensorName.Text = r.widget.sensorName
	r.station.Text = r.widget.sensorStation
	r.temp.Text = strconv.FormatFloat(r.widget.temp, 'f', 1, 64) + "º"
	r.humidity.Text = strconv.FormatFloat(r.widget.humidity, 'f', 1, 64) + "%"
	r.highTemp.Text = "Hi " + strconv.FormatFloat(r.widget.highTemp, 'f', 1, 64)
	r.lowTemp.Text = "Lo " + strconv.FormatFloat(r.widget.lowTemp, 'f', 1, 64)
	r.highHumidity.Text = "Hi " + strconv.FormatFloat(r.widget.highHumidity, 'f', 1, 64) + "%"
	r.lowHumidity.Text = "Lo " + strconv.FormatFloat(r.widget.lowHumidity, 'f', 1, 64) + "%"
	r.latestUpdate.Text = "Updated:   " + r.widget.latestUpdate
	if !r.widget.hasHumidity {
		r.lowHumidity.Hide()
		r.highHumidity.Hide()
		r.humidity.Hide()
	} else {
		r.lowHumidity.Show()
		r.highHumidity.Show()
		r.humidity.Show()
	}
}

/************************************
 * WeatherWidget Methods
 ************************************/

func (ww *weatherWidget) CreateRenderer() fyne.WidgetRenderer {
	r := newWeatherWidgetRenderer(ww)
	return r
}

func (ww *weatherWidget) Refresh() {
	if ww == nil {
		return
	}
	ww.BaseWidget.Refresh()
	if ww.renderer != nil {
		ww.renderer.Refresh()
	}
}

func (ww *weatherWidget) SetSensorName(name string) {
	ww.sensorName = name
}

func (ww *weatherWidget) SetTemp(temp float64) {
	ww.temp = temp
}

func (ww *weatherWidget) SetHumidity(humidity float64) {
	ww.humidity = humidity
}

func (ww *weatherWidget) SetHighTemp(temp float64) {
	ww.highTemp = temp
}

func (ww *weatherWidget) SetLowTemp(temp float64) {
	ww.lowTemp = temp
}

func (ww *weatherWidget) SetHighHumidity(humidity float64) {
	ww.highHumidity = humidity
}

func (ww *weatherWidget) SetLowHumidity(humidity float64) {
	ww.lowHumidity = humidity
}

func (ww *weatherWidget) SetLatestUpdate(latest string) {
	ww.latestUpdate = latest
}

func (ww *weatherWidget) UpdateDate() {
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	ww.latestUpdate = st
}

// Initialize fields of a weather widget using data from Sensor
func (ww *weatherWidget) Init(s *Sensor) {
	ww.sensorKey = s.Key
	ww.sensorName = s.Name
	ww.sensorStation = s.Station
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	ww.latestUpdate = st
	ww.temp = s.Temp
	ww.humidity = s.Humidity
	ww.hasHumidity = s.HasHumidity
	ww.highHumidity = s.HighHumidity
	ww.lowHumidity = s.LowHumidity
	ww.highTemp = activeSensors[s.Key].HighTemp
	ww.highTemp = s.HighTemp
	ww.lowTemp = s.LowTemp
	ww.latestUpdate = s.DataDate
	wwc := make(chan string, 5) // Buffered channel for this sensor
	ww.channel = wwc
	ww.goHandler = wwHandler
}

func (ww *weatherWidget) Tapped(*fyne.PointEvent) {
	// Call handler to edit the widget elements
	if ewwFlag {
		return // There is only one edit window at a time
	}
	editSpecificSensorHandler(ww.sensorKey)
}

func (mc *weatherWidget) TappedSecondary(*fyne.PointEvent) {}

// sortWeatherWidgets - returns array of keys for the sorted widgets
func sortWeatherWidgets() (sortedWWKeys []string) {

	keys := make([]string, 0, len(weatherWidgets))

	// Build array of widget keys to be sorted
	for key := range weatherWidgets {
		keys = append(keys, key)
	}

	// Sort the sensor key array using the Station and Name from each sensor
	sort.SliceStable(keys, func(i, j int) bool {
		ww1 := weatherWidgets[keys[i]].sensorStation + ":" + weatherWidgets[keys[i]].sensorName
		ww2 := weatherWidgets[keys[j]].sensorStation + ":" + weatherWidgets[keys[j]].sensorName
		return ww1 < ww2
	})

	return keys
}
