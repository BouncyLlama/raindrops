package azure

import (
	"bytes"
	"github.com/BouncyLlama/raindrops/internal/test"
	. "github.com/BouncyLlama/raindrops/internal/types"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// RoundTripFunc .

func TestHealthyAzure(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/azure_today_good.html")
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBuffer(dat)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	api := HttpClient{client, "http://example.com"}
	conf := Config{
		Report:   "all",
		Platform: "azure",
	}

	status := ScrapeStatus(conf, api)
	if status == nil || len(status) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	assert.Equal(t, 2956, len(status), "Should have 103 records")
	for idx, item := range status {
		assert.Equal(t, "OK", item.Status, `index `+string(idx)+` had status `+item.Status)
		assert.Equal(t, "azure", item.Platform, `index `+string(idx)+` had platform `+item.Platform)

	}

}
func TestUnHealthyAzure(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/azure_today_bad.html")
	client := test.NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBuffer(dat)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	api := HttpClient{client, "http://example.com"}
	conf := Config{
		Report:   "all",
		Platform: "azure",
	}

	status := ScrapeStatus(conf, api)
	if status == nil || len(status) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	assert.Equal(t, 2956, len(status), "Should have 103 records")
	assert.Equal(t, "Down", status[0].Status, "Expected service to be down")
	status = append(status[:0], status[0+1:]...)
	for idx, item := range status {
		assert.Equal(t, "OK", item.Status, `index `+strconv.Itoa(idx)+` had status `+item.Status)
		assert.Equal(t, "azure", item.Platform, `index `+strconv.Itoa(idx)+` had platform `+item.Platform)

	}

}
func TestIncidents(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/azure_incidents.html")

	client := test.NewTestClient(func(req *http.Request) *http.Response {
		var body io.ReadCloser

		body = ioutil.NopCloser(bytes.NewBuffer(dat))

		// Test request parameters
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: body,
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	api := HttpClient{client, "http://example.com"}
	conf := Config{
		Report:   "all",
		Platform: "amazon",
	}

	incidents := ScrapeIncidents(conf, api)
	if incidents == nil || len(incidents) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	time, _ := time.Parse("1/2/2006 at 3:04pm (MST)", "10/18/2019 at 3:04pm (MST)")

	assert.Equal(t, 17, len(incidents), "Should have 17 records")
	assert.Equal(t, time, incidents[5].Time)
	assert.Equal(t, "RCA - Authentication issues with Azure MFA in North America", incidents[5].Description)
	assert.Equal(t, "2019-10-18 15:04:00 +0000 MST RCA - Authentication issues with Azure MFA in North America", incidents[5].Identifier)

}
