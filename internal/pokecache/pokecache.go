package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu           sync.Mutex
	cacheEntries map[string]cacheEntry
	timer        *time.Ticker
	done         chan bool
}

func NewCache(interval time.Duration) *Cache {
	var c *Cache
	c = &Cache{
		mu:           sync.Mutex{},
		cacheEntries: map[string]cacheEntry{},
		done:         make(chan bool),
		timer:        time.NewTicker(interval),
	}

	go func() {
		for {
			select {
			case t := <-c.timer.C:
				c.mu.Lock()
				for key, entry := range c.cacheEntries {
					if t.Sub(entry.createdAt) >= interval {
						delete(c.cacheEntries, key)
					}
				}
				c.mu.Unlock()
			}
		}
	}()

	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheEntries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.cacheEntries) == 0 {
		return []byte{}, false
	}
	entry, ok := c.cacheEntries[key]
	return entry.val, ok
}
