# Design Decision Records (DDR)

## [2025-09-12] - RulesAccess Design Decision: Concurrent Access Strategy

**Decision**: Option C - VersioningUtility-Level Coordination

**Context**: RulesAccess must handle concurrent read/write operations safely while maintaining data consistency and performance.

**Options Considered**:
- **Option A: File-Level Locking**
  - Use file locks (flock) for rule file access
  - OS-level coordination across processes
  - Simple implementation
  - Potential performance bottleneck
  
- **Option B: In-Memory Mutex with Caching**
  - Mutex per directory path
  - Cache rule sets in memory with TTL
  - Better performance for read-heavy workloads
  - Memory usage and cache consistency concerns
  
- **Option C: VersioningUtility-Level Coordination**
  - Rely on VersioningUtility for concurrency control
  - Atomic commit operations handle conflicts
  - Consistent with other ResourceAccess components
  - Version control overhead for all operations

**Rationale**: Choose Option C for architectural consistency and leveraging existing infrastructure. VersioningUtility already provides atomic operations and conflict detection. This approach maintains consistency with BoardAccess and other ResourceAccess components while providing proper concurrency control through version control mechanisms.

**Consequences**:
- Consistent with other ResourceAccess layer components
- Leverages existing VersioningUtility concurrency control
- Atomic operations and conflict detection built-in
- Version control overhead for all operations (acceptable trade-off)
- Simplified RulesAccess implementation by delegating concurrency to VersioningUtility

**User Approval**: Approved on [2025-09-12]

## [2025-09-12] - RulesAccess Design Decision: Rule Validation Architecture

**Decision**: Option A - Embedded Schema Validation

**Context**: RulesAccess must validate rule syntax, semantics, dependencies, and conflicts. Need to determine validation architecture and extensibility approach.

**Options Considered**:
- **Option A: Embedded Schema Validation**
  - JSON schema validation built into RulesAccess
  - Schema defined as Go structs with validation tags
  - Simple implementation, fast validation
  - Schema changes require code changes
  
- **Option B: External Schema File**
  - JSON schema stored as separate file (rules-schema.json)
  - Runtime schema loading and validation
  - Schema updates without code changes
  - More complex validation logic
  
- **Option C: Plugin-Based Validation**
  - Extensible validation interface for different rule types
  - Support for custom validators per workflow methodology
  - Maximum flexibility for future extensions
  - Complex implementation and testing

**Rationale**: Choose Option A for simplicity and performance. Embedded schema validation using Go structs with validation tags provides fast, compile-time safety and straightforward implementation. Schema changes requiring code changes is acceptable trade-off for initial implementation.

**Consequences**:
- Fast validation performance with compile-time safety
- Simple implementation and testing
- Schema changes require code updates and recompilation
- Good starting point that can be enhanced later if needed
- Direct integration with Go type system

**User Approval**: Approved on [2025-09-12]

## [2025-09-12] - RulesAccess Design Decision: Rule Storage Structure

**Decision**: Option A - Single rules.json File

**Context**: RulesAccess needs to store rule sets for board directories in JSON format with version control. Need to determine the file organization and naming strategy within directories.

**Options Considered**:
- **Option A: Single rules.json File**
  - Store entire rule set in one `rules.json` file per directory
  - Simple atomic replacement for rule set changes
  - Easy to read/write complete rule set
  - Version control tracks entire rule set changes
  
- **Option B: Multiple Rule Category Files**
  - Separate files: `validation-rules.json`, `workflow-rules.json`, `automation-rules.json`, `notification-rules.json`
  - Granular version control per rule category
  - Smaller files for specific rule types
  - More complex atomic replacement logic
  
- **Option C: Individual Rule Files**
  - One file per rule: `rule-{id}.json`
  - Maximum granular version control
  - Complex rule set assembly and validation
  - Contradicts SRS requirement for atomic rule set operations

**Rationale**: Choose Option A to avoid coordination conflicts with BoardAccess. Originally considered storing rules in board.json, but that would require BoardAccess and RulesAccess to coordinate file access. A separate rules.json file provides clean separation of concerns while maintaining atomic rule set operations required by SRS.

**Consequences**:
- Clean separation between board configuration and rule data
- No coordination required with BoardAccess for file access
- Simple atomic rule set replacement implementation
- Version control tracks complete rule set changes as single units
- Easy to implement and maintain

**User Approval**: Approved on [2025-09-12]

## [2025-09-11] - BoardAccess Design Decision: Concurrency and Thread Safety Strategy

**Decision**: Service-Level Mutex Protection (Option A)  

**Context**: SRS requires concurrent operations without data corruption (REQ-PERFORMANCE-002) and data consistency under simultaneous operations (REQ-RELIABILITY-002). VersioningUtility provides repository-level locking, but BoardAccess needs operation-level coordination.

**Options Considered**:

### Option A: Service-Level Mutex Protection
- **Strategy**: Single mutex protecting all TaskAccess operations
- **Implementation**: RWMutex allowing multiple readers, exclusive writers
- **Advantages**:
  - Simple implementation
  - Guaranteed data consistency
  - No deadlock potential
- **Disadvantages**:
  - Limited concurrency (serializes all operations)
  - Suboptimal performance for read-heavy workloads
  - Doesn't leverage VersioningUtility's repository-level locking

### Option B: Operation-Level Locking
- **Strategy**: Different locks for read vs. write operations, with task-level granularity
- **Implementation**: Map of task ID mutexes for fine-grained locking
- **Advantages**:
  - Maximum concurrency for independent tasks
  - Optimal read/write separation
  - Better performance characteristics
- **Disadvantages**:
  - Complex lock management
  - Potential deadlock scenarios
  - Memory overhead for lock map

### Option C: Repository Delegation with Atomic Operations
- **Strategy**: Rely on VersioningUtility repository locking, make TaskAccess operations atomic
- **Implementation**: Each operation completes entirely within VersioningUtility transaction boundaries
- **Advantages**:
  - Leverages existing VersioningUtility thread safety
  - Consistent with architectural layering
  - No additional locking complexity
- **Disadvantages**:
  - Limited by VersioningUtility locking granularity
  - May not optimize for TaskAccess-specific access patterns
  - Potential performance bottleneck for bulk operations

**Rationale**: Option A chosen for guaranteed data consistency and simple implementation. Service-level RWMutex ensures file consistency under concurrent requests, which is critical for directory-structure-per-column approach. Simpler than complex lock management while providing reliable concurrent access.

**User Approval**: Option A approved on [2025-09-11]

## [2025-09-11] - BoardAccess Design Decision: Error Handling and Recovery Strategy  

**Decision**: Error Wrapping with Context (Option B)

**Context**: SRS requires structured error information (REQ-RELIABILITY-001) and graceful failure handling when VersioningUtility unavailable (REQ-RELIABILITY-003). Need consistent error response format and recovery mechanisms.

**Options Considered**:

### Option A: Structured Error Types with Recovery Actions
- **Error Structure**: Custom error types implementing structured format
  ```go
  type TaskAccessError struct {
      Code        string            // ERROR_TASK_NOT_FOUND, ERROR_STORAGE_FAILED
      Message     string            // Human-readable description
      Details     map[string]interface{} // Technical debugging info
      Suggestions []string          // Recovery action suggestions
      Cause       error            // Underlying error if any
  }
  ```
- **Recovery Strategy**: Return specific suggestions per error type
- **Advantages**:
  - Meets SRS structured error requirements precisely
  - Clear recovery guidance for callers
  - Rich debugging information
- **Disadvantages**:
  - More complex error handling implementation
  - Potential over-engineering for simple errors

### Option B: Error Wrapping with Context
- **Error Structure**: Go standard error wrapping with context
- **Strategy**: Use fmt.Errorf with contextual information
- **Advantages**:
  - Follows Go idioms
  - Simple implementation
  - Good error chain preservation
- **Disadvantages**:
  - Less structured than SRS requirements
  - Limited recovery action guidance

### Option C: Hybrid Approach - Structured for Domain Errors, Simple for System Errors
- **Strategy**: Structured errors for business logic, wrapped errors for system failures
- **Implementation**: Custom types for task-specific errors, standard wrapping for storage/logging errors
- **Advantages**:
  - Meets SRS requirements for important cases
  - Simpler handling for infrastructure errors
  - Balanced complexity
- **Disadvantages**:
  - Inconsistent error handling patterns
  - Callers need to handle multiple error types

**Rationale**: Option B chosen for simpler implementation following Go idioms while still providing good error chain preservation. Standard error wrapping with contextual information provides sufficient debugging capability without over-engineering.

**User Approval**: Option B approved on [2025-09-11]

## [2025-09-11] - BoardAccess Design Decision: Data Storage and File Organization Strategy

**Decision**: Directory Structure with Board Configuration (User-Specified Approach)

**Context**: BoardAccess requires efficient storage of board data, column configuration, and task data with JSON format (REQ-FORMAT-001), version control integration (REQ-INTEGRATION-001), and separate active/archived task organization (REQ-FORMAT-003). Need to optimize for minimal git diffs during common operations like priority changes (REQ-FORMAT-002).

**Options Considered**:

### Option A: Single Task Per File Approach
- **Structure**: Each task stored in separate JSON file (e.g., `tasks/active/task-12345.json`, `tasks/archived/task-12345.json`)
- **File Organization**: 
  ```
  tasks/
  ├── active/
  │   ├── task-12345.json
  │   └── task-67890.json
  └── archived/
      └── task-11111.json
  ```
- **Advantages**:
  - Optimal git diffs: only affected task file changes
  - Easy conflict resolution during merges
  - Simple task archiving (move file between directories)
  - Natural task history per file through VersioningUtility
  - No need for complex JSON manipulation
- **Disadvantages**:
  - More files to manage
  - Bulk queries require multiple file reads
  - Directory operations for task enumeration

### Option B: Priority-Grouped JSON Files
- **Structure**: Tasks grouped by Eisenhower matrix quadrant in separate files
- **File Organization**:
  ```
  tasks/
  ├── active/
  │   ├── urgent-important.json
  │   ├── urgent-not-important.json
  │   ├── not-urgent-important.json
  │   └── not-urgent-not-important.json
  └── archived/
      └── archived-tasks.json
  ```
- **JSON Format**: Array of tasks per priority level
- **Advantages**:
  - Fewer files to manage
  - Fast priority-based queries (single file read)
  - Natural grouping matches domain model
- **Disadvantages**:
  - Large git diffs when moving tasks between priorities
  - Complex JSON array manipulation
  - Potential merge conflicts on same file
  - Archive operations require JSON modification

### Option C: Hybrid Single Index + Individual Files
- **Structure**: Master index file with task metadata + individual task files
- **File Organization**:
  ```
  tasks/
  ├── index.json          # Master index: {id, priority, status, archived}
  ├── active/
  │   ├── task-12345.json
  │   └── task-67890.json
  └── archived/
      └── task-11111.json
  ```
- **Advantages**:
  - Fast bulk queries via index
  - Minimal diffs for priority changes (only index)
  - Individual task history preserved
- **Disadvantages**:
  - Index consistency challenges
  - Two-stage operations (index + task file)
  - Complex recovery if index corrupts

### Option D: Two Aggregate Files (Active + Archived)
- **Structure**: All active tasks in single file, all archived tasks in separate file
- **File Organization**:
  ```
  tasks/
  ├── active-tasks.json    # All active tasks as JSON array/object
  └── archived-tasks.json  # All archived tasks as JSON array/object
  ```
- **JSON Format**: Either array of tasks or object with task IDs as keys
- **Advantages**:
  - Minimal files to manage (only 2 files)
  - Simple bulk operations (single file read/write)
  - Easy backup and synchronization
  - Fast enumeration of all tasks
  - Directly meets REQ-FORMAT-003 (separate active/archived)
- **Disadvantages**:
  - Large git diffs for any task change
  - Potential merge conflicts on same file
  - No individual task history tracking
  - Entire file rewrite for single task changes
  - Poor performance for large task sets
  - Complex JSON manipulation for individual operations
  - File locking issues under high concurrency

### **User-Specified Approach: Directory Structure with Board Configuration**

**Structure**: 
- **Board Configuration**: `board.json` - contains column definitions and Eisenhower sections
- **Active Tasks**: `<column>[/<section>]/task-<id>.json` - tasks organized by column/section directory structure  
- **Archived Tasks**: `archived/task-<id>.json` - archived tasks in dedicated directory

**File Organization**:
```
board.json                           # Board and column configuration
todo/
├── urgent-important/
│   ├── task-12345.json
│   └── task-67890.json
├── urgent-not-important/
│   └── task-11111.json
├── not-urgent-important/
│   └── task-22222.json
└── not-urgent-not-important/
    └── task-33333.json
doing/
├── task-44444.json
└── task-55555.json
done/
├── task-66666.json
└── task-77777.json
archived/
├── task-99999.json
└── task-88888.json
```

**Advantages**:
- **Optimal git diffs**: Moving between sections = file move operation, minimal diff
- **Board configuration centralized**: Column definitions, Eisenhower setup in `board.json`
- **Natural directory queries**: List files in directory for column/section queries
- **Simple archiving**: Move file to `archive/` directory
- **Column context implicit**: Directory structure provides column/section information
- **Clean separation**: Board structure vs task content clearly separated

**Implementation Details**:
- `board.json` contains column definitions, section mappings, workflow rules
- Directory structure mirrors logical board organization
- Task files contain only task-specific data (no column redundancy)
- Archive operations are simple file moves
- Section queries become directory listings

**Rationale**: User-specified approach provides optimal git diff behavior for common operations (priority/column moves), centralizes board configuration, and uses directory structure as natural organizational mechanism.

**User Approval**: **APPROVED** - User specified this exact approach

---

**Design Review Status**: Complete design approved by user on [2025-09-11] - Ready for implementation

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