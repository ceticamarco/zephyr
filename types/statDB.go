package types

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// StatDB data type, representing a mapping between a location and its weather
type StatDB struct {
	mu sync.RWMutex
	db map[string]Weather
}

func InitDB() *StatDB {
	return &StatDB{
		db: make(map[string]Weather),
	}
}

func (statDB *StatDB) AddStatistic(cityName string, weather Weather) {
	statDB.mu.Lock()
	defer statDB.mu.Unlock()

	key := fmt.Sprintf("%s@%s", weather.Date.Date.Format("2006-01-02"), cityName)

	// Insert weather statistic into the database only if it isn't present
	if _, isPresent := statDB.db[key]; isPresent {
		return
	}

	statDB.db[key] = weather
}

func (statDB *StatDB) IsKeyInvalid(key string) bool {
	statDB.mu.RLock()
	defer statDB.mu.RUnlock()

	// A key is invalid if it has less than 2 entries within the last 2 days
	threshold := time.Now().AddDate(0, 0, -2)

	var validKeys uint = 0
	for storedKey, record := range statDB.db {
		if !strings.HasSuffix(storedKey, key) {
			continue
		}

		if !record.Date.Date.Before(threshold) {
			validKeys++

			// Early skip if we already found two valid keys
			if validKeys >= 2 {
				return false
			}
		}
	}

	return true
}

func (statDB *StatDB) GetCityStatistics(cityName string) []Weather {
	statDB.mu.RLock()
	defer statDB.mu.RUnlock()

	result := make([]Weather, 0)

	for key, record := range statDB.db {
		if strings.HasSuffix(key, cityName) {
			result = append(result, record)
		}
	}

	return result
}
