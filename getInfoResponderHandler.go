package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func getInfoResponderHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	responder := struct {
		APIToken string  `json:"api_token"`
		IncID    string  `json:"inc_id"`
		Lat      float64 `json:"latitude"`
		Lng      float64 `json:"longitude"`
	}{"", "", 0, 0}

	requesterlat := ""
	requesterlng := ""

	err := decoder.Decode(&responder)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusBadRequest), w, http.StatusBadRequest)
		return
	}

	queryString := "SELECT ST_X(init_req_location), ST_Y(init_req_location) FROM incidents WHERE inc_id = $1;"
	stmt, _ := db.Prepare(queryString)
	err = stmt.QueryRow(responder.IncID).Scan(&requesterlng, &requesterlat)

	if err != nil {
		failWithStatusCode(err, "failed to query database", w, http.StatusInternalServerError)
		return
	}

	urlString := "https://api.mapbox.com/directions/v5/mapbox/driving-traffic/" +
		strconv.FormatFloat(responder.Lng, 'f', 12, 64) + "," +
		strconv.FormatFloat(responder.Lat, 'f', 12, 64) + ";" +
		requesterlng + "," +
		requesterlat + ".json" +
		"?access_token=" + getMapBoxToken()

	fmt.Printf("mapbox request:\n%s\n", urlString)

	resp, err := http.Get(urlString)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	decoder = json.NewDecoder(resp.Body)
	MapboxResponse := struct {
		Routes []MapboxRoute `json:"routes"`
	}{[]MapboxRoute{}}
	err = decoder.Decode(&MapboxResponse)

	if err != nil {
		failWithStatusCode(err, http.StatusText(http.StatusInternalServerError), w, http.StatusInternalServerError)
		return
	}

	if len(MapboxResponse.Routes) < 1 {
		failWithStatusCode(err, "No Route from mapbox", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"dist\":\"%f\", \"time\":\"%f\"}", MapboxResponse.Routes[0].Distance, MapboxResponse.Routes[0].Duration)
}
