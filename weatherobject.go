/*
*    The container object to display the data from a given sensor
 */

package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type WeatherWidget struct {
	frame          *canvas.Rectangle
	header         *canvas.Text
	latestTemp     *canvas.Text
	latestHumidity *canvas.Text
	highTemp       *canvas.Text
	lowTemp        *canvas.Text
	latestUpdate   *canvas.Text
	dispContainer  *fyne.Container
}

func newWeatherWidget() *WeatherWidget {
	frame := &canvas.Rectangle{
		FillColor: color.RGBA{R: 173, G: 219, B: 156, A: 200},
	}
	header := canvas.NewText("Header", color.Black)
	header.Alignment = fyne.TextAlignCenter // The alignment of the text content

	newWW := &WeatherWidget{
		frame: frame,
	}
	return newWW
}
