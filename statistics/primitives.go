package statistics

import (
	"math"
	"slices"
	"strconv"

	"github.com/ceticamarco/zephyr/types"
)

func Mean(temperatures []float64) float64 {
	if len(temperatures) == 0 {
		return 0
	}

	var sum float64

	for _, val := range temperatures {
		sum += val
	}

	return sum / float64(len(temperatures))
}

func StdDev(temperatures []float64) float64 {
	if len(temperatures) == 0 {
		return 0
	}

	mean := Mean(temperatures)

	var variance float64

	for _, val := range temperatures {
		variance += math.Pow((val - mean), 2)
	}

	variance /= float64(len(temperatures))

	return math.Sqrt(variance)
}

func Median(temperatures []float64) float64 {
	if len(temperatures) == 0 {
		return 0
	}

	// Sort the array without mutating the original values
	sortedTemps := slices.Clone(temperatures)
	slices.Sort(sortedTemps)

	length := len(sortedTemps)
	midValue := length / 2

	if length%2 == 0 {
		return (sortedTemps[midValue-1] + sortedTemps[midValue]) / 2
	} else {
		return sortedTemps[midValue]
	}
}

// This method will always returns the largest mode
// on a multi-modal dataset
func Mode(temperatures []float64) float64 {
	if len(temperatures) == 0 {
		return 0
	}

	// Sort the array without mutating the original values
	sortedTemps := slices.Clone(temperatures)
	slices.Sort(sortedTemps)

	frequencies := make(map[float64]int)
	for _, val := range sortedTemps {
		frequencies[val]++
	}

	var mode float64 = 0
	var maxFreq int = 0

	for val, freq := range frequencies {
		if freq > maxFreq || (freq == maxFreq && val > mode) {
			mode = val
			maxFreq = freq
		}
	}

	return mode
}

// Detects statistical anomalies using the Robust Z-Score algorithm
//
// This method is based on the median and the Median Absolute Deviation(MAD),
// making it more robust to anomalies than the standard z-score which uses the arithmetical mean
// and standard deviation
//
// A value is considered an anomaly if its modified z-score exceeds a fixed threshold(4.5)
// and whether the absolute deviation surpasses another fixed parameter(8 degrees).
// These constants have been fine-tuned to work well with the weather data of a wide range of climates
// and to ignore daily temperature fluctuations while still being able to detect significant anomalies.
//
// The scaling constant Φ⁻¹(0.75) ≈ 0.6745 adjusts the MAD to be comparable to the standard deviation
// under the assumption of normal distribution (i.e. 75% of values lie within ~0.6745 standard deviations
// of the median).
//
// Daily temperatures collected over a short time window(1/2 months, but not less than a few days)
// *should* be normally distributed. This algorithm only work under this assumption.
func RobustZScore(temperatures []float64) []struct {
	Idx   int
	Value float64
} {
	const threshold = 4.5    // threshold for MAD ZScore algorithms
	const scale = 0.6745     // Φ⁻¹(3/4) ≈ 0.6745
	const minDeviation = 8.0 // outliers must deviate at least 8°C from the median
	const epsilon = 1e-10

	med := Median(temperatures)
	absDevs := make([]float64, len(temperatures))
	for idx, val := range temperatures {
		absDevs[idx] = math.Abs(val - med)
	}

	madAbsDev := Median(absDevs)
	if madAbsDev < epsilon {
		return nil
	}

	var anomalies []struct {
		Idx   int
		Value float64
	}

	for idx, val := range temperatures {
		z := scale * (val - med) / madAbsDev

		if math.Abs(z) > threshold && math.Abs(val-med) >= minDeviation {
			anomalies = append(anomalies, struct {
				Idx   int
				Value float64
			}{
				Idx:   idx,
				Value: val,
			})
		}
	}

	return anomalies
}

func DetectAnomalies(weatherArr []types.Weather) []types.WeatherAnomaly {
	temps := make([]float64, len(weatherArr))

	for idx, weather := range weatherArr {
		temp, _ := strconv.ParseFloat(weather.Temperature, 64)
		temps[idx] = temp
	}

	// Apply the Robust/MAD Z-Score anomaly detection algorithm
	anomalies := RobustZScore(temps)
	result := make([]types.WeatherAnomaly, 0, len(anomalies))
	for _, anomaly := range anomalies {
		result = append(result, types.WeatherAnomaly{
			Date: weatherArr[anomaly.Idx].Date,
			Temp: strconv.FormatFloat(anomaly.Value, 'f', -1, 64),
		})
	}

	return result
}
