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
		configuration.Twilio.Sid = os.Getenv("TWILIO_SID")
		configuration.Twilio.Token = os.Getenv("TWILIO_TOKEN")
		configuration.Twilio.Number = os.Getenv("TWILIO_NUMBER")
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
