# CacheUtility Software Requirements Specifications (SRS)

## 1. Service Overview

### 1.1 Purpose
CacheUtility provides efficient in-memory caching capabilities for UI components, enabling fast data access and reducing service calls while maintaining data consistency through coordinated invalidation within the EisenKan Client application.

### 1.2 Scope
This utility component encapsulates data caching, TTL management, pattern-based invalidation, and thread-safe operations to support UI responsiveness and minimize redundant service calls across Access layer components.

### 1.3 Context Integration
CacheUtility operates in the Utilities layer of the EisenKan Client architecture, serving as a shared component accessible to Access layer components (TaskManagerAccess, BoardAccess) and potentially Manager layer widgets for data caching needs.

---

## 2. Operations

CacheUtility shall support the following core operations for data caching within the UI context:

### 2.1 Data Storage Operations
- **Set**: Store data with optional TTL expiration
- **Get**: Retrieve cached data with existence check
- **Contains**: Check if key exists without retrieving data
- **Size**: Get current cache size and statistics

### 2.2 Data Invalidation Operations
- **Invalidate**: Remove specific cache entry by key
- **InvalidatePattern**: Remove multiple entries matching pattern
- **Clear**: Remove all cache entries
- **Cleanup**: Remove expired entries

### 2.3 Cache Management Operations
- **SetMaxSize**: Configure maximum cache size
- **SetDefaultTTL**: Set default expiration time
- **GetStats**: Retrieve cache performance statistics

---

## 3. Functional Requirements

1. **REQ-CACHE-001**: When UI components store data, CacheUtility shall accept any serializable data type with configurable TTL
2. **REQ-CACHE-002**: When UI components retrieve data, CacheUtility shall return cached data within 1 millisecond for memory-resident entries
3. **REQ-CACHE-003**: When cache entries expire, CacheUtility shall automatically remove expired entries during access operations
4. **REQ-CACHE-004**: When pattern-based invalidation is requested, CacheUtility shall efficiently remove all matching cache entries
5. **REQ-CACHE-005**: When cache size exceeds configured limits, CacheUtility shall evict least recently used entries
6. **REQ-CACHE-006**: When concurrent access occurs, CacheUtility shall maintain thread safety without data corruption
7. **REQ-CACHE-007**: When cache statistics are requested, CacheUtility shall provide hit ratio, miss ratio, and size information
8. **REQ-CACHE-008**: When manual cleanup is requested, CacheUtility shall remove all expired entries and compact storage

---

## 4. Quality Attributes

### 4.1 Performance Requirements
- **REQ-PERFORMANCE-001**: CacheUtility shall complete Get operations within 1 millisecond for cached entries
- **REQ-PERFORMANCE-002**: CacheUtility shall complete Set operations within 5 milliseconds including TTL setup
- **REQ-PERFORMANCE-003**: CacheUtility shall handle 1000+ concurrent cache operations without performance degradation

### 4.2 Reliability Requirements
- **REQ-RELIABILITY-001**: CacheUtility shall maintain data integrity under concurrent access from multiple goroutines
- **REQ-RELIABILITY-002**: CacheUtility shall gracefully handle memory pressure without system failure
- **REQ-RELIABILITY-003**: CacheUtility shall prevent memory leaks through automatic cleanup of expired entries

### 4.3 Usability Requirements
- **REQ-USABILITY-001**: CacheUtility shall provide simple key-value interface compatible with any data type
- **REQ-USABILITY-002**: CacheUtility shall support intuitive pattern matching for bulk invalidation operations
- **REQ-USABILITY-003**: CacheUtility shall provide clear cache statistics for performance monitoring

---

## 5. Interface Requirements

### 5.1 Service Contract

```go
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
```

### 5.2 Data Contracts

#### CacheStats
```go
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
```

#### CacheEntry (internal)
```go
type CacheEntry struct {
    Value      interface{} `json:"value"`
    ExpiresAt  time.Time   `json:"expires_at"`
    AccessedAt time.Time   `json:"accessed_at"`
    CreatedAt  time.Time   `json:"created_at"`
}
```

### 5.3 Pattern Matching Specification
Pattern matching for `InvalidatePattern` shall support:
- **Wildcard matching**: `tasks_*` matches `tasks_list`, `tasks_summary`, etc.
- **Prefix matching**: `board_` matches all keys starting with `board_`
- **Exact matching**: Keys without wildcards match exactly
- **Multiple patterns**: Comma-separated patterns like `tasks_*,board_*`

---

## 6. Technical Constraints

### 6.1 Architectural Constraints
- **REQ-ARCH-001**: CacheUtility shall operate as a shared Utility component accessible to all client layers
- **REQ-ARCH-002**: CacheUtility shall not contain any business logic and shall provide generic caching capabilities
- **REQ-ARCH-003**: CacheUtility shall be stateless regarding business domain knowledge
- **REQ-ARCH-004**: CacheUtility shall support the "coffee machine UI test" - any UI could use this utility

### 6.2 Technology Constraints
- **REQ-TECH-001**: CacheUtility shall use Go's native synchronization primitives for thread safety
- **REQ-TECH-002**: CacheUtility shall implement memory-efficient storage without external dependencies
- **REQ-TECH-003**: CacheUtility shall support automatic garbage collection integration
- **REQ-TECH-004**: CacheUtility shall be compatible with any Go interface{} data type

### 6.3 Resource Constraints
- **REQ-RESOURCE-001**: CacheUtility shall limit memory usage to configurable maximum size
- **REQ-RESOURCE-002**: CacheUtility shall implement LRU eviction policy when size limits are exceeded
- **REQ-RESOURCE-003**: CacheUtility shall automatically clean up expired entries to prevent memory leaks

---

## 7. Dependencies

### 7.1 Internal Dependencies
- **No Service Dependencies**: CacheUtility is a leaf utility component
- **No Layer Dependencies**: Must not depend on Access, Engine, Manager, or Client layers

### 7.2 External Dependencies
- **Go Standard Library**: sync, time, regexp packages for core functionality
- **Memory Management**: Go garbage collector for automatic memory cleanup

---

**Document Version**: 1.0  
**Created**: 2025-09-14  
**Status**: Accepted