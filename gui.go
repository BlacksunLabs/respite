// Copyright 2019 Blacksun Research Labs

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/jroimartin/gocui"
	"github.com/nlopes/slack"
)

var _Version = "1.0"
var _Tagline = "Blacksun Research Labs 2019"

// ANSI Color Codes
const (
	// 8 Bit colors
	ClrReset   = "\u001b[0m"
	ClrBlack   = "\u001b[30m"
	ClrRed     = "\u001b[31m"
	ClrGreen   = "\u001b[32m"
	ClrYellow  = "\u001b[33m"
	ClrBlue    = "\u001b[34m"
	ClrMagenta = "\u001b[35m"
	ClrCyan    = "\u001b[36m"
	ClrWhite   = "\u001b[37m"

	// 16 Bit colors
	ClrBrightBlack   = "\u001b[30;1m"
	ClrBrightRed     = "\u001b[31;1m"
	ClrBrightGreen   = "\u001b[32;1m"
	ClrBrightYellow  = "\u001b[33;1m"
	ClrBrightBlue    = "\u001b[34;1m"
	ClrBrightMagenta = "\u001b[35;1m"
	ClrBrightCyan    = "\u001b[36;1m"
	ClrBrightWhite   = "\u001b[37;1m"

	// 8 Bit Background colors
	BgClrBlack   = "\u001b[40m"
	BgClrRed     = "\u001b[41m"
	BgClrGreen   = "\u001b[42m"
	BgClrYellow  = "\u001b[43m"
	BgClrBlue    = "\u001b[44m"
	BgClrMagenta = "\u001b[45m"
	BgClrCyan    = "\u001b[46m"
	BgClrWhite   = "\u001b[47m"

	// 16 Bit Background colors
	BgClrBrightBlack   = "\u001b[40;1m"
	BgClrBrightRed     = "\u001b[41;1m"
	BgClrBrightGreen   = "\u001b[42;1m"
	BgClrBrightYellow  = "\u001b[43;1m"
	BgClrBrightBlue    = "\u001b[44;1m"
	BgClrBrightMagenta = "\u001b[45;1m"
	BgClrBrightCyan    = "\u001b[46;1m"
	BgClrBrightWhite   = "\u001b[47;1m"
)

var (
	curMainview = make(map[string]int)
	curLogview  = make(map[string]int)
	curChanlist = make(map[string]int)
)

// setCurrentViewOnTop sets the current view to the top position
func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		postToLog(g, fmt.Sprintf("setCurrentViewOnTop : %v", err))
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func update(g *gocui.Gui) error {
	return nil
}

// postToChat print a string to the `mainview`
func postToChat(g *gocui.Gui, msg string) (err error) {
	if msg == "" && filterChan != "" {
		// Handle filtered message processing here in future
		postToLog(g, "[i] Message filtered")
		return nil
	}

	var currentView = g.CurrentView()
	if v, err := g.SetCurrentView("mainview"); err == nil {
		if v.Name() == "mainview" {
			v.Autoscroll = true
			fmt.Fprintf(v, "%v\n", msg)
			g.Update(update)
		}
		if currentView != nil {
			_, err = g.SetCurrentView(currentView.Name())
			if err != nil {
				postToLog(g, err.Error())
			}
		}
		return nil
	}
	return err
}

// Writes input to `logview`
func postToLog(g *gocui.Gui, msg string) (err error) {
	var currentView = g.CurrentView()
	if v, err := g.SetCurrentView("logview"); err == nil {
		if v.Name() == "logview" {
			v.Autoscroll = true
			_, err = fmt.Fprintf(v, "%v\n", msg)
			if err != nil {
				return err
			}
			g.Update(update)
		}
		if currentView != nil {
			_, err = g.SetCurrentView(currentView.Name())
			if err != nil {
				postToLog(g, err.Error())
			}
		}
		return nil
	}
	return err
}

// addToChanview adds Slack channels to the sidebar view "`chanview`"
//
// - Parameters:
//   - g *gocui.Gui : Pointer to the parent Gui of chanview
//   - channels []slack.Channel : Collection of slack channels to add
//
// - Returns:
//   - err error - Any errors encountered
func addToChanview(g *gocui.Gui, channels []slack.Channel) (err error) {
	// Small procedure that adds a channel to a a channel map
	// if the channel does not currently exist.
	// Allows processing channels by type if needed. Currently
	// has support for public channels, private channels, and DMs
	for k := range channels {
		if _, ok := channelMap[channels[k].ID]; !ok {
			channelMap[channels[k].ID] = channels[k].Name
		}
	}

	// Buckets for different types of Slack channels
	var (
		publicChans  []string
		privateChans []string
		imChans      []string
	)

	// Iterate through channels parameter and toss channels
	// into their appropriate bucket
	for _, name := range channels {
		if name.IsPrivate {
			privateChans = append(privateChans, name.Name)
		} else if name.IsChannel {
			publicChans = append(publicChans, name.Name)
		} else if name.IsIM {
			imChans = append(imChans, name.ID)
		}
	}

	// TODO: Add option to toggle between sorted and
	//     : unsorted channel listing
	sort.Strings(publicChans)
	sort.Strings(privateChans)
	sort.Strings(imChans)

	// Set `chanlist` as primary view with focus and
	// update its buffer with channel listing
	if v, err := g.SetCurrentView("chanlist"); err == nil {
		if v.Name() == "chanlist" {
			v.Autoscroll = true

			for c := range publicChans {
				fmt.Fprintf(v, "%s%s%s\n", ClrGreen, publicChans[c], ClrReset)
			}
			for c := range privateChans {
				fmt.Fprintf(v, "%s%s%s\n", ClrYellow, privateChans[c], ClrReset)
			}
			for c := range imChans {
				fmt.Fprintf(v, "%s%s%s\n", ClrCyan, imChans[c], ClrReset)
			}
		}
		return nil
	}
	return err
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("chanlist", 0, 0, int(0.2*float32(maxX)), maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = "Channels"
		v.Editable = false
		v.Wrap = true

		channels, err := getConversations()
		if err != nil {
			postToLog(g, err.Error())
		}
		addToChanview(g, channels)
	}

	if v, err := g.SetView("mainview", int(0.2*float32(maxX)), 0, maxX, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf("\t\tRespite v%s - %s\t\t", _Version, _Tagline)
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true
		v.Overwrite = false
	}
	if v, err := g.SetView("logview", -1, maxY-5, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Logs"
		v.Editable = false
		v.Wrap = true
		v.Autoscroll = true
		v.Overwrite = false
	}
	return nil
}

func startTUI() (err error) {
	g, err = gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panic(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, globalTab); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("chanlist", gocui.KeyArrowDown, gocui.ModNone, arrowDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("chanlist", gocui.KeyArrowUp, gocui.ModNone, arrowUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("chanlist", gocui.KeyEnter, gocui.ModNone, chanlistEnter); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	} else if err == gocui.ErrQuit {
		fmt.Printf("%s", ClrReset)
		g.Close()
		os.Exit(0)
	}
	return nil
}
