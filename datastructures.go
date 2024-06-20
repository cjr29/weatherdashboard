package main

import (
	"strconv"
)

type WeatherDataRaw struct {
	Time          string        `json:"time"`          //"2024-06-11 10:33:52"
	Model         string        `json:"model"`         //"Acurite-5n1"
	Message_type  int           `json:"message_type"`  //56
	Id            int           `json:"id"`            //1997
	Channel       CustomChannel `json:"channel"`       //"A" or 1
	Sequence_num  int           `json:"sequence_num"`  //0
	Battery_ok    int           `json:"battery_ok"`    //1
	Wind_avg_mi_h float64       `json:"wind_avg_mi_h"` //4.73634
	Temperature_F float64       `json:"temperature_F"` //69.4
	Humidity      float64       `json:"humidity"`      // Can appear as integer or a decimal value
	Mic           string        `json:"mic"`           //"CHECKSUM"
}

type CustomChannel struct {
	Channel string
}

func (cc *CustomChannel) channel() string {
	return cc.Channel
}

type WeatherData struct {
	Time          string  `json:"time"`          //"2024-06-11 10:33:52"
	Model         string  `json:"model"`         //"Acurite-5n1"
	Message_type  int     `json:"message_type"`  //56
	Id            int     `json:"id"`            //1997
	Channel       string  `json:"channel"`       //"A" or 1
	Sequence_num  int     `json:"sequence_num"`  //0
	Battery_ok    int     `json:"battery_ok"`    //1
	Wind_avg_mi_h float64 `json:"wind_avg_mi_h"` //4.73634
	Temperature_F float64 `json:"temperature_F"` //69.4
	Humidity      float64 `json:"humidity"`      // Can appear as integer or a decimal value
	Mic           string  `json:"mic"`           //"CHECKSUM"
	Home          string  // Sensor station
}

type Sensor struct {
	Key       string // Sensor key used for map lookup
	Model     string
	Id        int
	Channel   string
	Name      string // Name given by user
	Location  string // Optional location of sensor
	Home      string // Station name, e.g., "House" or "Barn"
	DateAdded string
}

/**********************************************************************************
 *	Data Structures
 **********************************************************************************/

func (wd *WeatherData) GetSensorFromData() Sensor {
	var s Sensor
	s.Key = wd.BuildSensorKey()
	s.Home = wd.Home
	s.Model = wd.Model
	s.Id = wd.Id
	s.Channel = wd.Channel
	s.Name = ""
	s.Location = ""
	s.DateAdded = wd.Time
	return s
}

func (wd *WeatherData) CopyWDRtoWD(from WeatherDataRaw) {
	wd.Time = from.Time
	wd.Model = from.Model
	wd.Message_type = from.Message_type
	wd.Id = from.Id
	wd.Channel = from.Channel.channel()
	wd.Sequence_num = from.Sequence_num
	wd.Battery_ok = from.Battery_ok
	wd.Wind_avg_mi_h = from.Wind_avg_mi_h
	wd.Temperature_F = from.Temperature_F
	wd.Humidity = from.Humidity
	wd.Mic = from.Mic
}

// buildSensorKey - Generate the sensor key from the WeatehrData structure
func (wd *WeatherData) BuildSensorKey() string {
	key := wd.Home + ":" + wd.Model + ":" + strconv.Itoa(wd.Id) + ":" + wd.Channel
	return key
}

// Format Sensor string for writing
// Style = 0, newlines; Style = 1, comma-separated one line
func (s *Sensor) FormatSensor(style int) string {
	switch style {
	case 0:
		{
			str := "Sensor:\n"
			str = str + "   Name: " + s.Name + "\n"
			str = str + "   Home: " + s.Home + "\n"
			str = str + "   Location: " + s.Location + "\n"
			str = str + "   Model: " + s.Model + "\n"
			str = str + "   Id: " + strconv.Itoa(s.Id) + "\n"
			str = str + "   Channel: " + s.Channel + "\n"
			str = str + "   Date Added: " + s.DateAdded + "\n"
			return str
		}
	case 1:
		{
			str := "Name: " + s.Name + ","
			str = str + "Home: " + s.Home + ","
			str = str + "Location: " + s.Location + ","
			str = str + "Model: " + s.Model + ","
			str = str + "Id: " + strconv.Itoa(s.Id) + ","
			str = str + "Channel: " + s.Channel + ","
			str = str + "Date Added: " + s.DateAdded
			return str
		}
	default:
		{
			return "No format specified for sensor."
		}
	}
}
