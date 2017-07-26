package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func userStatusHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	req := struct {
		APIToken string `json:"api_token"`
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
		err = stmt.QueryRow(req.APIToken).Scan(&result)

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
		res, err := stmt.Exec(req.Status, req.APIToken)

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
