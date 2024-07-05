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

// Goroutine to run for each weather widget to watch for channel messages
var wwHandler func(key string) = func(key string) {
	//fmt.Printf("wwHandler for %s - %s started.\n", weatherWidgets[key].sensorName, key)
	// Loop forever
	var c = weatherWidgets[key].channel
	for {
		// Kill yourself if widget has been removed
		if weatherWidgets[key] == nil {
			return
		}
		key := <-c
		// fmt.Println("Received key: ", key)
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
		FillColor:   color.RGBA{R: 202, G: 230, B: 243, A: 255},
		StrokeColor: color.RGBA{R: 202, G: 230, B: 243, A: 255},
	}
	frame.SetMinSize(fyne.NewSize(widgetSizeX, widgetSizeY))

	header := canvas.NewText(ww.sensorName, color.Black)
	header.TextSize = 18

	tw := canvas.NewText(strconv.FormatFloat(ww.temp, 'f', 1, 64), color.Black)
	tw.TextSize = 40
	tw.TextStyle = fyne.TextStyle{Bold: true}

	hw := canvas.NewText("Humidity "+strconv.FormatFloat(ww.humidity, 'f', 1, 64)+"%", color.Black)
	hw.TextSize = 20
	hw.TextStyle = fyne.TextStyle{Italic: true}

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
	r.temp = tw
	r.humidity = hw
	r.highTemp = htw
	r.lowTemp = ltw
	r.highHumidity = hhw
	r.lowHumidity = lhw
	r.latestUpdate = latestUpdate
	r.objects = append(r.objects, frame, header, tw, hw, htw, ltw, hhw, lhw, latestUpdate)

	r.widget.ExtendBaseWidget(ww)

	return &r
}

func (r *weatherWidgetRenderer) Destroy() {

}

func (r *weatherWidgetRenderer) Layout(size fyne.Size) {
	r.frame.Move(fyne.NewPos(0, 0))
	r.sensorName.Move(fyne.NewPos(55, 0))
	r.temp.Move(fyne.NewPos(70, 25))
	r.humidity.Move(fyne.NewPos(50, 85))
	r.highTemp.Move(fyne.NewPos(0, 40))
	r.lowTemp.Move(fyne.NewPos(0, 55))
	r.highHumidity.Move(fyne.NewPos(0, 85))
	r.lowHumidity.Move(fyne.NewPos(0, 100))
	r.latestUpdate.Move(fyne.NewPos(17, 130))
}

func (r *weatherWidgetRenderer) MinSize() fyne.Size {
	return fyne.NewSize(widgetSizeX, widgetSizeY)
}

func (r *weatherWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *weatherWidgetRenderer) Refresh() {
	r.sensorName.Text = r.widget.sensorName
	r.temp.Text = strconv.FormatFloat(r.widget.temp, 'f', 1, 64)
	r.humidity.Text = "Humidity " + strconv.FormatFloat(r.widget.humidity, 'f', 1, 64) + "%"
	r.highTemp.Text = "Hi " + strconv.FormatFloat(r.widget.highTemp, 'f', 1, 64)
	r.lowTemp.Text = "Lo " + strconv.FormatFloat(r.widget.lowTemp, 'f', 1, 64)
	r.highHumidity.Text = "Hi " + strconv.FormatFloat(r.widget.highHumidity, 'f', 1, 64) + "%"
	r.lowHumidity.Text = "Lo " + strconv.FormatFloat(r.widget.lowHumidity, 'f', 1, 64) + "%"
	r.latestUpdate.Text = r.widget.latestUpdate
}

/************************************
 * WeatherWidget Methods
 ************************************/

/* func newWeatherWidget() *weatherWidget {
	ww := weatherWidget{
		sensorKey:     "",
		sensorName:    "",
		sensorStation: "",
		temp:          0,
		humidity:      0,
		highTemp:      0,
		lowTemp:       0,
		highHumidity:  0,
		lowHumidity:   0,
		latestUpdate:  "date created",
		channel:       make(chan string, 5),
		goHandler:     wwHandler,
	}
	// ww.BaseWidget.ExtendBaseWidget(&ww)
	return &ww
} */

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
	s_LastEdit_widget.SetText(st)
	ww.latestUpdate = st
	ww.temp = activeSensors[s.Key].Temp
	ww.humidity = activeSensors[s.Key].Humidity
	ww.highHumidity = activeSensors[s.Key].HighHumidity
	ww.lowHumidity = activeSensors[s.Key].LowHumidity
	ww.highTemp = activeSensors[s.Key].HighTemp
	ww.lowTemp = activeSensors[s.Key].LowTemp
	ww.latestUpdate = activeSensors[s.Key].DataDate
	wwc := make(chan string, 5) // Buffered channel for this sensor
	ww.channel = wwc
	ww.goHandler = wwHandler
}

// sortWeatherWidgets
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

	// for _, k := range keys {
	// 	fmt.Println(k)
	// }

	return keys
}
