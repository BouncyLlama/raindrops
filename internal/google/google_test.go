package google

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

func TestHealthy(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/google_today_good.html")
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
		Platform: "google",
	}

	status := ScrapeStatus(conf, api)
	if status == nil || len(status) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	assert.Equal(t, 33, len(status), "Should have 33 records")
	for idx, item := range status {
		assert.Equal(t, "OK", item.Status, `index `+string(idx)+` had status `+item.Status)
		assert.Equal(t, "google", item.Platform, `index `+string(idx)+` had platform `+item.Platform)

	}

}
func TestUnHealthy(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/google_today_bad.html")
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
		Platform: "google",
	}

	status := ScrapeStatus(conf, api)
	if status == nil || len(status) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	assert.Equal(t, 33, len(status), "Should have 33 records")
	assert.Equal(t, "Down", status[32].Status, "Expected service to be down")
	status = append(status[:32], status[32+1:]...)
	for idx, item := range status {
		assert.Equal(t, "OK", item.Status, `index `+strconv.Itoa(idx)+` had status `+item.Status)
		assert.Equal(t, "google", item.Platform, `index `+strconv.Itoa(idx)+` had platform `+item.Platform)

	}

}
func TestIncidents(t *testing.T) {

	dat, _ := ioutil.ReadFile("../testdata/google_incidents.html")
	dat2, _ := ioutil.ReadFile("../testdata/google_incident.html")

	client := test.NewTestClient(func(req *http.Request) *http.Response {
		var body io.ReadCloser
		if strings.Contains(req.URL.String(), "incident") {
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
	time, _ := time.Parse("2006-01-02 15:04 (MST)", "2019-09-03 09:08"+" (PST)")

	assert.Equal(t, 115, len(incidents), "Should have 33 records")
	assert.Equal(t, time, incidents[1].Time)
	assert.Equal(t, "Google App Engine GAE19010 Began 03 September 2019, lasting 4 hours 31 minutes", incidents[1].Description)
	assert.Equal(t, "GAE19010", incidents[1].Identifier)

}
