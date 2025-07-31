package model

import (
	"errors"
	"slices"
	"strconv"

	"github.com/ceticamarco/zephyr/statistics"
	"github.com/ceticamarco/zephyr/types"
)

func GetStatistics(cityName string, statDB *types.StatDB) (types.StatResult, error) {
	// Check whether there are sufficient and updated records for the given location
	if statDB.IsKeyInvalid(cityName) {
		return types.StatResult{}, errors.New("insufficient or outdated data to perform statistical analysis")
	}

	extractTemps := func(weatherArr []types.Weather) ([]float64, error) {
		temps := make([]float64, 0, len(weatherArr))

		for _, weather := range weatherArr {
			temperature, err := strconv.ParseFloat(weather.Temperature, 64)
			if err != nil {
				return nil, err
			}
			temps = append(temps, temperature)
		}

		return temps, nil
	}

	// Extract records from the database
	stats := statDB.GetCityStatistics(cityName)

	// Extract temperatures from weather statistics
	temps, err := extractTemps(stats)
	if err != nil {
		return types.StatResult{}, err
	}

	// Detect anomalies
	anomalies := statistics.DetectAnomalies(stats)
	if len(anomalies) == 0 {
		anomalies = nil
	}

	// Compute statistics
	return types.StatResult{
		Min:     strconv.FormatFloat(slices.Min(temps), 'f', -1, 64),
		Max:     strconv.FormatFloat(slices.Max(temps), 'f', -1, 64),
		Count:   len(stats),
		Mean:    strconv.FormatFloat(statistics.Mean(temps), 'f', -1, 64),
		StdDev:  strconv.FormatFloat(statistics.StdDev(temps), 'f', -1, 64),
		Median:  strconv.FormatFloat(statistics.Median(temps), 'f', -1, 64),
		Mode:    strconv.FormatFloat(statistics.Mode(temps), 'f', -1, 64),
		Anomaly: &anomalies,
	}, nil
}
