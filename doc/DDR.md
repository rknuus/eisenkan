# Design Decision Records (DDR)

## [2025-09-06] - LoggingUtility Design Decision: Structured Data Serialization Strategy

**Decision**: Type Switch + Interface-based Approach (Option B)

**Context**: The SRS requires support for arbitrary Go types with preserved type information and human-readable narratives. Need to choose between reflection-based, type-switch-based, or hybrid approaches.

**Options Considered**: 
- Option A: Reflection-based JSON + Custom Formatting
- Option B: Type Switch + Interface-based Approach  
- Option C: Hybrid Approach

**Rationale**: Choose Option B to keep implementation simple as a starting point. Type switches provide better performance for common types, and interface-based approach allows extensibility. Fallback to fmt.Sprintf ensures all types are handled, even if not optimally.

**Consequences**: 
- Better performance for common logging scenarios
- Simpler implementation and maintenance
- May require future enhancement for complex edge cases
- Interface adoption needed for optimal custom type logging

**User Approval**: [User] approved on [2025-09-06]

## [2025-09-06] - LoggingUtility Design Decision: Circular Reference Handling

**Decision**: Depth Limiting Only (Option B)

**Context**: STP requires handling of self-referencing structures without infinite loops. SRS specifies 5-level depth limit (REQ-FORMAT-003).

**Options Considered**:
- Option A: Visited Map with Pointer Tracking
- Option B: Depth Limiting Only
- Option C: Visited Map + Depth Limiting

**Rationale**: Choose Option B to keep implementation simple and directly meet the "5 levels" requirement from SRS. Depth limiting prevents infinite loops effectively while maintaining simple implementation.

**Consequences**:
- Simple implementation without pointer tracking overhead
- Directly satisfies SRS depth requirement
- May not detect circular references at shallow depths (acceptable trade-off)
- Deterministic truncation behavior

**User Approval**: [User] approved on [2025-09-06]

## [2025-09-06] - LoggingUtility Design Decision: Error Handling Strategy

**Decision**: Panic on Internal Failures

**Context**: Need to determine how to handle internal logging failures (file system errors, configuration issues, etc.).

**Options Considered**:
- Option A: Silent Failure with Internal Error Tracking
- Option B: Best-Effort with Fallback
- Option C: Error Return with Graceful Degradation
- Option D: Panic on Internal Failures

**Rationale**: User decision to use fail-fast behavior. Panicking on internal failures provides clear failure indication and simpler implementation. Logging failures are typically configuration/environment issues that should be addressed immediately. This is in accordance to **REQ-RELIABILITY-001**: If log output fails, then the LoggingUtility shall crash the application.

**Consequences**:
- Simpler error handling implementation
- Clear failure signals for debugging
- Callers must handle potential panics or fix logging configuration
- Removes requirement for complex fallback logic

**User Approval**: [User] approved on [2025-09-06]

## [2025-09-06] - LoggingUtility Design Decision: Interface Design Revision

**Decision**: 3-Operation Interface with Extended Log Method

**Context**: Original SRS suggestion specified 4 operations including separate LogWithStructuredData. User wants to simplify to 3 operations.

**Original Proposed Interface**:
- LogWithStructuredData(level, context, data)
- Log(level, component, message) 
- LogError(level, component, error, context)
- IsLevelEnabled(level)

**Revised Interface Decision**:
- Log(level, component, message, data interface{}) - Extended with data parameter
- LogError(component, error, data interface{}) - Removed level parameter, always logs at Error level
- IsLevelEnabled(level) - Unchanged

**Rationale**: User wants to keep it simple with only 3 operations. Extending Log() with data parameter allows arbitrary data logging without separate method. This removes the need for StructuredLogContext structure.

**Consequences**:
- Simpler interface with fewer methods
- Single Log method handles both simple and structured logging
- No need for separate StructuredLogContext structure
- May need to handle nil data parameter gracefully
- **SRS Update Required**: Interface contract section needs revision

**User Approval**: [User] approved on [2025-09-06]

## [2025-09-07] - VersioningUtility Design Decision: Repository Handle Pattern

**Decision**: Repository Handle Pattern (Option C)

**Context**: Need to choose repository object management approach for go-git integration, balancing performance and simplicity for 7 interface operations.

**Options Considered**:
- Option A: Repository Instance Per Operation - Simple but high overhead
- Option B: Repository Caching with Lifecycle Management - Complex state management
- Option C: Repository Handle Pattern - Explicit lifecycle control

**Rationale**: Repository Handle Pattern provides optimal performance for multi-operation workflows while giving callers explicit control over repository lifecycle. This aligns with the performance requirements (REQ-PERFORMANCE-001) and supports efficient resource management.

**Consequences**:
- Interface returns handles for multi-operation scenarios
- Callers manage repository lifecycle explicitly
- Better performance for batch operations
- Slight increase in interface complexity
- Requires careful handle cleanup in error scenarios

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-07] - VersioningUtility Design Decision: Direct Error Passthrough with Context

**Decision**: Direct go-git Error Passthrough with Annotations (Option C)

**Context**: Need structured error information (REQ-RELIABILITY-001) while integrating with go-git error handling.

**Options Considered**:
- Option A: Error Wrapping with Context - Custom error structures
- Option B: Error Translation to Domain Errors - Abstract away go-git errors
- Option C: Direct go-git Error Passthrough with Annotations - Preserve original errors with context

**Rationale**: Direct passthrough preserves all go-git error information while adding necessary context. This provides maximum debugging information and maintains compatibility with go-git error handling patterns.

**Consequences**:
- Rich error information preserved from go-git
- Context annotations provide operation and path information
- Callers can handle specific go-git error types if needed
- Error messages include full chain of context
- Maintains compatibility with Go error handling idioms

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-07] - VersioningUtility Design Decision: Selective Staging with Patterns

**Decision**: Selective Staging with Patterns (Option B)

**Context**: REQ-VERSION-003 requires staging "all modifications" but need flexibility for different staging scenarios.

**Options Considered**:
- Option A: Stage All Changes - Simple but inflexible
- Option B: Selective Staging with Patterns - Granular control with patterns
- Option C: Smart Staging with Conflict Detection - Complex logic

**Rationale**: Selective staging with patterns provides flexibility for different use cases while maintaining simplicity. Default pattern can stage all files, but callers can specify patterns for selective staging when needed.

**Consequences**:
- Interface supports both "stage all" and selective staging
- Pattern-based approach familiar to git users
- More flexible than simple stage-all approach
- Requires pattern validation and error handling
- Default behavior stages all changes for simplicity

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-07] - VersioningUtility Design Decision: Lazy Loading with Limits Plus Streaming

**Decision**: Combined Lazy Loading with Limits and Streaming Results (Hybrid Option A+C)

**Context**: REQ-PERFORMANCE-001 requires 5-second completion for repositories with 10,000 commits, need scalable approach for large repositories.

**Options Considered**:
- Option A: Lazy Loading with Limits - Good for bounded results
- Option B: Caching with Background Updates - Complex state management
- Option C: Streaming Results - Good for large result sets
- Hybrid: Combine A+C for optimal flexibility

**Rationale**: Combining lazy loading with streaming provides both immediate responsiveness for small requests and scalability for large ones. Interface can provide both synchronous (limited) and asynchronous (streaming) access patterns.

**Consequences**:
- Synchronous methods with limits for simple use cases
- Streaming methods for large result sets
- Optimal memory usage for different scenarios
- Dual interface approach requires careful design
- Performance meets requirements under various loads

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-07] - VersioningUtility Design Decision: Per-Repository Mutex Locking

**Decision**: Per-Repository Mutex Locking (Option A)

**Context**: Need thread-safe operations for concurrent access to repositories, as go-git repositories may not be inherently thread-safe.

**Options Considered**:
- Option A: Per-Repository Mutex Locking - Fine-grained locking by path
- Option B: Operation-Level Locking - Coarser locking approach
- Option C: go-git Native Concurrency - Rely on library thread safety

**Rationale**: Per-repository mutex locking provides optimal concurrency by allowing operations on different repositories to proceed simultaneously while protecting individual repositories from concurrent modifications.

**Consequences**:
- Maximum concurrency for multi-repository scenarios
- Path-based mutex map requires memory management
- Deadlock prevention through consistent lock ordering
- Lock cleanup needed for unused repositories
- Thread-safe access to repository operations

**User Approval**: [User] approved on [2025-09-07]