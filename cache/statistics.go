package cache

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ceticamarco/zephyr/types"
)

// statistic cache data type, representing a mapping between a location+date and its daily average temperature
type StatCache struct {
	mu sync.RWMutex
	db map[string]float64
}

func InitStatCache() *StatCache {
	return &StatCache{
		db: make(map[string]float64),
	}
}

func (cache *StatCache) AddStatistic(cityName string, statDate string, dailyTemp float64) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Format key as '<DATE>@<LOCATION>
	key := fmt.Sprintf("%s@%s", statDate, cityName)

	// Insert weather statistic into the database if it doesn't already exist
	if _, exists := cache.db[key]; exists {
		return
	}

	cache.db[key] = dailyTemp
}

func (cache *StatCache) IsKeyInvalid(key string) bool {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	// A key is invalid if it has less than 2 entries within the last 2 days
	threshold := time.Now().AddDate(0, 0, -2)

	var validEntries uint = 0
	for storedKey := range cache.db {
		if !strings.HasSuffix(storedKey, key) {
			continue
		}

		// Get <DATE> from <DATE>@<LOCATION>
		keyDate, err := time.Parse("2006-01-02", strings.Split(storedKey, "@")[0])
		if err != nil {
			keyDate = time.Now() // Add a fallback date if parsing fails
		}

		if !keyDate.Before(threshold) {
			validEntries++

			// Early skip if we already found two valid entries
			if validEntries >= 2 {
				return false
			}
		}
	}

	return true
}

func (cache *StatCache) GetCityStatistics(cityName string) []types.StatElement {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	result := make([]types.StatElement, 0)

	for key, record := range cache.db {
		if strings.HasSuffix(key, cityName) {
			// Get <DATE> from <DATE>@<LOCATION>
			keyDate, err := time.Parse("2006-01-02", strings.Split(key, "@")[0])
			if err != nil {
				keyDate = time.Now() // Add a fallback date if parsing fails
			}

			result = append(result, types.StatElement{
				Temperature: record,
				Date:        keyDate,
			})
		}
	}

	return result
}
