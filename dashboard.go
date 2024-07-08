/******************************************************************
 *
 * Dashboard functions - Used by the Fyne GUI
 *
 ******************************************************************/

package main

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

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

	dashboardWindow.CenterOnScreen()
	dashboardWindow.Show()
}

// ConsoleWrite - call this function to write a string to the scrolling console status window
func ConsoleWrite(text string) {
	Console.Add(&canvas.Text{
		Text:      text,
		Color:     th.Color(theme.ColorNameForeground, a.Settings().ThemeVariant()),
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
		Color:     th.Color(theme.ColorNameForeground, a.Settings().ThemeVariant()),
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
		Color:     th.Color(theme.ColorNameForeground, a.Settings().ThemeVariant()),
		TextSize:  12,
		TextStyle: fyne.TextStyle{Monospace: true},
	})
	for m := range t {
		msg := t[m]
		text := msg.Topic
		TopicDisplay.Add(&canvas.Text{
			Text:      text,
			Color:     th.Color(theme.ColorNameForeground, a.Settings().ThemeVariant()),
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
		Color:     th.Color(theme.ColorNameForeground, a.Settings().ThemeVariant()),
		TextSize:  12,
		TextStyle: fyne.TextStyle{Monospace: true},
	})
	for s := range m {
		sens := m[s]
		text := sens.FormatSensor(1)
		SensorDisplay.Add(&canvas.Text{
			Text:      text,
			Color:     th.Color(theme.ColorNameForeground, a.Settings().ThemeVariant()),
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
		Color:     th.Color(theme.ColorNameForeground, a.Settings().ThemeVariant()),
		TextSize:  12,
		TextStyle: fyne.TextStyle{Monospace: true},
	})
	for s := range m {
		sens := m[s]
		text := sens.FormatSensor(1)
		SensorDisplay2.Add(&canvas.Text{
			Text:      text,
			Color:     th.Color(theme.ColorNameForeground, a.Settings().ThemeVariant()),
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
