package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/nlopes/slack"
)

var _Version = "1.0"
var _Tagline = "Blacksun Research Labs 2019"

var api *slack.Client
var g *gocui.Gui

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

// Collections of channels, users, private messages, etc.
// Useful for lookups when converting between Slack's
// internally referenced object ID and its "human-friendly"
// representation which are familiar to users.
var (
	// channelMap contains name:ID of public and private channels
	channelMap = make(map[string]string)
	// userMap contains ID:Name of users
	userMap = make(map[string]string)
)

var filterChan = ""

func stripTS(ts string) string {
	return strings.Split(ts, ".")[0]
}

func getNameForUserID(id string) (username string, err error) {
	if _, ok := userMap[id]; ok {
		return userMap[id], nil
	}
	return "", fmt.Errorf("failed to map ID %s to a username %v", id, err)
}

func getNameForChanID(id string) (chanName string, err error) {
	channel, err := api.GetChannelInfo(id)
	if err != nil {
		postToLog(g, fmt.Sprintf("failed to get channel info for channel id %s : %v", id, err))
		return "", err
	}
	return channel.Name, nil
}

// messageFormatHumanReadable normalizes messages sent from Slack's
// RTM API in preparation for displaying to the user
func messageFormatHumanReadable(msg slack.Msg) (hrMsg string) {
	var username string
	user, err := api.GetUserInfo(msg.User)
	if err != nil {
		postToLog(g, fmt.Sprintf("failed to get user info from user id %s : %v", msg.User, err))
		username = ""
	} else {
		username, err = getNameForUserID(user.ID)
		if err != nil {
			username = ""
		}
	}

	channel, err := getNameForChanID(msg.Channel)
	if err != nil {
		postToLog(g, fmt.Sprintf("failed to get channel info from channel id %s : %v", msg.Channel, err))
	}

	text := msg.Text

	ts := stripTS(msg.Timestamp)
	tsInt64, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		postToLog(g, fmt.Sprintf("failed to convert timestamp to Int64: %v", err))
	}

	ut := time.Unix(tsInt64, 0)

	if filterChan == "" {
		hrMsg = fmt.Sprintf("[%s] #%s| [%s]> %s", ut, channel, username, text)
	} else if filterChan == channel {
		hrMsg = fmt.Sprintf("[%s] [%s]> %s", ut, username, text)
	} else {
		hrMsg = ""
	}
	return hrMsg
}

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

func getConversations() ([]slack.Channel, error) {
	var channels []slack.Channel
	var params = slack.GetConversationsParameters{Types: strings.Fields("private_channel public_channel im")}

	channelsAll, _, err := api.GetConversations(&params)
	if err != nil {
		postToLog(g, err.Error())
	}
	for i, j := range channelsAll {
		if j.IsMember {
			channels = append(channels, channelsAll[i])
		} else if j.IsIM {
			channels = append(channels, channelsAll[i])
		}
	}
	return channels, nil
}

func mapUsernamesToID(g *gocui.Gui) {
	users, err := api.GetUsers()
	if err != nil {
		postToLog(g, err.Error())
	}
	for _, u := range users {
		userMap[u.ID] = u.Name
	}
}

func main() {
	api = slack.New(
		os.Getenv("SLACK_TOKEN"),
		slack.OptionLog(log.New(os.Stdout, "respite: ", log.Lshortfile|log.LstdFlags)),
	)
	rtm := api.NewRTM()
	go startTUI()
	go rtm.ManageConnection()

	mapUsernamesToID(g)

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignored

		case *slack.ConnectedEvent:
			msg := fmt.Sprintf("Connected to %s (%s.slack.com) as user %s", ev.Info.Team.Name, ev.Info.Team.Domain, ev.Info.User.Name)
			postToLog(g, msg)

		case *slack.MessageEvent:
			if ev.Msg.Upload {
				// Might handle this specially later on
				continue
			}
			msg := messageFormatHumanReadable(ev.Msg)
			postToChat(g, msg)

		case *slack.RTMError:
			postToLog(g, fmt.Sprintf("error: %s", ev.Error()))

		case *slack.InvalidAuthEvent:
			log.Panicf("Invalid credentials!")
			return

		default:
			// Ignored
		}
	}
}
