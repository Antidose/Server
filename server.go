package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"database/sql"

	"github.com/sfreiberg/gotwilio"
)

// Configuration : Core config structure
type Configuration struct {
	Twilio   TwilioKey
	DB       DbCreds
	Mapbox   Mapbox
	Firebase Firebase
}

// TwilioKey : Config strucuture for Twilio
type TwilioKey struct {
	Sid    string
	Token  string
	Number string
}

// DbCreds : Cred structure for DB
type DbCreds struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DbName string
}

// Mapbox : Mapbox structure for config
type Mapbox struct {
	Token string
}

// Firebase : Structure for firrebase key
type Firebase struct {
	Key string
}

// Location : Location object
type Location struct {
	Type        string
	Coordinates []float64
	Crs         struct {
		Type       string
		Properties struct {
			Name string
		}
	}
}

// MapboxRoute : Route for mapbox structure
type MapboxRoute struct {
	Duration   float32
	Distance   float32
	Weight     float32
	WeightName string
	Geometry   string
}

// SocketMessage : Structure for socket message
type SocketMessage struct {
	IncidentID string
	UserID     string
}

// Globals
var (
	isHeroku       = checkHeroku()
	done           = make(chan struct{})
	configuration  = loadConfig()
	antidoseTwilio = loadTwilio()
	db             = loadDB()
)

func loadConfig() Configuration {
	configuration := Configuration{}
	if !isHeroku {
		file, err := os.Open("./config/conf.json")
		failOnError(err, "Config json not found. Make sure it is present.")
		decoder := json.NewDecoder(file)

		err = decoder.Decode(&configuration)
		if err != nil {
			fmt.Println("error:", err)
		}
	}
	return configuration
}

func loadTwilio() *gotwilio.Twilio {
	if isHeroku {
		return gotwilio.NewTwilioClient(os.Getenv("TWILIO_SID"), os.Getenv("TWILIO_TOKEN"))
	}
	return gotwilio.NewTwilioClient(configuration.Twilio.Sid, configuration.Twilio.Token)
}

func loadDB() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", configuration.DB.Host, configuration.DB.Port, configuration.DB.User, configuration.DB.Pass, configuration.DB.DbName)
	db, err := sql.Open("postgres", psqlInfo)
	if isHeroku {
		db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	}
	failOnError(err, "Failed to open Postgres")

	err = db.Ping()
	failOnError(err, "Failed to ping Postgres")
	return db
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	initRoutes()
	<-done
}
