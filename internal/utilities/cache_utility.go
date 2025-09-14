// Package utilities provides Utility layer components for the EisenKan system following iDesign methodology.
// This package contains reusable components that provide infrastructure services across all system layers.
package utilities

import (
	"container/list"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// ICacheUtility defines the interface for UI caching operations
type ICacheUtility interface {
	// Data Storage Operations
	Set(key string, value interface{}, ttl time.Duration)
	Get(key string) (interface{}, bool)
	Contains(key string) bool

	// Data Invalidation Operations
	Invalidate(key string)
	InvalidatePattern(pattern string)
	Clear()
	Cleanup()

	// Cache Management Operations
	SetMaxSize(size int)
	SetDefaultTTL(ttl time.Duration)
	GetStats() CacheStats
	Size() int
}

// CacheStats provides cache performance and usage statistics
type CacheStats struct {
	Size         int           `json:"size"`          // Current number of entries
	MaxSize      int           `json:"max_size"`      // Maximum configured size
	HitCount     int64         `json:"hit_count"`     // Total cache hits
	MissCount    int64         `json:"miss_count"`    // Total cache misses
	HitRatio     float64       `json:"hit_ratio"`     // Hit ratio (0.0 - 1.0)
	EvictCount   int64         `json:"evict_count"`   // Total evictions
	MemoryUsage  int64         `json:"memory_usage"`  // Approximate memory usage in bytes
	LastCleanup  time.Time     `json:"last_cleanup"`  // Last cleanup operation time
}

// CacheEntry represents a cached entry with metadata
type cacheEntry struct {
	Value      interface{} `json:"value"`
	ExpiresAt  time.Time   `json:"expires_at"`
	AccessedAt time.Time   `json:"accessed_at"`
	CreatedAt  time.Time   `json:"created_at"`
	listElement *list.Element // LRU list element for O(1) access
}

// CacheUtility implements ICacheUtility with thread-safe in-memory caching
type CacheUtility struct {
	// Core data structures
	data    map[string]*cacheEntry
	lruList *list.List // Most recently used at front, least at back
	mutex   sync.RWMutex

	// Configuration
	maxSize    int
	defaultTTL time.Duration

	// Statistics (atomic for lock-free access during reads)
	hitCount   int64
	missCount  int64
	evictCount int64

	// Background cleanup
	cleanupTicker *time.Ticker
	shutdown      chan struct{}
	cleanupDone   chan struct{}
}

// NewCacheUtility creates a new CacheUtility instance with default configuration
func NewCacheUtility() ICacheUtility {
	cache := &CacheUtility{
		data:          make(map[string]*cacheEntry),
		lruList:       list.New(),
		maxSize:       1000,           // Default max size
		defaultTTL:    5 * time.Minute, // Default TTL
		cleanupTicker: time.NewTicker(30 * time.Second), // Cleanup every 30 seconds
		shutdown:      make(chan struct{}),
		cleanupDone:   make(chan struct{}),
	}

	// Start background cleanup goroutine
	go cache.backgroundCleanup()

	return cache
}

// Set stores data with optional TTL expiration
func (c *CacheUtility) Set(key string, value interface{}, ttl time.Duration) {
	if key == "" {
		return // Ignore empty keys
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiresAt := now.Add(ttl)
	if ttl <= 0 {
		expiresAt = now.Add(c.defaultTTL)
	}

	// Check if key already exists
	if existing, exists := c.data[key]; exists {
		// Update existing entry
		existing.Value = value
		existing.ExpiresAt = expiresAt
		existing.AccessedAt = now
		// Move to front of LRU list
		c.lruList.MoveToFront(existing.listElement)
		return
	}

	// Create new entry
	entry := &cacheEntry{
		Value:      value,
		ExpiresAt:  expiresAt,
		AccessedAt: now,
		CreatedAt:  now,
	}

	// Add to LRU list (front = most recently used)
	entry.listElement = c.lruList.PushFront(key)
	c.data[key] = entry

	// Enforce size limit with LRU eviction
	c.enforceMaxSize()
}

// Get retrieves cached data with existence check
func (c *CacheUtility) Get(key string) (interface{}, bool) {
	if key == "" {
		atomic.AddInt64(&c.missCount, 1)
		return nil, false
	}

	c.mutex.RLock()
	entry, exists := c.data[key]
	c.mutex.RUnlock()

	if !exists {
		atomic.AddInt64(&c.missCount, 1)
		return nil, false
	}

	// Check expiration
	now := time.Now()
	if now.After(entry.ExpiresAt) {
		// Entry expired, remove it
		c.Invalidate(key)
		atomic.AddInt64(&c.missCount, 1)
		return nil, false
	}

	// Update access time and move to front of LRU list
	c.mutex.Lock()
	entry.AccessedAt = now
	c.lruList.MoveToFront(entry.listElement)
	c.mutex.Unlock()

	atomic.AddInt64(&c.hitCount, 1)
	return entry.Value, true
}

// Contains checks if key exists without retrieving data
func (c *CacheUtility) Contains(key string) bool {
	_, exists := c.Get(key)
	return exists
}

// Invalidate removes specific cache entry by key
func (c *CacheUtility) Invalidate(key string) {
	if key == "" {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entry, exists := c.data[key]; exists {
		// Remove from LRU list
		c.lruList.Remove(entry.listElement)
		// Remove from map
		delete(c.data, key)
	}
}

// InvalidatePattern removes multiple entries matching pattern
func (c *CacheUtility) InvalidatePattern(pattern string) {
	if pattern == "" {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var keysToRemove []string

	// Find all matching keys
	for key := range c.data {
		matched, err := filepath.Match(pattern, key)
		if err == nil && matched {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove matching entries
	for _, key := range keysToRemove {
		if entry, exists := c.data[key]; exists {
			c.lruList.Remove(entry.listElement)
			delete(c.data, key)
		}
	}
}

// Clear removes all cache entries
func (c *CacheUtility) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*cacheEntry)
	c.lruList = list.New()
}

// Cleanup removes expired entries
func (c *CacheUtility) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	var keysToRemove []string

	// Find expired entries
	for key, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove expired entries
	for _, key := range keysToRemove {
		if entry, exists := c.data[key]; exists {
			c.lruList.Remove(entry.listElement)
			delete(c.data, key)
		}
	}
}

// SetMaxSize configures maximum cache size
func (c *CacheUtility) SetMaxSize(size int) {
	if size <= 0 {
		return // Ignore invalid sizes
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.maxSize = size
	c.enforceMaxSize()
}

// SetDefaultTTL sets default expiration time
func (c *CacheUtility) SetDefaultTTL(ttl time.Duration) {
	if ttl <= 0 {
		return // Ignore invalid TTL
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.defaultTTL = ttl
}

// GetStats retrieves cache performance statistics
func (c *CacheUtility) GetStats() CacheStats {
	c.mutex.RLock()
	size := len(c.data)
	maxSize := c.maxSize
	c.mutex.RUnlock()

	hitCount := atomic.LoadInt64(&c.hitCount)
	missCount := atomic.LoadInt64(&c.missCount)
	evictCount := atomic.LoadInt64(&c.evictCount)

	var hitRatio float64
	if hitCount+missCount > 0 {
		hitRatio = float64(hitCount) / float64(hitCount+missCount)
	}

	return CacheStats{
		Size:        size,
		MaxSize:     maxSize,
		HitCount:    hitCount,
		MissCount:   missCount,
		HitRatio:    hitRatio,
		EvictCount:  evictCount,
		MemoryUsage: int64(size * 256), // Rough estimate: 256 bytes per entry
		LastCleanup: time.Now(),        // Simplified for now
	}
}

// Size returns current cache size
func (c *CacheUtility) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}

// enforceMaxSize evicts least recently used entries when size limit exceeded
// Must be called with write lock held
func (c *CacheUtility) enforceMaxSize() {
	for len(c.data) > c.maxSize {
		// Remove least recently used (back of list)
		if back := c.lruList.Back(); back != nil {
			key := back.Value.(string)
			c.lruList.Remove(back)
			delete(c.data, key)
			atomic.AddInt64(&c.evictCount, 1)
		} else {
			break // Should not happen, but prevent infinite loop
		}
	}
}

// backgroundCleanup runs periodic cleanup of expired entries
func (c *CacheUtility) backgroundCleanup() {
	defer close(c.cleanupDone)

	for {
		select {
		case <-c.cleanupTicker.C:
			c.Cleanup()
		case <-c.shutdown:
			c.cleanupTicker.Stop()
			return
		}
	}
}

// Shutdown gracefully stops the cache utility
func (c *CacheUtility) Shutdown() {
	close(c.shutdown)
	<-c.cleanupDone
}