package azure

import (
	"github.com/BouncyLlama/raindrops/internal/types"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"time"
)

func ScrapeStatus(conf types.Config, client types.HttpClient) []types.ServiceStatus {
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
	rows := doc.Find("table[data-zone-name=americas].region-status-table tr ").Not(".status-category")
	serviceStatus := []types.ServiceStatus{}
	rows.Find("td:first-child ").Each(
		func(i int, s *goquery.Selection) {
			text := s.Text()
			for strings.Contains(text, "\n") {
				text = strings.ReplaceAll(text, "\n", " ")
			}
			for strings.Contains(text, "  ") {
				text = strings.ReplaceAll(text, "  ", " ")
			}
			//only east us for now
			status := strings.TrimSpace(s.Parent().Find("td:nth-child(3)").Text())
			if status == "" || status == "Blank" {
				return
			}
			text = strings.TrimSpace(text)
			if status == "Good" {
				status = "OK"
			} else if status == "Warning" {
				status = "Degraded"
			} else if status == "Critical" {
				status = "Down"
			} else {
				status = "Info"
			}

			if status != "OK" && conf.Report != types.Up {
				serviceStatus = append(serviceStatus, types.ServiceStatus{
					Platform: types.Azure,
					Service:  text,
					Status:   status,
					Time:     time.Now(),
				})
			}
			if status == "OK" && conf.Report != types.Down {
				serviceStatus = append(serviceStatus, types.ServiceStatus{
					Platform: types.Azure,
					Service:  text,
					Status:   status,
					Time:     time.Now(),
				})
			}
		})
	return serviceStatus
}
