package azure

import (
	"fmt"
	. "github.com/BouncyLlama/raindrops/internal/types"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
	"strings"
	"time"
)

func ScrapeIncidents(conf Config, client HttpClient) []Incident {
	res, err := client.Client.Get(client.BaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	links := doc.Find(".column.small-11")
	incidents := []Incident{}
	links.Each(func(i int, selection *goquery.Selection) {
		incident := Incident{}
		description := selection.Find(".text-heading4").Text()
		incident.Description = description
		date := selection.Prev().Text()

		date = strings.TrimSpace(date)
		incident.Time, err = time.Parse("1/2/2006 at 3:04pm (MST)", date+"/"+strconv.Itoa(time.Now().Year())+" at 3:04pm (MST)")
		if err != nil {
			fmt.Println("unable to parse time for incident " + incident.Identifier)
			incident.Time = time.Now()
		}

		incident.Identifier = incident.Time.String() + " " + incident.Description
		incident.Text = selection.Text()
		incident.Platform = Azure

		incidents = append(incidents, incident)
	})

	return incidents
}
