package main

import (
	"net/http"
	"log"
	"encoding/json"
	"fmt"
)

func adminInfoHandler(w http.ResponseWriter, r *http.Request) {

	adminInfo := AdminInfo{}

	queryString := "SELECT u_id, first_name, last_name, ST_X(help_location), ST_Y(help_location) " +
				   "FROM users NATURAL JOIN location WHERE current_status = 'active';"
	rows, err := db.Query(queryString)
	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var r Responder
		err := rows.Scan(&r.Uid, &r.First, &r.Last, &r.Longitude, &r.Latitude)
		if err != nil {
			log.Fatal(err)
		}
		adminInfo.Responders = append(adminInfo.Responders, r)
	}
	rows.Close()

	queryString = "SELECT ST_X(init_req_location), ST_Y(init_req_location), time_start, time_end FROM incidents;"
	rows, err = db.Query(queryString)
	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var i Incident
		err := rows.Scan(&i.Longitude, &i.Latitude, &i.Start, &i.End)
		if err != nil {
			log.Fatal(err)
		}
		adminInfo.Incidents = append(adminInfo.Incidents, i)
	}
	rows.Close()

	b, err := json.Marshal(adminInfo)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s",b)
}