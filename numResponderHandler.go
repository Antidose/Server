package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
)

func numResponderHandler(w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)
    req := struct {
        APIToken string `json:"api_token"`
        IncID    string `json:"inc_id"`
    }{"", ""}

    err := decoder.Decode(&req)
    if err != nil {
        failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
        return
    }

    result := ""
    queryString := "SELECT count(response_val) FROM requests WHERE response_val = TRUE AND inc_id = $1;"
    stmt, _ := db.Prepare(queryString)
    err = stmt.QueryRow(req.IncID).Scan(&result)

    if err != nil {
        failWithStatusCode(err, "Server error", w, http.StatusInternalServerError)
        return
    }

    resultInt, err := strconv.Atoi(result)

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "{\"responders\":%d}", resultInt)
}
