package main

import (
	"fmt"

	"net/http"
	"os"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var (
	//Globals
	maxRand = 999999
	minRand = 100000

	targetNumCandidates  = 50
	initialSearchRange   = 1000
	maxSearchRange       = 10000
	searchRangeIncrement = 1000
	coordNames           = [9]string{"N", "NE", "E", "SE", "S", "SW", "W", "NW", "N"}
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var tokenCache = make(map[string]bool)

var userSocketCache = make(map[string]*websocket.Conn)

var incidentSocketCache = make(map[string][]*websocket.Conn)

func initRoutes() {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8088"
	} else {
		port = ":" + port
	}

	fmt.Printf("Started watching on port %s\n", port)

	fs := http.FileServer(http.Dir("admin"))
	http.Handle("/admin/", http.StripPrefix("/admin/", fs))
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/register", regHandler)
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/respondIncident", respondIncidentHandler)
	http.HandleFunc("/stopIncident", stopIncidentHandler)
	http.HandleFunc("/startIncident", startIncidentHandler)
	http.HandleFunc("/location", locationUpdateHandler)
	http.HandleFunc("/userStatus", userStatusHandler)
	http.HandleFunc("/updateFirebaseToken", updateFirebaseTokenHandler)
	http.HandleFunc("/deleteAccount", deleteAccountHandler)
	http.HandleFunc("/numResponders", numResponderHandler)
	http.HandleFunc("/getInfoResponder", getInfoResponderHandler)
	http.HandleFunc("/adminInfo", adminInfoHandler)
	http.ListenAndServe(port, nil)
}
