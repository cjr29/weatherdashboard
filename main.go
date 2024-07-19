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
	w               fyne.Window
	Client          mqtt.Client
	status          string
	Console         = container.NewVBox()
	ConsoleScroller = container.NewVScroll(Console)
	WeatherDataDisp = container.NewVBox()
	WeatherScroller = container.NewVScroll(WeatherDataDisp)
	TopicDisplay    = container.NewVBox()
	TopicScroller   = container.NewVScroll(TopicDisplay)

	th                 = weatherTheme{}
	statusContainer    *fyne.Container
	dashboardContainer *fyne.Container
	dataWindow         fyne.Window
	dashboardWindow    fyne.Window

	// Allow only one instance of any of these windows to be opened at a time
	dashFlag                 bool = false // Dashboard window flag. If true, window has been initialized.
	listActiveSensorsFlag    bool = false
	listAvailableSensorsFlag bool = false
	editActiveSensorsFlag    bool = false
	addActiveSensorsFlag     bool = false
	removeActiveSensorsFlag  bool = false
	ddflag                   bool = false // Data display flag. If true, window has been initialized.
	tflag                    bool = false // Topic display flag. If true, window has been initialized
	ewwFlag                  bool = false // Edit weather widget flag. If true, window has been initialized
	// end of window flags

	hideflag          bool      = false // Used by hideWidgetHandler DO NOT DELETE!
	logdata_flg       bool      = false
	windowScale       float64   = 1.0
	defaultWindowSize fyne.Size = fyne.NewSize(500, 300)
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
	a.Preferences().SetFloat("DEFAULT_SCALE", 0.65)
	w = a.NewWindow("Weather Dashboard")

	// w.Resize(fyne.NewSize(640, 460))
	w.Resize(defaultWindowSize)
	// windowScale = windowScale * .65
	windowScale = a.Preferences().FloatWithFallback("DEFAULT_SCALE", 1.0)
	scale := strconv.FormatFloat(windowScale, 'f', 1, 32)
	os.Setenv("FYNE_SCALE", scale)

	// weatherTheme support Light and Dark variants
	a.Settings().SetTheme(&th)
	// settings.NewSettings().LoadAppearanceScreen(w)
	w.SetMaster()

	//**********************************
	//  Prepare Menus
	//**********************************

	listActiveSensorsItem := fyne.NewMenuItem("List Active Sensors", func() {
		if !listActiveSensorsFlag {
			chooseSensors("Active Sensors", activeSensors, ListActive)
		}
	})
	listAvailableSensorsItem := fyne.NewMenuItem("List Available Sensors", func() {
		if !listAvailableSensorsFlag {
			chooseSensors("Available Sensors", availableSensors, ListAvail)
		}
	})
	addActiveSensorItem := fyne.NewMenuItem("Add Active Sensors", func() {
		if !addActiveSensorsFlag {
			chooseSensors("Select Sensors to Add", availableSensors, Add)
		}
	})
	removeActiveSensorItem := fyne.NewMenuItem("Remove Active Sensors", func() {
		if !removeActiveSensorsFlag {
			chooseSensors("Select Sensors to Remove", activeSensors, Remove)
		}
	})
	editActiveSensorItem := fyne.NewMenuItem("Edit Active Sensors", func() {
		if !editActiveSensorsFlag {
			chooseSensors("Select Sensors to Edit", activeSensors, Edit)
		}
	})

	sensorMenu := fyne.NewMenu("Sensors",
		listActiveSensorsItem,
		listAvailableSensorsItem,
		addActiveSensorItem,
		editActiveSensorItem,
		removeActiveSensorItem,
	)

	listTopicsItem := fyne.NewMenuItem("List", func() {
		if !listTopicsFlag {
			chooseTopics("Current Subscribed Topics", subscriptions, ListTopics)
		}
	})
	addTopicItem := fyne.NewMenuItem("New", addTopicHandler)
	removeTopicItem := fyne.NewMenuItem("Remove", removeTopicHandler)
	subscriptionsMenu := fyne.NewMenu("Subscriptions", listTopicsItem, addTopicItem, removeTopicItem)

	dataDisplayItem := fyne.NewMenuItem("Station Data Live Feed", scrollDataHandler)
	dashboardItem := fyne.NewMenuItem("Dashboard Widgets", dashboardHandler)
	dataMenuSeparator := fyne.NewMenuItemSeparator()
	toggleDataLoggingOnItem := fyne.NewMenuItem("Data Logging On", dataLoggingOnHandler)
	toggleDataLoggingOffItem := fyne.NewMenuItem("DataLogging Off", dataLoggingOffHandler)
	dataMenu := fyne.NewMenu("Data",
		dataDisplayItem,
		dashboardItem,
		dataMenuSeparator,
		toggleDataLoggingOnItem,
		toggleDataLoggingOffItem,
	)

	zoomPlusViewItem := fyne.NewMenuItem("Zoom +", zoomPlusHandler)
	zoomMinusViewItem := fyne.NewMenuItem("Zoom -", zoomMinusHandler)
	themeLightItem := fyne.NewMenuItem("Light", themeLightHandler)
	themeDarkItem := fyne.NewMenuItem("Dark", themeDarkHandler)
	viewMenu := fyne.NewMenu("View",
		zoomPlusViewItem,
		zoomMinusViewItem,
		themeLightItem,
		themeDarkItem,
	)

	menu := fyne.NewMainMenu(dataMenu, sensorMenu, subscriptionsMenu, viewMenu)

	w.SetMainMenu(menu)
	menu.Refresh()

	//*************************************************
	// Prepare Dashboard Containers
	//*************************************************

	ConsoleScroller.SetMinSize(fyne.NewSize(640, 400))
	WeatherScroller.SetMinSize(fyne.NewSize(700, 500))
	TopicScroller.SetMinSize(fyne.NewSize(300, 200))

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
	// Set configuration for MQTT
	//**********************************
	for _, b := range brokers {
		opts := mqtt.NewClientOptions()
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", b.Path, b.Port))
		opts.SetClientID(clientID)
		opts.SetUsername(b.Uid)
		opts.SetPassword(b.Pwd)
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
	}
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
