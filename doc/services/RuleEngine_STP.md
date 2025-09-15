# RuleEngine Software Test Plan (STP)

## 1. Test Overview

### 1.1 Purpose
This Software Test Plan defines destructive testing strategies and comprehensive requirements verification for the RuleEngine service. The plan emphasizes API boundary testing, error condition validation, and complete traceability to all EARS requirements specified in [RuleEngine_SRS.md](RuleEngine_SRS.md).

### 1.2 Scope
Testing covers destructive API testing, requirements verification, error condition handling, performance degradation scenarios, and graceful degradation validation for Kanban rule evaluation including WIP limits for tasks and subtasks, workflow transitions, definition of ready/done, subtask workflow coupling, parent-child dependency validation, and age-based task management capabilities.

### 1.3 Test Environment Requirements
- Go 1.24.3+ runtime environment with race detector support
- Memory and CPU profiling capabilities
- Rule set test data with various complexity levels (WIP limits for tasks and subtasks, workflow transitions, subtask workflow coupling rules, subtask cascading rules, parent-child dependency rules, age limits)
- Concurrent execution environment (goroutine support)
- LoggingUtility service for operational logging
- RulesAccess service for providing rule sets
- BoardAccess service for enriched rule evaluation context
- Mock rule sets for boundary condition testing
- Test board directories with git repositories for integration testing

## 2. Test Strategy

This STP emphasizes breaking the system through:
- **API Contract Violations**: Invalid TaskEvent structures, malformed board paths, type mismatches
- **Resource Exhaustion**: Memory limits, large rule sets, complex rule conditions, oversized board data
- **Rule Logic Edge Cases**: WIP limit boundary conditions, invalid workflow transitions, malformed rule expressions
- **Performance Degradation**: Large rule sets, complex evaluation scenarios, concurrent load, BoardAccess integration overhead
- **Requirements Verification Tests**: Validate all EARS requirements REQ-RULEENGINE-001 through REQ-RULEENGINE-007 and REQ-RULETYPE-001 through REQ-RULETYPE-009 with negative cases
- **Error Recovery Tests**: Test graceful degradation when BoardAccess fails, rule loading errors
- **Concurrency Stress Testing**: Test race conditions and consistency under concurrent rule evaluation with shared board access

## 3. Destructive API Test Cases

### 3.1 API Contract Violations

**Test Case DT-API-001**: Rule Evaluation with Invalid Inputs
- **Objective**: Test API contract violations for EvaluateTaskChange operation
- **Destructive Inputs**:
  - nil TaskEvent context
  - TaskEvent with missing required fields (event type, future state)
  - TaskEvent with invalid data types in Task fields
  - TaskEvent with extremely large task descriptions (>10KB)
  - TaskEvent with invalid unicode characters
  - TaskEvent with malformed Priority or WorkflowStatus values
  - Invalid or non-existent board path strings
  - Board paths pointing to non-git directories
  - Board paths with permission access issues
  - Empty or nil rule sets from RulesAccess
  - Rules with missing required Actions field
  - Rules with invalid trigger types not matching event types
  - Rules with malformed condition expressions
  - TaskEvent with invalid parent-child relationships (subtask without parent, circular references)
  - TaskEvent for subtask operations with non-existent parent tasks
  - TaskEvent attempting to create >2 level hierarchies
  - TaskEvent with invalid subtask workflow coupling data
  - TaskEvent with malformed hierarchical context information
- **Expected**:
  - Service handles nil TaskEvent gracefully without crashes
  - Missing required fields are detected and return structured errors
  - Invalid data types are validated and rejected with clear messages
  - Large inputs are handled appropriately without memory exhaustion
  - Invalid board paths return appropriate error messages
  - Malformed rules are rejected with detailed error information
  - Unicode handling is correct throughout evaluation
  - Invalid parent-child relationships are detected and rejected
  - Non-existent parent task references are validated and rejected
  - Hierarchy constraint violations are enforced (1-2 levels only)
  - Invalid subtask workflow coupling data is handled appropriately
  - Malformed hierarchical context is detected and rejected

### 3.2 Rule Logic Edge Cases

**Test Case DT-LOGIC-001**: Rule Condition and Configuration Edge Cases
- **Objective**: Test Kanban rule evaluation under complex conditions and edge cases
- **Edge Case Scenarios**:
  - WIP limit rules with zero or negative limits for tasks and subtasks
  - WIP limit rules with extremely high limits (>10000) for tasks and subtasks
  - WIP limit rules with mismatched task/subtask configurations
  - Required field rules referencing non-existent task properties
  - Workflow transition rules with invalid column names
  - Workflow transition rules with circular transition definitions
  - Age limit rules with zero or negative age thresholds
  - Age limit rules with timestamp parsing edge cases
  - Rules with conditions referencing non-existent board metadata
  - Rules with invalid priority values (negative, overflow)
  - Rules with malformed Actions field structures
  - Rules with unsupported category values
  - Rules targeting non-existent event types
  - Subtask workflow coupling rules with invalid parent-child combinations
  - Subtask completion dependency rules with circular dependencies
  - Subtask hierarchy rules with invalid depth specifications
  - Parent task archival rules with conflicting cascade policies
  - Subtask workflow coupling rules with invalid trigger conditions
  - Rules with disabled status but referenced by other rules
  - BoardAccess returning malformed WIP count data
  - BoardAccess returning invalid task history data
  - BoardAccess integration failures during rule evaluation
- **Expected**:
  - WIP limit edge cases are handled with appropriate error messages
  - Required field validation handles missing properties gracefully
  - Workflow transition validation rejects invalid configurations
  - Age limit calculations handle timestamp edge cases correctly
  - Invalid rule configurations are rejected with detailed diagnostics
  - BoardAccess integration failures result in partial evaluation results
  - System continues operating despite individual rule evaluation failures

**Test Case DT-LOGIC-002**: Rule Priority and Conflict Resolution
- **Objective**: Test rule evaluation when multiple rules conflict or have complex priorities
- **Conflict Scenarios**:
  - Multiple rules matching the same event with different verdicts
  - Rules with identical priority values
  - Rules with priority values exceeding integer ranges
  - Rules with negative priority values
  - Rules where higher priority rules contradict lower priority rules
  - Rules with dependencies that create execution order conflicts
- **Expected**:
  - Priority ordering is respected consistently
  - Identical priorities are handled deterministically
  - Invalid priority values are handled gracefully
  - Rule conflicts are detected and reported clearly
  - Dependency conflicts are resolved or reported appropriately

**Test Case DT-LOGIC-003**: Subtask Rule Evaluation Edge Cases
- **Objective**: Test subtask-specific rule evaluation scenarios including workflow coupling and dependency validation
- **Edge Case Scenarios**:
  - First subtask "todo"→"doing" transition with parent already in "doing" status
  - First subtask "todo"→"doing" transition with parent in "done" status (invalid)
  - Parent task "todo"→"done" and "doing"→"done" transition with non-done subtasks
  - Parent task "todo"→"done" and "doing"→"done" transition with all subtasks done
  - Simultaneous subtask completion attempts by multiple subtasks
  - Parent task deletion with mixed subtask completion states
  - Parent task archival with mixed subtask completion states
  - Subtask creation under non-existent parent tasks
  - Subtask creation under subtasks (violating 1-2 level constraint)
  - WIP limit evaluation with separate task and subtask limits
  - WIP limit evaluation when limits are exceeded for tasks but not subtasks
  - WIP limit evaluation when limits are exceeded for subtasks but not tasks
  - Concurrent subtask workflow changes affecting same parent
  - Rule evaluation with corrupted hierarchical relationship data
- **Expected**:
  - Workflow coupling rules are enforced correctly
  - Parent completion dependency rules prevent invalid transitions
  - Hierarchy constraint violations are detected and rejected
  - Separate WIP limits for tasks and subtasks are properly evaluated
  - Concurrent subtask operations maintain parent-child consistency
  - Corrupted hierarchical data is detected and handled gracefully
  - Invalid parent task references are validated and rejected

## 4. Performance and Resource Testing

### 4.1 Performance Degradation Testing

**Test Case DT-PERFORMANCE-001**: Large Rule Set Evaluation
- **Objective**: Test performance degradation with large numbers of rules
- **Method**:
  - Evaluate task changes against rule sets with 1, 10, 100, 1000, 10000 rules
  - Test rule sets with varying complexity (simple vs complex conditions)
  - Monitor: CPU usage, memory usage, evaluation time
  - Measure: Average latency and 99th percentile response times
  - Test concurrent evaluations from multiple goroutines
- **Expected**:
  - Evaluation completes within 500ms for up to 100 rules (SRS requirement)
  - Memory usage grows predictably with rule set size
  - CPU usage remains reasonable under concurrent load
  - Performance degrades gracefully beyond optimal rule counts
  - System remains responsive during evaluation

**Test Case DT-RESOURCE-001**: Memory and Resource Exhaustion
- **Objective**: Test behavior under memory pressure and resource limits
- **Method**:
  - Evaluate extremely large rule sets (50,000+ rules)
  - Test rules with very large condition expressions
  - Test concurrent evaluations across many goroutines (200+ concurrent)
  - Monitor memory usage, garbage collection, and resource cleanup
  - Test with limited available memory scenarios
- **Expected**:
  - System fails gracefully when resource limits are reached
  - No memory leaks detected during or after evaluation
  - Garbage collection doesn't cause excessive delays
  - Resource cleanup occurs properly after evaluations
  - Error messages indicate resource constraint issues clearly

## 5. Error Condition Testing

### 5.1 Runtime Error Testing

**Test Case DT-ERROR-001**: Runtime Evaluation Errors
- **Objective**: Test error handling during rule evaluation execution
- **Error Scenarios**:
  - Rules that throw exceptions during condition evaluation
  - Rules that access invalid memory or cause segmentation faults
  - Rules that cause arithmetic overflow/underflow
  - Rules that attempt to access restricted system resources
  - Rules that cause stack overflow through deep recursion
- **Expected**: Runtime errors are caught and reported without crashing the service

### 5.2 Concurrent Access Testing

**Test Case DT-CONCURRENT-001**: Race Condition Testing
- **Objective**: Verify thread safety and data consistency under concurrent access
- **Method**:
  - Concurrent rule evaluations from multiple goroutines (200+ concurrent)
  - Simultaneous evaluation of same task event across multiple threads
  - Mixed read operations with varying rule sets
  - Stress test with rapid evaluation requests
  - Use Go race detector to identify data races
- **Expected**: No race conditions detected, consistent evaluation results, thread-safe operation

## 6. Recovery and Degradation Testing

### 6.1 Graceful Degradation

**Test Case DT-RECOVERY-001**: Service Recovery from Failures
- **Objective**: Test recovery capabilities after various failure conditions
- **Recovery Scenarios**:
  - Recovery from invalid rule set loading
  - Recovery from memory exhaustion conditions
  - Recovery from evaluation timeout conditions
  - Recovery from logging service failures
- **Expected**: Service recovers automatically and continues normal operation

**Test Case DT-RECOVERY-002**: Partial Functionality Under Constraints
- **Objective**: Test continued operation under resource constraints
- **Constraint Scenarios**:
  - Limited memory availability
  - Reduced CPU resources
  - Logging service degradation
  - High concurrent load conditions
- **Expected**: Core rule evaluation functionality maintained, graceful performance degradation

## 7. Test Execution Requirements

### 7.1 Required Tools and Environment
- Go race detector (`go test -race`)
- Memory profiling tools (`go test -memprofile`)
- CPU Profiling: Enabled (`go test -cpuprofile`)
- Concurrent load generation tools
- Rule set generators for large-scale testing
- Performance monitoring utilities
- LoggingUtility service integration for testing

### 7.2 Success Criteria
- **100% Requirements Coverage**: All EARS requirements REQ-RULEENGINE-001 through REQ-RULEENGINE-005 have corresponding destructive tests
- **Zero Critical Failures**: No crashes, memory leaks, or data corruption
- **Race Detector Clean**: No race conditions detected under any scenario
- **Performance Requirements Met**: 100 rules evaluated within 500ms requirement maintained under adverse conditions
- **BoardAccess Integration**: All enriched context scenarios tested including WIP counts, task history, and board metadata
- **Kanban Rule Types**: All rule categories (validation, workflow, automation, notification) tested with destructive scenarios
- **Graceful Error Handling**: All error conditions handled without service failures, including BoardAccess failures
- **Complete Recovery**: Service recovers from all testable failure conditions including rule loading and board access errors
- **Rule Evaluation Consistency**: Task change decisions remain consistent across all failure and recovery scenarios

---

**Document Version**: 1.0  
**Created**: 2025-09-12
**Updated**: 2025-09-14
**Status**: Accepted