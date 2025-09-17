# TaskManager Software Requirements Specification (SRS)

## 1. Introduction

### 1.1 Purpose
This Software Requirements Specification defines the requirements for the TaskManager service, a Manager layer component that orchestrates task-related business workflows in the EisenKan Kanban application. The service coordinates task operations, enforces business rules, and manages workflow transitions.

### 1.2 Scope
TaskManager is responsible for:
- Single point of entry for all clients except external systems
- Business workflow orchestration for task and subtask lifecycle management
- Task and subtask data validation and business rule enforcement
- State transition control between Kanban workflow stages
- Coordination with Resource Access components for data persistence
- Integration with business rule engines for validation logic

### 1.3 System Context
TaskManager operates in the Manager layer of the EisenKan architecture, serving as the primary orchestrator for task-related business processes. It coordinates between Clients (providing task management interfaces) and lower layers including Engines (business logic) and ResourceAccess components (data persistence), while maintaining stateless workflow-focused responsibilities.

## 2. Operations

The following operations define the required behavior for TaskManager:

#### OP-1: Create Task
**Actors**: EisenKan Client
**Trigger**: When a user creates a new task or subtask through the interface
**Flow**:
1. Receive task creation request with task data, priority information, and optional parent task identifier
2. Verify request using rules engine
3. Delegate to BoardAccess for persistent storage
4. Return task creation confirmation with assigned identifier and return the task data containing information set by the backend like the creation date

#### OP-2: Modify Task
**Actors**: EisenKan Client
**Trigger**: When a user updates existing task information
**Flow**:
1. Receive task modification request with updated data
2. Verify request using rules engine
3. Delegate to BoardAccess for persistent storage
4. Return modification confirmation and return the task data containing information set by the backend like the changed date

#### OP-3: Change Task Workflow Status
**Actors**: EisenKan Client
**Trigger**: When a user moves a task or subtask between Kanban columns
**Flow**:
1. Receive task status change request (Todo → In Progress → Done)
2. Verify request using rules engine
3. If subtask workflow coupling applies, trigger additional parent task status changes
4. Delegate to BoardAccess for persistent storage of all affected tasks
5. Return status change confirmation and return the task data containing information set by the backend like the changed date

#### OP-4: Retrieve Task Data
**Actors**: EisenKan Client
**Trigger**: When task information is needed for display or operations
**Flow**:
1. Receive task data request with query parameters
2. Fetch data from BoardAccess
3. Apply any business logic for data presentation
4. Return formatted task data to requesting client

#### OP-5: Delete Task
**Actors**: EisenKan Client
**Trigger**: When a user removes a task from the system
**Flow**:
1. Receive task deletion request
2. Verify deletion permissions including subtask cascade policies
3. Coordinate with BoardAccess for data removal or archival with cascade handling
4. Return deletion confirmation

#### OP-6: Process Priority Promotions
**Actors**: EisenKan System (automated process)
**Trigger**: When system processes tasks with promotion dates that have been reached
**Flow**:
1. Query tasks with priority promotion dates on or before current date
2. Validate current priority allows escalation (not-urgent-important → urgent-important)
3. Update task priority to next urgency level in Eisenhower matrix
4. Clear promotion date after successful escalation
5. Log priority promotion action for audit trail

#### OP-7: Load Context (IContext Facet)
**Actors**: EisenKan Client
**Trigger**: When client needs to restore UI context and user preferences
**Flow**:
1. Receive context load request with context type specification
2. Delegate to git-based storage for context data retrieval
3. Parse and validate JSON context data
4. Return context data including window states, user preferences, view configurations, and session data

#### OP-8: Store Context (IContext Facet)
**Actors**: EisenKan Client
**Trigger**: When client needs to persist UI context and user preferences
**Flow**:
1. Receive context store request with context data and type specification
2. Validate context data structure and content
3. Serialize context data to JSON format
4. Delegate to git-based storage for atomic persistence with versioning
5. Return context storage confirmation

## 3. Functional Requirements

### 3.1 Task Creation Requirements

**REQ-TASKMANAGER-001**: When a task creation request is received with complete data, the TaskManager service shall validate the data and coordinate storage through BoardAccess.

**REQ-TASKMANAGER-002**: When task creation data is incomplete or violates business rules, the TaskManager service shall reject the request with structured error information.

**REQ-TASKMANAGER-003**: When creating a task, the TaskManager service shall preserve priority information received from the Client and ensure it is stored correctly.

**REQ-TASKMANAGER-015**: When creating a subtask with a parent task identifier, the TaskManager service shall validate parent task existence and enforce the 1-2 level hierarchy constraint.

### 3.2 Priority Promotion Requirements

**REQ-TASKMANAGER-019**: When creating or updating a task with priority promotion date, the TaskManager service shall validate the promotion date is in the future and store it for automated priority escalation.

**REQ-TASKMANAGER-020**: When a task's priority promotion date is reached, the TaskManager service shall automatically escalate the task's priority to the next urgency level in the Eisenhower matrix (not-urgent-important → urgent-important).

**REQ-TASKMANAGER-021**: The TaskManager service shall support querying tasks by priority promotion date to enable automated processing of priority escalations.

### 3.3 Task Modification Requirements

**REQ-TASKMANAGER-004**: When a valid task modification request is received, the TaskManager service shall validate changes and coordinate updates through BoardAccess.

**REQ-TASKMANAGER-005**: When task modification violates business rules or references non-existent tasks, the TaskManager service shall reject the modification and return appropriate error information.

**REQ-TASKMANAGER-006**: The TaskManager service shall support partial task updates without requiring complete task data in modification requests.

### 3.4 Workflow State Management Requirements

**REQ-TASKMANAGER-007**: When a task status change request is received, the TaskManager service shall validate the state transition against business rules before applying changes.

**REQ-TASKMANAGER-008**: The TaskManager service shall enforce valid Kanban workflow transitions (Todo → In Progress → Done) and reject invalid transitions.

**REQ-TASKMANAGER-009**: When applying workflow status changes, the TaskManager service shall coordinate with the RuleEngine to ensure business rule compliance.

**REQ-TASKMANAGER-016**: Depending on the active subtask policy checked by the RuleEngine, when the first subtask of a parent task moves from "todo" to "doing", the TaskManager service shall automatically move the parent task from "todo" to "doing" if the parent is currently in "todo" status.

**REQ-TASKMANAGER-017**: Depending on the active subtask policy checked by the RuleEngine, when a parent task is requested to move to "done" status, the TaskManager service shall automatically move subtasks to "done".

### 3.5 Data Retrieval Requirements

**REQ-TASKMANAGER-010**: When task data is requested, the TaskManager service shall coordinate with BoardAccess to retrieve and return complete task information.

**REQ-TASKMANAGER-011**: The TaskManager service shall support querying tasks by multiple criteria including priority, status, and other task attributes.

**REQ-TASKMANAGER-012**: When queried data does not exist, the TaskManager service shall return appropriate not-found responses without errors.

### 3.6 Task Deletion Requirements

**REQ-TASKMANAGER-013**: When a task deletion request is received, the TaskManager service shall validate deletion permissions and coordinate removal through BoardAccess.

**REQ-TASKMANAGER-014**: When deleting non-existent tasks, the TaskManager service shall handle the request gracefully without errors (idempotent operation).

**REQ-TASKMANAGER-018**: When a parent task is archived or deleted, the TaskManager service shall handle cascade operations for its subtasks according to configured cascade policy.

### 3.7 Context Management Requirements (IContext Facet)

**REQ-TASKMANAGER-019**: When a context load request is received, the TaskManager service shall retrieve context data from git-based JSON storage and return parsed context information.

**REQ-TASKMANAGER-021**: When a context store request is received, the TaskManager service shall pass it to the persistency component.

**REQ-TASKMANAGER-022**: When the persistency component rejects the context data, the TaskManager service shall return detailed error information.

## 4. Quality Attributes

### 4.1 Performance Requirements

**REQ-PERFORMANCE-001**: The TaskManager service shall complete all workflow operations within 3 seconds under normal load conditions.

**REQ-PERFORMANCE-002**: The TaskManager service shall maintain stateless operation to support concurrent request handling.

### 4.2 Reliability Requirements

**REQ-RELIABILITY-001**: When business rule validation fails, the TaskManager service shall return structured error information including specific rule violations.

**REQ-RELIABILITY-002**: When dependent services are unavailable, the TaskManager service shall fail gracefully with appropriate error messages and recovery suggestions.

**REQ-RELIABILITY-003**: The TaskManager service shall maintain workflow consistency even when multiple operations are performed simultaneously.

### 4.3 Usability Requirements

**REQ-USABILITY-001**: The TaskManager service shall provide clear, actionable error messages for all business rule violations and validation failures.

**REQ-USABILITY-002**: The TaskManager service shall accept and return task data in formats compatible with EisenKan Client interfaces.

**REQ-USABILITY-003**: All TaskManager operations shall provide confirmation responses that include relevant operation details for client feedback.

## 5. Service Contract Requirements

### 5.1 Interface Operations
The TaskManager service shall provide the following behavioral operations:

- **Create Task**: Accept task data with priority, optional priority promotion date, and optional parent task identifier, coordinate validation, rule checking, and storage
- **Update Task**: Accept task modifications including priority promotion date changes, validate changes, and coordinate persistence while maintaining hierarchical relationships
- **Change Task Status**: Accept workflow status changes, validate transitions, apply business rules including subtask workflow coupling
- **Get Task**: Accept task queries and coordinate data retrieval with proper formatting including priority promotion information
- **Delete Task**: Accept task removal requests, validate permissions, and coordinate deletion with cascade handling
- **List Tasks**: Accept query criteria including hierarchical filters and priority promotion date filters and return matching task collections
- **Validate Task**: Accept task data for validation without persistence (for client-side validation) including subtask constraints and priority promotion date validation
- **Process Priority Promotions**: Query and automatically escalate tasks with reached promotion dates from not-urgent-important to urgent-important priority level

#### IContext Facet Operations
- **Load Context**: Accept context type specification and return UI context data including window states, user preferences, view configurations, and session data from git-based JSON storage
- **Store Context**: Accept context data with type specification, validate content, and persist to git-based JSON storage with atomic operations and versioning

### 5.2 Data Contracts
The service shall work with these conceptual data entities:

**Task Request Entity**: Contains task identification, descriptive information, priority classification (from Client's Eisenhower matrix assignment), workflow status specification, categorization tags, optional deadline information, optional priority promotion date for Eisenhower matrix escalation, and optional parent task identifier for subtask creation.

**Task Response Entity**: Provides complete task information including all task attributes, current workflow status, priority promotion date for escalation tracking, hierarchical information (parent/subtask relationships), and operation confirmation details for client consumption.

**Workflow Status Entity**: Represents current position in Kanban workflow (Todo, In Progress, Done) with validation rules for valid state transitions.

**Validation Result Entity**: Contains business rule validation outcomes, error information, and suggested corrections for failed validations.

#### IContext Facet Data Contracts

**Context Request Entity**: Contains context type specification (window, preferences, views, sessions), optional context identifier, and operation metadata for context load/store operations.

**Context Data Entity**: Provides structured context information including window states (positions, sizes, monitor configurations), user preferences (themes, settings, customizations), view configurations (panel states, filters, sort orders), and session data (temporary states, recent items) in JSON-serializable format.

**Context Response Entity**: Contains context operation results, context data payload, version information from git storage, and operation confirmation details for client consumption.

### 5.3 Error Handling
All errors shall include:
- Business rule violation details
- Workflow validation failure information
- Technical error codes and messages
- Suggested corrective actions for common failures

## 6. Technical Constraints

### 6.1 Integration Requirements
**REQ-INTEGRATION-001**: The TaskManager service shall coordinate with BoardAccess for all task data persistence operations.

**REQ-INTEGRATION-002**: The TaskManager service shall use the RuleEngine service for business rule validation and workflow compliance checking.

**REQ-INTEGRATION-003**: The TaskManager service shall use the LoggingUtility service for all workflow operation logging.

**REQ-INTEGRATION-004**: The TaskManager service shall operate within the Manager architectural layer constraints and maintain stateless operation.

### 6.2 Business Rule Requirements
**REQ-BUSINESS-001**: The TaskManager service shall enforce Kanban workflow state transitions and reject invalid status changes.

**REQ-BUSINESS-002**: The TaskManager service shall preserve priority information as received from Clients without modification or reordering logic.

**REQ-BUSINESS-003**: The TaskManager service shall validate task data completeness and business rule compliance before persistence operations.

## 7. Acceptance Criteria

### 7.1 Functional Acceptance
- All requirements REQ-TASKMANAGER-001 through REQ-TASKMANAGER-021 are met
- All operations OP-1 through OP-6 are fully supported
- Priority promotion functionality works correctly for Eisenhower matrix escalation
- Workflow orchestration operates correctly with proper validation and error handling
- Business rule enforcement functions correctly with RuleEngine integration

### 7.2 Quality Acceptance
- All Quality Attribute requirements are met
- Service maintains stateless operation under concurrent load
- All error scenarios return structured, actionable information

### 7.3 Integration Acceptance
- Service integrates successfully with BoardAccess for data operations
- Service integrates successfully with RuleEngine for business rule validation
- Service integrates successfully with LoggingUtility for operational logging
- Service can be consumed by Client layer components without architectural violations

---

**Document Version**: 1.0  
**Created**: 2025-09-13
**Updated**: 2025-09-17
**Status**: Accepted