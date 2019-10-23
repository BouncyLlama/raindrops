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

const up = "up"
const down = "down"
const all = "all"
const google = "google"
const amazon = "amazon"
const azure = "azure"

var influx client.Client

func main() {
	parser := argparse.NewParser("monitor", "monitors cloud host status")
	// Create string flag
	report := parser.Selector("r", "report", []string{down, up, all}, &argparse.Options{Required: true, Help: "report if services are down, up, or all. For the Google platform, we only retrieve if the service is down."})
	platform := parser.Selector("c", "cloud", []string{google, amazon, azure, all}, &argparse.Options{Required: true, Help: "which platforms to report on"})
	influxCommand := parser.NewCommand("influx", "Dump statii to influxdb")
	influxUrl := influxCommand.String("i", "influx-url", &argparse.Options{Required: true, Help: "The url to influxdb"})
	influxUser := influxCommand.String("u", "username", &argparse.Options{Required: true, Help: "influxdb username"})
	influxPass := influxCommand.String("p", "pasword", &argparse.Options{Required: true, Help: "influxdb password"})
	influxdDB := influxCommand.String("d", "influxdb", &argparse.Options{Required: true, Help: "db name to use"})

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
	conf.influxDb = *influxdDB
	if influxUrl != nil && *influxUrl != "" {

		influx, err = client.NewHTTPClient(client.HTTPConfig{Addr: *influxUrl, Username: *influxUser, Password: *influxPass})
		defer influx.Close()
	}
	if err != nil {
		panic(err)
	}

	if conf.platform == azure || conf.platform == all {
		Azure(conf)
	}
	if conf.platform == google || conf.platform == all {
		Google(conf)
	}
	if conf.platform == amazon || conf.platform == all {
		Amazon(conf)
	}
}

func logProblem(conf config, platform string, service string, status string) {
	fmt.Println(platform + " " + service + " " + status)
	if conf.influxDb != "" {
		bpc := client.BatchPointsConfig{Database: conf.influxDb}
		bp, _ := client.NewBatchPoints(bpc)
		var data map[string]interface{}
		data = map[string]interface{}{"status": status}
		point, _ := client.NewPoint("platforms", map[string]string{"platform": platform, "service": service}, data, time.Now())
		bp.AddPoint(point)
		influx.Write(bp)
	}

}
func logOk(conf config, platform string, service string) {
	fmt.Println(platform + " " + service + " is fine")
	if conf.influxDb != "" {

		bpc := client.BatchPointsConfig{Database: conf.influxDb}
		bp, _ := client.NewBatchPoints(bpc)
		var data map[string]interface{}
		data = map[string]interface{}{"status": "OK"}
		point, _ := client.NewPoint("platforms", map[string]string{"platform": platform, "service": service}, data, time.Now())
		bp.AddPoint(point)
		influx.Write(bp)
	}
}
func Google(conf config) {
	res, err := http.Get("https://status.cloud.google.com")
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

			status := "Ok"
			if highResult != nil && len(highResult.Nodes) != 0 {
				status = "Down"
			} else if mediumResult != nil && len(mediumResult.Nodes) != 0 {
				status = "Disrupted"
			} else if okResult != nil && len(okResult.Nodes) != 0 {
				status = "Ok"
			}
			text = strings.TrimSpace(text)
			if text == "" || status == "" {
				return
			}
			if !strings.Contains(status, "Ok") && conf.report != up {
				logProblem(conf, google, text, status)
			}
			if strings.Contains(status, "Ok") && conf.report != down {
				logOk(conf, google, text)
			}

		})

}
func Amazon(conf config) {
	res, err := http.Get("https://status.aws.amazon.com")
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

	rows.Each(
		func(i int, s *goquery.Selection) {

			text := s.Find("td:nth-child(2)").Text()

			status := s.Find("td:nth-child(3)").Text()
			status = strings.TrimSpace(status)
			text = strings.TrimSpace(text)
			if text == "" || status == "" {
				return
			}
			if !strings.Contains(status, "operating normally") && conf.report != up {
				logProblem(conf, amazon, text, status)
			}
			if strings.Contains(status, "operating normally") && conf.report != down {
				logOk(conf, amazon, text)
			}

		})

}
func Azure(conf config) {
	res, err := http.Get("https://status.azure.com/en-us/status")
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
			if status != "Good" && conf.report != up {
				logProblem(conf, azure, text, status)
			}
			if status == "Good" && conf.report != down {
				logOk(conf, azure, text)
			}
		})
}
