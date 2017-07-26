package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func pushMessageToSubscribers(incidentID string, message string) {
	for _, socket := range incidentUserSocketMap[incidentID] {
		socket.WriteMessage(websocket.TextMessage, []byte(message))
	}
}

func updateIncidentUserCount(incidentID string) {
	numResponders := string(len(incidentUserSocketMap[incidentID]))
	pushMessageToSubscribers(incidentID, numResponders)
}

func closeIncidentSockets(incidentID string) {
	for _, socket := range incidentUserSocketMap[incidentID] {
		socket.Close()
	}
	delete(incidentUserSocketMap, incidentID)
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

	if len(message.UserID) == 16 {
		// add user to incident using token
		incidentUserSocketMap[message.IncidentID] = append(incidentUserSocketMap[message.IncidentID], conn)
		updateIncidentUserCount(message.IncidentID)
	}

	if len(message.UserID) == 15 {
		// new Incident from IMEI
		users, found := incidentUserSocketMap[message.IncidentID]
		if found {
			// Another request is being opened from the same IMEI. Das bad
			fmt.Print(users)
			//Probs close all sockets and start over
			// Close all sockets
			//userSocket.Close()
		}
		incidentUserSocketMap[message.IncidentID] = []*websocket.Conn{conn}
	}
	//conn.WriteMessage(websocket.TextMessage, []byte("4"))
	fmt.Printf("Handshake from client is %+v\n", message)
	fmt.Printf("Incident Table looks like %+v\n", incidentUserSocketMap)
}
