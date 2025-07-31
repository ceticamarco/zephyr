package model

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/ceticamarco/zephyr/types"
)

func GetCoordinates(cityName string, apiKey string) (types.City, error) {
	url, err := url.Parse(GEO_URL)
	if err != nil {
		return types.City{}, err
	}

	params := url.Query()
	params.Set("q", cityName)
	params.Set("limit", "1")
	params.Set("appid", apiKey)

	url.RawQuery = params.Encode()

	res, err := http.Get(url.String())
	if err != nil {
		return types.City{}, err
	}
	defer res.Body.Close()

	var geoArr []types.City
	if err := json.NewDecoder(res.Body).Decode(&geoArr); err != nil {
		return types.City{}, err
	}

	if len(geoArr) == 0 {
		return types.City{}, errors.New("cannot find this city")
	}

	return types.City{
		Name: geoArr[0].Name,
		Lat:  geoArr[0].Lat,
		Lon:  geoArr[0].Lon,
	}, nil
}
