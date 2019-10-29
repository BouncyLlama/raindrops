package google

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
	rows := doc.Find("table:nth-child(1) tr")
	serviceStatus := []types.ServiceStatus{}
	rows.Each(
		func(i int, s *goquery.Selection) {

			// drop the legend row
			if i == rows.Length()-1 {
				return
			}
			text := s.Find("td").Text()
			//this column is the current date
			statusNode := s.Find("td:nth-child(9)")
			okResult := statusNode.Find(".ok")
			highResult := statusNode.Find(".timeline-incident.high")
			mediumResult := statusNode.Find(".timeline-incident.medium")

			status := "OK"
			if highResult != nil && len(highResult.Nodes) != 0 {
				status = "Down"
			} else if mediumResult != nil && len(mediumResult.Nodes) != 0 {
				status = "Degraded"
			} else if okResult != nil && len(okResult.Nodes) != 0 {
				status = "OK"
			}
			text = strings.TrimSpace(text)
			if text == "" || status == "" {
				return
			}
			if !strings.Contains(status, "OK") && conf.Report != types.Up {
				serviceStatus = append(serviceStatus, types.ServiceStatus{
					Platform: types.Google,
					Service:  text,
					Status:   status,
					Time:     time.Now(),
				})
			}
			if strings.Contains(status, "OK") && conf.Report != types.Down {
				serviceStatus = append(serviceStatus, types.ServiceStatus{
					Platform: types.Google,
					Service:  text,
					Status:   status,
					Time:     time.Now(),
				})
			}

		})
	return serviceStatus
}
