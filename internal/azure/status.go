package azure

import (
	"github.com/BouncyLlama/raindrops/internal/types"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
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
	serviceStatus := []types.ServiceStatus{}

	for _, regionstr := range []string{"americas", "europe", "asia", "middle-east-africa", "azure-government"} {
		rows := doc.Find("table[data-zone-name=" + regionstr + "].region-status-table tr ").Not(".status-category")
		rows.Find("td:first-child ").Each(
			func(i int, s *goquery.Selection) {
				text := s.Text()
				for strings.Contains(text, "\n") {
					text = strings.ReplaceAll(text, "\n", " ")
				}
				for strings.Contains(text, "  ") {
					text = strings.ReplaceAll(text, "  ", " ")
				}
				cols := s.Parent().Find("td")
				for i := 2; i < cols.Length(); i++ {
					status := strings.TrimSpace(s.Parent().Find("td:nth-child(" + strconv.Itoa(i) + ")").Text())
					if status == "" || status == "Blank" {
						continue
					}
					header := doc.Find("table[data-zone-name=" + regionstr + "].region-status-table th:nth-child(" + strconv.Itoa(i) + ")").Text()
					header = strings.TrimSpace(header)
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
							Region:   &header,
							Time:     time.Now(),
						})
					}
					if status == "OK" && conf.Report != types.Down {
						serviceStatus = append(serviceStatus, types.ServiceStatus{
							Platform: types.Azure,
							Service:  text,
							Status:   status,
							Region:   &header,
							Time:     time.Now(),
						})
					}
				}

			})
	}

	return serviceStatus
}
