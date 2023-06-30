package cache

import (
	"sync"
	"time"

	v1 "github.com/cityhunteur/weather-service/api/v1"
)

const (
	defaultExpiry   time.Duration = 5 * time.Hour
	defaultCapacity               = 100
)

// Store is an in-memory cache backed by a map.
type Store struct {
	data map[string]*v1.Forecast

	mu sync.RWMutex
}

// NewStore creates a new Store.
func NewStore() *Store {
	return &Store{
		data: make(map[string]*v1.Forecast, defaultCapacity),
	}
}

// Set adds the given value v to the cache using the specified k.
func (c *Store) Set(k string, v *v1.Forecast) {
	c.mu.Lock()
	c.data[k] = v
	c.mu.Unlock()
}

// Get return the value with the specified k if it exists and has not expired.
func (c *Store) Get(k string) (*v1.Forecast, bool) {
	c.mu.RLock()
	entry, found := c.data[k]
	if !found {
		c.mu.RUnlock()
		return nil, false
	}
	if time.Time(entry.Detail[0].StartTime).UTC().Sub(time.Now().UTC()) > defaultExpiry {
		c.mu.RUnlock()
		return nil, false
	}
	c.mu.RUnlock()
	return entry, true
}
