package crawler

import (
	"time"
)

type FlightID string

type Crawler interface {
	Crawl(time.Time, time.Time) ([]Flight, error)
}
