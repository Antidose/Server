package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func respondIncidentHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		APIToken string `json:"api_token"`
		IncID    string `json:"inc_id"`
		HasKit   bool   `json:"has_kit"`
		IsGoing  bool   `json:"is_going"`
	}{"", "", false, false}

	err := decoder.Decode(&req)

	if err != nil || req.APIToken == "" {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
	}

	queryString := "UPDATE requests SET time_responded = $1, response_val = $2, has_kit = $3 WHERE inc_id = $4;"
	stmt, err := db.Prepare(queryString)
	res, err := stmt.Exec("now", req.IsGoing, req.HasKit, req.APIToken, req.IncID)

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
		incidentLat := 0.00
		incidentLng := 0.00

		queryString = "SELECT ST_X(init_req_location), ST_Y(init_req_location) FROM incidents WHERE inc_id = $1;"
		stmt, _ = db.Prepare(queryString)
		err = stmt.QueryRow(req.IncID).Scan(&incidentLng, &incidentLat)
		fmt.Fprintf(w, "{\"latitude\":\"%f\", \"longitude\":\"%f\"}", incidentLat, incidentLng)
	}

	if err != nil {
		failWithStatusCode(err, "failed to query database", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}