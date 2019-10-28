package amazon

import (
	. "github.com/BouncyLlama/raindrops/internal/types"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
)

func ScrapeStatus(conf Config, client HttpClient) []ServiceStatus {
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
	rows := doc.Find("div#NA_block table:nth-child(2) tr")
	serviceStatus := []ServiceStatus{}
	rows.Each(
		func(i int, s *goquery.Selection) {

			text := s.Find("td:nth-child(2)").Text()

			status := s.Find("td:nth-child(3)").Text()
			status = strings.TrimSpace(status)
			text = strings.TrimSpace(text)
			if text == "" || status == "" {
				return
			}
			if status == "Service is operating normally" {
				status = "OK"
			} else if status == "Service degradation" {
				status = "Degraded"
			} else if status == "Service disruption" {
				status = "Down"
			} else {
				status = "Info"
			}
			if status != "OK" && conf.Report != Up {
				serviceStatus = append(serviceStatus, ServiceStatus{
					Platform: Amazon,
					Service:  text,
					Status:   status,
				})
			}
			if status == "OK" && conf.Report != Down {
				serviceStatus = append(serviceStatus, ServiceStatus{
					Platform: Amazon,
					Service:  text,
					Status:   status,
				})
			}

		})
	return serviceStatus
}
