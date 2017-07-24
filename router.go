package main

import (
	"encoding/json"
	"fmt"

	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"database/sql"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var (
	//Globals
	maxRand = 999999
	minRand = 100000

	targetNumCandidates  = 4
	initialSearchRange   = 1000
	maxSearchRange       = 10000
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

func formatGeoSON(lat float64, lng float64) ([]byte) {
	Loc := Location{}
	Loc.Type = "Point"
	Loc.Coordinates = []float64{lat, lng}
	Loc.Crs.Type = "name"
	Loc.Crs.Properties.Name = "EPSG:4326"

	LocJSON, err := json.Marshal(Loc)

	failGracefully(err, "could not encode as geojson")

	return LocJSON
}

func getMapBoxToken() (string){
	if isHeroku {
		return os.Getenv("MAPBOX_TOKEN")
	} else {
		return configuration.Mapbox.Token
	}
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

	if err != nil {
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

var userSocketmap = make(map[string]*websocket.Conn) // Maps

var incidentUserSocketMap = make(map[string][]*websocket.Conn)

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

	if len(message.UserId) == 16 {
		// add user to incident using token
		incidentUserSocketMap[message.IncidentId] = append(incidentUserSocketMap[message.IncidentId], conn)
		updateUserSockets(message.IncidentId)
	}

	if len(message.UserId) == 15 {
		// new Incident from IMEI
		users, found := incidentUserSocketMap[message.IncidentId]
		if found {
			// Another request is being opened from the same IMEI. Das bad
			fmt.Print(users)
			//Probs close all sockets and start over
			// Close all sockets
			//userSocket.Close()
		}
		incidentUserSocketMap[message.IncidentId] = []*websocket.Conn{conn}
	}
	//conn.WriteMessage(websocket.TextMessage, []byte("4"))
	fmt.Printf("Handshake from client is %+v\n", message)
	fmt.Printf("Incident Table looks like %+v\n", incidentUserSocketMap)

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

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	var userID int
	err = stmt.QueryRow(newUser.PhoneNumber).Scan(&userID)

	queryString = "SELECT temp_u_id FROM temp_users WHERE phone_number = $1"
	stmt, err = db.Prepare(queryString)

	if err != nil {
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

		if err != nil {
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
		if err != nil {
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
		FirebaseId  string `json:"firebase_id"`
	}{"", "", ""}
	err := decoder.Decode(&Req)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	if Req.Token == "" || Req.PhoneNumber == "" || Req.FirebaseId == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request")
		return
	}

	Req.PhoneNumber = strings.Replace(Req.PhoneNumber, "-", "", -1)
	_, err = strconv.Atoi(Req.PhoneNumber)
	if (err != nil) || (len(Req.PhoneNumber) < 10 || len(Req.PhoneNumber) > 16) {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
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

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	err = stmt.QueryRow(Req.PhoneNumber).Scan(&User.FirstName, &User.LastName, &User.PhoneNumber, &User.Token)

	if User.Token == "" {
		failWithStatusCode(err, "Attempting to verify user that does not exist", w, http.StatusNotFound)
		return
	}

	if Req.Token == User.Token {
		queryString = "INSERT INTO users(first_name, last_name, phone_number, current_status, api_token, firebase_id) VALUES($1, $2, $3, $4, $5, $6)" +
			"ON CONFLICT (phone_number) DO UPDATE SET first_name = $1, last_name = $2, current_status = $4, api_token = $5, firebase_id = $6 WHERE EXCLUDED.phone_number = $3"
		stmt, err = db.Prepare(queryString)
		if err != nil {
			failWithStatusCode(err, "Error preparing query", w, http.StatusInternalServerError)
			return
		}
		var apiToken = randString(16)
		res, err := stmt.Exec(User.FirstName, User.LastName, User.PhoneNumber, "active", apiToken, Req.FirebaseId)
		if err != nil {
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
		if err != nil {
			failWithStatusCode(err, "Error preparing query", w, http.StatusInternalServerError)
			return
		}
		res, err = stmt.Exec(Req.PhoneNumber)
		if err != nil {
			failWithStatusCode(err, "Problem deleting temp entry", w, http.StatusInternalServerError)
			return
		}
		numRows, err = res.RowsAffected()
		if numRows < 1 {
			failWithStatusCode(err, "Problem deleting temp entry", w, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{\"api_token\":\"%s\"}", apiToken)

	} else {
		failWithStatusCode(err, "Token does not match", w, http.StatusUnauthorized)
	}

}

func numResponderHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		ApiToken string	`json:"api_token"`
		IncId    string	`json:"inc_id"`

	}{"", ""}

	err := decoder.Decode(&req)
	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	result := ""
	queryString := "SELECT count(response_val) FROM requests WHERE response_val = TRUE AND inc_id = $1;"
	stmt, _ := db.Prepare(queryString)
	err = stmt.QueryRow(req.IncId).Scan(&result)

	if err != nil {
		failWithStatusCode(err, "Server error", w, http.StatusInternalServerError)
		return
	}

	resultInt, err := strconv.Atoi(result)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"responders\":%d}", resultInt)
}

func startIncidentHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	alert := struct {
		IMEI	int			`json:"IMEI"`
		Lat		float64		`json:"latitude"`
		Lng		float64		`json:"longitude"`
	}{0, 0, 0}
	err := decoder.Decode(&alert)


	if err != nil || alert.IMEI == 0 || alert.Lat == 0 || alert.Lng == 0 {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	LocJSON := formatGeoSON(alert.Lat, alert.Lng)

	if err != nil{
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	var count int
	queryString := "SELECT count(*) FROM incidents WHERE requester_imei = $1 AND time_end IS NULL"
	stmt, err := db.Prepare(queryString)
	err = stmt.QueryRow(alert.IMEI).Scan(&count)
	if err != nil {
		failWithStatusCode(err, "Internal Error", w, http.StatusInternalServerError)
		return
	}

	if count > 0 {
		failWithStatusCode(err, "Requestor already has open incident", w, http.StatusBadRequest)
		return
	}

	incId := randString(12)
	queryString = "INSERT INTO incidents(inc_id, requester_imei, init_req_location, time_start) VALUES($1, $2, ST_GeomFromGeoJson($3), $4)"
	stmt, err = db.Prepare(queryString)
	res, err := stmt.Exec(incId, alert.IMEI, string(LocJSON), "now()")
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
		Uid      int
		Distance int
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
				Uid, err := strconv.Atoi(colArray[0])
				if err != nil {
					failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
				}
				Distance, err := strconv.Atoi(colArray[1])
				if err != nil {
					failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
				}
				responderCandidates[Uid] = Distance
			} else {
				break
			}
		}
		startRadius += searchRangeIncrement
	}
}

func respondIncidentHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		ApiToken string	`json:"api_token"`
		IncId    string	`json:"inc_id"`
		HasKit   bool	`json:"has_kit"`
		IsGoing  bool	`json:"is_going"`
	}{"","", false, false}

	err := decoder.Decode(&req)

	if err != nil || req.ApiToken == "" {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
	}

	queryString := "UPDATE requests SET time_responded = $1, response_val = $2, has_kit = $3 WHERE inc_id = $4;"
	stmt, err := db.Prepare(queryString)
	res, err := stmt.Exec("now", req.IsGoing, req.HasKit, req.ApiToken, req.IncId)

	numRows, _ := res.RowsAffected()

	if err != nil || numRows < 1 {
		failWithStatusCode(err, "Failed to process response", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if req.IsGoing == false {
		fmt.Fprintf(w, "Response processed")
		return
	}

	if req.IsGoing {
		incidentLat := 0
		incidentLng := 0

		queryString = "SELECT ST_X(init_req_location), ST_Y(init_req_location) FROM incidents WHERE inc_id = $1;"
		stmt, _ = db.Prepare(queryString)
		err = stmt.QueryRow(req.IncId).Scan(&incidentLng, &incidentLat)
		fmt.Fprintf(w, "{\"latitude\":\"%f\", \"longitude\":\"%f\"}", incidentLat, incidentLng)
	}

	if err != nil {
		failWithStatusCode(err, "failed to query database", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func stopIncidentHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		IMEI       string `json:"IMEI"`
		IsResolved bool   `json:"is_resolved"`
	}{"", false}
	err := decoder.Decode(&req)

	if err != nil || req.IMEI == "" {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	queryString := "UPDATE incidents SET time_end = 'now', is_resolved = $1 WHERE time_end IS NULL AND requester_imei = $2"
	stmt, err := db.Prepare(queryString)
	if err != nil {
		failWithStatusCode(err, "Server error", w, http.StatusInternalServerError)
	}
	res, err := stmt.Exec(req.IsResolved, req.IMEI)

	if err != nil {
		failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
	}

	numRows, err := res.RowsAffected()

	if numRows < 1 || err != nil {
		failWithStatusCode(err, "Failed to close incident", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Incident ended")
}

func locationUpdateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		ApiToken string		`json:"api_token"`
		Lat      float64		`json:"latitude"`
		Lng      float64		`json:"longitude"`
	}{"", 0, 0}

	err := decoder.Decode(&req)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	LocJSON := formatGeoSON(req.Lat, req.Lng)

	queryString := "INSERT INTO location (u_id, help_location) " +
		"SELECT u_id, ST_GeomFromGeoJSON($2) " +
		"FROM users where api_token LIKE $1 " +
		"ON CONFLICT (u_id) " +
		"DO UPDATE SET help_location = ST_GeomFromGeoJSON($2);"

	stmt, err := db.Prepare(queryString)
	_, err = stmt.Exec(req.ApiToken, LocJSON)

	if err != nil {
		failWithStatusCode(err, "failed to update location", w, http.StatusInternalServerError)
		return
	}
}

func getInfoResponderHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	responder := struct {
		ApiToken string 		`json:"api_token"`
		IncId    string 		`json:"inc_id"`
		Lat      float64		`json:"latitude"`
		Lng      float64		`json:"longitude"`
	}{"", "", 0, 0}

	requesterlat := ""
	requesterlng := ""

	err := decoder.Decode(&responder)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	queryString := "SELECT ST_X(init_req_location), ST_Y(init_req_location) FROM incidents WHERE inc_id = $1;"
	stmt, _ := db.Prepare(queryString)
	err = stmt.QueryRow(responder.IncId).Scan(&requesterlng, &requesterlat)

	if err != nil {
		failWithStatusCode(err, "failed to query database", w, http.StatusInternalServerError)
		return
	}

	urlString := "https://api.mapbox.com/directions/v5/mapbox/driving-traffic/" +
		strconv.FormatFloat(responder.Lng, 'f', 12, 64) + "," +
		strconv.FormatFloat(responder.Lat, 'f', 12, 64) + ";" +
		requesterlng + "," +
		requesterlat + ".json" +
		"?access_token=" + getMapBoxToken()

	fmt.Printf("mapbox request:\n%s\n", urlString)

	resp, err := http.Get(urlString)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	decoder = json.NewDecoder(resp.Body)
	MapboxResponse := struct {
		Routes []MapboxRoute `json:"routes"`
	}{[]MapboxRoute{}}
	err = decoder.Decode(&MapboxResponse)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	if len(MapboxResponse.Routes) < 1 {
		failWithStatusCode(err, "No Route from mapbox", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"dist\":\"%f\", \"time\":\"%f\"}", MapboxResponse.Routes[0].Distance, MapboxResponse.Routes[0].Duration)
}

func userStatusHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		ApiToken string `json:"api_token"`
		Status   string `json:"status"`
	}{"", ""}

	err := decoder.Decode(&req)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":

		result := ""
		queryString := "SELECT current_status FROM users WHERE api_token LIKE $1;"
		stmt, _ := db.Prepare(queryString)
		err = stmt.QueryRow(req.ApiToken).Scan(&result)

		if err == sql.ErrNoRows {
			failWithStatusCode(err, "could not find user", w, http.StatusNotFound)
			return
		} else if err != nil {
			failWithStatusCode(err, "failed to query database", w, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{\"user_status\":\"%s\"}", result)

	case "POST":

		queryString := "UPDATE users SET current_status = $1 WHERE api_token LIKE $2;"
		stmt, _ := db.Prepare(queryString)
		res, err := stmt.Exec(req.Status, req.ApiToken)

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

func deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		ApiToken string `json:"api_token"`
	}{""}

	err := decoder.Decode(&req)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	queryString := "DELETE FROM users WHERE api_token = $1"
	stmt, _ := db.Prepare(queryString)
	res, err := stmt.Exec(req.ApiToken)

	if err != nil {
		failWithStatusCode(err, "Database error", w, http.StatusInternalServerError)
	}

	numRows, err := res.RowsAffected()

	if numRows < 1 {
		failWithStatusCode(err, "Could not delete user", w, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
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
	http.HandleFunc("/respondIncident", respondIncidentHandler)
	http.HandleFunc("/stopIncident", stopIncidentHandler)
	http.HandleFunc("/startIncident", startIncidentHandler)
	http.HandleFunc("/location", locationUpdateHandler)
	http.HandleFunc("/userStatus", userStatusHandler)
	http.HandleFunc("/deleteAccount", deleteAccountHandler)
	http.HandleFunc("/numResponders", numResponderHandler)
	http.HandleFunc("/getInfoResponder", getInfoResponderHandler)
	http.ListenAndServe(port, nil)
}
