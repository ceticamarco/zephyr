package controller

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ceticamarco/zephyr/cache"
	"github.com/ceticamarco/zephyr/model"
	"github.com/ceticamarco/zephyr/types"
)

func jsonError(res http.ResponseWriter, key string, value string, status int) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	json.NewEncoder(res).Encode(map[string]string{key: value})
}

func jsonValue(res http.ResponseWriter, val any) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(val)
}

func fmtTemperature(temp string, isImperial bool) string {
	parsedTemp, _ := strconv.ParseFloat(temp, 64)

	if isImperial {
		return fmt.Sprintf("%d°F", int(math.Round(parsedTemp*(9/5)+32)))
	}

	return fmt.Sprintf("%d°C", int(math.Round(parsedTemp)))
}

func fmtStdDev(stdDev string, isImperial bool) string {
	parsedStdDev, _ := strconv.ParseFloat(stdDev, 64)

	if isImperial {
		return fmt.Sprintf("%.4f°F", (parsedStdDev*(9/5) + 32))
	}

	return fmt.Sprintf("%.4f°C", parsedStdDev)
}

func fmtWind(windSpeed string, isImperial bool) string {
	// Convert wind speed to mph or km/s from m/s
	// 1 m/s = 2.23694 mph
	// 1 m/s = 3.6 km/h
	parsedSpeed, _ := strconv.ParseFloat(windSpeed, 64)

	if isImperial {
		return fmt.Sprintf("%.1f mph", (parsedSpeed * 2.23694))
	}

	return fmt.Sprintf("%.1f km/h", (parsedSpeed * 3.6))
}

func fmtKey(key string) string {
	// Cache/database key is formatted by:
	// 1. Removing leading and trailing whitespaces
	// 2. Replacing in-between spaces using the '+' token
	// 3. Making the key uppercase
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(key), " ", "+"))
}

func fmtDailyForecast(forecast *types.DailyForecast, isImperial bool) {
	for idx := range forecast.Forecast {
		val := &forecast.Forecast[idx]
		val.Min = fmtTemperature(val.Min, isImperial)
		val.Max = fmtTemperature(val.Max, isImperial)
		val.FeelsLike = fmtTemperature(val.FeelsLike, isImperial)
		val.Wind.Speed = fmtWind(val.Wind.Speed, isImperial)
	}
}

func fmtHourlyForecast(forecast *types.HourlyForecast, isImperial bool) {
	for idx := range forecast.Forecast {
		val := &forecast.Forecast[idx]
		val.Temperature = fmtTemperature(val.Temperature, isImperial)
		val.Wind.Speed = fmtWind(val.Wind.Speed, isImperial)
	}
}

func deepCopyForecast[T types.DailyForecast | types.HourlyForecast](original T) T {
	var fc_copy T

	switch any(original).(type) {
	case types.DailyForecast:
		orig := any(original).(types.DailyForecast)
		fc_copy = any(types.DailyForecast{
			Forecast: append([]types.DailyForecastEntity(nil), orig.Forecast...),
		}).(T)
	case types.HourlyForecast:
		orig := any(original).(types.HourlyForecast)
		fc_copy = any(types.HourlyForecast{
			Forecast: append([]types.HourlyForecastEntity(nil), orig.Forecast...),
		}).(T)
	}

	return fc_copy
}

func GetWeather(res http.ResponseWriter, req *http.Request, cache *cache.MasterCache[types.Weather], statCache *cache.StatCache, vars *types.Variables) {
	if req.Method != http.MethodGet {
		jsonError(res, "error", "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract city name from '/weather/:city'
	path := strings.TrimPrefix(req.URL.Path, "/weather/")
	cityName := strings.Trim(path, "/") // Remove trailing slash if present

	if cityName == "" {
		jsonError(res, "error", "specify city name", http.StatusMethodNotAllowed)
		return
	}

	// Check whether the 'i' parameter(imperial mode) is specified
	isImperial := req.URL.Query().Has("i")

	cachedValue, found := cache.GetEntry(fmtKey(cityName), vars.TimeToLive)
	if found {
		// Format weather object and then return it
		cachedValue.Temperature = fmtTemperature(cachedValue.Temperature, isImperial)
		cachedValue.Min = fmtTemperature(cachedValue.Min, isImperial)
		cachedValue.Max = fmtTemperature(cachedValue.Max, isImperial)
		cachedValue.FeelsLike = fmtTemperature(cachedValue.FeelsLike, isImperial)

		jsonValue(res, cachedValue)
	} else {
		// Get city coordinates
		city, err := model.GetCoordinates(cityName, vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		// Get city weather
		weather, dailyTemp, err := model.GetWeather(&city, vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		// Add result to cache
		cache.AddEntry(weather, fmtKey(cityName))

		// Insert new statistic entry into the statistics database
		currentDate := time.Now().Format("2006-01-02")
		statCache.AddStatistic(fmtKey(cityName), currentDate, dailyTemp)

		// Format weather object and then return it
		weather.Temperature = fmtTemperature(weather.Temperature, isImperial)
		weather.Min = fmtTemperature(weather.Min, isImperial)
		weather.Max = fmtTemperature(weather.Max, isImperial)
		weather.FeelsLike = fmtTemperature(weather.FeelsLike, isImperial)

		jsonValue(res, weather)
	}
}

func GetMetrics(res http.ResponseWriter, req *http.Request, cache *cache.MasterCache[types.Metrics], vars *types.Variables) {
	if req.Method != http.MethodGet {
		jsonError(res, "error", "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract city name from '/metrics/:city'
	path := strings.TrimPrefix(req.URL.Path, "/metrics/")
	cityName := strings.Trim(path, "/") // Remove trailing slash if present

	if cityName == "" {
		jsonError(res, "error", "specify city name", http.StatusMethodNotAllowed)
		return
	}

	// Check whether the 'i' parameter(imperial mode) is specified
	isImperial := req.URL.Query().Has("i")

	cachedValue, found := cache.GetEntry(fmtKey(cityName), vars.TimeToLive)
	if found {
		// Format metrics object and then return it
		cachedValue.Humidity = fmt.Sprintf("%s%%", cachedValue.Humidity)
		cachedValue.Pressure = fmt.Sprintf("%s hPa", cachedValue.Pressure)
		cachedValue.DewPoint = fmtTemperature(cachedValue.DewPoint, isImperial)
		cachedValue.Visibility = fmt.Sprintf("%skm", cachedValue.Visibility)

		jsonValue(res, cachedValue)
	} else {
		// Get city coordinates
		city, err := model.GetCoordinates(cityName, vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		// Get city weather
		metrics, err := model.GetMetrics(&city, vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		// Add result to cache
		cache.AddEntry(metrics, fmtKey(cityName))

		// Format metrics object and then return it
		metrics.Humidity = fmt.Sprintf("%s%%", metrics.Humidity)
		metrics.Pressure = fmt.Sprintf("%s hPa", metrics.Pressure)
		metrics.DewPoint = fmtTemperature(metrics.DewPoint, isImperial)
		metrics.Visibility = fmt.Sprintf("%skm", metrics.Visibility)

		jsonValue(res, metrics)
	}
}

func GetWind(res http.ResponseWriter, req *http.Request, cache *cache.MasterCache[types.Wind], vars *types.Variables) {
	if req.Method != http.MethodGet {
		jsonError(res, "error", "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract city name from '/wind/:city'
	path := strings.TrimPrefix(req.URL.Path, "/wind/")
	cityName := strings.Trim(path, "/") // Remove trailing slash if present

	if cityName == "" {
		jsonError(res, "error", "specify city name", http.StatusMethodNotAllowed)
		return
	}

	// Check whether the 'i' parameter(imperial mode) is specified
	isImperial := req.URL.Query().Has("i")

	cachedValue, found := cache.GetEntry(fmtKey(cityName), vars.TimeToLive)
	if found {
		// Format wind object and then return it
		cachedValue.Speed = fmtWind(cachedValue.Speed, isImperial)

		jsonValue(res, cachedValue)
	} else {
		// Get city coordinates
		city, err := model.GetCoordinates(cityName, vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		// Get city wind
		wind, err := model.GetWind(&city, vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		// Add result to cache
		cache.AddEntry(wind, fmtKey(cityName))

		// Format wind object and then return it
		wind.Speed = fmtWind(wind.Speed, isImperial)

		jsonValue(res, wind)
	}
}

func GetForecast(
	res http.ResponseWriter,
	req *http.Request,
	dCache *cache.MasterCache[types.DailyForecast],
	hCache *cache.MasterCache[types.HourlyForecast],
	vars *types.Variables,
) {
	if req.Method != http.MethodGet {
		jsonError(res, "error", "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract city name from '/forecast/:city'
	path := strings.TrimPrefix(req.URL.Path, "/forecast/")
	cityName := strings.Trim(path, "/") // Remove trailing slash if present

	if cityName == "" {
		jsonError(res, "error", "specify city name", http.StatusMethodNotAllowed)
		return
	}

	// Check whether the 'i' parameter(imperial mode) is specified
	isImperial := req.URL.Query().Has("i")

	// Check whether the 'h' parameter(hourly forecast) is specified
	if req.URL.Query().Has("h") {
		cachedValue, found := hCache.GetEntry(fmtKey(cityName), vars.TimeToLive)
		if found {
			forecast := deepCopyForecast(cachedValue)
			fmtHourlyForecast(&forecast, isImperial)
			jsonValue(res, forecast)
			return
		}

		city, err := model.GetCoordinates(cityName, vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		forecast, err := model.GetForecast[types.HourlyForecast](&city, vars.Token, model.HOURLY)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		hCache.AddEntry(deepCopyForecast(forecast), fmtKey(cityName))
		fmtHourlyForecast(&forecast, isImperial)
		jsonValue(res, forecast)
	} else { // Daily forecast(default)
		cachedValue, found := dCache.GetEntry(fmtKey(cityName), vars.TimeToLive)
		if found {
			forecast := deepCopyForecast(cachedValue)
			fmtDailyForecast(&forecast, isImperial)
			jsonValue(res, forecast)
			return
		}

		city, err := model.GetCoordinates(cityName, vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		forecast, err := model.GetForecast[types.DailyForecast](&city, vars.Token, model.DAILY)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		dCache.AddEntry(deepCopyForecast(forecast), fmtKey(cityName))
		fmtDailyForecast(&forecast, isImperial)
		jsonValue(res, forecast)
	}
}

func GetMoon(res http.ResponseWriter, req *http.Request, cache *cache.MasterCache[types.Moon], vars *types.Variables) {
	if req.Method != http.MethodGet {
		jsonError(res, "error", "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cachedValue, found := cache.GetEntry(fmtKey("moon"), vars.TimeToLive)
	if found {
		// Format moon object and then return it
		cachedValue.Percentage = fmt.Sprintf("%s%%", cachedValue.Percentage)

		jsonValue(res, cachedValue)
	} else {
		// Get moon data
		moon, err := model.GetMoon(vars.Token)
		if err != nil {
			jsonError(res, "error", err.Error(), http.StatusBadRequest)
			return
		}

		// Add result to cache
		cache.AddEntry(moon, fmtKey("moon"))

		// Format moon object and then return it
		moon.Percentage = fmt.Sprintf("%s%%", moon.Percentage)

		jsonValue(res, moon)
	}
}

func addRandomStatistics(statDB *cache.StatCache, city string, n int, meanTemp, stdDev float64) {
	now := time.Now().AddDate(0, 0, -1) // Start from yesterday
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < n; i++ {
		date := now.AddDate(0, 0, -i)
		temp := r.NormFloat64()*stdDev + meanTemp

		statDB.AddStatistic(
			fmtKey(city),
			date.Format("2006-01-02"),
			temp,
		)
	}
}

func GetStatistics(res http.ResponseWriter, req *http.Request, statCache *cache.StatCache) {
	if req.Method != http.MethodGet {
		jsonError(res, "error", "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract city name from '/stats/:city'
	path := strings.TrimPrefix(req.URL.Path, "/stats/")
	cityName := strings.Trim(path, "/") // Remove trailing slash if present

	if cityName == "" {
		jsonError(res, "error", "specify city name", http.StatusMethodNotAllowed)
		return
	}

	// Check whether the 'i' parameter(imperial mode) is specified
	isImperial := req.URL.Query().Has("i")

	// Get city statistics
	stats, err := model.GetStatistics(fmtKey(cityName), statCache)
	if err != nil {
		jsonError(res, "error", err.Error(), http.StatusBadRequest)
		return
	}

	// Format statistics object and then return it
	stats.Min = fmtTemperature(stats.Min, isImperial)
	stats.Max = fmtTemperature(stats.Max, isImperial)
	stats.Mean = fmtTemperature(stats.Mean, isImperial)
	stats.StdDev = fmtStdDev(stats.StdDev, isImperial)
	stats.Median = fmtTemperature(stats.Median, isImperial)
	stats.Mode = fmtTemperature(stats.Mode, isImperial)
	if stats.Anomaly != nil {
		for idx, val := range *stats.Anomaly {
			(*stats.Anomaly)[idx].Temp = fmtTemperature(val.Temp, isImperial)
		}
	}

	jsonValue(res, stats)
}
