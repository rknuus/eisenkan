# BoardAccess Software Requirements Specification (SRS)

## 1. Introduction

### 1.1 Purpose
This Software Requirements Specification defines the requirements for the BoardAccess service, a ResourceAccess layer component that provides persistent storage and retrieval capabilities for EisenKan tasks and board management operations. The service encapsulates task data management, board lifecycle operations, and provides atomic business operations for task and board manipulation.

### 1.2 Scope
BoardAccess is responsible for:
- Persistent storage and retrieval of task data including hierarchical task relationships
- Version control integration for task history and change tracking
- Atomic operations for task lifecycle management with subtask support
- Data consistency and integrity enforcement for parent-child task relationships
- Board discovery, metadata extraction, and lifecycle management through IBoard facet
- Board validation and configuration management with RuleEngine integration
- Resource access abstraction for task-related and board-related operations

### 1.3 System Context
BoardAccess operates in the ResourceAccess layer of the EisenKan architecture, sitting between the business logic layers (Engines/Managers) and the resource layer (file system via VersioningUtility). It provides a stable API for task data operations while encapsulating the volatility of data storage mechanisms.

## 2. Operations

The following operations define the required behavior for BoardAccess:

#### OP-1: Store New Task
**Actors**: TaskManager, ValidationEngine
**Trigger**: When a new task is created in the system  
**Flow**:
1. Receive task data with required attributes and optional parent task identifier
2. Validate task data completeness and parent-child relationship constraints
3. Assign unique task identifier
4. Persist task to version-controlled storage with hierarchical relationship
5. Return task identifier and confirmation

#### OP-2: Retrieve Task
**Actors**: TaskManager
**Trigger**: When task data is needed for business operations  
**Flow**:
1. Receive task identifier request
2. Locate task in storage
3. Return complete task data or not found indication

#### OP-3: Update Task
**Actors**: TaskManager, ValidationEngine
**Trigger**: When task data needs modification  
**Flow**:
1. Receive task identifier and updated data
2. Validate update request
3. Apply changes to stored task
4. Create version history entry
5. Return update confirmation

#### OP-4: Archive or remove Task  
**Actors**: TaskManager  
**Trigger**: When task should be deleted from system  
**Flow**:
1. Receive task identifier for removal
2. Locate task and subtasks - if any - in storage
3. Archive or remove task and - depending on the policy - subtasks data
4. Return removal confirmation

#### OP-5: Query Tasks by Criteria
**Actors**: TaskManager
**Trigger**: When tasks need to be found by specific attributes
**Flow**:
1. Receive query criteria (priority, status, tags, parent task, etc.)
2. Search task storage using criteria including hierarchical filters
3. Return matching task identifiers and data with optional hierarchical information

#### OP-6: Discover Boards (IBoard Facet)
**Actors**: TaskManager
**Trigger**: When the system needs to identify available boards in a directory structure
**Flow**:
1. Receive directory path for board discovery
2. Validate directory existence and access permissions
3. Check for git repository presence and validity
4. Identify board configuration files and validate structure
5. Return list of discovered board locations with basic metadata

#### OP-7: Extract Board Metadata (IBoard Facet)
**Actors**: TaskManager
**Trigger**: When board information is needed for display or processing
**Flow**:
1. Receive board directory path
2. Access board configuration and data files
3. Extract metadata (title, description, creation date, last modified)
4. Calculate board statistics (task counts, column distributions)
5. Return comprehensive board metadata with caching optimization

#### OP-8: Validate Board Structure (IBoard Facet)
**Actors**: TaskManager
**Trigger**: When board integrity needs verification
**Flow**:
1. Receive board directory path
2. Validate git repository structure and accessibility
3. Verify board configuration file presence and format
4. Check data file integrity and schema version compatibility
5. Return validation result with detailed error information if issues found

#### OP-9: Load Board Configuration (IBoard Facet)
**Actors**: TaskManager
**Trigger**: When board-level configuration data needs to be retrieved
**Flow**:
1. Receive configuration load request with configuration type and identifier
2. Access git-based JSON configuration storage
3. Parse and validate configuration data structure
4. Return configuration data including board settings, column definitions, and workflow rules

#### OP-10: Store Board Configuration (IBoard Facet)
**Actors**: TaskManager
**Trigger**: When board-level configuration data needs to be persisted
**Flow**:
1. Receive configuration store request with configuration data and type
2. Validate configuration data against schema requirements
3. Serialize configuration to JSON format
4. Store configuration through git-based storage with atomic operations and versioning
5. Return configuration storage confirmation

#### OP-11: Create Board (IBoard Facet)
**Actors**: TaskManager
**Trigger**: When a new board needs to be initialized
**Flow**:
1. Receive board creation parameters (path, title, description, initial configuration)
2. Validate board configuration using RuleEngine
3. Initialize git repository structure if needed
4. Create board configuration and data files with proper schema
5. Return board creation confirmation with generated board identifier

#### OP-12: Delete Board (IBoard Facet)
**Actors**: TaskManager
**Trigger**: When a board needs to be removed from the system
**Flow**:
1. Receive board identifier and deletion parameters (trash vs permanent)
2. Validate deletion prerequisites and dependencies
3. Create backup of board data before deletion
4. If OS supports trash and user chooses trash option, move board to trash; otherwise permanently delete board files and directory structure
5. Return deletion confirmation with backup location and deletion method information

## 3. Functional Requirements

### 3.1 Task Storage Requirements

**REQ-BOARDACCESS-001**: When a valid task is provided, the BoardAccess service shall store the task data persistently with version control tracking.

**REQ-BOARDACCESS-002**: When storing a task, the BoardAccess service shall generate a unique task identifier and return it to the caller.

**REQ-BOARDACCESS-003**: When task data is incomplete or invalid, the BoardAccess service shall reject the storage request with a structured error message.

**REQ-BOARDACCESS-016**: When a task is created with a parent task identifier, the BoardAccess service shall validate that the parent task exists and enforce the 1-2 level hierarchy constraint (subtasks cannot have children).

**REQ-BOARDACCESS-017**: When storing a task with parent relationship, the BoardAccess service shall maintain referential integrity between parent and child tasks.

### 3.2 Task Retrieval Requirements  

**REQ-BOARDACCESS-004**: When a task identifier is provided, the BoardAccess service shall return the complete task data if it exists.

**REQ-BOARDACCESS-005**: When a non-existent task identifier is requested, the BoardAccess service shall return a not-found indication without error.

**REQ-BOARDACCESS-006**: The BoardAccess service shall support bulk retrieval of multiple tasks using a list of task identifiers.

**REQ-BOARDACCESS-018**: When a task identifier is provided with subtask inclusion parameter, the BoardAccess service shall return the task data along with its subtasks or parent task information.

### 3.3 Task Update Requirements

**REQ-BOARDACCESS-007**: When a valid task update request is provided, the BoardAccess service shall store the task data persistently with version control tracking.

**REQ-BOARDACCESS-008**: When task update data is invalid (e.g. non-existent task identifier), the BoardAccess service shall reject the update and leave the original data unchanged.

**REQ-BOARDACCESS-022**: When a subtask update request attempts to modify the parent task identifier, the BoardAccess service shall reject the update with a structured error message indicating that parent task relationships are immutable.

### 3.4 Task Query Requirements

**REQ-BOARDACCESS-009**: The BoardAccess service shall support bulk retrieval of all task identifiers.

**REQ-BOARDACCESS-010**: The BoardAccess service shall support querying tasks by priority level (urgent/important combinations).

**REQ-BOARDACCESS-011**: The BoardAccess service shall support querying tasks by workflow status.

**REQ-BOARDACCESS-012**: When query criteria match no tasks, the BoardAccess service shall return an empty result set without error.

**REQ-BOARDACCESS-019**: The BoardAccess service shall support querying tasks by parent task identifier to retrieve all subtasks of a given parent.

**REQ-BOARDACCESS-020**: The BoardAccess service shall support querying for top-level tasks only (tasks without parent task identifiers).

**REQ-BOARDACCESS-023**: The BoardAccess service shall support querying tasks by priority promotion date to enable automated priority escalation processing.

**REQ-BOARDACCESS-024**: When storing or updating tasks with priority promotion dates, the BoardAccess service shall validate the promotion date format and store it persistently with the task data.

### 3.5 Task Removal Requirements

**REQ-BOARDACCESS-013**: When a task archive request is received, the BoardAccess service shall archive the task instead of permanently deleting it.

**REQ-BOARDACCESS-015**: When a task removal request is received, the BoardAccess service shall permanently delete it.

**REQ-BOARDACCESS-014**: When removing a non-existent task, the BoardAccess service shall return success without error (idempotent operation).

**REQ-BOARDACCESS-021**: When a parent task is archived or deleted, the BoardAccess service shall handle cascade operations for all its subtasks according to configured cascade policy (archive subtasks, delete subtasks, or promote subtasks to top-level).

### 3.6 Board Management Requirements (IBoard Facet)

**REQ-BOARDACCESS-025**: When a board configuration load request is received, the BoardAccess service shall retrieve configuration data from git-based JSON storage and return parsed configuration information.

**REQ-BOARDACCESS-026**: When board configuration data is not found, the BoardAccess service shall return appropriate default configuration data without errors.

**REQ-BOARDACCESS-027**: When a board configuration store request is received with valid data, the BoardAccess service shall serialize the configuration to JSON format and persist it through git-based storage with atomic operations.

**REQ-BOARDACCESS-028**: When board configuration data validation fails, the BoardAccess service shall return detailed validation error information without persisting invalid data.

**REQ-BOARDACCESS-029**: When storing board configuration data, the BoardAccess service shall ensure atomic operations and leverage git versioning for data integrity and rollback capabilities.

**REQ-BOARDACCESS-030**: When a directory path is provided, the BoardAccess service shall validate the directory exists and is accessible for board operations.

**REQ-BOARDACCESS-031**: When discovering boards, the BoardAccess service shall identify git repositories and verify their validity for board storage.

**REQ-BOARDACCESS-032**: When a directory contains board configuration files, the BoardAccess service shall validate the configuration file format and structure.

**REQ-BOARDACCESS-033**: When discovering boards in a directory, the BoardAccess service shall return a list of valid board locations with basic identification metadata.

**REQ-BOARDACCESS-034**: When board discovery encounters invalid or corrupted board structures, the BoardAccess service shall continue processing other boards and report issues without failing completely.

**REQ-BOARDACCESS-035**: When a board path is provided, the BoardAccess service shall extract and return board metadata including title, description, creation date, and last modified timestamp.

**REQ-BOARDACCESS-036**: When extracting board statistics, the BoardAccess service shall calculate task counts, column distributions, and board activity metrics.

**REQ-BOARDACCESS-037**: When board configuration data is accessed, the BoardAccess service shall extract and return board settings, column definitions, and workflow configurations.

**REQ-BOARDACCESS-038**: When metadata extraction fails due to missing or corrupted files, the BoardAccess service shall return appropriate error information with recovery suggestions.

**REQ-BOARDACCESS-039**: When validating board structure, the BoardAccess service shall verify git repository integrity and accessibility.

**REQ-BOARDACCESS-040**: When validating boards, the BoardAccess service shall check configuration file presence, format validity, and schema version compatibility.

**REQ-BOARDACCESS-041**: When validation identifies issues, the BoardAccess service shall provide detailed diagnostic information including specific problems and recommended fixes.

**REQ-BOARDACCESS-042**: The BoardAccess service shall support validation of board data file integrity and consistency with configuration settings.

**REQ-BOARDACCESS-043**: When creating a new board, the BoardAccess service shall validate the board configuration using the RuleEngine before initialization.

**REQ-BOARDACCESS-044**: When creating a board, the BoardAccess service shall initialize the git repository structure and create proper board configuration and data files.

**REQ-BOARDACCESS-045**: When deleting a board, the BoardAccess service shall create a backup of board data and remove files safely with confirmation.

**REQ-BOARDACCESS-046**: When the operating system supports a trash/recycle bin, the BoardAccess service shall offer the user a choice between moving the board to trash (recoverable deletion) or permanent deletion, defaulting to trash for safety.

**REQ-BOARDACCESS-047**: When board lifecycle operations fail, the BoardAccess service shall maintain data integrity and provide rollback capabilities where possible.

**REQ-BOARDACCESS-048**: When performing git operations for board management, the BoardAccess service shall use the VersioningUtility service for all repository interactions.

**REQ-BOARDACCESS-049**: When validating board configurations, the BoardAccess service shall use the RuleEngine service for all validation operations.

## 4. Quality Attributes

### 4.1 Performance Requirements

**REQ-PERFORMANCE-001**: The BoardAccess service shall complete all single-task operations within 2 seconds under normal load conditions.

**REQ-PERFORMANCE-002**: The BoardAccess service shall support concurrent operations from multiple clients without data corruption.

### 4.2 Reliability Requirements  

**REQ-RELIABILITY-001**: When storage operations fail, the BoardAccess service shall return structured error information including failure reason and recovery suggestions.

**REQ-RELIABILITY-002**: The BoardAccess service shall maintain data consistency even when multiple operations are performed simultaneously.

**REQ-RELIABILITY-003**: When the underlying storage system is unavailable, the BoardAccess service shall fail gracefully with appropriate error messages.

### 4.3 Usability Requirements

**REQ-USABILITY-001**: The BoardAccess service shall provide clear error messages for all failure conditions that include specific information about what went wrong.

**REQ-USABILITY-002**: The BoardAccess service shall accept task data in a structured format that aligns with EisenKan domain models.

**REQ-USABILITY-003**: The change history generated by the TaskAccess shall allow tracing of creation, modification, and deletion of tasks.

**REQ-USABILITY-004**: The file format used to store data persistently shall not leak through the service interface.

## 5. Service Contract Requirements

### 5.1 Interface Operations
The BoardAccess service shall provide the following behavioral operations:

- **Create Task**: Accept task data with optional parent task identifier and optional priority promotion date, return unique identifier with success confirmation
- **Retrieve Single Task**: Accept task identifier and return complete task data
- **List Task Identifiers**: Return list with identifiers of all tasks with optional hierarchical filtering
- **Get Tasks Data**: Accept list of task identifiers and return corresponding task data with optional hierarchical information
- **Change Task Data**: Accept task identifier and updated data, apply changes with version history while maintaining parent-child relationships
- **Archive Task**: Accept task identifier and archive task data safely with cascade handling for subtasks
- **Remove Task**: Accept task identifier and remove task permanently with cascade handling for subtasks
- **Find Tasks**: Accept search criteria including parent task filters and priority promotion date filters and return matching tasks
- **Get Task History**: Accept task identifier and return version history information

#### IBoard Facet Operations
- **Discover Boards**: Accept directory path and return list of valid board locations with basic metadata
- **Extract Board Metadata**: Accept board path and return comprehensive board information including statistics
- **Validate Board Structure**: Accept board path and return validation results with detailed diagnostics
- **Load Board Configuration**: Accept configuration type and identifier, return board-level configuration data including board settings, column definitions, and workflow rules from git-based JSON storage
- **Store Board Configuration**: Accept configuration data with type specification, validate content, and persist to git-based JSON storage with atomic operations and versioning
- **Create Board**: Accept board parameters and configuration, initialize board structure with confirmation
- **Delete Board**: Accept board identifier and deletion parameters, create backup, and remove board to trash or permanently with confirmation
- **Get Board Statistics**: Accept board path and return calculated metrics for task distribution and activity

### 5.2 Data Contracts
The service shall work with these conceptual data entities:

**Task Data Entity**: Contains task identification, descriptive information, priority classification, workflow status, categorization tags, temporal tracking information, optional deadline specification, optional priority promotion date for Eisenhower matrix escalation, and optional parent task identifier for hierarchical relationships.

**Priority Classification**: Represents Eisenhower matrix categorization with urgent and important dimensions for task prioritization.

**Workflow Status**: Tracks current workflow position and maintains historical record of status transitions for task lifecycle management.

**Query Criteria**: Defines search parameters including priority filters, status constraints, tag selections, temporal range specifications, priority promotion date filters, parent task identifiers, and hierarchical level filters for task retrieval operations.

#### IBoard Facet Data Contracts

**Board Discovery Result Entity**: Contains discovered board paths, basic identification information, validation status, and git repository information for board location enumeration.

**Board Metadata Entity**: Provides comprehensive board information including title, description, creation and modification timestamps, task statistics, column distributions, configuration settings, and version information.

**Board Configuration Entity**: Contains board settings (title, description, ownership), column definitions (names, ordering, workflow mappings), workflow rules (transition constraints, validation rules), and schema version information.

**Board Statistics Entity**: Provides calculated metrics including task counts by status, column distributions, activity levels, last modification dates, and board health indicators.

**Board Validation Result Entity**: Contains validation status, detailed diagnostic information, identified issues with severity levels, recommended fixes, and structural integrity assessment.

**Board Configuration Request Entity**: Contains configuration type specification (board, columns, workflows, rules), optional configuration identifier, and operation metadata for configuration load/store operations.

**Board Configuration Data Entity**: Provides structured board-level configuration including board settings (name, description, ownership), column definitions (names, ordering, workflow mappings), workflow rules (transition constraints, validation rules), and visual settings (themes, layouts) in JSON-serializable format.

**Board Configuration Response Entity**: Contains configuration operation results, configuration data payload, version information from git storage, validation status, and operation confirmation details for service consumption.

**Board Deletion Request Entity**: Contains deletion parameters including board identifier, deletion method preference (trash/permanent), OS trash capability detection, user confirmation status, and backup requirements for safe board removal operations.

### 5.3 Error Handling
All errors shall include:
- Error code classification  
- Human-readable error message
- Technical details for debugging
- Suggested recovery actions where applicable

## 6. Technical Constraints

### 6.1 Integration Requirements
**REQ-INTEGRATION-001**: The BoardAccess service shall use the VersioningUtility service for all persistent storage operations.

**REQ-INTEGRATION-002**: The BoardAccess service shall use the LoggingUtility service for all operational logging.

**REQ-INTEGRATION-003**: The BoardAccess service shall operate within the ResourceAccess architectural layer constraints.

### 6.2 Data Format Requirements
**REQ-FORMAT-001**: The BoardAccess service shall store task data in JSON format for human readability and version control compatibility.

**REQ-FORMAT-002**: The BoardAccess service shall use a JSON data structure optimized to keep the differences between two file versions minimal for the common operations, e.g. when moving a task to another column.

**REQ-FORMAT-003**: The BoardAccess service shall organize data of active tasks in one and data of archived tasks in another file.

**REQ-FORMAT-004**: The BoardAccess service shall maintain parent-child relationships in the directory structure to support hierarchical queries.

## 7. Acceptance Criteria

### 7.1 Functional Acceptance
- All requirements REQ-BOARDACCESS-001 through REQ-BOARDACCESS-049 are met
- All operations OP-1 through OP-12 are fully supported
- Priority promotion date functionality is fully supported for storage, retrieval, and querying
- Board management functionality through IBoard facet is fully operational including configuration management
- Service operations complete within performance requirements
- Error conditions are handled gracefully with appropriate messaging

### 7.2 Quality Acceptance  
- All Quality Attribute requirements are met
- All error scenarios return structured, actionable error information

### 7.3 Integration Acceptance
- Service integrates successfully with VersioningUtility for storage operations
- Service integrates successfully with LoggingUtility for operational visibility
- Service integrates successfully with RuleEngine for board configuration validation
- Service can be consumed by business logic layers (Engines/Managers) without coupling
- IBoard facet supports TaskManager board operations (OP-9 to OP-13) from PROJECT_PLAN.md

---

**Document Version**: 1.1
**Created**: 2025-09-07
**Updated**: 2025-09-20
**Changes**: Moved IConfiguration to IBoard, added operations to support board management
**Status**: Accepted