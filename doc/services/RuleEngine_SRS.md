# RuleEngine Software Requirements Specification (SRS)

## 1. Introduction

### 1.1 Purpose
This Software Requirements Specification defines the requirements for the RuleEngine service, an Engines layer component that provides business rule processing capabilities for the EisenKan system. The service encapsulates rule evaluation logic and provides atomic rule execution operations, enabling workflow customization and automated task management based on configurable business rules.

### 1.2 Scope
RuleEngine is responsible for:
- Business rule evaluation for Kanban workflow management (WIP limits, workflow transitions, definition of ready/done)
- Task change validation against configurable business rules with comprehensive board context
- Rule priority ordering and violation aggregation for complete violation reporting
- Support for extensible rule categories (validation, workflow, automation, notification)
- Integration with BoardAccess for enriched rule evaluation context (WIP counts, task history, column timestamps)
- Age-based task management rules and subtask dependency validation

### 1.3 System Context
RuleEngine operates in the Engines layer of the EisenKan architecture, accessing RulesAccess for rule definitions and BoardAccess for enriched board context. It provides stateless rule evaluation services to the Manager layer, supporting Kanban-specific business rules including WIP limits, workflow transitions, definition of ready/done criteria, and age-based task management.

## 2. Operations

The following operations define the required behavior for RuleEngine:

#### OP-1: Evaluate Task Change
**Actors**: TaskManager
**Trigger**: When a task state change is requested
**Flow**:
1. Receive TaskEvent with current and future task states
2. Fetch applicable rule set from RulesAccess for the board
3. Filter rules based on event type and enabled status
4. Enrich evaluation context with board data from BoardAccess (WIP counts, task history, column timestamps)
5. Evaluate all applicable rules sequentially and aggregate violations
6. Return RuleEvaluationResult with allowed status and violation details

#### OP-2: Close Engine Resources
**Actors**: TaskManager
**Trigger**: When shutting down RuleEngine
**Flow**:
1. Receive close request
2. Release any held resources
3. Log shutdown completion
4. Return success status

## 3. Functional Requirements

### 3.1 Rule Evaluation Requirements

**REQ-RULEENGINE-001**: When a task change request is submitted with a TaskEvent, the RuleEngine shall evaluate all applicable rules within 500ms and return a RuleEvaluationResult indicating whether the change is allowed.

**REQ-RULEENGINE-002**: Where rule violations are detected during task change evaluation, the RuleEngine shall return violation details including rule ID, priority, message, category, and optional details.

**REQ-RULEENGINE-003**: When evaluating rules, the RuleEngine shall access BoardAccess to obtain enriched context including WIP counts, task history, column timestamps, and board metadata for comprehensive rule evaluation.

**REQ-RULEENGINE-004**: When multiple rules apply to the same task event, the RuleEngine shall evaluate all applicable rules and aggregate violations sorted by priority (higher priority first).

**REQ-RULEENGINE-005**: When no applicable rules are found for a task event, the RuleEngine shall allow the task change by default.

## 4. Quality Attributes

### 4.1 Performance Requirements

**REQ-PERFORMANCE-001**: The RuleEngine shall complete rule evaluation for up to 100 rules within 500 milliseconds under normal load conditions on a MacAir M4.

### 4.2 Reliability Requirements

**REQ-RELIABILITY-001**: When rule evaluation encounters invalid rule definitions, the RuleEngine shall fail the operation with a helpful error.

**REQ-RELIABILITY-002**: The RuleEngine shall maintain stateless operation to ensure consistent rule evaluation results for identical inputs.

### 4.3 Usability Requirements

**REQ-USABILITY-001**: The RuleEngine shall provide clear error messages for all rule evaluation failures that include specific rule identification and condition details.

**REQ-USABILITY-002**: The RuleEngine shall accept rule data in the structured format defined by RulesAccess without format conversion.


### 4.4 Extensibility Requirements

**REQ-EXTENSIBILITY-001**: The RuleEngine shall support extensible trigger types to accommodate different workflow events (task transitions, due dates, status changes).

## 5. Service Contract Requirements

### 5.1 Interface Operations
The RuleEngine service shall provide the following behavioral operations:

- **EvaluateTaskChange**: Accept TaskEvent and board path, return RuleEvaluationResult with allowed status and violation details
- **Close**: Release any resources held by the engine and perform cleanup

### 5.2 Data Contracts
The service shall work with these conceptual data entities:

**TaskEvent**: Contains event type (task_transition, task_create, task_update), current state (TaskWithTimestamps), future state (TaskState), and timestamp.

**TaskState**: Contains Task, Priority, and WorkflowStatus representing the intended state.

**RuleEvaluationResult**: Contains Allowed boolean and Violations array with detailed rule violation information.

**RuleViolation**: Contains RuleID, Priority, Message, Category, and optional Details for specific violation context.

**EnrichedContext**: Contains TaskEvent, WIP counts, task history, subtasks, column tasks, column enter times, and board metadata for comprehensive rule evaluation.

### 5.3 Error Handling
All errors shall include:
- Error code classification
- Human-readable error message  
- Technical details for debugging
- Suggested recovery actions where applicable
- Rule identification and context information when applicable

## 6. Technical Constraints

### 6.1 Integration Requirements
**REQ-INTEGRATION-001**: The RuleEngine service shall receive rule sets from RulesAccess component for rule definitions.

**REQ-INTEGRATION-002**: The RuleEngine service shall use BoardAccess component to obtain enriched board context including WIP counts, task history, and column metadata.

**REQ-INTEGRATION-003**: The RuleEngine service shall use the LoggingUtility service for all operational logging.

**REQ-INTEGRATION-004**: The RuleEngine service shall operate within the Engines architectural layer constraints, maintain stateless operation, and only access ResourceAccess and Utilities components.

### 6.2 Data Format Requirements
**REQ-FORMAT-001**: The RuleEngine service shall process rule data in the format provided by RulesAccess without format transformation.

**REQ-FORMAT-002**: The RuleEngine service shall support rule categories including validation (WIP limits, required fields), workflow (allowed transitions), automation (age limits), and notification rules.

**REQ-FORMAT-003**: The RuleEngine service shall return rule evaluation results in structured JSON-serializable format suitable for Manager layer orchestration.

### 6.3 Rule Type Requirements
**REQ-RULETYPE-001**: The RuleEngine service shall support WIP limit rules that prevent exceeding configurable task counts per column.

**REQ-RULETYPE-002**: The RuleEngine service shall support required field rules that validate task completeness before column transitions.

**REQ-RULETYPE-003**: The RuleEngine service shall support workflow transition rules that enforce allowed column-to-column movements.

**REQ-RULETYPE-004**: The RuleEngine service shall support age limit rules that warn when tasks remain in columns beyond configurable time thresholds.

## 7. Acceptance Criteria

### 7.1 Functional Acceptance
- All functional requirements REQ-RULEENGINE-001 through REQ-RULEENGINE-002 are met
- Operation OP-1 is fully supported  
- Service operations complete within performance requirements
- Error conditions are handled gracefully with appropriate messaging
- Rule evaluation produces consistent and deterministic results

### 7.2 Quality Acceptance
- All Quality Attribute requirements are met
- All error scenarios return structured, actionable error information
- Rule processing maintains stateless operation across concurrent requests
- Rule evaluation performance meets specified latency requirements

### 7.3 Integration Acceptance  
- Service integrates successfully with TaskManager for rule processing workflow
- Service integrates successfully with LoggingUtility for operational visibility
- Service can process rule sets from RulesAccess without data transformation
- Service follows iDesign Engines layer patterns and maintains architectural compliance
- Service supports extensible trigger types for workflow customization

---

**Document Version**: 1.0  
**Created**: 2025-09-12
**Updated**: 2025-09-13
**Status**: Accepted