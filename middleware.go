package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		req := struct {
			APIToken string `json:"api_token"`
		}{""}

		err := decoder.Decode(&req)

		if err != nil {
			failWithStatusCode(err, "Token Error", w, http.StatusForbidden)
		}

		if userSocketmap[req.APIToken] == nil {
			// Have to get from DB
			result := ""
			queryString := "SELECT current_status FROM users WHERE api_token LIKE $1;"
			stmt, _ := db.Prepare(queryString)
			err = stmt.QueryRow(req.APIToken).Scan(&result)

			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "Token Validation failed")
			}
		}
		next.ServeHTTP(w, r)
	})
}
