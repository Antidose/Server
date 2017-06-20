package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sfreiberg/gotwilio"
)

// Globals
var (
	done           = make(chan struct{})
	antidoseTwilio = gotwilio.NewTwilioClient("ACbf63e163b500ca960f648e79e24e9100", "8b5045f6891b48a3d858ba5e89f0a4d4")
	antidoseNumber = "+17784004161"
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

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	initRoutes()
	<-done
}
