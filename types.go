package main

import "database/sql"

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


type Incident struct {
	IncID      string
	Longitude  float64
	Latitude   float64
	Start      string
	End        sql.NullString
}

type Responder struct {
	Uid          int
	First        string
	Last         string
	Longitude    float64
	Latitude     float64
	RespondingTo string
}

type AdminInfo struct {
	Incidents []Incident
	Responders []Responder
}

// DataStruct : Structure for sending Notification Data
type DataStruct struct {
	Notification string  `json:"notification"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	Max          int     `json:"max"`
	IncidentID   string  `json:"incident_id"`
}

// Notification : Structure for Notifications with Data
type Notification struct {
	To         string     `json:"to"`
	Priority   string     `json:"priority"`
	Data       DataStruct `json:"data"`
	TimeToLive int        `json:"time_to_live"`
}
