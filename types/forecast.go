package types

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
