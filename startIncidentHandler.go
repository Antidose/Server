package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func startIncidentHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	alert := struct {
		IMEI     string  `json:"IMEI"`
		ApiToken string  `json:"api_token"`
		Lat      float64 `json:"latitude"`
		Lng      float64 `json:"longitude"`
	}{"", "", 0, 0}
	err := decoder.Decode(&alert)

	if err != nil || alert.IMEI == "" || alert.Lat == 0 || alert.Lng == 0 {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	LocJSON := formatGeoSON(alert.Lng, alert.Lat)

	if err != nil {
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

	incID := randString(12)
	queryString = "INSERT INTO incidents(inc_id, requester_imei, init_req_location, time_start) VALUES($1, $2, ST_GeomFromGeoJson($3), $4)"
	stmt, err = db.Prepare(queryString)
	res, err := stmt.Exec(incID, alert.IMEI, string(LocJSON), "now()")
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
		UID      int
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
			return
		}

		for rows.Next() {
			if len(responderCandidates) < targetNumCandidates {
				tuple := ""
				err = rows.Scan(&tuple)
				if err != nil {
					failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
					return
				}
				tuple = strings.Replace(tuple, "(", "", 1)
				tuple = strings.Replace(tuple, ")", "", 1)
				colArray := strings.Split(tuple, ",")
				UID, err := strconv.Atoi(colArray[0])
				if err != nil {
					failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
					return
				}
				Distance, err := strconv.Atoi(colArray[1])
				if err != nil {
					failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
					return
				}
				responderCandidates[UID] = Distance
			} else {
				break
			}
		}
		startRadius += searchRangeIncrement
	}

	var maxRange = 0

	for _, v := range responderCandidates {
		if v > maxRange {
			maxRange = v
		}
	}

	for userID := range responderCandidates {
		notification := &Notification{
			To:       "",
			Priority: "",
			Data: DataStruct{
				Notification: "",
				Lat:          0,
				Lon:          0,
				Max:          0,
				IncidentID:   "",
			},
			TimeToLive: 0,
		}

		var numOpenRequests int
		queryString := "SELECT count(*) FROM requests WHERE u_id = $1 AND time_responded IS NULL"
		stmt, err := db.Prepare(queryString)
		err = stmt.QueryRow(userID).Scan(&numOpenRequests)

		if err != nil {
			failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
			return
		}

		if numOpenRequests > 0 {
			continue
		}

		selfUID := 0
		queryString = "SELECT u_id FROM users WHERE api_token = $1"
		stmt, err = db.Prepare(queryString)
		err = stmt.QueryRow(alert.ApiToken).Scan(&selfUID)

		if err != nil {
			failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
			return
		}

		if selfUID == userID {
			continue
		}

		var lon float64
		var lat float64
		var firebaseID string
		queryString = "SELECT ST_X(help_location), ST_Y(help_location), firebase_id FROM users NATURAL JOIN location WHERE u_id = $1"
		stmt, err = db.Prepare(queryString)
		err = stmt.QueryRow(userID).Scan(&lon, &lat, &firebaseID)

		if err != nil {
			failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
			return
		}

		notification.Data.Lat = lat
		notification.Data.Lon = lon
		notification.Data.Notification = "help"
		notification.Data.Max = maxRange
		notification.To = firebaseID
		notification.Priority = "high"
		notification.Data.IncidentID = incID

		firebaseJSON, err := json.Marshal(notification)

		if err != nil {
			failWithStatusCode(err, "Server error", w, http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer([]byte(firebaseJSON)))
		if err != nil {
			failWithStatusCode(err, "Unable to create notification", w, http.StatusInternalServerError)
		}

		firebaseKey := ""
		if isHeroku {
			firebaseKey = os.Getenv("FIREBASE_AUTH")
		} else {
			firebaseKey = configuration.Firebase.Key
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", firebaseKey)
		http.DefaultClient.Do(req)

		queryString = "INSERT INTO requests(u_id, init_time, inc_id, init_help_location) VALUES($1, $2, $3, ST_GeomFRomGeoJson($4))"
		stmt, _ = db.Prepare(queryString)
		res, err := stmt.Exec(userID, "now", incID, string(LocJSON))

		if err != nil {
			failWithStatusCode(err, "Database Error", w, http.StatusInternalServerError)
			return
		}

		numRows, _ := res.RowsAffected()

		if numRows < 1 {
			failWithStatusCode(err, "Server Error", w, http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"incident_id\":\"%s\",\"num_notified\":%d, \"radius\":%d}", incID, len(responderCandidates), startRadius)
}
