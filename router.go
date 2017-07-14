package main

import (
	"encoding/json"
	"fmt"
	
	"math/rand"
	"net/http"
	"os"
	"strings"
	"strconv"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"time"
	"database/sql"
)

var (
	//Globals
	maxRand = 999999
	minRand = 100000

	targetNumCandidates = 4
	initialSearchRange = 1000
	maxSearchRange = 10000
	searchRangeIncrement = 1000
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)


func jsonToString() {}

func randString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

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

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	userToken := minRand + rand.Intn(maxRand-minRand)

	// Uncomment this out when we want to account send phone verification. It works.
	sendText(cmd.Number, fmt.Sprintf("Welcome to Antidose! Your verification token is %d", userToken))
	fmt.Fprintf(w, "%d", userToken)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "welcome to antidose")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var userSocketmap = make(map[string]*websocket.Conn)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}
	// frontend handshake to get user and hook them into the userMap for sockets
	_, message, err := conn.ReadMessage()
	if err != nil{
		failWithStatusCode(err, "Failed to handshake", w, http.StatusInternalServerError)
		return
	}
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

	if err != nil || newUser.FirstName == "" || newUser.LastName == "" || newUser.PhoneNumber == "" {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	newUser.PhoneNumber = strings.Replace(newUser.PhoneNumber, "-", "", -1)
	_, err = strconv.Atoi(newUser.PhoneNumber)
	if (err != nil) || (len(newUser.PhoneNumber) < 10 || len(newUser.PhoneNumber) > 16) {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	//	Check both tables for the supplied phone number
	queryString := "SELECT u_id FROM users WHERE phone_number = $1"
	stmt, err := db.Prepare(queryString)

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	var userID int
	err = stmt.QueryRow(newUser.PhoneNumber).Scan(&userID)

	queryString = "SELECT temp_u_id FROM temp_users WHERE phone_number = $1"
	stmt, err = db.Prepare(queryString)

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	var tempUserID int
	err = stmt.QueryRow(newUser.PhoneNumber).Scan(&tempUserID)

	if tempUserID == 0 {
		//	Not in temp

		token := minRand + rand.Intn(maxRand-minRand)

		//	Insert the new row into the scratch table
		queryString = "INSERT INTO temp_users(first_name, last_name, phone_number, token, init_time) VALUES($1, $2, $3, $4, current_timestamp)"
		stmt, err = db.Prepare(queryString)
		res, err := stmt.Exec(newUser.FirstName, newUser.LastName, newUser.PhoneNumber, token)

		if err != nil{
			failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
			return
		}

		numRows, err := res.RowsAffected()

		if numRows < 1 {
			failWithStatusCode(err, "Unable to insert new user", w, http.StatusInternalServerError)
			return
		}

		sendText(newUser.PhoneNumber, fmt.Sprintf("Welcome to Antidose! Your verification token is %d", token)) // Send the text containing the token

		//	Send response to the app
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Registation Recieved. Need to verify")

	} else {
		//	Is in temp

		token := minRand + rand.Intn(maxRand-minRand)

		queryString = "UPDATE temp_users SET token = $1 WHERE phone_number = $2"
		stmt, err = db.Prepare(queryString)
		res, err := stmt.Exec(token, newUser.PhoneNumber)
		if err != nil{
			failWithStatusCode(err, "Problem with update query", w, http.StatusInternalServerError)
			return
		}
		numRows, err := res.RowsAffected()
		if numRows < 1 {
			failWithStatusCode(err, "Unable to update new user", w, http.StatusConflict)
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

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

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

	if err != nil{
		failWithStatusCode(err, "Error preparing query", w, http.StatusInternalServerError)
		return
	}

	err = stmt.QueryRow(Req.PhoneNumber).Scan(&User.FirstName, &User.LastName, &User.PhoneNumber, &User.Token)

	if User.Token == "" {
		failWithStatusCode(err, "Attempting to verify user that does not exist", w, http.StatusNotFound)
		return
	}

	if Req.Token == User.Token {
		queryString = "INSERT INTO users(first_name, last_name, phone_number, current_status, api_token) VALUES($1, $2, $3, $4, $5)" +
			"ON CONFLICT (phone_number) DO UPDATE SET first_name = $1, last_name = $2, current_status = $4, api_token = $5 WHERE EXCLUDED.phone_number = $3"
		stmt, err = db.Prepare(queryString)
		if err != nil{
			failWithStatusCode(err, "Error preparing query", w, http.StatusInternalServerError)
			return
		}
		var api_token = randString(16)
		res, err := stmt.Exec(User.FirstName, User.LastName, User.PhoneNumber, "active", api_token)
		if err != nil{
			failWithStatusCode(err, "Error Inserting User", w, http.StatusInternalServerError)
			return
		}
		numRows, err := res.RowsAffected()
		if numRows < 1 {
			failWithStatusCode(err, "Error Inserting User", w, http.StatusConflict)
			return
		}

		queryString = "DELETE FROM temp_users WHERE phone_number = $1"
		stmt, err = db.Prepare(queryString)
		if err != nil{
			failWithStatusCode(err, "Error preparing query", w, http.StatusInternalServerError)
			return
		}
		res, err = stmt.Exec(Req.PhoneNumber)
		if err != nil{
			failWithStatusCode(err, "Problem deleting temp entry", w, http.StatusInternalServerError)
			return
		}
		numRows, err = res.RowsAffected()
		if numRows < 1 {
			failWithStatusCode(err, "Problem deleting temp entry", w, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{\"api_token\":\"%s\"}", api_token)

	} else {
		failWithStatusCode(err, "Token does not match", w, http.StatusUnauthorized)
	}

}


func alertHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	alert := struct {
		IMEI	int			`json:"IMEI"`
		Loc		Location	`json:"location"`
	}{0, Location{}}
	err := decoder.Decode(&alert)

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	LocJSON, err := json.Marshal(alert.Loc)

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}
	
	queryString := "INSERT INTO incidents(requester_imei, init_req_location, time_start) VALUES($1, ST_GeomFromGeoJson($2), $3)"
	stmt, err := db.Prepare(queryString)
	res, err := stmt.Exec(alert.IMEI, LocJSON, "now")
	if err != nil {
		failWithStatusCode(err, "Failed to initiate incident", w, http.StatusInternalServerError)
		return
	}

	numRows, err := res.RowsAffected()
	if numRows < 1 {
		failWithStatusCode(err, "Unable to initiate incident", w, http.StatusInternalServerError)
		return
	}

	type responder struct {
		U_id 		int
		Distance	int
	}

	var responderCandidates = make(map[int]int)
	startRadius := initialSearchRange

	for len(responderCandidates) < targetNumCandidates {
		if startRadius > maxSearchRange {
			break
		}

		queryString = "SELECT nearest_helpers($1, $2)"
		stmt, err = db.Prepare(queryString)
		rows, err := stmt.Query(LocJSON, startRadius)
		if err != nil {
			failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
		}

		for rows.Next() {
			if len(responderCandidates) < targetNumCandidates {
				tuple := ""
				err = rows.Scan(&tuple)
				if err != nil {
					failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
				}
				tuple = strings.Replace(tuple, "(", "", 1)
				tuple = strings.Replace(tuple, ")", "", 1)
				colArray := strings.Split(tuple, ",")
				U_id, err := strconv.Atoi(colArray[0])
				if err != nil {
					failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
				}
				Distance, err := strconv.Atoi(colArray[1])
				if err != nil {
					failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
				}
				responderCandidates[U_id] = Distance
			} else {
				break
			}
		}
		startRadius += searchRangeIncrement
	}
}

func locationUpdateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		Api_token string `json:"api_token"`
		Loc Location	 `json:"location"`
	}{"", Location{}}

	err := decoder.Decode(&req)

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	LocJSON, err := json.Marshal(req.Loc)

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	queryString :=  "INSERT INTO location (u_id, help_location) " +
						"SELECT u_id, ST_GeomFromGeoJSON($2) " +
						"FROM users where api_token LIKE $1 " +
					"ON CONFLICT (u_id) " +
						"DO UPDATE SET help_location = ST_GeomFromGeoJSON($2);"

	stmt, err := db.Prepare(queryString)
	_, err = stmt.Exec(req.Api_token, LocJSON)

	if err != nil{
		failWithStatusCode(err, "failed to update location", w, http.StatusInternalServerError)
		return
	}
}

func userStatusHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		Api_token string `json:"api_token"`
		Status string `json:"status"`
	}{"", ""}

	err := decoder.Decode(&req)

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	switch r.Method{
	case "GET":

		result := ""
		queryString := "SELECT current_status FROM users WHERE api_token LIKE $1;"
		stmt, _ := db.Prepare(queryString)
		err = stmt.QueryRow(req.Api_token).Scan(&result)

		if err == sql.ErrNoRows {
			failWithStatusCode(err, "could not find user", w, http.StatusNotFound)
			return
		}else if err != nil {
			failWithStatusCode(err, "failed to query database", w, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{\"user_status\":\"%s\"}", result)

	case "POST":

		queryString := "UPDATE users SET current_status = $1 WHERE api_token LIKE $2;"
		stmt, _ := db.Prepare(queryString)
		res, err := stmt.Exec(req.Status, req.Api_token)

		if err != nil {
			failWithStatusCode(err, "failed to query database", w, http.StatusInternalServerError)
			return
		}

		numRows, err := res.RowsAffected()

		if numRows < 1 {
			failWithStatusCode(err, "could not find user", w, http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
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
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/register", regHandler)
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/alert", alertHandler)
	http.HandleFunc("/location", locationUpdateHandler)
	http.HandleFunc("/userStatus", userStatusHandler)
	http.ListenAndServe(port, nil)
}
