package main

import (
	"fmt"
	"net/http"

	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func addUserToIncident(incidentID string, userSocket *websocket.Conn) {
	IncidentEventObj := IncidentEvent{
		Requester:  incidentSocketCache[incidentID].Requester,
		Responders: append(incidentSocketCache[incidentID].Responders, userSocket),
	}
	incidentSocketCache[incidentID] = IncidentEventObj
	updateIncidentUserCount(incidentID)
}

func pushMessageToSubscribers(incidentID string, message string) {
	// Push message to responders
	for _, socket := range incidentSocketCache[incidentID].Responders {
		fmt.Printf("Pushing message %s", message)
		socket.WriteMessage(websocket.TextMessage, []byte(message))
		fmt.Printf("Pushed")
	}
	// Push message to requester
	fmt.Printf("Now Requester")
	incidentSocketCache[incidentID].Requester.WriteMessage(websocket.TextMessage, []byte(message))
	fmt.Printf("Pushed")
}

func updateIncidentUserCount(incidentID string) {
	numResponders := strconv.Itoa(len(incidentSocketCache[incidentID].Responders))
	pushMessageToSubscribers(incidentID, numResponders)
}

func closeIncidentSockets(incidentID string) {
	for _, socket := range incidentSocketCache[incidentID].Responders {
		socket.Close()
	}
	delete(incidentSocketCache, incidentID)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}
	message := SocketMessage{}
	// frontend handshake to get user and hook them into the userMap for sockets
	err = conn.ReadJSON(&message)
	if err != nil {
		failWithStatusCode(err, "Failed to handshake", w, http.StatusInternalServerError)
		return
	}

	if len(message.UserID) == 15 {
		// new Incident from IMEI
		incident, found := incidentSocketCache[message.IncidentID]
		if found {
			// Another request is being opened from the same IMEI. Das bad
			fmt.Print(incident.Responders)
			//Probs close all sockets and start over
			// Close all sockets
			//userSocket.Close()
		}
		IncidentEventObj := IncidentEvent{Requester: conn}
		incidentSocketCache[message.IncidentID] = IncidentEventObj
	}
	userSocketCache[message.UserID] = conn
	//conn.WriteMessage(websocket.TextMessage, []byte("4"))
	fmt.Printf("Handshake from client is %+v\n", message)
	fmt.Printf("Incident Table looks like %+v\n", incidentSocketCache)
}
