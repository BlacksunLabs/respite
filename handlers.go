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
