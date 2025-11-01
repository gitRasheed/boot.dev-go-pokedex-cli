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
    mu       sync.Mutex
    entries  map[string]cacheEntry
    interval time.Duration
}

func NewCache(interval time.Duration) *Cache {
    c := &Cache{
        entries:  make(map[string]cacheEntry),
        interval: interval,
    }
    go c.reapLoop()
    return c
}

func (c *Cache) Add(key string, val []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.entries[key] = cacheEntry{createdAt: time.Now(), val: append([]byte(nil), val...)}
}

func (c *Cache) Get(key string) ([]byte, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    e, ok := c.entries[key]
    if !ok {
        return nil, false
    }
    // return a copy to avoid external mutation
    return append([]byte(nil), e.val...), true
}

func (c *Cache) reapLoop() {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()
    for range ticker.C {
        cutoff := time.Now().Add(-c.interval)
        c.mu.Lock()
        for k, e := range c.entries {
            if e.createdAt.Before(cutoff) {
                delete(c.entries, k)
            }
        }
        c.mu.Unlock()
    }
}
