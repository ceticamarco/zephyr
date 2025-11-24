package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ceticamarco/zephyr/cache"
	"github.com/ceticamarco/zephyr/controller"
	"github.com/ceticamarco/zephyr/types"
)

func main() {
	// Retrieve listening port, API token and cache time-to-live from environment variables
	var (
		host   = os.Getenv("ZEPHYR_ADDR")
		port   = os.Getenv("ZEPHYR_PORT")
		token  = os.Getenv("ZEPHYR_TOKEN")
		ttl, _ = strconv.ParseInt(os.Getenv("ZEPHYR_CACHE_TTL"), 10, 8)
	)

	if host == "" || port == "" || token == "" || ttl == 0 {
		log.Fatalf("Environment variables not set")
	}

	// Initialize cache, statDB and vars
	masterCache := cache.InitMasterCache()
	statCache := cache.InitStatCache()
	vars := types.Variables{
		Token:      token,
		TimeToLive: int8(ttl),
	}

	// API endpoints
	http.HandleFunc("/weather/", func(res http.ResponseWriter, req *http.Request) {
		controller.GetWeather(res, req, &masterCache.WeatherCache, statCache, &vars)
	})

	http.HandleFunc("/metrics/", func(res http.ResponseWriter, req *http.Request) {
		controller.GetMetrics(res, req, &masterCache.MetricsCache, &vars)
	})

	http.HandleFunc("/wind/", func(res http.ResponseWriter, req *http.Request) {
		controller.GetWind(res, req, &masterCache.WindCache, &vars)
	})

	http.HandleFunc("/forecast/", func(res http.ResponseWriter, req *http.Request) {
		controller.GetForecast(res, req, &masterCache.DailyForecastCache, &masterCache.HourlyForecastCache, &vars)
	})

	http.HandleFunc("/moon", func(res http.ResponseWriter, req *http.Request) {
		controller.GetMoon(res, req, &masterCache.MoonCache, &vars)
	})

	http.HandleFunc("/stats/", func(res http.ResponseWriter, req *http.Request) {
		controller.GetStatistics(res, req, statCache)
	})

	listenAddr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Server listening on %s", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}
