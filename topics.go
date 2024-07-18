/******************************************************************
 *
 * Topic functions
 *
 ******************************************************************/

package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DisplayTopics - Call this function to display a list of topics in a scrolling window
func displayTopics(t map[int]*Subscription) {
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

// LIST TOPIC
var listTopicsHandler = func() {
	displayTopics(subscriptions)
	if !tflag {
		topicWindow = a.NewWindow("Subscribed Topics")
		// Get displayable list of subscribed topics
		topicWindow.SetContent(TopicScroller)
		topicWindow.SetOnClosed(func() {
			tflag = false
		})
		tflag = true
		topicWindow.Show()
	} else {
		topicWindow.Show()
		tflag = true
	}
}

// ADD TOPIC
var addTopicHandler = func() {
	addTopicWindow := a.NewWindow("New Topic")
	inputT := widget.NewEntry()
	inputS := widget.NewEntry()
	inputT.SetPlaceHolder("Topic")
	inputS.SetPlaceHolder("Station")
	addTopicContainer := container.NewVBox(
		widget.NewLabel("Enter the full topic and its station name to which you want to subscribe."),
		inputT,
		inputS,
		widget.NewButton("Submit", func() {
			SetStatus(fmt.Sprintf("Added Topic: %s, Station: %s", inputT.Text, inputS.Text))
			// Add input text to topics[]
			var m Subscription
			m.Topic = inputT.Text
			m.Station = inputS.Text
			key := rand.Int()
			subscriptions[key] = &m
			Client.Subscribe(m.Topic, 0, messageHandler)
			SetStatus(fmt.Sprintf("Subscribed to Topic: %s", m.Topic))
			addTopicWindow.Close()
		}),
	)
	addTopicWindow.SetContent(addTopicContainer)
	addTopicWindow.Show()
}

// REMOVE TOPIC
var removeTopicHandler = func() {
	delTopicWindow := a.NewWindow("Delete Topic")
	var choices []string
	tlist := buildSubscriptionsList(subscriptions)
	for _, m := range tlist {
		choices = append(choices, m.Display)
	}
	selections := make([]string, 0)
	pickTopics := widget.NewCheckGroup(choices, func(c []string) {
		selections = append(selections, c...)
	})
	delTopicContainer := container.NewVBox(
		widget.NewLabel("Select the topic you want to remove from subscribing."),
		pickTopics,
		widget.NewButton("Submit", func() {
			for i := 0; i < len(selections); i++ {
				j, err := strconv.ParseInt(strings.Split(selections[i], ":")[0], 10, 32) // Index of choice at head of string "0: ", "1: ", ...
				check(err)
				k := int(j) // k is index into the tlist array of []ChoicesIntKey where .Key is the Message key
				// Verify message is in map before rying to delete
				if checkMessage(tlist[k].Key, subscriptions) {
					// Delete using key
					key := tlist[k].Key
					unsubscribe(Client, subscriptions[key])
					delete(subscriptions, key)
				}
			}
			delTopicWindow.Close()
		}),
	)
	delTopicWindow.SetContent(delTopicContainer)
	delTopicWindow.Show()
}

// Create subscription (topic) list
func buildSubscriptionsList(m map[int]*Subscription) []ChoicesIntKey {
	var list []ChoicesIntKey
	i := 0
	for key, message := range m {
		var c ChoicesIntKey
		c.Display = strconv.Itoa(i) + ": " + message.FormatSubscription(1)
		c.Key = key
		list = append(list, c)
		i++
	}
	return list
}
