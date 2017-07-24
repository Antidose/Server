package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

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
