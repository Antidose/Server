package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var (
	//Globals
	maxRand = 999999
	minRand = 100000
)

var userAuthStore = make(map[string]string)

func textHandler(w http.ResponseWriter, r *http.Request) {
	// Send a text to a user. Response is the code which is checked.
	decoder := json.NewDecoder(r.Body)
	cmd := struct{ Number string }{""}
	err := decoder.Decode(&cmd)
	failGracefully(err, "Failed to decode phone number")
	userToken := minRand + rand.Intn(maxRand-minRand)

	// Uncomment this out when we want to account send phone verification. It works.
	//antidoseTwilio.SendSMS(antidoseNumber, cmd.Number, fmt.Sprintf("Welcome to Antidose! Your verification token is %d", userToken), "", "")
	fmt.Fprintf(w, "%d", userToken)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "welcome to antidose")
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	cmd := struct {
		Pass string
		User string
	}{"", ""}
	err := decoder.Decode(&cmd)
	failOnError(err, "Failed to decode request")
	pass, found := userAuthStore[cmd.User]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "User %s does not exist", cmd.User) // SET THE RIGHT STATUS CODES!
		return
	}
	if pass != cmd.Pass {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Password for User %s is incorrect", cmd.User)
		return
	}
	fmt.Fprintf(w, "User %s successfully logged in", cmd.User)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var userSocketmap = make(map[string]*websocket.Conn)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// frontend handshake to get user and hook them into the userMap for sockets
	_, message, err := conn.ReadMessage()
	failOnError(err, "Failed to handshake")
	fmt.Printf("Handshake from client is %s\n", message)
	userSocket, found := userSocketmap[string(message)]
	if found {
		userSocket.Close()
	}
	userSocketmap[string(message)] = conn
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	newUser := struct {
		FirstName string `json:"first_name"`
		LastName string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
	}{"", "", ""}
	err := decoder.Decode(&newUser)
	failOnError(err, "Failed to decode body")
	
	//	Check both tables for the supplied phone number
	queryString := "SELECT u_id FROM users WHERE phone_number = $1"
	stmt, err := db.Prepare(queryString)
	failOnError(err, "Failed to prepare query")
	var u_id int
	err = stmt.QueryRow(newUser.PhoneNumber).Scan(&u_id)

	queryString = "SELECT temp_u_id FROM temp_users WHERE phone_number = $1"
	stmt, err = db.Prepare(queryString)
	failOnError(err, "Error preparing query")
	var temp_u_id int
	err = stmt.QueryRow(newUser.PhoneNumber).Scan(&temp_u_id)
	
	if (u_id == 0 && temp_u_id == 0) {
		//	Not present in either table
		
		var characterRunes = []rune("abcdefghijklmnopqrstuvwrxyz1234567890")
		tokenArray := make([]rune, 6)
		for i := range tokenArray {
			tokenArray[i] = characterRunes[rand.Intn(len(characterRunes))]
		}
		token := string(tokenArray)

		//	Insert the new row into the scratch table
		queryString = "INSERT INTO temp_users(first_name, last_name, phone_number, token, init_time) VALUES($1, $2, $3, $4, current_timestamp)"
		stmt, err = db.Prepare(queryString)
		res, err := stmt.Exec(newUser.FirstName, newUser.LastName, newUser.PhoneNumber, token)
		failOnError(err, "Problem with insert query")
		numRows, err := res.RowsAffected()
		if numRows < 1 {
			failOnError(err, "Unable to insert new user")
		}

	} else if (u_id == 0 && temp_u_id != 0) {
		//	In users, not in scratch

	} else if (u_id != 0 && temp_u_id == 0) {
		//	Not in users, is in scratch

	} else if (u_id != 0 && temp_u_id != 0) {
		//	In scratch and users

	}

}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	user := struct {
		Token string `json:"token"`
		PhoneNumber string `json:"phone_number"`
	}{"", ""}
	err := decoder.Decode(&user)

	var serverToken string
	queryString := "SELECT token FROM temp_users WHERE phone_number = ?"
	err = db.QueryRow(queryString, user.PhoneNumber).Scan(&serverToken)

	failOnError(err, "Query execution error")
	
	//if err == sql.ErrNoRows{
		//	Send response indicating error
	//} else {
		//failOnError(err, "Problem selecting from scratch table")
	//}

	if user.Token == serverToken {
		//	Move row to users table, delete temp_users row
	} else {
		//	Send response indicating failure
	}

}

func postgresTest(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	cmd := struct{ Command string }{""}
	err := decoder.Decode(&cmd)
	fmt.Println(cmd)
	failOnError(err, "Failed to decode body")

	if cmd.Command == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad command")

		return
	}

	rows, err := db.Query(cmd.Command)
	failOnError(err, "Failed in query")
	defer rows.Close()

	numRows := 0
	for rows.Next() {
		numRows++
	}

	fmt.Fprintf(w, "Query ran successfully!")

}

func alertHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	//TODO geojson for location
	alert := struct{
		IMEI int `json:"IMEI"`
		location string `json:"locaion"`
	}{0,""}
	err := decoder.Decode(&alert)
	failOnError(err, "Failed to decode body")

	//TODO socket

	queryString := "INSERT INTO incidents(requester_imei, init_req_location, time_start) VALUES($1, $2, $3)"
	stmt, err := db.Prepare(queryString)
	_, err = stmt.Exec(alert.IMEI, alert.location, "now")
	failOnError(err, "Failed to insert new user")
}

func initRoutes() {
	port := ":8088"
	fmt.Printf("Started watching on port %s\n", port)
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/register", regHandler)
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/postgres", postgresTest)
	http.HandleFunc("/alert", alertHandler)
	http.ListenAndServe(port, nil)
}
