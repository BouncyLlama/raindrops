package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

// RoundTripFunc .

func TestHealthyAzure(t *testing.T) {

	dat, _ := ioutil.ReadFile("./testdata/azure_today_good.html")
	client := NewTestClient(func(req *http.Request) *http.Response {
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
	conf := config{
		report:   "all",
		platform: "azure",
		influxDb: "",
	}

	status := Azure(conf, api)
	if status == nil || len(status) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	assert.Equal(t, 55, len(status), "Should have 33 records")
	for idx, item := range status {
		assert.Equal(t, "OK", item.Status, `index `+string(idx)+` had status `+item.Status)
		assert.Equal(t, "azure", item.Platform, `index `+string(idx)+` had platform `+item.Platform)

	}

}
func TestUnHealthyAzure(t *testing.T) {

	dat, _ := ioutil.ReadFile("./testdata/azure_today_bad.html")
	client := NewTestClient(func(req *http.Request) *http.Response {
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
	conf := config{
		report:   "all",
		platform: "azure",
		influxDb: "",
	}

	status := Azure(conf, api)
	if status == nil || len(status) == 0 {
		assert.Fail(t, "Should have returned results")
	}
	assert.Equal(t, 55, len(status), "Should have 33 records")
	assert.Equal(t, "Bad", status[0].Status, "Expected service to be down")
	status = append(status[:0], status[0+1:]...)
	for idx, item := range status {
		assert.Equal(t, "OK", item.Status, `index `+strconv.Itoa(idx)+` had status `+item.Status)
		assert.Equal(t, "azure", item.Platform, `index `+strconv.Itoa(idx)+` had platform `+item.Platform)

	}

}
