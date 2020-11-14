package driverlochttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(baseURL string) *Client {
	// Improvement: make http timeout configurable
	httpClient := &http.Client{Timeout: 5 * time.Second}
	return &Client{httpClient: httpClient, baseURL: baseURL}
}

func (c *Client) GetLocations(ctx context.Context, driverID string, timeInterval time.Duration) (
	[]*driverloc.Location, error) {
	timeIntervalMinutes := int(math.Round(timeInterval.Minutes()))
	urlQuery := url.Values{}
	urlQuery.Add("minutes", strconv.Itoa(timeIntervalMinutes))
	urlQuery.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil /* body */)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't initialize a new http request with base url")
	}
	req.URL.Path = path.Join(req.URL.Path, fmt.Sprintf("drivers/%s/locations", driverID))
	req.URL.RawQuery = urlQuery.Encode()
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
