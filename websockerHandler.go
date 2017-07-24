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

func updateUserSockets(incidentID string) {
	numResponders := []byte(string(len(incidentUserSocketMap[incidentID])))
	fmt.Printf("Number of responders %s\n", numResponders)
	for _, socket := range incidentUserSocketMap[incidentID] {
		// Notify all users that
		socket.WriteMessage(websocket.TextMessage, numResponders)
		// index is the index where we are
		// element is the element from someSlice for where we are
	}
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
		updateUserSockets(message.IncidentID)
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
