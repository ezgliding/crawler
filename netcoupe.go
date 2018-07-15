// Copyright 2018 The ezgliding authors. All rights reserverd.

package crawler

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

// DailyUrl is the main page to list netcoupe flights.
const DailyUrl = "http://www.netcoupe.net/Results/DailyResults.aspx"

// FlightBaseUrl is the base path to fetch flight details from a flight ID.
const FlightBaseUrl = "http://www.netcoupe.net/Results/FlightDetail.aspx?FlightID="

// TrackBaseUrl is the base path to download the flight track from a track ID.
const TrackBaseUrl = "http://www.netcoupe.net/Download/DownloadIGC.aspx?FileID="

// This is a constant map.
var httpHeaders = map[string][]string{
	"Accept-Encoding":           []string{"gzip, deflate"},
	"Cache-Control":             []string{"max-age=0"},
	"Upgrade-Insecure-Requests": []string{"1"},
	"DNT":             []string{"1"},
	"Origin":          []string{"http://www.netcoupe.net"},
	"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/66.0.3359.181 Chrome/66.0.3359.181 Safari/537.36"},
	"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
	"Accept-Language": []string{"en-US,en;q=0.9,de;q=0.8,fr;q=0.7,pt;q=0.6,es;q=0.5,it;q=0.4,ny;q=0.3"},
	"Connection":      []string{"keep-alive"},
	"Referer":         []string{"http://www.netcoupe.net/Results/DailyResults.aspx"}}

// Netcoupe implements a crawler for http://netcoupe.net.
type Netcoupe struct {
	DownloadTrack bool
	BasePath      string
}

// Crawl checks for flights on netcoupe.net.
//
// It works by querying flights for specific days, making it easier to iterate
// through past data. The rules for flight submission are defined here:
// http://www.planeur.net/_download/netcoupe/2018_np4.2.pdf
// """
// Chaque performance doit être enregistrée dans un délai de 15 jours sur le
// site de la NetCoupe (www.netcoupe.net) par le commandant de bord ou par le
// Responsable de la NetCoupe de l’association avec l'accord du pilote.
// """
// Which means that it's only worth to crawl for new flights back to 2 weeks max.
func (n Netcoupe) Crawl(start time.Time, end time.Time) ([]Flight, error) {
	var flights []Flight

	// Do not allow start > end
	if end.Before(start) {
		return nil, errors.New("Invalid start end date pair")
	}

	r, _ := regexp.Compile(`.*DisplayFlightDetail\('(?P<ID>[0-9]+)'\).*`)

	c := colly.NewCollector()
	c.AllowURLRevisit = true
	c.UserAgent = httpHeaders["User-Agent"][0]
	c.OnRequest(func(r *colly.Request) {
		log.WithFields(log.Fields{
			"url":     r.URL.String(),
			"headers": r.Headers}).Trace("Visiting flight list")
	})
	c.OnError(func(r *colly.Response, err error) {
		log.WithFields(log.Fields{
			"response": r,
			"error":    err}).Error("Failed to visit url")
	})
	d := c.Clone()
	d.OnRequest(func(r *colly.Request) {
		log.WithFields(log.Fields{
			"url":     r.URL.String(),
			"headers": r.Headers}).Trace("Visiting flight details")
	})
	t := c.Clone()
	t.OnRequest(func(r *colly.Request) {
		log.WithFields(log.Fields{
			"url":     r.URL.String(),
			"headers": r.Headers}).Trace("Visiting flight track")
	})

	c.OnHTML("table tr td:nth-child(4) a[href]", func(e *colly.HTMLElement) {
		id := r.FindStringSubmatch(e.Attr("href"))
		log.WithFields(log.Fields{
			"flight_id": id[1]}).Trace("Scheduling flight details")
		if len(id) == 2 {
			d.Visit(fmt.Sprintf("%v%v", FlightBaseUrl, id[1]))
		}

	})

	d.OnHTML("div center table[width]", func(e *colly.HTMLElement) {
		f := Flight{}
		f.URL = e.Request.URL.String()
		f.ID = e.Request.URL.Query()["FlightID"][0]
		f.Pilot = e.ChildText("tbody tr:nth-child(3) td:nth-child(2) a")
		f.Club = e.ChildText("tbody tr:nth-child(5) td:nth-child(2) a")
		f.Date, _ = time.Parse("02/01/2006", e.ChildText("tbody tr:nth-child(8) td:nth-child(2) div"))
		f.Takeoff = e.ChildText("tbody tr:nth-child(9) td:nth-child(2) div")
		f.Region = e.ChildText("tbody tr:nth-child(10) td:nth-child(2) div")
		f.Country = e.ChildText("tbody tr:nth-child(11) td:nth-child(2) div")
		f.Distance = parseFloat(e.ChildText("tbody tr:nth-child(12) td:nth-child(2) div"))
		f.Points = parseFloat(e.ChildText("tbody tr:nth-child(13) td:nth-child(2) div"))
		f.Glider = e.ChildText("tbody tr:nth-child(14) td:nth-child(2) div table tbody tr td")
		f.Type = e.ChildText("tbody tr:nth-child(15) td:nth-child(2) div")
		// FIXME(rochaporto): handle competition field
		trackUrl, _ := url.Parse(e.ChildAttr("tbody tr:nth-child(16) td:nth-child(2) div a", "href"))
		f.TrackID = trackUrl.Query()["FileID"][0]
		f.TrackURL = fmt.Sprintf("%v%v", TrackBaseUrl, f.TrackID)
		f.Speed = parseFloat(e.ChildText("tbody tr:nth-child(17) td:nth-child(2) div"))
		f.Comments = e.ChildText("tbody tr:nth-child(23) td:nth-child(2) div")

		flights = append(flights, f)
		if n.DownloadTrack {
			if _, err := os.Stat(fmt.Sprintf("%v/%v", n.BasePath, f.TrackID)); os.IsNotExist(err) {
				t.Visit(fmt.Sprintf("%v%v", TrackBaseUrl, f.TrackID))
			}
		}
	})

	t.OnResponse(func(r *colly.Response) {
		trackID := r.Request.URL.Query()["FileID"][0]
		ioutil.WriteFile(fmt.Sprintf("%v/%v", n.BasePath, trackID), r.Body, 0644)
	})

	current := time.Date(start.Year(), start.Month(), start.Day(), 12, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 12, 0, 0, 0, time.UTC)
	for ; end.After(current.AddDate(0, 0, -1)); current = current.AddDate(0, 0, 1) {
		data := n.sessionHeaders(c)
		data["ddlDisplayRange"] = "0"
		data["ddlDisplayDate"] = current.Format("02/01/2006")
		data["rbgDisplayMode"] = "rbDisplayByDate"
		n.post(c, DailyUrl, data)
	}

	log.WithFields(log.Fields{
		"start":       start,
		"end":         end,
		"flights":     flights,
		"num_flights": len(flights),
	}).Trace("Finishing crawling flights")
	return flights, nil
}

func (n Netcoupe) sessionHeaders(c *colly.Collector) map[string]string {
	headers := map[string]string{
		"__EVENTARGUMENT": "",
		"__LASTFOCUS":     "",
		"__EVENTTARGET":   "ddlDisplayDate",
		//FIXME(rochaporto): handle single page listing to include all flights
		//"__EVENTTARGET": "dgDailyResults$ctl01$ctl01",
	}

	t := c.Clone()
	t.OnRequest(func(r *colly.Request) {
		log.WithFields(log.Fields{
			"url":     r.URL.String(),
			"headers": r.Headers}).Trace("Visiting for session data collection")

	})
	t.OnHTML("input", func(e *colly.HTMLElement) {
		switch e.Attr("name") {
		case "__EVENTVALIDATION":
			headers["__EVENTVALIDATION"] = e.Attr("value")
		case "__VIEWSTATE":
			headers["__VIEWSTATE"] = e.Attr("value")
		case "__VIEWSTATEGENERATOR":
			headers["__VIEWSTATEGENERATOR"] = e.Attr("value")
		}
	})
	t.Visit(DailyUrl)

	return headers
}

func (n Netcoupe) post(c *colly.Collector, url string, data map[string]string) {
	cookies := c.Cookies(url)
	c.SetCookies(url, cookies)
	dur, _ := time.ParseDuration("1m")
	c.SetRequestTimeout(dur)
	log.WithFields(log.Fields{
		"url":  url,
		"data": data}).Trace("Post request")
	c.Request("POST", url, createFormReader(data), nil, httpHeaders)
}

func (n Netcoupe) get(c *colly.Collector, url string, data map[string]string) {
	cookies := c.Cookies(url)
	c.SetCookies(url, cookies)
	dur, _ := time.ParseDuration("1m")
	c.SetRequestTimeout(dur)
	log.WithFields(log.Fields{
		"url":  url,
		"data": data}).Trace("Post request")
	c.Request("GET", url, createFormReader(data), nil, httpHeaders)
}

func parseFloat(s string) float64 {
	rs := strings.Replace(strings.TrimSpace(strings.Split(s, " ")[0]), ",", ".", -1)
	r, _ := strconv.ParseFloat(rs, 32)
	return r
}

func createFormReader(data map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}
	return strings.NewReader(form.Encode())
}
