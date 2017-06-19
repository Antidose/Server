package main

import "fmt"

// Globals
var (
	done = make(chan struct{})
)

func main() {
	initRoutes()
	<-done
	fmt.Println("Hello World")
}
