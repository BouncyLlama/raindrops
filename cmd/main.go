package main

import (
	"fmt"
	"github.com/BouncyLlama/raindrops/internal/amazon"
	"github.com/BouncyLlama/raindrops/internal/azure"
	"github.com/BouncyLlama/raindrops/internal/google"
	"github.com/BouncyLlama/raindrops/internal/types"
	"github.com/akamensky/argparse"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"net/http"
	"os"
	"time"
)

func main() {
	parser := argparse.NewParser("monitor", "monitors cloud host status")
	// Create string flag
	report := parser.Selector("r", "report", []string{types.Down, types.Up, types.All}, &argparse.Options{Required: true, Help: "report if services are types.Down, types.Up, or types.All. For the Google platform, we only retrieve if the service is types.Down."})
	platform := parser.Selector("c", "cloud", []string{types.Google, types.Amazon, types.Azure, types.All}, &argparse.Options{Required: true, Help: "which platforms to report on"})
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
	conf := types.Config{}
	conf.Report = *report
	conf.Platform = *platform
	conf.InfluxDb = *influxDB
	reporters := types.Reporters{}
	if influxUrl != nil && *influxUrl != "" {

		influx, _ := client.NewHTTPClient(client.HTTPConfig{Addr: *influxUrl, Username: *influxUser, Password: *influxPass})
		reporters.Influx = influx
		bpc := client.BatchPointsConfig{Database: conf.InfluxDb}
		reporters.Bpc = bpc
		defer influx.Close()
	}
	if err != nil {
		panic(err)
	}

	if conf.Platform == types.Azure || conf.Platform == types.All {
		httpClient := types.HttpClient{
			Client:  &http.Client{},
			BaseUrl: "https://status.azure.com/en-us/status",
		}
		HandleResult(azure.ScrapeStatus(conf, httpClient), reporters)
	}
	if conf.Platform == types.Google || conf.Platform == types.All {
		httpClient := types.HttpClient{
			Client:  &http.Client{},
			BaseUrl: "https://status.cloud.google.com",
		}
		HandleResult(google.ScrapeStatus(conf, httpClient), reporters)
	}
	if conf.Platform == types.Amazon || conf.Platform == types.All {
		httpClient := types.HttpClient{
			Client:  &http.Client{},
			BaseUrl: "https://status.aws.amazon.com",
		}
		HandleResult(amazon.ScrapeStatus(conf, httpClient), reporters)
	}
}

func HandleResult(result []types.ServiceStatus, reporters types.Reporters) {
	for _, item := range result {
		if item.Status != "OK" {
			logProblem(item, reporters)
		} else {
			logOk(item, reporters)
		}
	}
}

func logProblem(result types.ServiceStatus, reporters types.Reporters) {
	fmt.Println(result.Platform + " " + result.Service + " " + result.Status)
	if reporters.Influx != nil {
		bp, _ := client.NewBatchPoints(reporters.Bpc)
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
		reporters.Influx.Write(bp)
	}

}
func logOk(result types.ServiceStatus, reporters types.Reporters) {
	fmt.Println(result.Platform + " " + result.Service + " is fine")
	if reporters.Influx != nil {

		bp, _ := client.NewBatchPoints(reporters.Bpc)
		var data map[string]interface{}
		data = map[string]interface{}{"status": 0}
		point, _ := client.NewPoint("platforms", map[string]string{"platform": result.Platform, "service": result.Service, "status-label": result.Status}, data, time.Now())
		bp.AddPoint(point)
		reporters.Influx.Write(bp)
	}
}
