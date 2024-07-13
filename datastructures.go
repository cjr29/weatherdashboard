/******************************************************************
 *
 * Data Structures - All constants and structures used by the
 *                   weatherdashboard
 *
 ******************************************************************/

package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	widgetSizeX   float32 = 250
	widgetSizeY   float32 = 175
	widgetPadding float32 = 5 // separation between widgets
	numColumns            = 5
	cornerRadius  float32 = 10
	strokeWidth   float32 = 2
)
const (
	// YYYY-MM-DD: 2022-03-23
	YYYYMMDD = "2006-01-02"
	// 24h hh:mm:ss: 14:23:20
	HHMMSS24h = "15:04:05"
	// 12h hh:mm:ss: 2:23:20 PM
)

//IconNameSettings fyne.ThemeIconName = "settings"

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
	Time           string  `json:"time"`          //"2024-06-11 10:33:52"
	Model          string  `json:"model"`         //"Acurite-5n1"
	Message_type   int     `json:"message_type"`  //56
	Id             int     `json:"id"`            //1997
	Channel        string  `json:"channel"`       //"A" or 1
	Sequence_num   int     `json:"sequence_num"`  //0
	Battery_ok     int     `json:"battery_ok"`    //1
	Wind_avg_mi_h  float64 `json:"wind_avg_mi_h"` //4.73634
	Temperature_F  float64 `json:"temperature_F"` //69.4
	Humidity       float64 `json:"humidity"`      // Can appear as integer or a decimal value
	Mic            string  `json:"mic"`           //"CHECKSUM"
	Station        string  `json:"station"`       // Sensor station
	SensorName     string  `json:"sensorName"`
	SensorLocation string  `json:"sensorLocation"`
}

type Sensor struct {
	Key       string `json:"Key"` // Sensor key used for map lookup
	Model     string `json:"Model"`
	Id        int    `json:"Id"`
	Channel   string `json:"Channel"`
	Station   string `json:"Station"`  // Station name, e.g., "Home" or "Barn"
	Name      string `json:"Name"`     // Name given by user
	Location  string `json:"Location"` // Optional location of sensor
	DateAdded string `json:"DateAdded"`
	LastEdit  string `json:"LastEdit"`
	// Latest sensor data received
	Temp         float64 `json:"Temp"`
	Humidity     float64 `json:"Humidity"`
	DataDate     string  `json:"Date"`
	HighTemp     float64 `json:"HighTemp"`
	LowTemp      float64 `json:"LowTemp"`
	HighHumidity float64 `json:"HighHumidity"`
	LowHumidity  float64 `json:"LowHumidity"`
	// Visibility of sensor to menus and displays
	Hide        bool `json:"Hide"`        // If set true, do not include in the list of weatherWidgets in dashboard
	HasHumidity bool `json:"HasHumidity"` // If sensor does not provide humidity, set to false
}

type Broker struct {
	Path string `json:"Path"`
	Port int    `json:"Port"`
	Uid  string `json:"Uid"`
	Pwd  string `json:"Pwd"`
}

type latestData struct {
	Temp         float64 `json:"Temp"`
	Humidity     float64 `json:"Humidity"`
	Date         string  `json:"Date"`
	HighTemp     float64 `json:"HighTemp"`
	LowTemp      float64 `json:"LowTemp"`
	HighHumidity float64 `json:"HighHumidity"`
	LowHumidity  float64 `json:"LowHumidity"`
}

type newData struct {
	key      string
	temp     float64
	humidity float64
	date     string
}

type Message struct {
	Topic   string `json:"Topic"`
	Station string `json:"Station"`
}

type DataFile struct {
	file *os.File
	path string
}

type Configuration struct {
	Brokers       []Broker
	Messages      map[int]Message
	ActiveSensors map[string]Sensor
}

type ChoicesIntKey struct {
	Key     int    // Actual map key value
	Display string // String to display in selection menu
}

type weatherWidget struct {
	widget.BaseWidget // Inherit from BaseWidget
	sensorKey         string
	sensorStation     string
	sensorName        string
	temp              float64
	humidity          float64
	highTemp          float64
	lowTemp           float64
	highHumidity      float64
	lowHumidity       float64
	latestUpdate      string
	hasHumidity       bool
	channel           chan string
	goHandler         func(key string)
	renderer          *weatherWidgetRenderer
}

type weatherWidgetRenderer struct {
	widget       *weatherWidget
	frame        *canvas.Rectangle
	station      *canvas.Text
	sensorName   *canvas.Text
	temp         *canvas.Text
	humidity     *canvas.Text
	highTemp     *canvas.Text
	lowTemp      *canvas.Text
	highHumidity *canvas.Text
	lowHumidity  *canvas.Text
	latestUpdate *canvas.Text
	objects      []fyne.CanvasObject
}

var (
	availableSensors   = make(map[string]*Sensor)        // Visible sensors table, no dups allowed
	activeSensorsMutex sync.Mutex                        // Use to lock reads and writes to the map
	activeSensors      = make(map[string]*Sensor)        // Active sensors table, indirect
	messages           = make(map[int]Message)           // Topics to be subscribed
	weatherWidgets     = make(map[string]*weatherWidget) // Key is the Sensor key associated with the WW
	dataFiles          = make(map[string]DataFile)       // Home:DataFile
	brokers            = []Broker{
		// {"path", 1883, "uid", "pwd"},
	}
)

/**********************************************************************************
 *	Data Structure Functions
 **********************************************************************************/

func (wd *WeatherData) GetSensorFromData() Sensor {
	var s Sensor
	s.Key = wd.BuildSensorKey()
	s.Station = wd.Station
	s.Model = wd.Model
	s.Id = wd.Id
	s.Channel = wd.Channel
	s.Name = ""
	s.Station = ""
	s.Location = ""
	s.DateAdded = wd.Time
	s.LastEdit = wd.Time
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

// buildSensorKey - Generate the sensor key from the WeatherData structure
func (wd *WeatherData) BuildSensorKey() string {
	key := wd.Station + ":" + wd.Model + ":" + strconv.Itoa(wd.Id) + ":" + wd.Channel
	return key
}

// Initialize sensor
func (s *Sensor) init(key string) {
	s.Key = key
	s.Station = "Station"
	s.Name = "Name"
	s.Location = "Location"
	s.Model = "Model"
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	s.DateAdded = st
	s.LastEdit = ""
	// Latest sensor data received
	s.Temp = 0
	s.Humidity = 0
	s.DataDate = ""
	s.HighTemp = 0
	s.LowTemp = 0
	s.HighHumidity = 0
	s.LowHumidity = 0
	s.HasHumidity = true
	// Visibility of sensor to menus and displays
	s.Hide = true
}

// Format Sensor string for writing
// Style = 0, newlines; Style = 1, comma-separated one line
func (s *Sensor) FormatSensor(style int) string {
	switch style {
	case 0:
		{
			str := "Sensor:\n"
			str = str + "   Station: " + s.Station + "\n"
			str = str + "   Name: " + s.Name + "\n"
			str = str + "   Location: " + s.Location + "\n"
			str = str + "   Hidden: " + fmt.Sprintf("%t", s.Hide) + "\n"
			str = str + "   Model: " + s.Model + "\n"
			str = str + "   Id: " + strconv.Itoa(s.Id) + "\n"
			str = str + "   Channel: " + s.Channel + "\n"
			str = str + "   Date Added: " + s.DateAdded + "\n"
			str = str + "   Last Edit: " + s.LastEdit + "\n"
			str = str + "Key: " + s.Key + "\n"
			return str
		}
	case 1:
		{
			str := "Station: " + s.Station + ","
			str = str + "Name: " + s.Name + ","
			str = str + "Location: " + s.Location + ","
			str = str + "   Hidden: " + fmt.Sprintf("%t", s.Hide) + ","
			str = str + "Model: " + s.Model + ","
			str = str + "Id: " + strconv.Itoa(s.Id) + ","
			str = str + "Channel: " + s.Channel + ","
			str = str + "Date Added: " + s.DateAdded + ","
			str = str + "Last Edit: " + s.LastEdit + ","
			str = str + "Key: " + s.Key
			return str
		}
	default:
		{
			return "No format specified for sensor."
		}
	}
}

// Format Message string for writing
// Style = 0, newlines; Style = 1, comma-separated one line
func (m *Message) FormatMessage(style int) string {
	switch style {
	case 0:
		{
			str := "Message:\n"
			str = str + "   Station: " + m.Station + "\n"
			str = str + "   Topic: " + m.Topic + "\n"
			return str
		}
	case 1:
		{
			str := "Station: " + m.Station + ", "
			str = str + "Topic: " + m.Topic
			return str
		}
	default:
		{
			return "No format specified for topic."
		}
	}
}

// SortActiveSensors
func sortActiveSensors() (sortedSensorKeys []string) {

	keys := make([]string, 0, len(activeSensors))

	// Build array of sensor keys to be sorted
	for key := range activeSensors {
		keys = append(keys, key)
	}

	// Sort the sensor key array using the Station and Name from each sensor
	sort.SliceStable(keys, func(i, j int) bool {
		s1 := activeSensors[keys[i]].Station + ":" + activeSensors[keys[i]].Name
		s2 := activeSensors[keys[j]].Station + ":" + activeSensors[keys[j]].Name
		return s1 < s2
	})

	// for _, k := range keys {
	// 	fmt.Println(k)
	// }

	return keys
}
