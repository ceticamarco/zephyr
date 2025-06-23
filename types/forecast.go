package types

// The ForecastEntity data type, representing the weather forecast
// of a single day
type ForecastEntity struct {
	Date      ZephyrDate `json:"date"`
	Min       string     `json:"min"`
	Max       string     `json:"max"`
	Condition string     `json:"condition"`
	Emoji     string     `json:"emoji"`
	FeelsLike string     `json:"feelsLike"`
	Wind      Wind       `json:"wind"`
}

// The Forecast data type, representing a set of ForecastEntity
type Forecast struct {
    Forecast []ForecastEntity `json:"forecast"`
}
