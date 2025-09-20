# VersioningUtility Software Test Plan (STP)

## 1. Test Overview

### 1.1 Purpose
This Software Test Plan defines destructive testing strategies for the VersioningUtility service. The plan emphasizes API boundary testing, error condition validation, and destructive testing approaches for all EARS requirements specified in [VersioningUtility_SRS.md](VersioningUtility_SRS.md). Requirements verification and actual test execution results are documented in [VersioningUtility_STR.md](VersioningUtility_STR.md).

### 1.2 Scope
Testing covers destructive API testing, requirements verification, error condition handling, resource exhaustion scenarios, and graceful degradation validation for all interface operations and version control capabilities.

### 1.3 Test Environment Requirements
- Go 1.24.3+ runtime environment with race detector support
- File system with permission control capabilities
- Git repository testing utilities
- Memory and resource monitoring tools
- Concurrent execution environment (goroutine support)
- Large file generation capabilities (up to 100MB)

## 2. Test Strategy

This STP emphasizes breaking the system through:
- **API Contract Violations**: Invalid, extreme, and malformed inputs, boundary violations, type mismatches
- **Resource Exhaustion**: Memory limits, file handle exhaustion, repository size limits
- **External Dependency Failures**: File system errors, permission issues, repository corruption
- **Configuration Corruption**: Invalid repository paths, corrupted git structures
- **Requirements Verification Tests**: Validate all EARS requirements with negative cases
- **Error Recovery Tests**: Test graceful degradation and recovery
- **Concurrency Stress Testing**: Test race conditions and repository corruption under stress

## 3. Destructive API Test Cases

### 3.1 API Contract Violations

**Test Case DT-API-001**: InitializeRepository with invalid or unusual inputs
- **Objective**: Test API contract violations for repository initialization
- **Destructive Inputs**:
  - nil/empty repository paths
  - Paths with invalid unicode characters
  - Paths with binary data and control characters
  - Non-existent directory paths
  - Paths exceeding filesystem limits (>4096 chars)
  - Read-only directory paths
  - Directory paths with insufficient permissions
  - Paths to existing files (not directories)
  - Paths with special characters (@, #, %, etc.)
  - Relative paths with invalid traversal (../../../etc/passwd)
  - Network paths when not supported
- **Expected**:
  - Service handles nil gracefully without crashes
  - Unicode and binary data are properly encoded
  - Invalid paths return structured error information
  - Permission issues are detected and reported
  - File vs directory conflicts are identified
  - Path validation prevents security issues
  - All errors include recovery suggestions

**Test Case DT-API-002**: Repository operations with invalid states
- **Objective**: Test operations on corrupted or invalid repository states
- **Destructive Inputs**:
  - Operations on uninitialized paths
  - Operations during partial git operations
  - Operations with locked git files
  - Concurrent initialization attempts
  - Invalid version hashes passed to GetFileDifferences
- **Expected**:
  - Corrupted repositories are detected safely
  - Operations fail gracefully with structured errors
  - Repository state is preserved or safely recovered
  - Lock conflicts are handled appropriately
  - Concurrent operations maintain consistency

**Test Case DT-API-003**: CommitChanges with excessive or invalid data
- **Objective**: Test commit operations under extreme conditions
- **Destructive Inputs**:
  - Empty commit messages
  - Commit messages >100KB
  - Commit messages with binary data
  - Invalid author information (nil, empty, invalid email)
  - Author names with unicode/binary characters
  - Email addresses with special characters
  - Stages and commits with 10,000+ files
  - Stages and commits with files >100MB each
  - Stages and commits during repository conflicts
  - Operations on file paths with: Unicode paths, extremely long paths (500+ chars), special characters
  - Invalid email formats, extremely long author names, special characters
  - File differences on binary files, large binary commits  
- **Expected**:
  - Large commits are handled or rejected gracefully
  - Invalid author data is validated and rejected
  - Unicode handling is correct in all metadata
  - Memory usage remains bounded
  - Repository integrity is maintained
  - Stages and commits rejected if repository has conflicts

**Test Case DT-API-004**: History operations with boundary violations
- **Objective**: Test history retrieval under extreme conditions
- **Destructive Inputs**:
  - Requests for non-existent files
  - Invalid commit hash formats
  - Requests with negative limits
  - Requests with extremely large limits (MaxInt)
  - History requests on empty repositories
  - File history for binary files >100MB
  - History requests during active operations
  - Concurrent history requests
- **Expected**:
  - Invalid requests return structured errors
  - Large requests are handled or limited safely
  - Memory usage is controlled for large histories
  - Binary file handling is appropriate
  - Concurrent requests maintain consistency

**Test Case DT-API-005**: Repository and path validation with boundary violations
- **Objective**: Test repository validation under extreme and malicious conditions
- **Destructive Inputs**:
  - **Invalid repository paths**: nil/empty paths, binary data paths, paths with control characters, non-existent directories, paths exceeding filesystem limits, read-only directories, paths pointing to files instead of directories
  - **Corrupted repositories**: Missing .git directories, corrupted .git/config files, invalid repository references, truncated git objects, repositories with active lock files
  - **Invalid path lists**: nil path arrays, paths with binary data, extremely long path lists (10,000+ entries), paths with invalid characters, paths with directory traversal attempts (../../../etc), absolute paths mixed with relative paths, paths pointing to non-existent parent directories
  - **Filesystem boundary conditions**: Symbolic link cycles, paths to mounted filesystems, paths with special filesystem features (case sensitivity issues), paths with unicode normalization conflicts
- **Expected**:
  - Invalid repository paths return structured errors without crashes
  - Corrupted repositories are detected and reported safely
  - Path validation handles malformed input gracefully
  - Directory traversal attempts are blocked
  - Memory usage remains bounded for large path lists
  - Unicode and special character handling is correct
  - Repository state corruption is prevented

### 3.2 Resource Exhaustion and Performance Testing

**Test Case DT-RESOURCE-001**: Memory Exhaustion
- **Objective**: Test behavior under memory pressure
- **Method**:
  - Repository operations with 100,000+ commits
  - File differences on files >100MB
  - History operations returning 50,000+ commits
  - Monitor memory usage and garbage collection
  - Verify graceful degradation
- **Expected**:
  - GC pressure doesn't cause excessive delays
  - No memory leaks detected
  - Large operations complete or fail gracefully

**Test Case DT-RESOURCE-002**: File Handle Exhaustion
- **Objective**: Test repository operations under resource constraints
- **Method**:
  - Open multiple repositories simultaneously
  - Exhaust available file handles
  - Test recovery when handles become available
- **Expected**: Proper file handle management, structured error responses

**Test Case DT-RESOURCE-003**: Disk Exhaustion
- **Objective**: Test repository operations under disk pressure
- **Method**:
  - Fill disk
- **Expected**: Structured error responses

**Test Case DT-RESOURCE-004**: Repository Validation Performance Exhaustion
- **Objective**: Test repository validation under extreme load
- **Method**:
  - Validate repositories with 100,000+ files in path lists
  - Concurrent validation requests on same repository
  - Validation on repositories with extremely deep directory structures (100+ levels)
  - Validation with path lists containing symbolic link cycles
  - Multiple validation requests with overlapping file/directory lists
- **Expected**:
  - Path list processing remains bounded in time and memory
  - Concurrent validations don't corrupt each other
  - Deep directory traversal is handled efficiently
  - Symbolic link cycles are detected and handled safely
  - Performance remains within 5-second requirement

**Test Case DT-PERFORMANCE-001**: Performance Degradation Under Load
- **Objective**: Validate 5-second performance requirement under stress
- **Method**:
  - Concurrent operations from multiple goroutines
  - Monitor: CPU usage, memory usage, I/O wait times
  - Measure: Average latency and 99th percentile response times
- **Expected**:
  - All operations complete within 5 seconds
  - System remains responsive under sustained load
  - No performance degradation over time
  - Memory usage stabilizes

**Test Case DT-PERFORMANCE-002**: Repository Size Limits
- **Objective**: Test behavior with extremely large repositories
- **Method**:
  - Repository with 10,000 commits
  - epository with 10,000 files
  - Individual files >100MB
  - Total repository size >10GB
  - Measure operation times and resource usage
- **Expected**:
  - Operations complete within reasonable time
  - Resource usage scales appropriately
  - Error conditions are handled gracefully

## 4. Error Condition Testing

### 4.1 External Dependency Failures

**Test Case DT-ERROR-001**: File System Failures
- **Objective**: Test resilience to file system issues
- **Failure Scenarios**:
  - Repository directory deleted during operation
  - File permissions removed during operation
  - Disk full conditions during commits
  - I/O errors during git operations
- **Expected**: Structured error responses, repository integrity maintained

**Test Case DT-ERROR-002**: Repository Corruption
- **Objective**: Test handling of corrupted repository structures
- **Corruption Scenarios**:
  - Corrupted .git/config files
  - Missing or corrupted git objects
  - Truncated commit objects
  - Invalid repository references
  - Corrupted index files
  - Invalid HEAD
- **Expected**: Corruption detection, structured error reporting, no crashes

**Test Case DT-ERROR-003**: Repository Validation Error Conditions
- **Objective**: Test repository validation error handling and recovery
- **Error Scenarios**:
  - **Repository access errors**: Permission denied during validation, repository directory deleted during validation, filesystem unmounted during validation
  - **Path validation errors**: Files/directories created or deleted during validation, permission changes during path checking, symbolic links modified during traversal
  - **Concurrent modification errors**: Repository structure modified during validation, .git directory corruption during validation, file locks acquired during path checking
  - **Resource constraint errors**: Out of memory during large path list processing, file handle exhaustion during deep directory traversal, disk I/O errors during validation
- **Expected**:
  - All error conditions return structured error information
  - Partial validation results are handled appropriately
  - No repository state corruption occurs
  - Validation state is cleaned up after errors
  - Recovery is possible without service restart

### 4.2 Concurrent Access Violations

**Test Case DT-CONCURRENT-001**: Race Condition Testing
- **Objective**: Verify thread safety under stress
- **Method**:
  - Concurrent repository operations
  - Simultaneous commits and history operations
  - Parallel initialization attempts
  - Concurrent repository validation requests
  - Simultaneous validation and commit operations
  - Parallel validation requests on same repository with different path lists
- **Expected**: No race conditions detected by Go race detector

**Test Case DT-CONCURRENT-002**: Repository Lock Conflicts
- **Objective**: Test behavior with git lock conflicts
- **Method**:
  - Multiple processes accessing same repository
  - Operations during git maintenance
  - Lock file cleanup testing
  - Repository validation during active commits
  - Path validation while repository locks are held
- **Expected**: Lock conflicts handled gracefully, operations retry or fail safely

## 5. Recovery and Degradation Testing

### 5.1 Graceful Degradation

**Test Case DT-RECOVERY-001**: Service Recovery After Failures
- **Objective**: Test recovery capabilities after various failures
- **Recovery Scenarios**:
  - File system recovery after disk full
  - Permission restoration
  - Repository cleanup after corruption
  - Lock file cleanup after process termination
- **Expected**: Automatic recovery without restart required

**Test Case DT-RECOVERY-002**: Partial Functionality Under Constraints
- **Objective**: Test continued operation under resource constraints
- **Constraint Scenarios**:
  - Limited memory availability
  - Restricted file system access
  - High concurrent load
  - Repository size limitations
- **Expected**: Core functionality maintained, non-essential features gracefully degraded

## 6. Test Execution Requirements

### 6.1 Required Tools and Environment
- Go race detector (`go test -race`)
- Memory profiling tools (`go test -memprofile`)
- CPU Profiling: Enabled (`go test -cpuprofile`)
- Git command-line tools for repository setup
- File system permission control
- Resource monitoring utilities (disk space and file handles)
- Concurrent load generation tools

### 6.2 Success Criteria
- **100% Requirements Coverage**: Every EARS requirement has corresponding destructive tests
- **Zero Critical Failures**: No crashes, memory leaks, or repository corruption
- **Race Detector Clean**: No race conditions detected under any scenario
- **Graceful Error Handling**: All error conditions handled without caller failures
- **Performance Under Stress**: 5-second performance requirement maintained under adverse conditions
- **Complete Recovery**: Service recovers from all testable failure conditions

---

**Document Version**: 1.2
**Created**: 2025-09-07
**Updated**: 2025-09-20
**Changes**: Extended with repository validation destructive testing
**Status**: Accepted