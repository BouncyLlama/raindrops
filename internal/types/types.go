package types

import (
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"net/http"
)

type Config struct {
	Report   string
	Platform string
	InfluxDb string
}
type HttpClient struct {
	Client  *http.Client
	BaseUrl string
}
type ServiceStatus struct {
	Platform string
	Service  string
	Status   string
}
type Reporters struct {
	Influx client.Client
	Bpc    client.BatchPointsConfig
}
