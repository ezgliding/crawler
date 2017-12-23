// Copyright 2018 The ezgliding authors. All rights reserverd.

package crawler

import (
	"time"

	igc "github.com/ezgliding/goigc"
)

//FlightURL ...
type FlightURL string

// FlightGetter ...
type FlightGetter interface {
	FlightFetch() ([]Flight, error)
	FlightGet(url FlightURL) (Flight, error)
}

// Flight ...
type Flight struct {
	Pilot    string
	Club     string
	Date     time.Time
	Takeoff  string
	Region   string
	Country  string
	Distance float64
	Points   float64
	Glider   string
	Type     string
	Speed    float64
	Task     []igc.Point
	Comments string
}
