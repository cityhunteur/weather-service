package openstreetmap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

const (
	defaultBaseURL = "https://nominatim.openstreetmap.org/"
)

// Client is an openstreetmap API client.
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

// GetOptions specifies the parameters to query on.
type GetOptions struct {
	// q specifies a simple query
	Query string `url:"q"`
	// format specifies the encoding, e.g. json.
	Format string `url:"format"`
}

// GetPlace retrieve all places that matches the given list options.
func (c *Client) GetPlace(ctx context.Context, opts *GetOptions) ([]*Place, error) {
	u := "search"
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, fmt.Errorf("adding query params to url: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, fmt.Errorf("creating api request: %w", err)
	}

	var out []*Place
	err = c.doRequest(ctx, req, &out)
	if err != nil {
		return nil, fmt.Errorf("calling api to search places: %w", err)
	}

	return out, nil
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

	if err := checkResponse(resp); err != nil {
		return fmt.Errorf("http error response: %w", err)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("decoding json response: %w", err)
	}

	return nil
}

// addOptions adds the parameters in opts as URL query parameters.
func addOptions(s string, opts interface{}) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return s, fmt.Errorf("parsing url: %w", err)
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, fmt.Errorf("encoding query params: %w", err)
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	return fmt.Errorf("unexpected http status code: %d", r.StatusCode)
}
