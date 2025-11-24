package model

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ceticamarco/zephyr/types"
)

func GetEmoji(condition string, isNight bool) string {
	switch condition {
	case "Thunderstorm":
		return "⛈️"
	case "Drizzle":
		return "🌦️"
	case "Rain":
		return "🌧️"
	case "Snow":
		return "☃️"
	case "Mist", "Smoke", "Haze", "Dust", "Fog", "Sand", "Ash", "Squall", "Clouds":
		return "☁️"
	case "Tornado":
		return "🌪️"
	case "Clear":
		{
			if isNight {
				return "🌙"
			} else {
				return "☀️"
			}
		}
	case "SunWithCloud":
		return "🌤️"
	case "CloudWithSun":
		return "🌥️"
	}

	return "❓"
}

func GetWeather(city *types.City, apiKey string) (types.Weather, float64, error) {
	url, err := url.Parse(WTR_URL)
	if err != nil {
		return types.Weather{}, 0, err
	}

	params := url.Query()
	params.Set("lat", strconv.FormatFloat(city.Lat, 'f', -1, 64))
	params.Set("lon", strconv.FormatFloat(city.Lon, 'f', -1, 64))
	params.Set("appid", apiKey)
	params.Set("units", "metric")
	params.Set("exclude", "minutely,hourly")

	url.RawQuery = params.Encode()

	res, err := http.Get(url.String())
	if err != nil {
		return types.Weather{}, 0, err
	}
	defer res.Body.Close()

	// Structure representing the *current* weather
	type WeatherRes struct {
		Current struct {
			FeelsLike   float64 `json:"feels_like"`
			Temperature float64 `json:"temp"`
			Timestamp   int64   `json:"dt"`
			Weather     []struct {
				Title       string `json:"main"`
				Description string `json:"description"`
				Icon        string `json:"icon"`
			} `json:"weather"`
		} `json:"current"`
		Daily []struct {
			Temp struct {
				Daily float64 `json:"day"`
				Min   float64 `json:"min"`
				Max   float64 `json:"max"`
			} `json:"temp"`
		} `json:"daily"`
		Alerts []struct {
			Event       string `json:"event"`
			Start       int64  `json:"start"`
			End         int64  `json:"end"`
			Description string `json:"description"`
		} `json:"alerts"`
	}

	var weather WeatherRes
	if err := json.NewDecoder(res.Body).Decode(&weather); err != nil {
		return types.Weather{}, 0, err
	}

	// Format UNIX timestamp as 'YYYY-MM-DD'
	utcTime := time.Unix(int64(weather.Current.Timestamp), 0)
	weatherDate := types.ZephyrDate{Date: utcTime.UTC()}

	// Set condition accordingly to weather description
	var condition string
	switch weather.Current.Weather[0].Description {
	case "few clouds":
		condition = "SunWithCloud"
	case "broken clouds":
		condition = "CloudWithSun"
	default:
		condition = weather.Current.Weather[0].Title
	}

	// Get emoji from weather condition
	isNight := strings.HasSuffix(weather.Current.Weather[0].Icon, "n")
	emoji := GetEmoji(condition, isNight)

	// Format weather alerts
	var alerts []types.WeatherAlert
	for _, alert := range weather.Alerts {
		// Format both start and end timestamp as 'YYYY-MM-DD'
		utcStartDate := time.Unix(int64(alert.Start), 0)
		startDate := types.ZephyrAlertDate{Date: utcStartDate}

		utcEndDate := time.Unix(int64(alert.End), 0)
		endDate := types.ZephyrAlertDate{Date: utcEndDate}

		// Extract the first line of alert description
		eventDescription := strings.Split(alert.Description, "\n")[0]

		alerts = append(alerts, types.WeatherAlert{
			Event:       alert.Event,
			Start:       startDate,
			End:         endDate,
			Description: eventDescription,
		})
	}

	return types.Weather{
		Date:        weatherDate,
		Temperature: strconv.FormatFloat(weather.Current.Temperature, 'f', -1, 64),
		Min:         strconv.FormatFloat(weather.Daily[0].Temp.Min, 'f', -1, 64),
		Max:         strconv.FormatFloat(weather.Daily[0].Temp.Max, 'f', -1, 64),
		FeelsLike:   strconv.FormatFloat(weather.Current.FeelsLike, 'f', -1, 64),
		Condition:   weather.Current.Weather[0].Title,
		Emoji:       emoji,
		Alerts:      alerts,
	}, weather.Daily[0].Temp.Daily, nil
}
