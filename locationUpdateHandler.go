package main

import (
    "encoding/json"
    "net/http"
)

func locationUpdateHandler(w http.ResponseWriter, r *http.Request) {
    decoder := json.NewDecoder(r.Body)
    req := struct {
        APIToken string  `json:"api_token"`
        Lat      float64 `json:"latitude"`
        Lng      float64 `json:"longitude"`
    }{"", 0, 0}

    err := decoder.Decode(&req)

    if err != nil {
        failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
        return
    }

    LocJSON := formatGeoSON(req.Lng, req.Lat)

    queryString :=  "INSERT INTO location (u_id, help_location) " +
                    "SELECT u_id, ST_GeomFromGeoJSON($2) " +
                    "FROM users where api_token LIKE $1 " +
                    "ON CONFLICT (u_id) " +
                    "DO UPDATE SET help_location = ST_GeomFromGeoJSON($2);"

    stmt, err := db.Prepare(queryString)
    _, err = stmt.Exec(req.APIToken, LocJSON)

    if err != nil {
        failWithStatusCode(err, "failed to update location", w, http.StatusInternalServerError)
        return
    }
}
