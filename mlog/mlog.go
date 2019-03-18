package mlog

import (
	"log"
	"os"
)

// Message logs messagse from slack
var message *log.Logger

// Init initializes a logger which outputs to a file
func Init(logfile string) *log.Logger {

	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	message = log.New(f, "", 0)
	return message
}
