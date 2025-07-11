package types

// The Metrics data type, representing the humidity, pressure and
// similar miscellaneous values
type Metrics struct {
	Humidity   string `json:"humidity"`
	Pressure   string `json:"pressure"`
	DewPoint   string `json:"dewPoint"`
	UvIndex    string `json:"uvIndex"`
	Visibility string `json:"visibility"`
}
