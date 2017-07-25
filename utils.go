package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func round(val float64) int {
	intVal := int(val)
	if val-float64(intVal) > 0.5 {
		return intVal + 1
	}
	return intVal
}

func getBearing(lat1 float64, lon1 float64, lat2 float64, lon2 float64) string {

	rads := math.Atan2((lon1 - lon2), (lat1 - lat2))

	compass := rads * (180 / math.Pi)

	coordIndex := round(compass / 45)

	if coordIndex < 0 {
		coordIndex += 8
	}

	return coordNames[coordIndex]
}

func jsonToString() {}

func randString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func formatGeoSON(lng float64, lat float64) []byte {
	Loc := Location{}
	Loc.Type = "Point"
	Loc.Coordinates = []float64{lng, lat}
	Loc.Crs.Type = "name"
	Loc.Crs.Properties.Name = "EPSG:4326"

	LocJSON, err := json.Marshal(Loc)

	failGracefully(err, "could not encode as geojson")

	return LocJSON
}

func sendText(phoneNumber string, message string) {
	myPhoneNumber := configuration.Twilio.Number
	if isHeroku {
		myPhoneNumber = os.Getenv("TWILIO_NUMBER")
	}
	antidoseTwilio.SendSMS(myPhoneNumber, phoneNumber, message, "", "")
}

func getMapBoxToken() string {
	if isHeroku {
		return os.Getenv("MAPBOX_TOKEN")
	}
	return configuration.Mapbox.Token
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
		panic(err)
	}
}

func failWithStatusCode(err error, msg string, w http.ResponseWriter, statusCode int) {
	failGracefully(err, msg)
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, msg)
}

func failGracefully(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
	}
}

func checkHeroku() bool {
	if os.Getenv("IS_HEROKU") != "" {
		fmt.Printf("this is running on heroku")
		return true
	}
	return false
}
