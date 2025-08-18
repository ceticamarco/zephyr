package types

import (
	"strings"
	"sync"
	"time"
)

// cacheType, representing the abstract value of a CacheEntity
type cacheType interface {
	Weather | Metrics | Wind | DailyForecast | HourlyForecast | Moon
}

// CacheEntity, representing the value of the cache
type CacheEntity[T cacheType] struct {
	element   T
	timestamp time.Time
}

// Cache, representing a mapping between a key(str) and a CacheEntity
type Cache[T cacheType] struct {
	mu   sync.RWMutex
	Data map[string]CacheEntity[T]
}

// Caches, representing a grouping of the various caches
type Caches struct {
	WeatherCache        Cache[Weather]
	MetricsCache        Cache[Metrics]
	WindCache           Cache[Wind]
	DailyForecastCache  Cache[DailyForecast]
	HourlyForecastCache Cache[HourlyForecast]
	MoonCache           Cache[Moon]
}

func InitCache() *Caches {
	return &Caches{
		WeatherCache:        Cache[Weather]{Data: make(map[string]CacheEntity[Weather])},
		MetricsCache:        Cache[Metrics]{Data: make(map[string]CacheEntity[Metrics])},
		WindCache:           Cache[Wind]{Data: make(map[string]CacheEntity[Wind])},
		DailyForecastCache:  Cache[DailyForecast]{Data: make(map[string]CacheEntity[DailyForecast])},
		HourlyForecastCache: Cache[HourlyForecast]{Data: make(map[string]CacheEntity[HourlyForecast])},
		MoonCache:           Cache[Moon]{Data: make(map[string]CacheEntity[Moon])},
	}
}

func (cache *Cache[T]) GetEntry(cityName string, ttl int8) (T, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	val, isPresent := cache.Data[strings.ToUpper(cityName)]

	// If key is not present, return a zero value
	if !isPresent {
		return val.element, false
	}

	// Otherwise check whether cache element is expired
	currentTime := time.Now()
	expired := currentTime.Sub(val.timestamp) > (time.Duration(ttl) * time.Hour)
	if expired {
		return val.element, false
	}

	return val.element, true
}

func (cache *Cache[T]) AddEntry(entry T, cityName string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	currentTime := time.Now()

	cache.Data[strings.ToUpper(cityName)] = CacheEntity[T]{
		element:   entry,
		timestamp: currentTime,
	}
}
