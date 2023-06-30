// Package v1 contains types used by the API.
package v1

type Detail struct {
	StartTime   Time3339 `json:"startTime"`
	EndTime     Time3339 `json:"endTime"`
	Description string   `json:"description"`
}

// Forecast represents the current forecast and forecasts for the next two days for a given city.
type Forecast struct {
	Name   string    `json:"name"`
	Detail []*Detail `json:"detail"`
}

// ListWeatherResponse represents the response for the v1 API.
type ListWeatherResponse struct {
	Forecast []*Forecast `json:"forecast"`
}
