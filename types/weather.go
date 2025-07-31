package types

// The Weather data type, representing the weather of a certain location
type Weather struct {
	Date        ZephyrDate `json:"date"`
	Temperature string     `json:"temperature"`
	Min         string     `json:"min"`
	Max         string     `json:"max"`
	Condition   string     `json:"condition"`
	FeelsLike   string     `json:"feelsLike"`
	Emoji       string     `json:"emoji"`
}
