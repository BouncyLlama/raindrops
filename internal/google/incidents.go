package google

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
	links := doc.Find("tr").FilterFunction(func(i int, selection *goquery.Selection) bool {
		return len(selection.Find(".empty").Nodes) == 0
	})
	incidents := []Incident{}
	current := ""
	links.Each(func(i int, selection *goquery.Selection) {
		header := selection.Find("h1")
		if len(header.Nodes) > 0 {
			current = strings.TrimSpace(header.Text())
			return
		}
		//TODO: verify this is actually the way to see if its resolved or not
		bubble := selection.Find(".bubble")
		resolved := true
		if len(bubble.Nodes) > 0 {
			resolved = false
		}
		link := selection.Find("a")
		identifier := link.Text()
		href, _ := link.Attr("href")
		description := strings.TrimSpace(selection.Find(".description").Text())
		incident := Incident{}
		incident.Identifier = identifier
		incident.Service = current
		incident.Resolved = resolved
		incident.Description = current + " " + identifier + " " + description

		client.BaseUrl = "https://status.cloud.google.com" + href
		incident.Text, incident.Time = getIncident(client)
		incident.Platform = Google
		incidents = append(incidents, incident)
	})

	return incidents
}
func getIncident(client HttpClient) (string, time.Time) {
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

	timeStr := doc.Find(".secondary strong").First().Text()
	t, err := time.Parse("2006-01-02 15:04 (MST)", timeStr+" (PST)")
	if err != nil {
		fmt.Println(err.Error())
	}
	table := doc.Find("table")
	return table.Text(), t
}
