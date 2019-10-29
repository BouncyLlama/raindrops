package amazon

import (
	"fmt"
	. "github.com/BouncyLlama/raindrops/internal/types"
	"github.com/PuerkitoBio/goquery"
	"log"
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
	links := doc.Find("a").FilterFunction(func(i int, selection *goquery.Selection) bool {
		return strings.Contains(selection.Text(), "Summary of the")
	})
	incidents := []Incident{}
	links.Each(func(i int, selection *goquery.Selection) {
		incident := Incident{}
		description := selection.Text()
		incident.Description = description
		date := selection.Parent().Text()
		date = strings.ReplaceAll(date, description, "")
		date = strings.ReplaceAll(date, "&nbsp;", "")
		date = strings.ReplaceAll(date, ".", "")
		date = strings.TrimLeft(date, ",")
		date = strings.TrimSpace(date)
		incident.Time, err = time.Parse("January _2, 2006 at 3:04pm (MST)", date+" at 12:00am (UTC)")
		if err != nil {
			fmt.Println("unable to parse time for incident " + incident.Identifier)
			incident.Time = time.Now()
		}
		identifier, _ := selection.Attr("href")
		identifierStr := strings.Split(identifier, "/")[2]
		incident.Identifier = identifierStr
		client.BaseUrl = "https://aws.amazon.com/message/" + identifierStr
		incident.Text = getIncident(client)

		incidents = append(incidents, incident)
	})

	return incidents
}
func getIncident(client HttpClient) string {
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
	return doc.Find(".aws-text-box.section").Text()
}
