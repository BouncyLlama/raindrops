package main

import (
	"context"
	"fmt"
	"github.com/BouncyLlama/raindrops/internal/amazon"
	"github.com/BouncyLlama/raindrops/internal/azure"
	"github.com/BouncyLlama/raindrops/internal/google"
	"github.com/BouncyLlama/raindrops/internal/types"
	"github.com/akamensky/argparse"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	parser := argparse.NewParser("monitor", "monitors cloud host status")
	scrapeStatusCommand := parser.NewCommand("status", "scrape current statii")
	scrapeIncidentCommand := parser.NewCommand("incidents", "scrape incident descriptions")
	report := scrapeStatusCommand.Selector("r", "report", []string{types.Down, types.Up, types.All}, &argparse.Options{Required: true, Help: "report if services are types.Down, types.Up, or types.All. For the Google platform, we only retrieve if the service is types.Down."})
	platform := parser.Selector("c", "cloud", []string{types.Google, types.Amazon, types.Azure, types.All}, &argparse.Options{Required: true, Help: "which platforms to report on"})
	mongoUrl := parser.String("m", "mongo-url", &argparse.Options{Required: false, Help: "The url to mongo"})
	mongoUser := parser.String("u", "username", &argparse.Options{Required: false, Help: "mongo username"})
	mongoPass := parser.String("p", "password", &argparse.Options{Required: false, Help: "mongo password"})
	mongoDb := parser.String("d", "mongo-dbname", &argparse.Options{Required: false, Help: "db name to use"})

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
	reporters := types.Reporters{}
	if *mongoUrl != "" && *mongoUser != "" && *mongoPass != "" && *mongoDb != "" {

		// Set client options
		clientOptions := options.Client().ApplyURI("mongodb://" + *mongoUrl)
		clientOptions.Auth = &options.Credential{
			AuthSource: *mongoDb,
			Username:   *mongoUser,
			Password:   *mongoPass,
		}

		// Connect to MongoDB
		mongoClient, err := mongo.Connect(context.TODO(), clientOptions)
		defer mongoClient.Disconnect(context.TODO())
		statusCollection := mongoClient.Database(*mongoDb).Collection("platformStatus")
		incidentsCollection := mongoClient.Database(*mongoDb).Collection("platformIncidents")
		reporters.IncidentsCollection = incidentsCollection
		reporters.StatusCollection = statusCollection
		if err != nil {
			log.Fatal(err)
		}

		// Check the connection
		err = mongoClient.Ping(context.TODO(), nil)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Connected to MongoDB!")
	}

	if err != nil {
		panic(err)
	}

	if scrapeStatusCommand.Happened() {
		var statii []types.ServiceStatus
		if conf.Platform == types.Azure || conf.Platform == types.All {
			httpClient := types.HttpClient{
				Client:  &http.Client{},
				BaseUrl: "https://status.azure.com/en-us/status",
			}
			statii = append(statii, azure.ScrapeStatus(conf, httpClient)...)
		}
		if conf.Platform == types.Google || conf.Platform == types.All {
			httpClient := types.HttpClient{
				Client:  &http.Client{},
				BaseUrl: "https://status.cloud.google.com",
			}
			statii = append(statii, google.ScrapeStatus(conf, httpClient)...)
		}
		if conf.Platform == types.Amazon || conf.Platform == types.All {
			httpClient := types.HttpClient{
				Client:  &http.Client{},
				BaseUrl: "https://status.aws.amazon.com",
			}
			statii = append(statii, amazon.ScrapeStatus(conf, httpClient)...)

		}
		HandleResult(statii, reporters)

	}

	if scrapeIncidentCommand.Happened() {
		var incidents []types.Incident

		if conf.Platform == types.Amazon || conf.Platform == types.All {
			httpClient := types.HttpClient{
				Client:  &http.Client{},
				BaseUrl: "https://aws.amazon.com/premiumsupport/technology/pes/",
			}
			incidents = append(incidents, amazon.ScrapeIncidents(conf, httpClient)...)

		}
		for _, item := range incidents {
			logIncident(item, reporters)
		}
	}
}

func HandleResult(result []types.ServiceStatus, reporters types.Reporters) {
	for _, item := range result {
		logStatus(item, reporters)
	}

}
func logIncident(incident types.Incident, reporters types.Reporters) {
	fmt.Println(incident.Description)
	if reporters.IncidentsCollection != nil {
		filter := bson.D{{"identifier", bson.D{{"$eq", incident.Identifier}}}}
		res, err := reporters.IncidentsCollection.UpdateOne(context.TODO(), filter, bson.M{"$set": incident}, options.Update().SetUpsert(true))
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println(res)
		}
	}
}

func logStatus(result types.ServiceStatus, reporters types.Reporters) {
	fmt.Println(result.Platform + " " + result.Service + " " + result.Status)
	if reporters.StatusCollection != nil {
		t := time.Now()
		today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		doc := types.ServiceStatusDocument{
			Platform: result.Platform,
			Service:  result.Service,
			Day:      today,
			Samples:  []types.ServiceStatus{result},
		}

		filter := bson.M{"platform": doc.Platform, "service": doc.Service, "day": doc.Day}
		res := reporters.StatusCollection.FindOne(context.TODO(), filter)
		if res != nil {
			reporters.StatusCollection.UpdateOne(context.TODO(), filter, bson.M{"$push": bson.M{"samples": bson.M{"$each": doc.Samples}}}, options.Update().SetUpsert(true))

		} else {
			reporters.StatusCollection.InsertOne(context.TODO(), doc)
		}
	}

}
