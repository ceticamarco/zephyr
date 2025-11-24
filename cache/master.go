package cache

import (
	"strings"
	"sync"
	"time"

	"github.com/ceticamarco/zephyr/types"
)

// cacheType, representing the abstract value of a CacheEntity
type cacheType interface {
	types.Weather | types.Metrics | types.Wind | types.DailyForecast | types.HourlyForecast | types.Moon
}

// CacheEntity, representing the value of the cache
type CacheEntity[T cacheType] struct {
	element   T
	timestamp time.Time
}

// MasterCache, representing a mapping between a key(str) and a CacheEntity
type MasterCache[T cacheType] struct {
	mu   sync.RWMutex
	Data map[string]CacheEntity[T]
}

// MasterCaches, representing a grouping of the various caches
type MasterCaches struct {
	WeatherCache        MasterCache[types.Weather]
	MetricsCache        MasterCache[types.Metrics]
	WindCache           MasterCache[types.Wind]
	DailyForecastCache  MasterCache[types.DailyForecast]
	HourlyForecastCache MasterCache[types.HourlyForecast]
	MoonCache           MasterCache[types.Moon]
}

func InitMasterCache() *MasterCaches {
	return &MasterCaches{
		WeatherCache:        MasterCache[types.Weather]{Data: make(map[string]CacheEntity[types.Weather])},
		MetricsCache:        MasterCache[types.Metrics]{Data: make(map[string]CacheEntity[types.Metrics])},
		WindCache:           MasterCache[types.Wind]{Data: make(map[string]CacheEntity[types.Wind])},
		DailyForecastCache:  MasterCache[types.DailyForecast]{Data: make(map[string]CacheEntity[types.DailyForecast])},
		HourlyForecastCache: MasterCache[types.HourlyForecast]{Data: make(map[string]CacheEntity[types.HourlyForecast])},
		MoonCache:           MasterCache[types.Moon]{Data: make(map[string]CacheEntity[types.Moon])},
	}
}

func (cache *MasterCache[T]) GetEntry(cityName string, ttl int8) (T, bool) {
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

func (cache *MasterCache[T]) AddEntry(entry T, cityName string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	currentTime := time.Now()

	cache.Data[strings.ToUpper(cityName)] = CacheEntity[T]{
		element:   entry,
		timestamp: currentTime,
	}
}
