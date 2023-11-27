// Package gacli is a client for Google Analytics to send data points for hint-type=event.
package gacli

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	endpoint = "https://www.google-analytics.com/collect"
)

// Client is an analytics client.
type Client struct {
	id string // Google Analytics ID
}

// New creates a new analytics client for Segment.io with Segment's
// endpoint and access key.
func New(id string) Client {
	return Client{
		id: id,
	}
}

// Metric represents a data point.
type Metric struct {
	Category string
	Action   string
	Label    string
	Value    string
	User     string
	Version  string
}

// Send sends metrics to GA.
func (c Client) Send(metric Metric) error {
	v := url.Values{
		"v":   {"1"},
		"tid": {c.id},
		"cid": {metric.User},
		"t":   {"event"},
		"ec":  {metric.Category},
		"ea":  {metric.Action},
		"ua":  {"Opera/9.80 (Windows NT 6.0) Presto/2.12.388 Version/12.14"},
	}
	if metric.Label != "" {
		v.Set("el", metric.Label)
	}
	if metric.Value != "" {
		v.Set("ev", metric.Value)
	}
	if metric.Version != "" {
		v.Set("an", metric.Version)
		v.Set("av", metric.Version)
	}
	// Create an HTTP client and send the hit
	client := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	req.URL.RawQuery = v.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid gacli status code: %d", resp.StatusCode)
	}
	return nil
}
