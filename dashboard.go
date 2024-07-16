/******************************************************************
 *
 * Dashboard functions - Used by the Fyne GUI
 *
 ******************************************************************/

package main

import (
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

// UpdateAll - updates selected containers
func UpdateAll() {
	if dashboardContainer != nil {
		dashboardContainer.Refresh()
	}
}
