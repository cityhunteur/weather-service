package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	v1 "github.com/cityhunteur/weather-service/api/v1"
	"github.com/cityhunteur/weather-service/internal/cache"
	"github.com/cityhunteur/weather-service/internal/handler"
	"github.com/cityhunteur/weather-service/internal/handler/mocks"
	"github.com/cityhunteur/weather-service/internal/pkg/openstreetmap"
	"github.com/cityhunteur/weather-service/internal/pkg/weathergov"
)

func TestGetForecastHandler_GetForecast(t *testing.T) {
	now := time.Now().UTC()
	oneDayLater := now.AddDate(0, 0, 1)
	tests := []struct {
		name                         string
		query                        string
		wantStatusCode               int
		openStreetMapAPIExpectations func(api *mocks.OpenStreetMapAPI)
		weatherGovAPIExpectations    func(api *mocks.WeatherGovAPI)
		cacheExpectations            func(api *mocks.Cache)
	}{
		{
			name:           "success",
			query:          "?city=new%20york",
			wantStatusCode: 200,
			openStreetMapAPIExpectations: func(mockAPI *mocks.OpenStreetMapAPI) {
				mockAPI.On("GetPlace", mock.Anything, mock.Anything).Return([]*openstreetmap.Place{
					{
						ID:          366998854,
						Lat:         "40.7127281",
						Lon:         "-74.0060152",
						DisplayName: "City of New York, New York, United States",
					},
				}, nil)
			},
			weatherGovAPIExpectations: func(mockAPI *mocks.WeatherGovAPI) {
				mockAPI.On("GetPoints", mock.Anything, mock.Anything).Return(&weathergov.Points{
					ID: "https://api.weather.gov/points/40.7127,-74.006",
					Properties: weathergov.PointsProperties{
						Forecast: "https://api.weather.gov/gridpoints/OKX/33,35/forecast",
					},
				}, nil)
				start, _ := time.Parse(time.RFC3339, "2023-06-29T17:00:00-04:00")
				end, _ := time.Parse(time.RFC3339, "2023-06-29T18:00:00-04:00")
				mockAPI.On("GetForecast", mock.Anything, mock.Anything).Return(&weathergov.Forecast{
					Properties: weathergov.ForecastProperties{
						Periods: []weathergov.Periods{
							{
								Name:             "New York",
								StartTime:        v1.Time3339(start),
								EndTime:          v1.Time3339(end),
								DetailedForecast: "Haze. Partly sunny, with a high near 85. Southwest wind around 6 mph.",
							},
						},
					},
				}, nil)
			},
		},
		{
			name:                         "invalid query param",
			query:                        "?city=",
			wantStatusCode:               400,
			openStreetMapAPIExpectations: func(mockAPI *mocks.OpenStreetMapAPI) {},
			weatherGovAPIExpectations:    func(mockAPI *mocks.WeatherGovAPI) {},
		},
		{
			name:           "uses cache",
			query:          "?city=new%20york",
			wantStatusCode: 200,
			openStreetMapAPIExpectations: func(mockAPI *mocks.OpenStreetMapAPI) {
				mockAPI.On("GetPlace", mock.Anything, mock.Anything).Return([]*openstreetmap.Place{
					{
						ID:          366998854,
						Lat:         "40.7127281",
						Lon:         "-74.0060152",
						DisplayName: "City of New York, New York, United States",
					},
				}, nil)
			},
			weatherGovAPIExpectations: func(mockAPI *mocks.WeatherGovAPI) {
				mockAPI.On("GetPoints", mock.Anything, mock.Anything).Return(&weathergov.Points{
					ID: "https://api.weather.gov/points/40.7127,-74.006",
					Properties: weathergov.PointsProperties{
						Forecast: "https://api.weather.gov/gridpoints/OKX/33,35/forecast",
					},
				}, nil)
				start, _ := time.Parse(time.RFC3339, now.Format(time.RFC3339))
				end, _ := time.Parse(time.RFC3339, oneDayLater.Format(time.RFC3339))
				mockAPI.On("GetForecast", mock.Anything, mock.Anything).Return(&weathergov.Forecast{
					Properties: weathergov.ForecastProperties{
						Periods: []weathergov.Periods{
							{
								Name:             "New York",
								StartTime:        v1.Time3339(start),
								EndTime:          v1.Time3339(end),
								DetailedForecast: "Haze. Partly sunny, with a high near 85. Southwest wind around 6 mph.",
							},
						},
					},
				}, nil)
			},
			cacheExpectations: func(mockCache *mocks.Cache) {
				forecast := &weathergov.Forecast{
					Properties: weathergov.ForecastProperties{
						Periods: []weathergov.Periods{
							{
								Name:             "New York",
								StartTime:        v1.Time3339(now),
								EndTime:          v1.Time3339(oneDayLater),
								DetailedForecast: "Haze. Partly sunny, with a high near 85. Southwest wind around 6 mph.",
							},
						},
					},
				}
				mockCache.On("Set", "New York", forecast).Return(nil).Times(1)
				mockCache.On("Get", mock.Anything, mock.Anything).Return(forecast, true)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			mockOpenStreetMapAPI := mocks.NewOpenStreetMapAPI(t)
			tt.openStreetMapAPIExpectations(mockOpenStreetMapAPI)
			mockWeatherGovAPI := mocks.NewWeatherGovAPI(t)
			tt.weatherGovAPIExpectations(mockWeatherGovAPI)
			h := handler.NewGetForecastHandler(logger, mockOpenStreetMapAPI, mockWeatherGovAPI, cache.NewStore())

			resp := httptest.NewRecorder()
			_, router := gin.CreateTestContext(resp)
			router.GET("/v1/weather", h.GetForecast)
			req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/weather%s", tt.query), nil)
			router.ServeHTTP(resp, req)

			if tt.wantStatusCode != 200 {
				assert.Equal(t, tt.wantStatusCode, resp.Code)
				return
			}

			assert.Equal(t, tt.wantStatusCode, resp.Code)

			var got v1.ListWeatherResponse
			err := json.NewDecoder(resp.Body).Decode(&got)
			assert.NoError(t, err)
			assert.Len(t, got.Forecast, 1)
			assert.Equal(t, "New York", got.Forecast[0].Name)
			assert.Len(t, got.Forecast[0].Detail, 1)

			t.Cleanup(func() {
				mockOpenStreetMapAPI.AssertExpectations(t)
				mockWeatherGovAPI.AssertExpectations(t)
			})
		})
	}
}
