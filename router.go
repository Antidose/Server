package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
<<<<<<< HEAD
	"os"
=======
>>>>>>> 64884c1299e2bef665ca3a8ee38b7a94cb15b555
	"strings"
	"strconv"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var (
	//Globals
	maxRand = 999999
	minRand = 100000
)

var userAuthStore = make(map[string]string)

func sendText(phoneNumber string, message string) {
	if isHeroku {
		phoneNumber = os.Getenv("TWILIO_NUMBER")
	}
	antidoseTwilio.SendSMS(configuration.Twilio.Number, phoneNumber, message, "", "")
}

func textHandler(w http.ResponseWriter, r *http.Request) {
	// Send a text to a user. Response is the code which is checked.
	decoder := json.NewDecoder(r.Body)
	cmd := struct{ Number string }{""}
	err := decoder.Decode(&cmd)
	failGracefully(err, "Failed to decode phone number")
	userToken := minRand + rand.Intn(maxRand-minRand)

	// Uncomment this out when we want to account send phone verification. It works.
	sendText(cmd.Number, fmt.Sprintf("Welcome to Antidose! Your verification token is %d", userToken))
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
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
	}{"", "", ""}
	err := decoder.Decode(&newUser)
	failOnError(err, "Failed to decode body")

	if newUser.FirstName == "" || newUser.LastName == "" || newUser.PhoneNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}

	if newUser.FirstName == "" || newUser.LastName == "" || newUser.PhoneNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}

	newUser.PhoneNumber = strings.Replace(newUser.PhoneNumber, "-", "", -1)
	_, err = strconv.Atoi(newUser.PhoneNumber)
	if (err != nil) || (len(newUser.PhoneNumber) < 10 || len(newUser.PhoneNumber) > 16) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}

	//	Check both tables for the supplied phone number
	queryString := "SELECT u_id FROM users WHERE phone_number = $1"
	stmt, err := db.Prepare(queryString)
	failOnError(err, "Failed to prepare query")
	var userID int
	err = stmt.QueryRow(newUser.PhoneNumber).Scan(&userID)

	queryString = "SELECT temp_u_id FROM temp_users WHERE phone_number = $1"
	stmt, err = db.Prepare(queryString)
	failOnError(err, "Error preparing query")
	var tempUserID int
	err = stmt.QueryRow(newUser.PhoneNumber).Scan(&tempUserID)

	if tempUserID == 0 {
		//	Not in temp

		token := minRand + rand.Intn(maxRand-minRand)

		//	Insert the new row into the scratch table
		queryString = "INSERT INTO temp_users(first_name, last_name, phone_number, token, init_time) VALUES($1, $2, $3, $4, current_timestamp)"
		stmt, err = db.Prepare(queryString)
		res, err := stmt.Exec(newUser.FirstName, newUser.LastName, newUser.PhoneNumber, token)
		failOnError(err, "Problem with insert query")
		numRows, err := res.RowsAffected()
		if numRows < 1 {
			failOnError(err, "Unable to insert new user")
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "Server Error")
			return
		}

		sendText(newUser.PhoneNumber, fmt.Sprintf("Welcome to Antidose! Your verification token is %d", token)) // Send the text containing the token

		//	Send response to the app
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Registation Success")

	} else {
		//	Is in temp

		token := minRand + rand.Intn(maxRand-minRand)

		queryString = "UPDATE temp_users SET token = $1 WHERE phone_number = $2"
		stmt, err = db.Prepare(queryString)
		res, err := stmt.Exec(token, newUser.PhoneNumber)
		failOnError(err, "Problem with update query")
		numRows, err := res.RowsAffected()
		if numRows < 1 {
			failOnError(err, "Unable to update new user")
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "Server Error")
			return
		}

		sendText(newUser.PhoneNumber, fmt.Sprintf("Welcome to Antidose! Your verification token is %d", token)) // Send the text containing the token

		if userID != 0 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "New token sent")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Account unverified")

	}

}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	Req := struct {
		Token       string `json:"token"`
		PhoneNumber string `json:"phone_number"`
	}{"", ""}
	err := decoder.Decode(&Req)

	if Req.Token == "" || Req.PhoneNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}

	User := struct {
		FirstName   string
		LastName    string
		PhoneNumber string
		Token       string
	}{"", "", "", ""}

	queryString := "SELECT first_name, last_name, phone_number, token FROM temp_users WHERE phone_number = $1"
	stmt, err := db.Prepare(queryString)
	failOnError(err, "Error preparing query")
	err = stmt.QueryRow(Req.PhoneNumber).Scan(&User.FirstName, &User.LastName, &User.PhoneNumber, &User.Token)

	if User.Token == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Attempting to verify user that does not exist")
		return
	}

	if Req.Token == User.Token {
		queryString = "INSERT INTO users(first_name, last_name, phone_number, current_status, token) VALUES($1, $2, $3, $4, $5)" +
			"ON CONFLICT (phone_number) DO UPDATE SET first_name = $1, last_name = $2, current_status = $4, token = $5 WHERE EXCLUDED.phone_number = $3"
		stmt, err = db.Prepare(queryString)
		failOnError(err, "Error preparing query")
		res, err := stmt.Exec(User.FirstName, User.LastName, User.PhoneNumber, "active", User.Token)
		failOnError(err, "Problem inserting new user")
		numRows, err := res.RowsAffected()
		if numRows < 1 {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "Error inserting new user")
			return
		}

		queryString = "DELETE FROM temp_users WHERE phone_number = $1"
		stmt, err = db.Prepare(queryString)
		failOnError(err, "Error preparing query")
		res, err = stmt.Exec(Req.PhoneNumber)
		failOnError(err, "Problem deleting temp entry")
		numRows, err = res.RowsAffected()
		if numRows < 1 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Did not remove temp entry")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "New user verified")

	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Tokens do not match")
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
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Query ran successfully!")

}

func alertHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	//TODO geojson for location
	alert := struct {
		IMEI     int    `json:"IMEI"`
		Location string `json:"locaion"`
	}{0, ""}
	err := decoder.Decode(&alert)
	failOnError(err, "Failed to decode body")

	//TODO socket

	queryString := "INSERT INTO incidents(requester_imei, init_req_location, time_start) VALUES($1, $2, $3)"
	stmt, err := db.Prepare(queryString)
	_, err = stmt.Exec(alert.IMEI, alert.Location, "now")
	failOnError(err, "Failed to insert new user")
}

func initRoutes() {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8088"
	} else {
		port = ":" + port
	}
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
