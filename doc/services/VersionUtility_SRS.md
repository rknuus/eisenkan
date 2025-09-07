# VersionUtility Software Requirements Specification (SRS)

## 1. Service Overview

### 1.1 Purpose
The VersionUtility shall provide simple version control capabilities for data repositories within the EisenKan system, enabling TasksAccess and RulesAccess components to manage data versions without concerning themselves with the specifics of version control implementation.

### 1.2 Architectural Classification
- **Layer**: Utilities
- **Type**: Utility Service
- **Volatility**: Infrastructure (Low volatility - version control patterns are stable)
- **Namespace**: `eisenkan.Utilities.VersionUtility`

## 2. Business Requirements

### 2.1 Core Version Control Requirements

**REQ-VERSION-001**: When repository initialization is requested for a path, the VersionUtility shall initialize or open a repository at the specified location.

**REQ-VERSION-002**: When repository status is requested, the VersionUtility shall return information about modified files, staged changes, and current branch state.

**REQ-VERSION-003**: When file changes need to be staged, the VersionUtility shall prepare all modifications for version control operations.

**REQ-VERSION-004**: When a commit is requested with staged changes, the VersionUtility shall create a commit with the provided message and author information.

**REQ-VERSION-005**: When repository history is requested, the VersionUtility shall return a chronological list of commits with metadata including hash, author, timestamp, and message.

**REQ-VERSION-006**: When file history is requested for a specific file, the VersionUtility shall return a chronological list of commits with metadata including hash, author, timestamp, and message specific to that file.

**REQ-VERSION-007**: When file differences are requested between two versions, the VersionUtility shall return the differences between the specified versions.

### 2.2 Quality Attributes

**REQ-PERFORMANCE-001**: While the system is operational, repository operations shall complete within 5 seconds for repositories with up to 10,000 commits on a MacAir M4 under regular load.

**REQ-RELIABILITY-001**: If repository operations fail due to corruption or file system issues, then the VersionUtility shall return structured error information without crashing the application.

**REQ-RELIABILITY-002**: If a merge conflict exists in the repository, then the VersionUtility shall reject any modifying operation (staging and committing changes).

**REQ-USABILITY-001**: While the system is operational, the VersionUtility shall accept both absolute and relative repository paths.

## 3. Service Contract Requirements

### 3.1 Interface Operations
The VersionUtility shall provide the following operations:

1. **InitializeRepository**: Initialize or open a repository at specified path
2. **GetRepositoryStatus**: Retrieve the current status of a repository
3. **StageChanges**: Stage all new and changed files for commit
4. **CommitChanges**: Commit all staged changes with message and author
5. **GetRepositoryHistory**: Retrieve the commit history of the repository
6. **GetFileHistory**: Retrieve the version history of a specific file
7. **GetFileDifferences**: Retrieve differences between two versions of a file

Note: Raw file access is handled by standard file operations and is not provided by this interface. The interface intentionally provides a focused set of operations, with advanced operations like conflict resolution handled by external tools.

### 3.2 Data Contracts

**CommitInfo Structure**:
- ID: Unique commit identifier (e.g. hash)
- Author: Commit author name
- Email: Author email address
- Timestamp: Commit creation time
- Message: Commit description

**RepositoryStatus Structure**:
- CurrentBranch: Active branch name
- ModifiedFiles: List of changed files
- StagedFiles: List of files ready for commit
- UntrackedFiles: List of unversioned files
- HasConflicts: Indicates presence of merge conflicts

## 4. Technical Constraints

### 4.1 Technology Requirements
- Implementation Language: Go (matching project requirements)
- Dependencies:
  - `LoggingUtility` for logging
  - `go-git` for repository operations
  - Go standard library

### 4.2 Integration Requirements
- The VersionUtility shall be callable from all architectural layers
- The VersionUtility shall not create dependencies on business logic components
- The VersionUtility shall support graceful resource cleanup

### 4.3 Operational Requirements
- Local repository support is required
- Thread-safe operation under concurrent access
- Structured error reporting for failure scenarios

## 5. Acceptance Criteria

### 5.1 Functional Acceptance
- All interface operations work as specified in contract
- Repository initialization handles both new and existing repositories
- Version history and file differences are accurately retrieved
- Commit operations preserve author information and timestamps

### 5.2 Quality Acceptance
- Repository operations complete within specified performance limits
- Error conditions return structured information without system crashes
- Concurrent operations maintain data integrity
- File path flexibility supports both absolute and relative paths

### 5.3 Integration Acceptance
- TasksAccess and RulesAccess can successfully use versioning capabilities
- Operations integrate smoothly with existing logging infrastructure

## 6. Design Constraints

### 6.1 Architectural Constraints
- Must follow iDesign utility service patterns
- Must not contain business logic or domain-specific functionality
- Must be stateless regarding business data (repository state only)
- Must support interface-based design for testability

### 6.2 Version Control Constraints
- Must handle local repositories
- Must provide essential version control operations without complexity
- Must support standard version control workflows
- Must maintain data integrity under concurrent access

---

**Document Version**: 1.0
**Released**: 2025-09-07
**Status**: Accepted