package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/sfreiberg/gotwilio"
)

// Configuration : Core config structure
type Configuration struct {
	Twilio TwilioKey
}

// TwilioKey : Config strucuture for Twilio
type TwilioKey struct {
	Sid    string
	Token  string
	Number string
}

// Globals
var (
	done           = make(chan struct{})
	configuration  = loadConfig()
	antidoseTwilio = loadTwilio()
)

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
		panic(err)
	}
}

func failGracefully(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
	}
}

func loadConfig() Configuration {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}

func loadTwilio() *gotwilio.Twilio {
	return gotwilio.NewTwilioClient(configuration.Twilio.Sid, configuration.Twilio.Token)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	initRoutes()
	<-done
}
