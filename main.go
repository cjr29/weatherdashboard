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
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
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

	th                 = weatherTheme{}
	statusContainer    *fyne.Container
	dashboardContainer *fyne.Container
	dataWindow         fyne.Window
	sensorWindow2      fyne.Window
	// sensorSelectWindow fyne.Window
	topicWindow              fyne.Window
	dashboardWindow          fyne.Window
	dashFlag                 bool = false // Dashboard window flag. If true, window has been initialized.
	listActiveSensorsFlag    bool = false
	listAvailableSensorsFlag bool = false // Sensor window2 flag. If true, window has been initilized.
	addActiveSensorsFlag     bool = false
	ddflag                   bool = false // Data display flag. If true, window has been initialized.
	tflag                    bool = false // Topic display flag. If true, window has been initialized
	hideflag                 bool = false // Used by hideWidgetHandler DO NOT DELETE!
	logdata_flg              bool = false
	// selections      []string
)

/**********************************************************************************
 *	Program Control
 **********************************************************************************/

func main() {

	//**********************************
	// Set up Fyne window before trying to write to Status line!!!
	//**********************************
	os.Setenv("FYNE_THEME", "light")
	a = app.NewWithID("github.com/cjr29/weatherdashboard")
	w := a.NewWindow("Weather Dashboard")
	w.Resize(fyne.NewSize(640, 460))

	// weatherTheme support Light and Dark variants
	a.Settings().SetTheme(&th)
	settings.NewSettings().LoadAppearanceScreen(w)
	w.SetMaster()

	//**********************************
	//  Prepare Menus
	//**********************************
	listActiveSensorsItem := fyne.NewMenuItem("List Active Sensors", func() {
		chooseSensors("Active Sensors", activeSensors, false) // Results in global slice resultKeys
	})
	listAvailableSensorsItem := fyne.NewMenuItem("List Available Sensors", listAvailableSensorsHandler)
	addActiveSensorItem := fyne.NewMenuItem("Add Active Sensors", addSensorsHandler)
	removeActiveSensorItem := fyne.NewMenuItem("Remove Active Sensors", removeSensorsHandler)
	editActiveSensorItem := fyne.NewMenuItem("Edit Active Sensors", func() {
		chooseSensors("Select Sensors to Edit", activeSensors, true) // Results in global slice resultKeys
	})
	newActiveSensorItem := fyne.NewMenuItem("New Sensor List", newSensorDisplayListHandler)
	sensorMenu := fyne.NewMenu("Sensors",
		listActiveSensorsItem,
		listAvailableSensorsItem,
		addActiveSensorItem,
		editActiveSensorItem,
		removeActiveSensorItem,
		newActiveSensorItem)

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

	ConsoleScroller.SetMinSize(fyne.NewSize(640, 400))
	WeatherScroller.SetMinSize(fyne.NewSize(700, 500))
	SensorScroller.SetMinSize(fyne.NewSize(550, 500))
	SensorScroller2.SetMinSize(fyne.NewSize(550, 500))
	TopicScroller.SetMinSize(fyne.NewSize(300, 200))
	sensorSelectScroller.SetMinSize(fyne.NewSize(600, 50))

	statusContainer = container.NewVBox(
		ConsoleScroller,
	)

	mainContainer := container.NewVBox(
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
