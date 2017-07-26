package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func updateFirebaseTokenHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		ApiToken      string `json:"api_token"`
		FirebaseToken string `json:"firebase_token"`
	}{"", ""}
	err := decoder.Decode(&req)

	if err != nil || firebase_token == "" || api_token == "" {
		failWithStatusCode(err, "Bad Request", w, http.StatusBadRequest)
		return
	}

	queryString := "UPDATE USERS SET firebase_id = $1 WHERE api_token = $2"
	stmt, err := db.Prepare(queryString)
	if err != nil {
		failWithStatusCode(err, "Server error", w, http.StatusInternalServerError)
		return
	}
	res, err := stmt.Exec(req.FirebaseToken, req.ApiToken)
	if err != nil {
		failWithStatusCode(err, "Server error", w, httpStatusInternalServerError)
		return
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		failWithStatusCode(err, "Bad request", w, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ID Updated")
}