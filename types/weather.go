package types

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
