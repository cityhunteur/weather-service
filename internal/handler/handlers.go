package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	v1 "github.com/cityhunteur/weather-service/api/v1"
	"github.com/cityhunteur/weather-service/internal/pkg/openstreetmap"
	"github.com/cityhunteur/weather-service/internal/pkg/weathergov"
)

const (
	defaultTimeout = 10 * time.Second
	defaultFormat  = "json"

	defaultCountry = "USA"
)

//go:generate mockery --name OpenStreetMapAPI
type OpenStreetMapAPI interface {
	GetPlace(ctx context.Context, opts *openstreetmap.GetOptions) ([]*openstreetmap.Place, error)
}

//go:generate mockery --name WeatherGovAPI
type WeatherGovAPI interface {
	GetPoints(ctx context.Context, coord *weathergov.Coordinates) (*weathergov.Points, error)
	GetForecast(ctx context.Context, forecastURL string) (*weathergov.Forecast, error)
}

//go:generate mockery --name Cache
type Cache interface {
	Set(k string, v *v1.Forecast)
	Get(k string) (*v1.Forecast, bool)
}

type GetForecastHandler struct {
	logger *zap.SugaredLogger

	openStreetMapAPI OpenStreetMapAPI
	weatherGovAPI    WeatherGovAPI

	cache Cache
}

// NewGetForecastHandler creates an API handler to get weather forecast.
func NewGetForecastHandler(logger *zap.SugaredLogger, openStreetMapAPI OpenStreetMapAPI, weatherGovAPI WeatherGovAPI, cache Cache) *GetForecastHandler {
	return &GetForecastHandler{
		logger:           logger,
		openStreetMapAPI: openStreetMapAPI,
		weatherGovAPI:    weatherGovAPI,
		cache:            cache,
	}
}

func (h *GetForecastHandler) GetForecast(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c, defaultTimeout)
	defer cancel()

	citiesStr := c.DefaultQuery("city", "")
	if citiesStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Query param 'city' missing."})
		return
	}

	h.logger.Debugw("Getting forecasts", "cities", citiesStr)

	cities := strings.Split(citiesStr, ",")

	forecasts := make([]*v1.Forecast, 0)
	for _, city := range cities {
		h.logger.Debugw("Getting forecast for city", "city", city)

		// use value from cache if present
		if forecast, exists := h.cache.Get(city); exists {
			forecasts = append(forecasts, forecast)
			continue
		}

		q := fmt.Sprintf("%s,%s", city, defaultCountry)
		places, err := h.openStreetMapAPI.GetPlace(ctx, &openstreetmap.GetOptions{
			Query:  q,
			Format: defaultFormat,
		})
		if err != nil {
			h.logger.Errorw("Failed to retrieve coordinates for city",
				"error", err,
				"city", city,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to retrieve weather forecast"})
			return
		}

		if len(places) < 1 {
			h.logger.Errorw("Failed to retrieve place details for city", "city", city)
			// graceful degradation; continue with other cities
			continue
		}

		// IMPORTANT: defaults to first record; requires further analysis
		// of response to narrow down result
		place := places[0]

		forecast := &v1.Forecast{
			Name: cases.Title(language.English).String(city),
		}

		points, err := h.weatherGovAPI.GetPoints(ctx, &weathergov.Coordinates{
			Lat: place.Lat,
			Lon: place.Lon,
		})
		if err != nil {
			// graceful degradation; continue with other places despite err
			h.logger.Errorw("Failed to retrieve weather point details",
				"error", err,
				"city", city,
			)
		}

		forecastResp, err := h.weatherGovAPI.GetForecast(ctx, points.Properties.Forecast)
		if err != nil {
			// graceful degradation; continue with other places despite err
			h.logger.Errorw("Failed to retrieve weather point details for city",
				"error", err,
				"city", city,
			)
			continue
		}

		twoDaysLater := time.Now().UTC().AddDate(0, 0, 2)
		for _, period := range forecastResp.Properties.Periods {
			// we only include day time forecast for today and next two days
			if period.IsDaytime && time.Time(period.StartTime).After(twoDaysLater) {
				break
			}
			forecast.Detail = append(forecast.Detail, &v1.Detail{
				StartTime:   period.StartTime,
				EndTime:     period.EndTime,
				Description: period.DetailedForecast,
			})
		}

		forecasts = append(forecasts, forecast)
		h.cache.Set(city, forecast)
	}

	c.JSON(http.StatusOK, &v1.ListWeatherResponse{
		Forecast: forecasts,
	})
}
