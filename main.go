/******************************************************************
 *
 * Main - Initialization and Program Control
 *
 ******************************************************************/

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	a               fyne.App
	Client          mqtt.Client
	status          string
	Console         = container.NewVBox()
	ConsoleScroller = container.NewVScroll(Console)
	WeatherDataDisp = container.NewVBox()
	WeatherScroller = container.NewVScroll(WeatherDataDisp)
	SensorDisplay   = container.NewVBox()
	SensorScroller  = container.NewVScroll(SensorDisplay)
	SensorDisplay2  = container.NewVBox()
	SensorScroller2 = container.NewVScroll(SensorDisplay2)
	TopicDisplay    = container.NewVBox()
	TopicScroller   = container.NewVScroll(TopicDisplay)

	EditSensorContainer *fyne.Container
	statusContainer     *fyne.Container
	//buttonContainer     *fyne.Container
	dashboardContainer *fyne.Container
	dataWindow         fyne.Window
	sensorWindow       fyne.Window
	sensorWindow2      fyne.Window
	selectSensorWindow fyne.Window
	editSensorWindow   fyne.Window
	topicWindow        fyne.Window
	dashboardWindow    fyne.Window
	dashFlag           bool          = false // Dashboard window flag. If true, window has been initialized.
	swflag             bool          = false // Sensor window flag. If true, window has been initilized.
	swflag2            bool          = false // Sensor window2 flag. If true, window has been initilized.
	ddflag             bool          = false // Data display flag. If true, window has been initialized.
	tflag              bool          = false // Topic display flag. If true, window has been initialized
	hideflag           bool          = false // Used by hideWidgetHandler DO NOT DELETE!
	cancelEdit         bool          = false // Set to true to cancel current edit
	s_Station_widget   *widget.Entry = widget.NewEntry()
	s_Name_widget      *widget.Entry = widget.NewEntry()
	s_Location_widget  *widget.Entry = widget.NewEntry()
	s_Model_widget     *widget.Label = widget.NewLabel("")
	s_Id_widget        *widget.Label = widget.NewLabel("")
	s_Channel_widget   *widget.Label = widget.NewLabel("")
	s_LastEdit_widget  *widget.Label = widget.NewLabel("")
	s_Hide_widget      *widget.Check = widget.NewCheck("Check to hide sensor on weather dashboard", hideWidgetHandler)
	selectedValue      string        = ""
	logdata_flg        bool          = false
	selections         []string
	sav_Station        string // to restore in case of edit cancel
	sav_Name           string
	sav_Location       string
	sav_Hide           bool
)

/**********************************************************************************
 *	Program Control
 **********************************************************************************/

func main() {

	//**********************************
	// Set up Fyne window before trying to write to Status line!!!
	//**********************************
	a = app.NewWithID("github.com/cjr29/weatherdashboard")
	w := a.NewWindow("Weather Dashboard")
	w.Resize(fyne.NewSize(640, 460))
	w.SetMaster()
	os.Setenv("FYNE_THEME", "light")

	//**********************************
	//  Prepare Menus
	//**********************************
	listActiveSensorsItem := fyne.NewMenuItem("List Active Sensors", listSensorsHandler)
	listAvailableSensorsItem := fyne.NewMenuItem("List Available Sensors", listAvailableSensorsHandler)
	addActiveSensorItem := fyne.NewMenuItem("Add Active Sensor", addSensorHandler)
	removeActiveSensorItem := fyne.NewMenuItem("Remove Active Sensor", removeSensorHandler)
	editActiveSensorItem := fyne.NewMenuItem("Edit Active Sensor", editSensorHandler)
	sensorMenu := fyne.NewMenu("Sensors",
		listActiveSensorsItem,
		listAvailableSensorsItem,
		addActiveSensorItem,
		editActiveSensorItem,
		removeActiveSensorItem)

	listTopicsItem := fyne.NewMenuItem("List", listTopicsHandler)
	addTopicItem := fyne.NewMenuItem("New", addTopicHandler)
	removeTopicItem := fyne.NewMenuItem("Remove", removeTopicHandler)
	topicMenu := fyne.NewMenu("Topics", listTopicsItem, addTopicItem, removeTopicItem)

	dataDisplayItem := fyne.NewMenuItem("Weather Data Scroller", scrollDataHandler)
	dashboardItem := fyne.NewMenuItem("Dashboard Widgets", dashboardHandler)
	dataMenuSeparator := fyne.NewMenuItemSeparator()
	toggleDataLoggingOnItem := fyne.NewMenuItem("Data Logging On", dataLoggingOnHandler)
	toggleDataLoggingOffItem := fyne.NewMenuItem("DataLogging Off", dataLoggingOffHandler)
	weatherMenu := fyne.NewMenu("Display Data",
		dataDisplayItem,
		dashboardItem,
		dataMenuSeparator,
		toggleDataLoggingOnItem,
		toggleDataLoggingOffItem,
	)

	menu := fyne.NewMainMenu(weatherMenu, sensorMenu, topicMenu)

	w.SetMainMenu(menu)
	menu.Refresh()

	//*************************************************
	// Prepare Dashboard Containers
	//*************************************************

	// editSensor
	//
	EditSensorContainer = container.NewVBox(
		widget.NewLabel("Update Home, Name, Location, and Visibility, then press Submit to save."),
		s_Station_widget,
		s_Name_widget,
		s_Location_widget,
		s_Hide_widget,
		s_Model_widget,
		s_Id_widget,
		s_Channel_widget,
		s_LastEdit_widget,
		widget.NewButton("Submit", func() {
			cancelEdit = false
			editSensorWindow.Close()
		}),
		widget.NewButton("Cancel", func() {
			cancelEdit = true
			editSensorWindow.Close()
		}),
	)

	ConsoleScroller.SetMinSize(fyne.NewSize(640, 400))
	WeatherScroller.SetMinSize(fyne.NewSize(700, 500))
	SensorScroller.SetMinSize(fyne.NewSize(550, 500))
	SensorScroller2.SetMinSize(fyne.NewSize(550, 500))
	TopicScroller.SetMinSize(fyne.NewSize(300, 200))

	statusContainer = container.NewVBox(
		ConsoleScroller,
	)

	mainContainer := container.NewVBox(
		//buttonContainer,
		widget.NewLabel("Dashboard Status Scrolling Window"),
		statusContainer,
	)

	// Put main container in the primary window
	w.SetContent(mainContainer)

	//**********************************
	// Read configuration file
	//**********************************
	readConfig()

	// Build the widgets used by the dashboard before activating the GUI
	generateWeatherWidgets()

	//**********************************
	// Set configuration for MQTT, read from config.ini file in local directory
	//**********************************
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", brokers[0].Path, brokers[0].Port))
	opts.SetClientID(clientID)
	opts.SetUsername(brokers[0].Uid)
	opts.SetPassword(brokers[0].Pwd)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	//**********************************
	// Initialize MQTT client
	//**********************************
	Client = mqtt.NewClient(opts)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		SetStatus("Error connecting with broker. Closing program.")
		log.Println("Error connecting with broker. Closing program.")
		panic(token.Error())
	}
	t := time.Now().Local()
	st := t.Format(YYYYMMDD + " " + HHMMSS24h)
	SetStatus(fmt.Sprintf("%s : Client connected to broker %s", st, brokers[0].Path+":"+strconv.Itoa(brokers[0].Port)))

	//**********************************
	// Turn over control to the GUI
	//**********************************
	w.SetOnClosed(exitHandler)

	w.ShowAndRun()

	//*************************************************
	// NOTE! Program blocked until GUI closes
	//*************************************************
}

//*************************************************
// Support functions
//*************************************************

func check(e error) {
	if e != nil {
		log.Println(e)
		panic(e)
	}
}
