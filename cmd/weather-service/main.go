package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cityhunteur/weather-service/internal/cache"
	"github.com/cityhunteur/weather-service/internal/handler"
	"github.com/cityhunteur/weather-service/internal/pkg/openstreetmap"
	"github.com/cityhunteur/weather-service/internal/pkg/weathergov"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func main() {
	flag.Parse()

	zlog, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialise logger: %v", err)
	}
	defer func() { _ = zlog.Sync() }()
	logger = zlog.Sugar()

	osmClient := openstreetmap.NewClient(http.DefaultClient)
	wgClient := weathergov.NewClient(http.DefaultClient)
	store := cache.NewStore()

	h := handler.NewGetForecastHandler(logger, osmClient, wgClient, store)
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	// setup API routes
	router := gin.Default()
	router.GET("/v1/weather", h.GetForecast)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Infof("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Infof("Server exiting")
}
