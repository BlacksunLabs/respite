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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BlacksunLabs/respite/mlog"
	"github.com/jroimartin/gocui"
	"github.com/nlopes/slack"
)

var logger *log.Logger
var api *slack.Client
var g *gocui.Gui

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

	team, err := api.GetTeamInfo()
	if err != nil {
		postToLog(g, err.Error())
	}
	logger.Printf("[%s] %s.slack.com #%s| [%s]> %s\n", ut, team.Domain, channel, username, text)

	if filterChan == "" {
		hrMsg = fmt.Sprintf("[%s] #%s| [%s]> %s", ut, channel, username, text)
	} else if filterChan == channel {
		hrMsg = fmt.Sprintf("[%s] [%s]> %s", ut, username, text)
	} else {
		hrMsg = ""
	}
	return hrMsg
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

var flagLogfile string

func init() {
	flag.StringVar(&flagLogfile, "log", "", "path to log file")
	flag.Parse()
}

func main() {
	if flagLogfile != "" {
		logger = mlog.Init(flagLogfile)
	} else {
		logger = log.New(ioutil.Discard, "", 0)
		log.Println("Logging disabled")
	}
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
