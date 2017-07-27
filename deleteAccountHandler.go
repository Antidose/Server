package main

import (
	"encoding/json"
	"net/http"
)

func deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		APIToken string `json:"api_token"`
	}{""}

	err := decoder.Decode(&req)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	queryString := "UPDATE users SET current_status = 'deleted' WHERE api_token = $1"
	stmt, _ := db.Prepare(queryString)
	res, err := stmt.Exec(req.APIToken)

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
