package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

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
