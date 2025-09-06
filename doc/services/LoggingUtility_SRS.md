# LoggingUtility Software Requirements Specification (SRS)

## 1. Service Overview

### 1.1 Purpose
The LoggingUtility shall provide structured logging capabilities for all layers of the EisenKan system, enabling consistent event recording, contextual information capture, and arbitrary data type logging following Google's structured logging principles.

### 1.2 Architectural Classification
- **Layer**: Utilities
- **Type**: Utility Service
- **Volatility**: Infrastructure (Low volatility - logging patterns are stable)
- **Namespace**: `rknuus.Utility.eisenkan.Utilities.LoggingUtility`

## 2. Business Requirements (EARS Format)

### 2.1 Core Logging Requirements

**REQ-LOG-001**: The LoggingUtility shall record events with severity levels (Debug, Info, Warning, Error, Fatal) to enable filtering based on operational needs.

**REQ-LOG-002**: When a component calls the logging service with structured context, the LoggingUtility shall capture all contextual information including component and operation.

**REQ-LOG-003**: The LoggingUtility shall support multiple output destinations (console, file) simultaneously to accommodate different deployment environments.

**REQ-LOG-004**: When an error occurs, the LoggingUtility shall automatically capture stack trace information to facilitate rapid problem resolution.

**REQ-LOG-005**: The LoggingUtility shall provide level-based filtering checks to prevent expensive debug operations when not needed.

**REQ-LOG-006**: The LoggingUtility shall add a timestamp of when a log request was received by the utility to avoid skewing timestamps if the requests are processed asynchronously.

### 2.2 Structured Logging Requirements (Based on Google Research)

**REQ-STRUCT-001**: The LoggingUtility shall support logging of arbitrary Go types (structs, maps, slices, primitives) as structured data.

**REQ-STRUCT-002**: When logging structured data, the LoggingUtility shall preserve type information and hierarchical relationships to enable programmatic analysis.

**REQ-STRUCT-003**: The LoggingUtility shall support logging of plain messages (e.g. a descriptive error message) without any runtime-formatted data, because all additional data shall be passed as in key-value pairs.

**REQ-STRUCT-004**: The LoggingUtility shall generate human-readable messages while maintaining machine-parseable structured data.

### 2.3 Performance and Quality Requirements

**REQ-PERF-001**: The LoggingUtility shall introduce less than 4x performance overhead compared to baseline operations without logging.

**REQ-THREAD-001**: The LoggingUtility shall handle concurrent access from multiple goroutines without data races or deadlocks.

**REQ-RELIABILITY-001**: If log output fails, then the LoggingUtility shall crash the application.

**REQ-CONFIG-001**: The LoggingUtility shall read configuration from environment variables (EISENKAN_LOG_LEVEL, EISENKAN_LOG_FILE) to support stateless design.

## 3. Service Contract Requirements

### 3.1 Interface Operations
The LoggingUtility shall provide exactly 4 operations (following iDesign contract guidelines):

1. **Log**: Record event
2. **LogError**: Record error with automatic stack trace capture
3. **IsLevelEnabled**: Check if log level would be output (performance optimization)

### 4.2 Data Contracts

**LogLevel Enumeration**:
- Debug (detailed development information)
- Info (general informational events)  
- Warning (concerning but non-critical issues)
- Error (error conditions requiring attention)
- Fatal (critical errors causing system failure)

**StructuredLogContext Structure**:
- Message: String narrative describing the logged event
- Component: String identifying the calling component
- Data: interface{} for arbitrary structured data

### 4.3 Structured Logging Format Requirements

**REQ-FORMAT-001**: The LoggingUtility shall format structured logs as: `[timestamp] [level] narrative_message | component=X operation=Y [structured_data]`

**REQ-FORMAT-002**: When logging complex types, the LoggingUtility shall use JSON encoding for machine readability while preserving human narrative.

**REQ-FORMAT-003**: The LoggingUtility shall limit nested object depth to 5 levels to prevent output verbosity issues.

## 5. Technical Constraints

### 5.1 Technology Requirements
- Implementation Language: Go (matching project requirements)
- Output Formats: Human-readable with embedded JSON for structured data
- Configuration Method: Environment variables (stateless design)
- Dependencies: Go standard library only

### 5.2 Integration Requirements
- The LoggingUtility shall be callable from all architectural layers
- The LoggingUtility shall not create dependencies on other system components
- The LoggingUtility shall support graceful resource cleanup

### 5.3 Operational Requirements
- Environment Variables:
  - `LOG_LEVEL`: Controls minimum log level
  - `LOG_FILE`: Optional file path for file logging
- Default Behavior: INFO level to console if no configuration provided

## 6. Acceptance Criteria

### 6.1 Functional Acceptance
- All interface operations work as specified in contract
- Arbitrary Go types (structs, maps, slices, primitives) can be logged as structured data
- Environment variable configuration is properly applied
- Multiple output destinations work simultaneously

### 6.2 Quality Acceptance  
- Performance overhead is less than 4x baseline for typical operations
- No data races or deadlocks under concurrent load (100 goroutines, 1000 messages each)
- Service handles invalid file paths and permissions gracefully
- Structured logs contain narrative messages with embedded structured data

### 6.3 Structured Logging Acceptance
- Complex business objects are logged with preserved type information
- Machine-parseable structured data is embedded in human-readable format
- Nested object depth is limited to prevent verbosity

## 7. Design Constraints

### 7.1 Architectural Constraints
- Must follow iDesign utility service patterns
- Must not contain business logic or domain-specific functionality
- Must be stateless (configuration only, no business state)
- Must support interface-based design for testability

### 7.2 Structured Logging Constraints
- Must handle arbitrary types through reflection or type switching
- Must balance human readability with machine parseability
- Must prevent infinite recursion in self-referencing data structures

---

**Document Version**: 1.0  
**Created**: 2025-09-06  
**Status**: Accepted
**Based on**: Google Research "Structured Logging: Crafting Useful Message Content"