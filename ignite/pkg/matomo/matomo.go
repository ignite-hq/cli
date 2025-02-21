// Package matomo is a client for Matomo to send data points for hint-type=event.
package matomo

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-querystring/query"

	"github.com/ignite/cli/v29/ignite/pkg/errors"
)

type (
	// Client is an analytics client.
	Client struct {
		endpoint   string
		idSite     uint   // Matomo ID Site.
		tokenAuth  string // Matomo Token Auth.
		source     string
		httpClient http.Client
	}

	// Params analytics metrics body.
	Params struct {
		IDSite      uint   `url:"idsite"`
		Rec         uint   `url:"rec"`
		ActionName  string `url:"action_name"`
		APIVersion  uint   `url:"apiv"`
		TokenAuth   string `url:"token_auth,omitempty"`
		Rand        uint64 `url:"rand,omitempty"`
		URL         string `url:"url,omitempty"`
		UTMSource   string `url:"utm_source,omitempty"`
		UTMMedium   string `url:"utm_medium,omitempty"`
		UTMCampaign string `url:"utm_campaign,omitempty"`
		UTMContent  string `url:"utm_content,omitempty"`
		UserID      string `url:"uid,omitempty"`
		UserAgent   string `url:"ua,omitempty"`
		Hour        int    `url:"h,omitempty"`
		Minute      int    `url:"m,omitempty"`
		Second      int    `url:"s,omitempty"`

		// Dimension1 development mode boolean.
		// 1 = devMode ON | 0 = devMode OFF.
		Dimension1 uint `url:"dimension1"`

		// Dimension2 internal boolean.
		// 1 = internal ON not supported at present | 0 = internal OFF.
		Dimension2 uint `url:"dimension2"`

		// Dimension3 is deprecated.
		// Should always be 0.
		Dimension3 uint `url:"dimension3"`

		// Dimension4 ignite version
		Dimension4 string `url:"dimension4,omitempty"`

		// Dimension6 ignite config version
		Dimension6 string `url:"dimension6,omitempty"`

		// Dimension7 full cli command
		Dimension7 string `url:"dimension7,omitempty"`

		// Dimension11 scaffold customization type
		Dimension11 string `url:"dimension11,omitempty"`

		// Dimension13 command level 1.
		Dimension13 string `url:"dimension13,omitempty"`

		// Dimension14 command level 2.
		Dimension14 string `url:"dimension14,omitempty"`

		// Dimension15 command level 3.
		Dimension15 string `url:"dimension15,omitempty"`

		// Dimension16 command level 4.
		Dimension16 string `url:"dimension16,omitempty"`

		// Dimension17 cosmos-sdk version.
		Dimension17 string `url:"dimension17,omitempty"`

		// Dimension18 operational system.
		Dimension18 string `url:"dimension18,omitempty"`

		// Dimension19 system architecture.
		Dimension19 string `url:"dimension19,omitempty"`

		// Dimension20 golang version.
		Dimension20 string `url:"dimension20,omitempty"`

		// Dimension21 command level 5.
		Dimension21 string `url:"dimension21,omitempty"`

		// Dimension22 command level 6.
		Dimension22 string `url:"dimension22,omitempty"`
	}
	// Metric represents a custom data.
	Metric struct {
		Name            string
		Cmd             string
		OS              string
		Arch            string
		Version         string
		CLIVersion      string
		GoVersion       string
		SDKVersion      string
		BuildDate       string
		SourceHash      string
		ConfigVersion   string
		Uname           string
		CWD             string
		ScaffoldType    string
		BuildFromSource bool
		IsCI            bool
	}
)

// Option configures code generation.
type Option func(*Client)

// WithIDSite adds an id site.
func WithIDSite(idSite uint) Option {
	return func(c *Client) {
		c.idSite = idSite
	}
}

// WithTokenAuth adds a matomo token authentication.
func WithTokenAuth(tokenAuth string) Option {
	return func(c *Client) {
		c.tokenAuth = tokenAuth
	}
}

// WithSource adds a matomo URL source.
func WithSource(source string) Option {
	return func(c *Client) {
		c.source = source
	}
}

// New creates a new Matomo client.
func New(endpoint string, opts ...Option) Client {
	c := Client{
		endpoint: endpoint,
		source:   endpoint,
		httpClient: http.Client{
			Timeout: 1500 * time.Millisecond,
		},
	}
	// apply analytics options.
	for _, o := range opts {
		o(&c)
	}
	return c
}

// Send sends metric event to analytics.
func (c Client) Send(params Params) error {
	requestURL, err := url.Parse(c.endpoint)
	if err != nil {
		return err
	}

	// encode request parameters.
	queryParams, err := query.Values(params)
	if err != nil {
		return err
	}
	requestURL.RawQuery = queryParams.Encode()

	// Create an HTTP request with the payload.
	resp, err := c.httpClient.Get(requestURL.String())
	if err != nil {
		return errors.Wrapf(err, "error creating HTTP request: %s", requestURL.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("error to add matomo analytics metric. Status code: %d", resp.StatusCode)
	}

	return nil
}

// SendMetric build the metrics and send to analytics.
func (c Client) SendMetric(sessionID string, metric Metric) error {
	var (
		now       = time.Now()
		r, _      = rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		utmMedium = "dev"
	)
	if !metric.BuildFromSource {
		utmMedium = "binary"
	}

	cmd := splitCommand(metric.Cmd)

	return c.Send(Params{
		IDSite:      c.idSite,
		Rec:         1,
		APIVersion:  1,
		TokenAuth:   c.tokenAuth,
		Rand:        r.Uint64(),
		URL:         c.metricURL(metric.Cmd),
		UTMSource:   "source-code-github",
		UTMMedium:   utmMedium,
		UTMCampaign: metric.CLIVersion,
		UTMContent:  fmt.Sprintf("commit-%s", metric.SourceHash),
		UserID:      sessionID,
		UserAgent:   "Go-http-client",
		ActionName:  metric.Cmd,
		Hour:        now.Hour(),
		Minute:      now.Minute(),
		Second:      now.Second(),
		Dimension1:  0,
		Dimension2:  formatBool(metric.IsCI),
		Dimension4:  metric.Version,
		Dimension6:  metric.ConfigVersion,
		Dimension7:  metric.Cmd,
		Dimension11: metric.ScaffoldType,
		Dimension13: cmd[0],
		Dimension14: cmd[1],
		Dimension15: cmd[2],
		Dimension16: cmd[3],
		Dimension17: metric.SDKVersion,
		Dimension18: metric.OS,
		Dimension19: metric.Arch,
		Dimension20: metric.GoVersion,
		Dimension21: cmd[4],
		Dimension22: cmd[5],
	})
}

// formatBool returns "1" or "0" according to the value of b.
func formatBool(b bool) uint {
	if b {
		return 1
	}
	return 0
}

// splitCommand splice the command into a slice with length 6.
func splitCommand(cmd string) []string {
	var (
		splitCmd  = strings.Split(cmd, " ")
		cmdLevels = make([]string, 6)
	)
	for i := 0; i < len(cmdLevels); i++ {
		if i >= len(splitCmd) {
			break
		}
		cmdLevels[i] = splitCmd[i]
	}
	return cmdLevels
}

// metricURL build the metric URL.
func (c Client) metricURL(cmd string) string {
	return fmt.Sprintf("%s/%s", c.source, strings.ReplaceAll(cmd, " ", "_"))
}
