package amazon

import (
	"bytes"
	"github.com/BouncyLlama/raindrops/internal/test"
	. "github.com/BouncyLlama/raindrops/internal/types"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// RoundTripFunc .

func TestHealthyAmazon(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/amazon_today_good.html")
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
		Platform: "amazon",
	}

	status := ScrapeStatus(conf, api)
	if status == nil || len(status) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	assert.Equal(t, 555, len(status), "Should have 33 records")
	for idx, item := range status {
		assert.Equal(t, "OK", item.Status, `index `+string(idx)+` had status `+item.Status)
		assert.Equal(t, "amazon", item.Platform, `index `+string(idx)+` had platform `+item.Platform)

	}

}
func TestUnHealthyAmazon(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/amazon_today_bad.html")
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
		Platform: "amazon",
	}

	status := ScrapeStatus(conf, api)
	if status == nil || len(status) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	assert.Equal(t, 555, len(status), "Should have 33 records")
	assert.Equal(t, "Down", status[0].Status, "Expected service to be down")
	status = append(status[:0], status[0+1:]...)
	for idx, item := range status {
		assert.Equal(t, "OK", item.Status, `index `+strconv.Itoa(idx)+` had status `+item.Status)
		assert.Equal(t, "amazon", item.Platform, `index `+strconv.Itoa(idx)+` had platform `+item.Platform)

	}

}
func TestIncidents(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/amazon_incidents.html")
	dat2, _ := ioutil.ReadFile("../testdata/amazon_incident.html")

	client := test.NewTestClient(func(req *http.Request) *http.Response {
		var body io.ReadCloser
		if strings.Contains(req.URL.String(), "message") {
			body = ioutil.NopCloser(bytes.NewBuffer(dat2))
		} else {
			body = ioutil.NopCloser(bytes.NewBuffer(dat))

		}

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
	time, _ := time.Parse("January _2, 2006 at 3:04pm (MST)", "June 13, 2014"+" at 12:00am (UTC)")

	assert.Equal(t, 12, len(incidents), "Should have 33 records")
	assert.Equal(t, time, incidents[5].Time)
	assert.Equal(t, "Summary of the Amazon SimpleDB Service Disruption", incidents[5].Description)
	assert.Equal(t, "65649", incidents[5].Identifier)

}
