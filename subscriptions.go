/******************************************************************
 *
 * subscriptions - Manages the Subscriptions map of topics
 *		Includes the custom widget to display a single topic
 *		and the scrolling container to hold them.
 *
 ******************************************************************/

package main

import (
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	topicDisplayWidgetSizeX   float32 = 800
	topicDisplayWidgetSizeY   float32 = 40
	topicDisplayWidgetPadding float32 = 0 // separation between widgets
	topicDisplayCornerRadius  float32 = 5
	topicDisplayStrokeWidth   float32 = 1
)

var (
	topicWindow                       fyne.Window
	topicDisplay                      = container.NewVBox()
	topicScroller                     = container.NewVScroll(TopicDisplay)
	resultKeysTopics                  []int // storage for topics keys selected by user
	topicDisplayWidgetBackgroundColor       = color.RGBA{R: 214, G: 240, B: 246, A: 255}
	topicDisplayWidgetForegroundColor       = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // Black
	topicDisplayWidgetFrameColor            = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF} // Black
	showTopicsCheckBoxesFlag          bool  = false
	// Allow only one instance of any of these windows to be opened at a time
	listTopicsFlag   bool = false
	editTopicsFlag   bool = false
	addTopicsFlag    bool = false
	removeTopicsFlag bool = false
)

type topicDisplayWidget struct {
	widget.BaseWidget // Inherit from BaseWidget
	key               int
	topic             string
	station           string
	check             bool
	renderer          *topicDisplayWidgetRenderer
	sync.Mutex
}

type topicDisplayWidgetRenderer struct {
	widget   *topicDisplayWidget
	frame    *canvas.Rectangle
	checkbox *widget.Check
	station  *canvas.Text
	topic    *canvas.Text
	objects  []fyne.CanvasObject
}

/******************************************
 * Renderer Methods
 ******************************************/

func newTopicDisplayWidgetRenderer(tdw *topicDisplayWidget) fyne.WidgetRenderer {
	r := topicDisplayWidgetRenderer{}
	tdw.renderer = &r

	frame := &canvas.Rectangle{
		FillColor:   topicDisplayWidgetBackgroundColor,
		StrokeColor: topicDisplayWidgetFrameColor,
		StrokeWidth: topicDisplayStrokeWidth,
	}
	frame.SetMinSize(fyne.NewSize(topicDisplayWidgetSizeX, topicDisplayWidgetSizeY))
	frame.Resize(fyne.NewSize(topicDisplayWidgetSizeX, topicDisplayWidgetSizeY))
	frame.CornerRadius = topicDisplayCornerRadius

	/******************************
	* Add a Check Box widget
	******************************/
	check := widget.NewCheck("", func(b bool) {
		r.widget.check = b
	})

	st := canvas.NewText(tdw.station, topicDisplayWidgetForegroundColor)
	st.TextSize = 14
	st.TextStyle = fyne.TextStyle{Bold: true}

	tn := canvas.NewText(tdw.topic, topicDisplayWidgetForegroundColor)
	tn.TextSize = 14

	r.widget = tdw
	r.frame = frame
	r.checkbox = check
	r.station = st
	r.topic = tn
	if showCheckBoxesFlag {
		r.objects = append(r.objects, frame, check, st, tn)
	} else {
		r.objects = append(r.objects, frame, st, tn)
	}

	r.widget.ExtendBaseWidget(tdw)

	return &r
}

func (r *topicDisplayWidgetRenderer) Destroy() {
}

func (r *topicDisplayWidgetRenderer) Layout(size fyne.Size) {
	r.frame.Move(fyne.NewPos(0, 0))
	r.checkbox.Resize(fyne.NewSize(20, 20))
	r.checkbox.Move(fyne.NewPos(5, 10))
	ypos := ((topicDisplayWidgetSizeY - r.topic.TextSize) / 2) + 5
	r.station.Move(fyne.NewPos(40, ypos))
	r.topic.Move(fyne.NewPos(120, ypos))
}

func (r *topicDisplayWidgetRenderer) MinSize() fyne.Size {
	return fyne.NewSize(topicDisplayWidgetSizeX, topicDisplayWidgetSizeY)
}

func (r topicDisplayWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *topicDisplayWidgetRenderer) Refresh() {
	r.frame.Resize(fyne.NewSize(sensorDisplayWidgetSizeX, sensorDisplayWidgetSizeY)) // This is critical or frame won't appear
	r.frame.Show()
	r.checkbox.Resize(fyne.NewSize(20, 20))
	r.checkbox.Move(fyne.NewPos(5, 10))
	r.checkbox.Show()
	r.topic.Text = r.widget.topic
	r.station.Text = r.widget.station
}

/************************************
 * sensorDisplayWidget Methods
 ************************************/

func (tdw *topicDisplayWidget) CreateRenderer() fyne.WidgetRenderer {
	r := newTopicDisplayWidgetRenderer(tdw)
	return r
}

func (tdw *topicDisplayWidget) Hide() {
}

func (tdw *topicDisplayWidget) Checked() bool {
	return tdw.check
}

func (tdw *topicDisplayWidget) Refresh() {
	if tdw == nil {
		return
	}
	tdw.BaseWidget.Refresh()
	if tdw.renderer != nil {
		tdw.renderer.Refresh()
	}
}

// Initialize fields of a sensorDisplayWidget using data from Sensor
func (tdw *topicDisplayWidget) init(s *Subscription) {
	tdw.key = s.Key
	tdw.topic = s.Topic
	tdw.station = s.Station
	tdw.check = false
}

/************************************
 * Topic Selection Methods
 ************************************/

// chooseTopics
//
//	if showChecks, display check boxes
//	Use Action type to determine which activity to initiate after selections are made
func chooseTopics(title string, subscriptions map[int]*Subscription, action Action) {
	var topicSelectDisp = container.NewVBox()
	var topicSelectScroller = container.NewVScroll(topicSelectDisp)
	topicSelectScroller.SetMinSize(fyne.NewSize(600, 50))

	// Only display one instance of a window at a time
	if listTopicsFlag && (action == ListTopics) {
		return
	}
	if editTopicsFlag && (action == EditTopics) {
		return
	}
	if addTopicsFlag && (action == AddTopics) {
		return
	}
	if removeTopicsFlag && (action == RemoveTopics) {
		return
	}

	// Depending on action, turn check boxes on or off by setting the global flag
	switch action {
	case ListTopics:
		listTopicsFlag = true
		showTopicsCheckBoxesFlag = false // Set the global flag for widgets to see
	case EditTopics:
		editActiveSensorsFlag = true
		showTopicsCheckBoxesFlag = true
	case AddTopics:
		addActiveSensorsFlag = true
		showTopicsCheckBoxesFlag = true
	case RemoveTopics:
		removeActiveSensorsFlag = true
		showTopicsCheckBoxesFlag = true
	default:
		showTopicsCheckBoxesFlag = false
	}

	// Prepare the containers and widgets to display the selection list
	topicSelectWindow := a.NewWindow(title)
	topicSelectWindow.SetOnClosed(func() {
		clearSelectedFlag(action)
	})

	topicSelectWindow.Resize(fyne.NewSize(sensorDisplayWidgetSizeX+20, sensorDisplayWidgetSizeY*10))

	// Make a slice with capacity of all topics passed in the argument
	resultKeysTopics = make([]int, 0, len(subscriptions))
	var choices []*topicDisplayWidget
	for key, top := range subscriptions {
		temp := topicDisplayWidget{}
		temp.Lock()
		temp.init(top)
		a := subscriptions[key]
		temp.station = a.Station
		temp.topic = a.Topic
		topicSelectDisp.Add(&temp)
		choices = append(choices, &temp)
		temp.Unlock()
	}
	topicSelectScroller.ScrollToBottom()
	topicSelectDisp.Refresh()

	var buttonContainer *fyne.Container
	// Show Submit only if a choice is required
	if showTopicsCheckBoxesFlag {
		buttonContainer = container.NewHBox(
			widget.NewButton("Submit", func() {
				for _, value := range choices {
					value.Lock()
					if value.check {
						resultKeysTopics = append(resultKeysTopics, value.key)
					}
					value.Unlock()
				}
				clearSelectedFlag(action)
				topicSelectWindow.Close()
			}),
			widget.NewButton("Cancel", func() {
				clearSelectedFlag(action)
				topicSelectWindow.Close()
			}),
		)
	} else {
		buttonContainer = container.NewHBox(
			widget.NewButton("Cancel", func() {
				topicSelectWindow.Close()
			}),
		)
	}
	overallContainer := container.NewBorder(
		buttonContainer,     // top
		nil,                 // bottom
		nil,                 // left
		nil,                 // right
		topicSelectScroller, // middle
	)
	topicSelectWindow.SetContent(overallContainer)
	topicSelectWindow.Show()
}
