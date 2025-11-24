package types

import "time"

// Variables type, representing values read from environment variables
type Variables struct {
	Token      string
	TimeToLive int8
}

// The City data type, representing the name, the latitude and the longitude
// of a location
type City struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// The DailyForecastEntity data type, representing the weather forecast
// of a single day
type DailyForecastEntity struct {
	Date      ZephyrDate `json:"date"`
	Min       string     `json:"min"`
	Max       string     `json:"max"`
	Condition string     `json:"condition"`
	Emoji     string     `json:"emoji"`
	FeelsLike string     `json:"feelsLike"`
	Wind      Wind       `json:"wind"`
	RainProb  string     `json:"rainProbability"`
}

// The DailyForecast data type, representing a set of DailyForecastEntity
type DailyForecast struct {
	Forecast []DailyForecastEntity `json:"forecast"`
}

// The HourlyForecastEntity data type, representing the weather forecast
// of a single hour
type HourlyForecastEntity struct {
	Time        ZephyrTime `json:"time"`
	Temperature string     `json:"temperature"`
	Condition   string     `json:"condition"`
	Emoji       string     `json:"emoji"`
	Wind        Wind       `json:"wind"`
	RainProb    string     `json:"rainProbability"`
}

// The HourlyForecast data type, representing a set of HourlyForecastEntity
type HourlyForecast struct {
	Forecast []HourlyForecastEntity `json:"forecast"`
}

// The Metrics data type, representing the humidity, pressure and
// similar miscellaneous values
type Metrics struct {
	Humidity   string `json:"humidity"`
	Pressure   string `json:"pressure"`
	DewPoint   string `json:"dewPoint"`
	UvIndex    string `json:"uvIndex"`
	Visibility string `json:"visibility"`
}

// The Moon data type, representing the moon phase,
// the moon phase icon and the moon progress(%).
type Moon struct {
	Icon       string `json:"icon"`
	Phase      string `json:"phase"`
	Percentage string `json:"percentage"`
}

// The WeatherAnomaly data type, representing
// skewed meteorological events
type WeatherAnomaly struct {
	Date ZephyrDate `json:"date"`
	Temp string     `json:"temperature"`
}

// The StateElement data type, representing a statistical record
// This type is for internal usage
type StatElement struct {
	Temperature float64
	Date        time.Time
}

// The StatResult data type, representing weather statistics
// of past meteorological events
type StatResult struct {
	Min     string            `json:"min"`
	Max     string            `json:"max"`
	Count   int               `json:"count"`
	Mean    string            `json:"mean"`
	StdDev  string            `json:"stdDev"`
	Median  string            `json:"median"`
	Mode    string            `json:"mode"`
	Anomaly *[]WeatherAnomaly `json:"anomaly"`
}

// The WeatherAlert data type, representing a
// weather alert
type WeatherAlert struct {
	Event       string          `json:"event"`
	Start       ZephyrAlertDate `json:"startDate"`
	End         ZephyrAlertDate `json:"endDate"`
	Description string          `json:"description"`
}

// The Weather data type, representing the weather of a certain location
type Weather struct {
	Date        ZephyrDate     `json:"date"`
	Temperature string         `json:"temperature"`
	Min         string         `json:"min"`
	Max         string         `json:"max"`
	Condition   string         `json:"condition"`
	FeelsLike   string         `json:"feelsLike"`
	Emoji       string         `json:"emoji"`
	Alerts      []WeatherAlert `json:"alerts"`
}

// The Wind data type, representing the wind of a certain location
type Wind struct {
	Arrow     string `json:"arrow"`
	Direction string `json:"direction"`
	Speed     string `json:"speed"`
}
