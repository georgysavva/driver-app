package driverlochttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc"
)

type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func NewClient(baseURL string) (*Client, error) {
	// Improvement: make http timeout configurable
	httpClient := &http.Client{Timeout: 5 * time.Second}
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse driver-location service base url")
	}
	return &Client{httpClient: httpClient, baseURL: baseURLParsed}, nil
}

func (c *Client) GetLocations(ctx context.Context, driverID string, timeInterval time.Duration) (
	[]*driverloc.Location, error) {
	timeIntervalMinutes := int(math.Round(timeInterval.Minutes()))
	queryParams := url.Values{}
	queryParams.Set("minutes", strconv.Itoa(timeIntervalMinutes))
	reqURL := c.baseURL.ResolveReference(&url.URL{
		Path:     fmt.Sprintf("drivers/%s/locations", driverID),
		RawQuery: queryParams.Encode(),
	})

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil /* body */)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't initialize a new http request with base url")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http get request to driver locations endpoint failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read http response content")
	}
	if err := resp.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "couldn't close http response")
	}
	if resp.StatusCode != http.StatusOK {
		err := &url.Error{
			Op:  "Get",
			URL: req.URL.String(),
			Err: errors.Errorf("not OK http status code %d: %s", resp.StatusCode, body),
		}
		return nil, errors.WithStack(err)
	}
	var locations []*driverloc.Location
	if err := json.Unmarshal(body, &locations); err != nil {
		return nil, errors.Wrapf(err, "failed to decode json into locations list: %s", body)
	}
	return locations, nil
}
