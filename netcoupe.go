// Copyright 2018 The ezgliding authors. All rights reserverd.

package crawler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	// FlightDetail ...
	FlightDetail string = "http://netcoupe.net/Results/FlightDetail.aspx?FlightID="
)

// NetcoupeFetcher ...
type NetcoupeFetcher struct {
	lastID int
	n      int
}

// FlightFetch ...
func (nf *NetcoupeFetcher) FlightFetch() ([]Flight, error) {
	return FlightFetchN(nf.n)

}

// FlightFetchN ...
func (nf *NetcoupeFetcher) FlightFetchN(n int) ([]Flight, error) {
	flights := make([]Flight, n)
	newID := lastID + 1
	for i := 0; i < n; i++ {
		resp, err := http.Get(FlightDetaild + newID)
		if err != nil {
			log.Fatal("")
			break
		}
		// Stop when Status.MovedPermanently or Status.NotFound
		if resp.StatusCode == 302 || resp.StatusCode == 400 {
			break
		}
		flight[1], err = ParseFlight(resp)
		if err != nil {
			log.Fatal("")
			break
		}
		time.Sleep(nf.queryInterval)
		newID := lastID + 1
	}
}

// ParseFlight ...
func ParseFlight(content string) (Flight, error) {
	root, err := html.Parse(content)
	if err != nil {
		return err
	}
}

func main() {
	// request and parse the front page
	resp, err := http.Get("https://news.ycombinator.com/")
	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.A && n.Parent != nil && n.Parent.Parent != nil {
			return scrape.Attr(n.Parent.Parent, "class") == "athing"
		}
		return false
	}
	// grab all articles and print them
	articles := scrape.FindAll(root, matcher)
	for i, article := range articles {
		fmt.Printf("%2d %s (%s)\n", i, scrape.Text(article), scrape.Attr(article, "href"))
	}
}
