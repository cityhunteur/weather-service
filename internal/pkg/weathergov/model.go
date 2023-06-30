package weathergov

import v1 "github.com/cityhunteur/weather-service/api/v1"

type PointsProperties struct {
	Forecast string `json:"forecast"`
}

// Points holds the properties of a geolocation.
type Points struct {
	ID         string           `json:"id"`
	Properties PointsProperties `json:"properties"`
}

type Periods struct {
	Name             string      `json:"name"`
	StartTime        v1.Time3339 `json:"startTime"`
	EndTime          v1.Time3339 `json:"endTime"`
	IsDaytime        bool        `json:"isDaytime"`
	DetailedForecast string      `json:"detailedForecast"`
}

type ForecastProperties struct {
	Periods []Periods `json:"periods"`
}

// Forecast represents the forecast for a given point.
type Forecast struct {
	Properties ForecastProperties `json:"properties"`
}
