package model

import (
	"errors"
	"slices"
	"strconv"

	"github.com/ceticamarco/zephyr/cache"
	"github.com/ceticamarco/zephyr/statistics"
	"github.com/ceticamarco/zephyr/types"
)

func GetStatistics(cityName string, statCache *cache.StatCache) (types.StatResult, error) {
	// Check whether there are sufficient and updated records for the given location
	if statCache.IsKeyInvalid(cityName) {
		return types.StatResult{}, errors.New("insufficient or outdated data to perform statistical analysis")
	}

	// Extract records from the database
	stats := statCache.GetCityStatistics(cityName)
	// Extract temperatures from statistics
	temps := make([]float64, len(stats))
	for idx, stat := range stats {
		temps[idx] = stat.Temperature
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
