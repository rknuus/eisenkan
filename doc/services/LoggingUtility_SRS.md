# LoggingUtility Software Requirements Specification (SRS)

## 1. Introduction

### 1.1 Purpose
This Software Requirements Specification defines the requirements for the LoggingUtility service, a Utilities layer component that provides structured logging capabilities for all layers of the EisenKan system. The service enables consistent event recording, contextual information capture, and arbitrary data type logging following Google's structured logging principles.

### 1.2 Scope
LoggingUtility is responsible for:
- Structured logging with severity levels and contextual information
- Multiple output destinations (console, file) simultaneously
- Automatic stack trace capture for error conditions
- Level-based filtering for performance optimization
- Support for arbitrary data types as structured data
- Thread-safe concurrent logging operations

### 1.3 System Context
LoggingUtility operates in the Utilities layer of the EisenKan architecture, providing logging services to all other layers (Clients, Managers, Engines, ResourceAccess, Resources). It provides a stable API for structured logging while encapsulating the volatility of output formatting and destination management.

## 2. Operations

The following operations define the required behavior for LoggingUtility:

#### OP-1: Record Event
**Actors**: All system components
**Trigger**: When a component needs to record an event with structured data
**Flow**:
1. Receive log request with level, message, component, and structured data
2. Check if log level is enabled for performance optimization
3. Add timestamp and format structured data
4. Output to configured destinations (console, file)
5. Return success confirmation

#### OP-2: Record Error with Stack Trace
**Actors**: All system components
**Trigger**: When a component encounters an error condition
**Flow**:
1. Receive error log request with message, component, and error data
2. Automatically capture current stack trace information
3. Format error with stack trace and structured data
4. Output to configured destinations with ERROR level
5. Return success confirmation or crash application if output fails

#### OP-3: Check Log Level
**Actors**: All system components
**Trigger**: When a component needs to determine if expensive debug operations should run
**Flow**:
1. Receive log level check request
2. Compare requested level with current configuration
3. Return boolean indicating if level would be output

## 3. Functional Requirements

### 3.1 Event Recording Requirements

**REQ-LOG-001**: While the system is operational, the LoggingUtility shall record events with severity levels (Debug, Info, Warning, Error, Fatal) to enable filtering based on operational needs.

**REQ-LOG-002**: When a component calls the logging service with structured context, the LoggingUtility shall capture all contextual information including component and operation.

**REQ-LOG-003**: While the system is operational, the LoggingUtility shall support multiple output destinations (console, file) simultaneously to accommodate different deployment environments.

**REQ-LOG-004**: When an error logging operation is requested, the LoggingUtility shall automatically capture stack trace information to facilitate rapid problem resolution.

**REQ-LOG-005**: While the system is operational, the LoggingUtility shall provide level-based filtering checks to prevent expensive debug operations when not needed.

**REQ-LOG-006**: When a log request is received, the LoggingUtility shall add a timestamp to avoid skewing timestamps if the requests are processed asynchronously.

### 3.2 Structured Data Requirements

**REQ-STRUCT-001**: While the system is operational, the LoggingUtility shall support logging of arbitrary data types (structs, maps, slices, primitives) as structured data.

**REQ-STRUCT-002**: When logging structured data, the LoggingUtility shall preserve type information and hierarchical relationships to enable programmatic analysis.

**REQ-STRUCT-003**: While the system is operational, the LoggingUtility shall support logging of plain messages without any runtime-formatted data, because all additional data shall be passed as structured data.

**REQ-STRUCT-004**: When generating log output, the LoggingUtility shall generate human-readable messages while maintaining machine-parseable structured data.

**REQ-FORMAT-001**: When generating log output, the LoggingUtility shall format structured logs with timestamp, level, message, and structured data.

**REQ-FORMAT-003**: While processing structured data, the LoggingUtility shall limit nested object depth to 5 levels to prevent output verbosity issues.

## 4. Quality Attributes

### 4.1 Performance Requirements

**REQ-PERF-001**: While the system is operational, the LoggingUtility shall introduce less than 4x performance overhead compared to baseline operations without logging.

### 4.2 Reliability Requirements

**REQ-RELIABILITY-001**: If log output fails, then the LoggingUtility shall crash the application.

**REQ-THREAD-001**: While the system is operational, the LoggingUtility shall handle concurrent access from multiple execution contexts without data races or deadlocks.

### 4.3 Usability Requirements

**REQ-CONFIG-001**: When the LoggingUtility starts, it shall read configuration from environment variables to support stateless design.

## 5. Service Contract Requirements

### 5.1 Interface Operations
The LoggingUtility shall provide the following operations:

1. **Log**: Record event
2. **LogError**: Record error with automatic stack trace capture
3. **IsLevelEnabled**: Check if log level would be output (performance optimization)

### 5.2 Data Contracts

**LogLevel Enumeration**:
- Debug (detailed development information)
- Info (general informational events)  
- Warning (concerning but non-critical issues)
- Error (error conditions requiring attention)
- Fatal (critical errors causing system failure)

**Log Structure**:
- Message: String narrative describing the logged event
- Component: String identifying the calling component
- Data: interface{} for arbitrary structured data

### 5.3 Error Handling
All errors shall include:
- Error code classification
- Human-readable error message
- Technical details for debugging
- Suggested recovery actions where applicable

## 6. Technical Constraints

### 6.1 Integration Requirements
**REQ-INTEGRATION-001**: The LoggingUtility shall be callable from all architectural layers.

**REQ-INTEGRATION-002**: The LoggingUtility shall not create dependencies on other system components.

**REQ-INTEGRATION-003**: The LoggingUtility shall support graceful resource cleanup.

### 6.2 Data Format Requirements
**REQ-FORMAT-001**: The LoggingUtility shall store log output in human-readable format with embedded JSON for structured data.

**REQ-FORMAT-002**: The LoggingUtility shall support multiple output destinations (console, file) simultaneously.

**REQ-FORMAT-003**: The LoggingUtility shall use environment variables for configuration:
- `LOG_LEVEL`: Controls minimum log level
- `LOG_FILE`: Optional file path for file logging
- Default Behavior: INFO level to console if no configuration provided

## 7. Acceptance Criteria

### 7.1 Functional Acceptance
- All interface operations work as specified in contract
- Arbitrary Go types (structs, maps, slices, primitives) can be logged as structured data
- Environment variable configuration is properly applied
- Multiple output destinations work simultaneously

### 7.2 Quality Acceptance  
- Performance overhead is less than 4x baseline for typical operations
- No data races or deadlocks under concurrent load (100 goroutines, 1000 messages each)
- Service handles invalid file paths and permissions gracefully
- Structured logs contain narrative messages with embedded structured data

### 7.3 Integration Acceptance
- Complex business objects are logged with preserved type information
- Machine-parseable structured data is embedded in human-readable format
- Nested object depth is limited to prevent verbosity
- Service can be consumed by all system layers without coupling
- Service follows iDesign utility service patterns
- Service maintains stateless operation (configuration only, no business state)

---

**Document Version**: 1.1
**Released**: 2025-09-07
**Updated**: 2025-09-12
**Status**: Accepted
**Based on**: Google Research "Structured Logging: Crafting Useful Message Content"

**Superseded requirement IDs (must not be reused for tracability reasons)**:
- REQ-FORMAT-002