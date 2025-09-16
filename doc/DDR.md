# Design Decision Records (DDR)

## [2025-09-17] - UIStateAccess Design Decision: Implementation Architecture

**Decision**: Option C - Hybrid Platform-Optimized Approach

**Context**: Need to determine the implementation approach for UIStateAccess while providing cross-platform UI state persistence that leverages native OS storage mechanisms. The service requires optimal platform integration, excellent performance (<10ms state access), and robust error recovery while maintaining data integrity across Windows, macOS, and Linux platforms.

**Options Considered**:
- **Option A: Simple File-Based Approach** - Direct JSON file storage with basic error handling, simple implementation but limited performance optimization and platform differences exposed
- **Option B: Database-Backed Approach** - Embedded SQLite database with structured schema, excellent concurrency and ACID transactions but additional dependency and complexity overhead
- **Option C: Hybrid Platform-Optimized Approach** - Platform-specific storage backends with common interface, optimal platform integration and performance but more complex implementation
- **Option D: Layered Cache Architecture** - In-memory cache with persistent backend, excellent performance but memory usage concerns and cache coherency complexity

**Rationale**: Choose Option C to provide optimal balance of platform integration, performance, and maintainability. This approach leverages native OS storage mechanisms (Windows Registry + JSON files, macOS Preferences + plist, Linux XDG + JSON) for best user experience while maintaining cross-platform interface consistency. Platform-specific optimizations provide superior performance and reliability compared to generic approaches.

**Consequences**:
- Platform-specific storage implementations: Windows (Registry + AppData JSON), macOS (Preferences API + plist), Linux (XDG directories + JSON)
- Common interface abstracts platform differences through Strategy pattern
- Core components: StateManager (main interface), PlatformStorage (platform backends), StateValidator (data validation), BackupManager (multi-tier recovery), CacheLayer (performance optimization), LoggingIntegration (comprehensive monitoring)
- Key design patterns: Strategy (platform storage), Template Method (common validation), Observer (state change notifications), Command (atomic operations), Factory (backend creation)
- Multi-layer backup strategy with platform-specific optimizations
- Excellent cross-platform compatibility while leveraging each platform's strengths
- Performance targets met through platform-native optimizations and intelligent caching

**User Approval**: Approved on [2025-09-17]

## [2025-09-16] - FyneUtility Design Decision: Implementation Architecture

**Decision**: Option C - Hybrid Approach with Smart Defaults

**Context**: Need to determine the implementation approach for FyneUtility while providing foundational Fyne framework abstraction that enables consistent widget creation, theme management, and resource handling across all client UI components. The service requires cross-platform compatibility, high performance (widget creation <1ms), and seamless integration with ValidationUtility and FormatUtility.

**Options Considered**:
- **Option A: Functional Library with Typed Configuration** - Pure functional approach with strongly-typed configuration structs, explicit parameters, complete type safety but complex function signatures
- **Option B: Builder Pattern with Fluent Interface** - Builder pattern for complex configurations with method chaining, flexible but potentially stateful and complex state management
- **Option C: Hybrid Approach with Smart Defaults** - Simple functions with smart defaults plus functional options pattern, clean API for common cases with power when needed

**Rationale**: Choose Option C to provide excellent developer experience with simple API for common cases (`CreateButton("OK")`) while maintaining flexibility through functional options for advanced scenarios. This approach maintains stateless functional design for thread safety, enables easy extension without breaking changes, and provides self-documenting configuration through option function names. The pattern works well with ValidationUtility and FormatUtility integration requirements.

**Consequences**:
- Simple API for common cases: `CreateButton("Save")`
- Advanced configuration through functional options: `CreateButton("Save", WithButtonStyle(PrimaryButton), WithButtonIcon(icon))`
- 30+ core functions across 10 functional areas (widget creation, layout management, theme operations, etc.)
- Maintains stateless design for thread safety
- Easy extension with new option functions without breaking existing code
- Integration points with ValidationUtility (`WithValidation()`) and FormatUtility (`WithFormatter()`)
- Performance optimizations through resource caching, widget factories, and lazy loading
- Cross-platform compatibility with graceful platform-specific fallbacks

**User Approval**: Approved on [2025-09-16]

## [2025-09-16] - ValidationUtility Design Decision: Implementation Architecture

**Decision**: Option A - Simple Functional Approach

**Context**: Need to determine the implementation approach for ValidationUtility while maintaining consistency with FormatUtility's successful functional design and supporting all 13 SRS interface operations with proper error handling and validation result structures.

**Options Considered**:
- **Option A: Simple Functional Approach** - Direct functions, validation result structs, minimal types, standard error handling
- **Option B: Interface-Based Service Pattern** - IValidationUtility interface, struct implementation, rich error handling, consistent with other services
- **Option C: Hybrid Approach** - Core interface with functional helpers, mixed complexity

**Rationale**: Choose Option A to maintain consistency with FormatUtility's proven functional design. ValidationUtility operations are stateless validation functions that benefit from direct function calls without interface overhead. The functional approach aligns with the universal utility pattern and provides optimal performance for validation operations.

**Consequences**:
- Direct functions for all 13 operations (ValidateString, ValidateNumber, etc.)
- Rich data contracts (ValidationResult, ValidationRule, StringConstraints, NumericConstraints structs)
- Standard Go error handling with structured validation results
- No interface overhead - direct function calls
- Consistent with FormatUtility design patterns
- Thread-safe by design (stateless functions)
- Easy to use without service instantiation

**User Approval**: Accepted

## [2025-09-16] - FormatUtility Design Decision: Implementation Architecture

**Decision**: Option A - Simple Functional Approach

**Context**: Need to determine the implementation approach for FormatUtility while maintaining consistency with existing utility patterns (LoggingUtility, CacheUtility) and supporting all 11 SRS interface operations with proper error handling and extensibility.

**Options Considered**:
- **Option A: Simple Functional Approach** - Direct functions, no state, minimal types, simple error handling
- **Option B: Interface-Based Service Pattern** - IFormatUtility interface, struct implementation, rich error handling, consistent with existing utilities
- **Option C: Hybrid Approach** - Core interface with functional helpers, mixed complexity

**Rationale**: Choose Option A because FormatUtility operations are purely functional/stateless with no need for mocking in tests. Direct functions provide simpler implementation, easier testing, and optimal performance without interface overhead. The stateless nature makes interfaces unnecessary for abstraction.

**Consequences**:
- Direct functions for all 11 operations (TrimText, ConvertCase, etc.)
- Minimal type definitions (TextCaseType, FileSizeUnit, ValidationRule enums/structs)
- Standard Go error handling with contextual information
- No interface overhead - direct function calls
- Simpler implementation and testing
- Thread-safe by design (stateless functions)
- Easy to use without service instantiation

**User Approval**: Approved on [2025-09-16]

## [2025-09-14] - TaskManagerAccess: Implementation Architecture

**Decision**: Option A - Simple Channel-based Implementation

**Context**: Need to determine the internal implementation approach for TaskManagerAccess async operations while maintaining interface contract requirements and ensuring proper error handling, data transformation, and cache coordination.

**Options Considered**:
- **Option A: Simple Channel-based Implementation** - Direct channel returns from all async methods, minimal internal state management, direct TaskManager service calls with error wrapping
- **Option B: Worker Pool with Request Queue** - Internal worker goroutines handling service calls, request queuing for batching and optimization, more complex but potentially better performance
- **Option C: Hybrid Approach with Smart Caching** - Channel-based interface with internal caching logic, request deduplication and batching where beneficial, balance between simplicity and performance

**Rationale**: Choose Option A to start with simplicity and clean interface implementation. Direct channel-based approach provides straightforward async operations without internal complexity. Performance optimizations can be added later without changing the interface contract. This approach aligns with iDesign principles of starting simple and adding complexity only when needed.

**Consequences**:
- Direct async method implementation with immediate channel returns
- Minimal internal state and complexity
- Straightforward error translation and data transformation logic
- Simple cache coordination without internal queuing
- Easy to test and debug
- Performance optimization opportunities preserved for future enhancement
- Clear separation between interface contract and implementation details

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - Native GUI Framework Selection: Fyne

**Decision**: Fyne for native GUI implementation

**Context**: User prefers native GUIs over web UIs and needs drag-and-drop support for task management. Evaluated Go native GUI frameworks for cross-platform compatibility, drag-and-drop capabilities, and development simplicity.

**Options Considered**:
- **Fyne**: Built-in drag-and-drop via `fyne.Draggable` interface, cross-platform, Material Design-inspired, active development
- **Wails v2**: HTML5 drag/drop API support, web frontend with native packaging, modern architecture
- **Walk**: Native Windows drag/drop, Windows-only, lightweight, true native look
- **Gio**: Gesture-based drag support, GPU-accelerated immediate mode GUI, steeper learning curve
- **GTK (go-gtk)**: Full GTK drag/drop capabilities, Linux-native, comprehensive widgets
- **Qt (therecipe/qt)**: Complete Qt drag/drop framework, professional feature set, complex setup

**Rationale**: Fyne provides the optimal balance of simplicity, cross-platform support, and built-in drag-and-drop capabilities for a task management application. Material Design-inspired interface aligns well with modern user expectations, and the framework has active development with good documentation.

**Consequences**:
- Cross-platform native GUI (Windows, macOS, Linux, mobile)
- Built-in drag-and-drop support via `fyne.Draggable` interface
- Material Design-inspired UI components
- Single dependency with good Go ecosystem integration
- Active community and development support
- Simple deployment and distribution

**User Approval**: Approved on [2025-09-14]

## [2025-09-15] - CacheUtility: Implementation Architecture

**Decision**: Option A - Map-Based with RWMutex

**Context**: Need to determine the internal implementation approach for CacheUtility thread-safe caching operations while maintaining performance requirements (1ms Get, 5ms Set) and supporting TTL management, pattern-based invalidation, and LRU eviction.

**Options Considered**:
- **Option A: Map-Based with RWMutex** - Single map with read-write mutex protection, LRU tracking using doubly-linked list, simple implementation
- **Option B: Segmented Cache with Fine-Grained Locking** - Multiple cache segments with separate locks, reduced contention but more complex
- **Option C: Channel-Based Actor Model** - Single goroutine handling all operations via channels, serialized operations but channel overhead
- **Option D: Sync.Map with Custom TTL Management** - Go's built-in concurrent map with separate TTL tracking, good read performance but complex TTL management

**Rationale**: Choose Option A to align with the "start simple" approach used for TaskManagerAccess. Map-based implementation with RWMutex provides straightforward thread safety, meets performance requirements for expected cache sizes, and allows for clear LRU tracking and pattern invalidation implementation. Background cleanup goroutine handles TTL expiration efficiently.

**Consequences**:
- Simple implementation with clear thread safety guarantees
- LRU tracking using doubly-linked list with map pointers for O(1) access
- Background cleanup goroutine for automated expired entry removal
- Pattern matching using filepath.Match() for wildcard support
- Atomic operations for statistics to avoid mutex overhead during reads
- Easy to test and debug with predictable behavior
- Performance optimization opportunities preserved for future enhancement

**User Approval**: Approved on [2025-09-15]

## [2025-09-14] - TaskManager: Interface Contract Design

**Decision**: Option A - Single Comprehensive Interface

**Context**: Need to determine how TaskManager should organize its operations and data contracts while maintaining interface consistency with SRS requirements.

**Options Considered**:
- **Option A: Single Comprehensive Interface** - All 7 operations in one interface with rich data contracts
- **Option B: Separated CRUD and Workflow Interfaces** - Split between TaskCRUD and TaskWorkflow interfaces  
- **Option C: Operation-Grouped Interfaces** - Three focused interfaces (TaskManagement, TaskQuery, TaskWorkflow)

**Rationale**: Choose Option A because all operations concern the same "facet" of the manager.

**Consequences**:
- Single TaskManager interface with 7 operations as specified in SRS
- Rich data contracts: TaskRequest, TaskResponse, WorkflowStatus, ValidationResult entities
- Direct alignment with SRS service contract requirements
- Simplified client integration with single interface
- Future interface segregation possible if needed

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - TaskManager: Internal Architecture Structure

**Decision**: Option A - Simple Manager with Direct Dependencies

**Context**: Need to determine internal component organization for TaskManager while maintaining Manager layer architectural constraints.

**Options Considered**:
- **Option A: Simple Manager with Direct Dependencies** - TaskManager directly coordinates with BoardAccess, RuleEngine, LoggingUtility
- **Option B: Manager with Internal Workflow Orchestrator** - Additional WorkflowOrchestrator internal component
- **Option C: Manager with Specialized Internal Services** - Multiple internal components (TaskValidator, WorkflowOrchestrator, CascadeHandler)

**Rationale**: Choose Option A because the TaskManager with its dependencies builds a subsystem accordign to iDesign.

**Consequences**:
- Single TaskManager service implementation
- Direct calls to BoardAccess for data persistence
- Direct calls to RuleEngine for business rule validation
- Direct calls to LoggingUtility for operational logging
- Embedded workflow orchestration logic within TaskManager
- Simple implementation and maintenance approach

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - TaskManager: Subtask Workflow Coupling Strategy

**Decision**: TaskManager-Orchestrated with RuleEngine Compliance Verification

**Context**: Need to determine responsibility allocation between TaskManager and RuleEngine for subtask workflow coupling rules implementation.

**User Specification**: "The RuleEngine shall verify compliance with the rules as specified in the RuleEngine SRS and the RuleEngine code. All the remaining responsibilities shall be covered by the TaskManager."

**Rationale**: Clear separation of concerns where RuleEngine validates business rule compliance (as per its SRS) while TaskManager implements the actual workflow coupling orchestration. This maintains the Manager layer's orchestration responsibilities while leveraging RuleEngine for rule validation.

**Consequences**:
- TaskManager implements subtask workflow coupling logic (REQ-TASKMANAGER-016, REQ-TASKMANAGER-017)
- RuleEngine validates rule compliance before TaskManager applies changes
- TaskManager coordinates parent-child status transitions based on active policies
- TaskManager handles cascade operations for task deletion/archival
- Clear responsibility boundaries between rule validation and workflow orchestration

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - TaskManager: Error Handling Strategy

**Decision**: Option B - Error Wrapping with Context

**Context**: Need to choose error handling approach that provides structured information per SRS requirements while maintaining consistency with existing DDR patterns.

**Options Considered**:
- **Option A: Structured Error Types** - Custom error types with code, message, details, suggestions
- **Option B: Error Wrapping with Context** - Go standard error wrapping following BoardAccess DDR pattern
- **Option C: Hybrid Domain-Specific Errors** - Structured for workflow violations, wrapped for system failures

**Rationale**: Choose Option B for consistency with existing BoardAccess DDR decision and Go idioms. Error wrapping with contextual information provides sufficient debugging capability while maintaining implementation simplicity and architectural consistency across ResourceAccess components.

**Consequences**:
- Go standard error wrapping with contextual information
- Error chain preservation from dependencies (BoardAccess, RuleEngine)
- Contextual annotations for TaskManager operations
- Consistency with established DDR patterns
- Good error debugging capability without over-engineering

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - TaskManager: Concurrency and Thread Safety Strategy

**Decision**: Option A - Service-Level Mutex

**Context**: Need to ensure thread-safe operations for concurrent task workflow orchestration while maintaining data consistency per SRS performance and reliability requirements.

**Options Considered**:
- **Option A: Service-Level Mutex** - Single RWMutex following BoardAccess DDR pattern
- **Option B: Operation-Level Locking** - Fine-grained locking by operation type
- **Option C: Stateless with Dependency Coordination** - Rely on dependency locking

**Rationale**: Choose Option A for consistency with BoardAccess DDR decision and guaranteed data consistency. Service-level RWMutex ensures safe concurrent access to workflow orchestration operations while maintaining architectural consistency with other Manager components.

**Consequences**:
- Single RWMutex protecting all TaskManager operations
- Multiple concurrent readers, exclusive writers
- Guaranteed workflow consistency for parent-child operations
- Consistency with established BoardAccess concurrency pattern
- Simple implementation with reliable thread safety

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - Subtask Position Storage Decision: Position in Filename Not Content

**Decision**: Store task and subtask position information in the filename prefix, not in the JSON content

**Context**: Need to determine where to store position information for tasks and subtasks within columns/sections for proper ordering while maintaining optimal git diff behavior and data consistency.

**Options Considered**:

### Option A: Position in JSON Content
- **Structure**: Task JSON contains position field: `{"id": "12345", "position": 1, "title": "..."}`
- **Filename**: Static names like `task-12345.json`, `subtask-67890.json`
- **Advantages**:
  - Position data travels with task content
  - Simpler filename management
  - No filename changes for position updates
- **Disadvantages**:
  - Position changes require JSON file content modification
  - Larger git diffs for position updates
  - Risk of position data inconsistency with file location
  - Complex validation between JSON content and directory location

### Option B: Position in Filename Prefix
- **Structure**: Position encoded in filename prefix
- **Task Filename**: `<position>-task-<id>.json` (e.g., `001-task-12345.json`)
- **Subtask Filename**: `<position>-subtask-<id>.json` (e.g., `001-subtask-67890.json`)
- **Directory Names**: No position prefix (e.g., `task-12345/` for subtask container)
- **Advantages**:
  - Optimal git diffs for position changes (file rename vs content change)
  - Position immediately visible in directory listings
  - Natural sorting by filename gives correct position order
  - No risk of position data inconsistency
  - JSON content focuses purely on task data

**Rationale**: Choose Option B to optimize for REQ-FORMAT-002 (minimal git diffs for common operations). Position changes are common operations in Kanban boards (reordering tasks within columns/sections). Using filename prefixes makes position changes into git file renames rather than content modifications, resulting in cleaner version history and better merge conflict resolution.

**Consequences**:
- Task files: `001-task-12345.json`, `002-task-67890.json` etc.
- Subtask files: `001-subtask-11111.json`, `002-subtask-22222.json` etc.  
- Subtask directories: `task-12345/` (no position prefix to minimize directory moves)
- Position changes become file rename operations (optimal for git)
- Directory listings naturally sort by position
- JSON content remains focused on task attributes only
- File management operations must handle position prefix updates

**User Approval**: Approved on [2025-09-14]

## [2025-09-13] - ValidationEngine Service Decision: Service Not Required

**Decision**: Do not implement ValidationEngine service for EisenKan

**Context**: After analyzing the ValidationEngine scope and architectural requirements, determined that the validity of requests is already sufficiently verified by existing components in the system. The planned ValidationEngine would primarily duplicate validation logic already present in BoardAccess (data integrity validation) and RuleEngine (business rule validation).

**Options Considered**:
- **Option A: Implement ValidationEngine** - Add new Engine component for orchestrated validation scenarios
- **Option B: Enhance existing components** - Extend BoardAccess and RuleEngine validation capabilities as needed
- **Option C: No ValidationEngine** - Leave validation distributed across appropriate components (BoardAccess for data integrity, RuleEngine for business rules)

**Rationale**: Existing architecture already provides comprehensive validation:
- BoardAccess handles data integrity validation (required fields, formats, consistency)
- RuleEngine handles business rule validation (WIP limits, workflow transitions, dependencies)
- Cross-task dependency validation can be implemented in BoardAccess when needed
- External system integration validation can be added to specific components as requirements emerge
- No compelling use cases identified that require orchestrated validation beyond existing capabilities

**Consequences**:
- Simplified architecture with fewer components
- Validation logic remains close to domain responsibility (data validation in data layer, business rules in rule engine)
- Future validation needs addressed incrementally in appropriate components
- No validation orchestration layer - complex validations handled by extending existing components
- Reduced system complexity and maintenance overhead

**User Approval**: Approved on [2025-09-13]

## [2025-09-13] - RuleEngine Design Decision: Rule Context Data Access Strategy

**Decision**: Option C - Rule Engine with ResourceAccess Integration

**Context**: RuleEngine needs broader board context for Kanban rules (WIP limits, subtask status, column timestamps, other tasks' priorities) beyond single task event context defined in current SRS.

**Options Considered**:
- **Option A: Minimal Context** - Single task only, simple but cannot implement complex Kanban rules
- **Option B: Rich Context** - Manager provides board context, keeps Engine pure but increases Manager complexity
- **Option C: ResourceAccess Integration** - RuleEngine calls ResourceAccess directly for board data when needed

**Rationale**: Option C provides complete rule evaluation capabilities while maintaining architectural compliance (Engines can access ResourceAccess components per iDesign). This enables all identified Kanban rule types without overcomplicating the Manager layer.

**Consequences**:
- RuleEngine can implement WIP limits, age limits, subtask dependencies, and priority flow rules
- Direct access to BoardAccess for task counts and timestamps
- Direct access to task hierarchy information for subtask rules  
- RuleEngine becomes more capable but retains stateless operation
- Cleaner separation between rule logic (Engine) and workflow orchestration (Manager)

**User Approval**: Approved on [2025-09-13]

## [2025-09-13] - RuleEngine Design Decision: Rule Evaluation Architecture

**Decision**: Option B - Complete Sequential Processor

**Context**: Need to meet REQ-RULEENGINE-002 requirement to report all violations in one evaluation.

**Options Considered**:
- **Option A: Eager Sequential Processor** - Simple but cannot report all violations at once, because it stops at first violation
- **Option B: Complete Sequential Processor** - Simple and can report all violations at once, because it aggregates all violations
- **Option C: Parallel Evaluator** - Evaluate all rules in parallel, aggregate results to report all violations
- **Option D: Hybrid Priority Groups** - Complex implementation with partial violation reporting

**Rationale**: Option B meets REQ-RULEENGINE-002 requirement to evaluate all matching rules and report all violations. Parallel evaluation of option C) would provide better performance for large rule sets, but that's an implementation detail not considered critical and could be changed later on without breaking the interface.

**Consequences**:
- Can report all rule violations in single evaluation (meets SRS requirement)
- Stateless evaluation per individual rule
- Requires result aggregation
- All applicable rules evaluated regardless of priority (violations sorted by priority in results)

**User Approval**: Approved on [2025-09-13]

## [2025-09-13] - Fractal Design Decision

**Decision**: Option B - Integrate subtasks directly into existing interfaces

**Context**: Whether to use fractal design (treating parent tasks as boards with separate instances) or integrate subtasks into existing interfaces.

**Options Considered**:
- **Option A: Fractal design** - Separate system instances per task with subtasks, elegant API but complex routing and UI integration
- **Option B: Direct integration** - Single system instance managing both tasks and subtasks, simpler architecture

**Rationale**: Could not solve routing/instance issues cleanly in fractal approach while maintaining iDesign principles.

**Consequences**:
- Interfaces are moderately more complicated and explicit
- Subtask special cases mean the implementation is more complicated
- Can still use the "task as a board" concept as mental model

**User Approval**: Approved on [2025-09-13]

## [2025-09-13] - Subtasks Design Decision

**Decision**: Option A - Support subtasks

**Context**: Need to organize related tasks within boards. Previous workarounds (extra columns, tags) were tedious for tracking task groups with same goal.

**Options Considered**:
- **Option A: Support subtasks** - Fulfills user need but increases complexity
- **Option B: No subtasks** - Simpler but doesn't meet user need

**Rationale**: User need trumps implementation effort.

**Consequences**:
- More complex interfaces requiring subtask-aware operations
- Implementation complexity in storage, querying, and state management
- Enhanced data model with parent-child relationships
- Additional validation logic for subtask hierarchies
- UI complexity for nested task visualization

**User Approval**: Approved on [2025-09-13]

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