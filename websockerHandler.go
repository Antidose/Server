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
    incidentSocketCache[incidentID] = append(incidentSocketCache[incidentID], userSocket)
    updateIncidentUserCount(incidentID)
}

func pushMessageToSubscribers(incidentID string, message string) {
    for _, socket := range incidentSocketCache[incidentID] {
        socket.WriteMessage(websocket.TextMessage, []byte(message))
    }
}

func updateIncidentUserCount(incidentID string) {
    numResponders := strconv.Itoa(len(incidentSocketCache[incidentID]) - 1) // -1 Because Requester is in here
    pushMessageToSubscribers(incidentID, numResponders)
}

func closeIncidentSockets(incidentID string) {
    for _, socket := range incidentSocketCache[incidentID] {
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
            fmt.Print(incident)
            //Probs close all sockets and start over
            // Close all sockets
            //userSocket.Close()
        }
        //IncidentEventObj := &IncidentEvent{Requester: conn}
        //incidentSocketCache[message.IncidentID] = &IncidentEvent{}
        incidentSocketCache[message.IncidentID] = append(incidentSocketCache[message.IncidentID], conn)
        fmt.Printf("%+v", incidentSocketCache[message.IncidentID])
    }
    userSocketCache[message.UserID] = conn
    //conn.WriteMessage(websocket.TextMessage, []byte("4"))
    fmt.Printf("Handshake from client is %+v\n", message)
    fmt.Printf("Incident Table looks like %+v\n", incidentSocketCache)
}
