package model

import (
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ceticamarco/zephyr/types"
)

type FCType int

const (
	DAILY FCType = iota
	HOURLY
)

// Structures representing the daily forecast
type dailyRes struct {
	Temp struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"temp"`
	FeelsLike struct {
		Day float64 `json:"day"`
	} `json:"feels_like"`
	Weather []struct {
		Title       string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	WindSpeed float64 `json:"wind_speed"`
	WindDeg   float64 `json:"wind_deg"`
	RainProb  float64 `json:"pop"`
	Timestamp int64   `json:"dt"`
}

type dailyForecastRes struct {
	Daily []dailyRes `json:"daily"`
}

// Structure representing the hourly forecast
type hourlyRes struct {
	Temperature float64 `json:"temp"`
	Weather     []struct {
		Title       string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	WindSpeed float64 `json:"wind_speed"`
	WindDeg   float64 `json:"wind_deg"`
	RainProb  float64 `json:"pop"`
	Timestamp int64   `json:"dt"`
}

type hourlyForecastRes struct {
	Hourly []hourlyRes `json:"hourly"`
}

func getForecastEntity[T types.DailyForecastEntity | types.HourlyForecastEntity, K dailyRes | hourlyRes](forecast K) T {
	switch fc := any(forecast).(type) {
	case dailyRes:
		// Format UNIX timestamp as 'YYYY-MM-DD'
		utcTime := time.Unix(int64(fc.Timestamp), 0)
		weatherDate := types.ZephyrDate{Date: utcTime.UTC()}

		// Set condition accordingly to weather description
		var condition string
		switch fc.Weather[0].Description {
		case "few clouds":
			condition = "SunWithCloud"
		case "broken clouds":
			condition = "CloudWithSun"
		default:
			condition = fc.Weather[0].Title
		}

		// Get emoji from weather condition
		emoji := GetEmoji(condition, false)

		// Get cardinal direction and wind arrow
		windDirection, windArrow := GetCardinalDir(fc.WindDeg)

		// Round rain probability to the nearest integer
		rainProb := int64(math.Round(fc.RainProb * 100))

		return any(types.DailyForecastEntity{
			Date:      weatherDate,
			Min:       strconv.FormatFloat(fc.Temp.Min, 'f', -1, 64),
			Max:       strconv.FormatFloat(fc.Temp.Max, 'f', -1, 64),
			Condition: fc.Weather[0].Title,
			Emoji:     emoji,
			FeelsLike: strconv.FormatFloat(fc.FeelsLike.Day, 'f', -1, 64),
			Wind: types.Wind{
				Arrow:     windArrow,
				Direction: windDirection,
				Speed:     strconv.FormatFloat(fc.WindSpeed, 'f', 2, 64),
			},
			RainProb: strconv.FormatInt(rainProb, 10) + "%",
		}).(T)
	case hourlyRes:
		// Format UNIX timestamp as 'YYYY-MM-DD'
		utcTime := time.Unix(int64(fc.Timestamp), 0)
		weatherTime := types.ZephyrTime{Time: utcTime.UTC()}

		// Set condition accordingly to weather condition
		var condition string
		switch fc.Weather[0].Description {
		case "few clouds":
			condition = "SunWithCloud"
		case "broken clouds":
			condition = "CloudWithSun"
		default:
			condition = fc.Weather[0].Title
		}

		// Get emoji from weather condition
		isNight := strings.HasSuffix(fc.Weather[0].Icon, "n")
		emoji := GetEmoji(condition, isNight)

		// Get cardinal direction and wind arrow
		windDirection, windArrow := GetCardinalDir(fc.WindDeg)

		// Round rain probability to the nearest integer
		rainProb := int64(math.Round(fc.RainProb * 100))

		return any(types.HourlyForecastEntity{
			Time:        weatherTime,
			Temperature: strconv.FormatFloat(fc.Temperature, 'f', -1, 64),
			Condition:   fc.Weather[0].Title,
			Emoji:       emoji,
			Wind: types.Wind{
				Arrow:     windArrow,
				Direction: windDirection,
				Speed:     strconv.FormatFloat(fc.WindSpeed, 'f', 2, 64),
			},
			RainProb: strconv.FormatInt(rainProb, 10) + "%",
		}).(T)
	default:
		var zero T
		return zero
	}
}

func GetForecast[T types.DailyForecast | types.HourlyForecast](city *types.City, apiKey string, fcType FCType) (T, error) {
	var forecast T

	baseURL, err := url.Parse(WTR_URL)
	if err != nil {
		var zero T
		return zero, err
	}

	params := baseURL.Query()
	params.Set("lat", strconv.FormatFloat(city.Lat, 'f', -1, 64))
	params.Set("lon", strconv.FormatFloat(city.Lon, 'f', -1, 64))
	params.Set("appid", apiKey)
	params.Set("units", "metric")

	switch fcType {
	case DAILY:
		params.Set("exclude", "current,minutely,hourly,alerts")
		baseURL.RawQuery = params.Encode()

		res, err := http.Get(baseURL.String())
		if err != nil {
			var zero T
			return zero, err
		}
		defer res.Body.Close()

		var dailyRes dailyForecastRes
		if err := json.NewDecoder(res.Body).Decode(&dailyRes); err != nil {
			var zero T
			return zero, err
		}

		// We skip the first element since it represents the current day
		// We also ignore forecasts after the fourth day
		var forecastEntities []types.DailyForecastEntity
		for _, val := range dailyRes.Daily[1:5] {
			forecastEntities = append(forecastEntities, getForecastEntity[types.DailyForecastEntity](val))
		}
		forecast = any(types.DailyForecast{Forecast: forecastEntities}).(T)

	case HOURLY:
		params.Set("exclude", "current,minutely,daily,alerts")
		baseURL.RawQuery = params.Encode()

		res, err := http.Get(baseURL.String())
		if err != nil {
			var zero T
			return zero, err
		}
		defer res.Body.Close()

		var hourlyRes hourlyForecastRes
		if err := json.NewDecoder(res.Body).Decode(&hourlyRes); err != nil {
			var zero T
			return zero, err
		}

		// Get hourly forecast of a time window of 9 hours
		var forecastEntries []types.HourlyForecastEntity
		for _, val := range hourlyRes.Hourly[:9] {
			forecastEntries = append(forecastEntries, getForecastEntity[types.HourlyForecastEntity](val))
		}
		forecast = any(types.HourlyForecast{Forecast: forecastEntries}).(T)
	}

	return any(forecast).(T), nil
}
