// Copyright 2018 The ezgliding authors. All rights reserverd.

package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestNetcoupeCrawler(t *testing.T) {
	start := time.Date(2018, time.December, 24, 12, 0, 0, 0, time.UTC)
	end := time.Date(2018, time.December, 25, 12, 0, 0, 0, time.UTC)

	var n Crawler = NewNetcoupe()
	flights, err := n.Crawl(start, end)
	if err != nil {
		t.Errorf("%v", err)
	}

	if len(flights) <= 0 {
		t.Errorf("no flights returned")
	}

	jsonFlights, _ := json.MarshalIndent(flights, "", "   ")
	fmt.Printf("%v\n", string(jsonFlights))
}

func TestNetcoupeCrawlerDownload(t *testing.T) {

	year := 2016
	start := time.Date(year, 12, 1, 12, 0, 0, 0, time.UTC)
	end := time.Date(year, 12, 31, 12, 0, 0, 0, time.UTC)

	var n Netcoupe = NewNetcoupeYear(year)
	current := start
	for ; end.After(current.AddDate(0, 0, -1)); current = current.AddDate(0, 0, 1) {
		var flights []Flight
		dbFile := fmt.Sprintf("db/%v/%v.json", year, current.Format("02-01-2006"))
		if _, err := os.Stat(dbFile); os.IsNotExist(err) {
			flights, err = n.Crawl(current, current)
			if err != nil {
				t.Errorf("%v", err)
			}
			jsonFlights, _ := json.MarshalIndent(flights, "", "   ")
			ioutil.WriteFile(dbFile, jsonFlights, 0644)
		}

		for _, f := range flights {
			flightFile := fmt.Sprintf("db/%v/flights/%v", year, f.TrackID)
			if _, err := os.Stat(flightFile); os.IsNotExist(err) {
				url := fmt.Sprintf("%v%v", n.trackBaseUrl(), f.TrackID)
				data, _ := n.Get(url)
				ioutil.WriteFile(flightFile, data, 0644)
			}
		}
	}

}
