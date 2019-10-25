package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/akamensky/argparse"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type config struct {
	report   string
	platform string
	influxDb string
}
type HttpClient struct {
	Client  *http.Client
	baseUrl string
}
type ServiceStatus struct {
	Platform string
	Service  string
	Status   string
}
type Reporters struct {
	influx client.Client
	bpc    client.BatchPointsConfig
}

const up = "up"
const down = "down"
const all = "all"
const google = "google"
const amazon = "amazon"
const azure = "azure"

func main() {
	parser := argparse.NewParser("monitor", "monitors cloud host status")
	// Create string flag
	report := parser.Selector("r", "report", []string{down, up, all}, &argparse.Options{Required: true, Help: "report if services are down, up, or all. For the Google platform, we only retrieve if the service is down."})
	platform := parser.Selector("c", "cloud", []string{google, amazon, azure, all}, &argparse.Options{Required: true, Help: "which platforms to report on"})
	influxCommand := parser.NewCommand("influx", "Dump statii to influxdb")
	influxUrl := influxCommand.String("i", "influx-url", &argparse.Options{Required: true, Help: "The url to influxdb"})
	influxUser := influxCommand.String("u", "username", &argparse.Options{Required: true, Help: "influxdb username"})
	influxPass := influxCommand.String("p", "pasword", &argparse.Options{Required: true, Help: "influxdb password"})
	influxDB := influxCommand.String("d", "influxdb", &argparse.Options{Required: true, Help: "db name to use"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}
	conf := config{}
	conf.report = *report
	conf.platform = *platform
	conf.influxDb = *influxDB
	reporters := Reporters{}
	if influxUrl != nil && *influxUrl != "" {

		influx, _ := client.NewHTTPClient(client.HTTPConfig{Addr: *influxUrl, Username: *influxUser, Password: *influxPass})
		reporters.influx = influx
		bpc := client.BatchPointsConfig{Database: conf.influxDb}
		reporters.bpc = bpc
		defer influx.Close()
	}
	if err != nil {
		panic(err)
	}

	if conf.platform == azure || conf.platform == all {
		httpClient := HttpClient{
			Client:  &http.Client{},
			baseUrl: "https://status.azure.com/en-us/status",
		}
		HandleResult(Azure(conf, httpClient), reporters)
	}
	if conf.platform == google || conf.platform == all {
		httpClient := HttpClient{
			Client:  &http.Client{},
			baseUrl: "https://status.cloud.google.com",
		}
		HandleResult(Google(conf, httpClient), reporters)
	}
	if conf.platform == amazon || conf.platform == all {
		httpClient := HttpClient{
			Client:  &http.Client{},
			baseUrl: "https://status.aws.amazon.com",
		}
		HandleResult(Amazon(conf, httpClient), reporters)
	}
}

func HandleResult(result []ServiceStatus, reporters Reporters) {
	for _, item := range result {
		if item.Status != "OK" {
			logProblem(item, reporters)
		} else {
			logOk(item, reporters)
		}
	}
}

func logProblem(result ServiceStatus, reporters Reporters) {
	fmt.Println(result.Platform + " " + result.Service + " " + result.Status)
	if reporters.influx != nil {
		bp, _ := client.NewBatchPoints(reporters.bpc)
		var data map[string]interface{}
		status := -1
		if result.Status == "Degraded" {
			status = 2
		} else if result.Status == "Down" {
			status = 3
		} else if result.Status == "Info" {
			status = 1
		}
		data = map[string]interface{}{"status": status}
		point, _ := client.NewPoint("platforms", map[string]string{"platform": result.Platform, "service": result.Service, "status-label": result.Status}, data, time.Now())
		bp.AddPoint(point)
		reporters.influx.Write(bp)
	}

}
func logOk(result ServiceStatus, reporters Reporters) {
	fmt.Println(result.Platform + " " + result.Service + " is fine")
	if reporters.influx != nil {

		bp, _ := client.NewBatchPoints(reporters.bpc)
		var data map[string]interface{}
		data = map[string]interface{}{"status": 0}
		point, _ := client.NewPoint("platforms", map[string]string{"platform": result.Platform, "service": result.Service, "status-label": result.Status}, data, time.Now())
		bp.AddPoint(point)
		reporters.influx.Write(bp)
	}
}
func Google(conf config, client HttpClient) []ServiceStatus {
	res, err := client.Client.Get(client.baseUrl)
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
	serviceStatus := []ServiceStatus{}
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
			if !strings.Contains(status, "OK") && conf.report != up {
				serviceStatus = append(serviceStatus, ServiceStatus{
					Platform: google,
					Service:  text,
					Status:   status,
				})
			}
			if strings.Contains(status, "OK") && conf.report != down {
				serviceStatus = append(serviceStatus, ServiceStatus{
					Platform: google,
					Service:  text,
					Status:   status,
				})
			}

		})
	return serviceStatus
}
func Amazon(conf config, client HttpClient) []ServiceStatus {
	res, err := client.Client.Get(client.baseUrl)
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
			if status != "OK" && conf.report != up {
				serviceStatus = append(serviceStatus, ServiceStatus{
					Platform: amazon,
					Service:  text,
					Status:   status,
				})
			}
			if status == "OK" && conf.report != down {
				serviceStatus = append(serviceStatus, ServiceStatus{
					Platform: amazon,
					Service:  text,
					Status:   status,
				})
			}

		})
	return serviceStatus
}
func Azure(conf config, client HttpClient) []ServiceStatus {
	res, err := client.Client.Get(client.baseUrl)
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
	serviceStatus := []ServiceStatus{}
	rows.Find("td:first-child ").Each(
		func(i int, s *goquery.Selection) {
			text := s.Text()
			for strings.Contains(text, "\n") {
				text = strings.ReplaceAll(text, "\n", " ")
			}
			for strings.Contains(text, "  ") {
				text = strings.ReplaceAll(text, "  ", " ")
			}
			status := strings.TrimSpace(s.Parent().Find("td:nth-child(2)").Text())
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

			if status != "OK" && conf.report != up {
				serviceStatus = append(serviceStatus, ServiceStatus{
					Platform: azure,
					Service:  text,
					Status:   status,
				})
			}
			if status == "OK" && conf.report != down {
				serviceStatus = append(serviceStatus, ServiceStatus{
					Platform: azure,
					Service:  text,
					Status:   status,
				})
			}
		})
	return serviceStatus
}
