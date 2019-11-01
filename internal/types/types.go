package types

import (
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

type Config struct {
	Report    string
	Platform  string
	Reporters *Reporters
}
type HttpClient struct {
	Client  *http.Client
	BaseUrl string
}
type ServiceStatus struct {
	Platform string `-`
	Service  string `-`
	Region   *string
	Status   string
	Time     time.Time
}
type ServiceStatusDocument struct {
	Platform string
	Service  string
	Day      time.Time
	Samples  []ServiceStatus
}
type Reporters struct {
	IncidentsCollection *mongo.Collection
	StatusCollection    *mongo.Collection
}
type Incident struct {
	Time        time.Time
	Description string
	Text        string
	Service     string
	Identifier  string
	Platform    string
	Resolved    bool
}
