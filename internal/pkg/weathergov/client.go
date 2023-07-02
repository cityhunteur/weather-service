package weathergov

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL = "https://api.weather.gov/"
)

// Client is a Weather Gov API client.
type Client struct {
	client *http.Client

	baseURL *url.URL
}

// NewClient creates a new Client using the given http client if provided.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		client:  httpClient,
		baseURL: baseURL,
	}
}

// Coordinates represents the geo coordinates of a place.
type Coordinates struct {
	Lat string
	Lon string
}

// GetPoints get details about the Points at the specified coordinates.
func (c *Client) GetPoints(ctx context.Context, coord *Coordinates) (*Points, error) {
	u := fmt.Sprintf("points/%s,%s", coord.Lat, coord.Lon)

	req, err := c.newRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, fmt.Errorf("creating api request: %w", err)
	}

	var p Points
	err = c.doRequest(ctx, req, &p)
	if err != nil {
		return nil, fmt.Errorf("calling api to get points: %w", err)
	}

	return &p, nil
}

// GetForecast retrieves the forecast for a given place.
func (c *Client) GetForecast(ctx context.Context, forecastURL string) (*Forecast, error) {
	req, err := c.newRequest(ctx, http.MethodGet, forecastURL)
	if err != nil {
		return nil, fmt.Errorf("creating api request: %w", err)
	}

	var f Forecast
	err = c.doRequest(ctx, req, &f)
	if err != nil {
		return nil, fmt.Errorf("calling api to get forecast: %w", err)
	}

	return &f, nil
}

func (c *Client) newRequest(ctx context.Context, method, urlStr string) (*http.Request, error) {
	u, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) doRequest(ctx context.Context, req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return fmt.Errorf("sending http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("decoding json response: %w", err)
	}

	return nil
}
