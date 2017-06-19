package main

import "fmt"

// Globals
var (
	done = make(chan struct{})
)

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
		panic(err)
	}
}

func main() {
	initRoutes()
	<-done
}
