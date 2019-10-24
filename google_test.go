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
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestHealthy(t *testing.T) {

	dat, _ := ioutil.ReadFile("./testdata/google_today_good.html")
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
		platform: "google",
		influxDb: "",
	}

	status := Google(conf, api)
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

	dat, _ := ioutil.ReadFile("./testdata/google_today_bad.html")
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
		platform: "google",
		influxDb: "",
	}

	status := Google(conf, api)
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
