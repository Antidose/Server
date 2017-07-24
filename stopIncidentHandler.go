package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
