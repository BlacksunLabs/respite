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

	"github.com/jroimartin/gocui"
)

func quit(g *gocui.Gui, v *gocui.View) error {
	postToLog(g, fmt.Sprintf("Quitting ..."))
	return gocui.ErrQuit
}

func globalTab(g *gocui.Gui, v *gocui.View) error {
	filterChan = ""
	toggleOmniscient(g)
	return nil
}

func arrowDown(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(0, 1, false)
	return nil
}

func arrowUp(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(0, -1, false)
	return nil
}

func chanlistEnter(g *gocui.Gui, v *gocui.View) error {
	cursorX, cursorY := v.Cursor()
	chanWanted, err := v.Word(cursorX, cursorY)
	if err != nil {
		postToLog(g, err.Error())
		return err
	}
	filterChan = chanWanted
	toggleOmniscient(g)
	return nil
}

func toggleOmniscient(g *gocui.Gui) {
	mainview, err := g.View("mainview")
	if err != nil {
		postToLog(g, err.Error())
	}
	var title string
	if filterChan == "" {
		postToLog(g, "Disabling channel filter")
		title = fmt.Sprintf("\t\tRespite v%s - %s\t\t", _Version, _Tagline)
	} else {
		title = fmt.Sprintf("\t\tRespite v%s - %s\t\t(%s)\t", _Version, _Tagline, filterChan)
		postToLog(g, fmt.Sprintf("Enabling channel filter on %s", filterChan))
	}
	mainview.Title = fmt.Sprintf("%s", title)
}
