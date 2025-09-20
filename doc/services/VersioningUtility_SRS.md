# VersioningUtility Software Requirements Specification (SRS)

## 1. Introduction

### 1.1 Purpose
This Software Requirements Specification defines the requirements for the VersioningUtility service, a Utilities layer component that provides simple version control capabilities for data repositories within the EisenKan system. The service enables BoardAccess and RulesAccess components to manage data versions without concerning themselves with the specifics of version control implementation.

### 1.2 Scope
VersioningUtility is responsible for:
- Repository initialization and management operations
- Generic file and directory existence validation within repositories
- Staging and committing file changes with version control tracking
- Repository status monitoring and conflict detection
- Commit history retrieval for repositories and specific files
- File difference analysis between versions
- Thread-safe concurrent repository operations

### 1.3 System Context
VersioningUtility operates in the Utilities layer of the EisenKan architecture, providing version control services to ResourceAccess components (BoardAccess and RulesAccess). It provides a stable API for version control operations while encapsulating the volatility of version control implementation details and storage mechanisms. The utility maintains application-agnostic interfaces, allowing higher-layer components to specify application-specific file and directory requirements.

## 2. Operations

The following operations define the required behavior for VersioningUtility:

#### OP-1: Initialize Repository
**Actors**: BoardAccess, RulesAccess components
**Trigger**: When a component needs version control for a data directory
**Flow**:
1. Receive repository initialization request with directory path
2. Create new repository or open existing repository at specified location
3. Configure repository with appropriate settings
4. Return initialization confirmation or error details

#### OP-2: Commit Changes
**Actors**: BoardAccess, RulesAccess components
**Trigger**: When component has made data changes that need to be versioned
**Flow**:
1. Receive commit request with message and author information
2. Stage all modified and new files in the repository
3. Create commit with provided metadata (message, author, timestamp)
4. Return commit confirmation with commit hash or error details

#### OP-3: Retrieve History
**Actors**: BoardAccess, RulesAccess components
**Trigger**: When component needs access to version history for analysis
**Flow**:
1. Receive history request (repository-wide or file-specific)
2. Query version control system for chronological commit history
3. Return commit metadata (hash, author, timestamp, message)
4. Include file differences if requested between versions

#### OP-4: Validate Repository and Paths
**Actors**: BoardAccess components
**Trigger**: When component needs to verify a directory contains a valid git repository and optionally check for specific files/directories
**Flow**:
1. Receive validation request with directory path and optional list of required files/directories
2. Check if directory exists and is accessible
3. Verify presence of valid .git directory structure
4. If file/directory list provided, check existence of each specified path within repository
5. Return validation result with repository status and path existence details

## 3. Functional Requirements

### 3.1 Repository Management Requirements

**REQ-VERSION-001**: When repository initialization is requested for a path, the VersioningUtility shall initialize or open a repository at the specified location.

**REQ-VERSION-002**: When repository status is requested, the VersioningUtility shall return information about modified files, staged changes, and current branch state.

**REQ-VERSION-003**: When file changes need to be staged, the VersioningUtility shall prepare all modifications for version control operations.

**REQ-VERSION-004**: When a commit is requested with staged changes, the VersioningUtility shall create a commit with the provided message and author information.

**REQ-VERSION-005**: When repository history is requested, the VersioningUtility shall return a chronological list of commits with metadata including hash, author, timestamp, and message.

**REQ-VERSION-006**: When file history is requested for a specific file, the VersioningUtility shall return a chronological list of commits with metadata including hash, author, timestamp, and message specific to that file.

**REQ-VERSION-007**: When file differences are requested between two versions, the VersioningUtility shall return the differences between the specified versions.

**REQ-VERSION-008**: When repository and path validation is requested for a directory path, the VersioningUtility shall verify the directory exists and contains a valid git repository structure.

**REQ-VERSION-009**: When repository and path validation is requested with a list of required files and directories, the VersioningUtility shall check existence of each specified path and return validation results.

### 3.2 Repository Discovery Requirements

**REQ-VERSION-010**: When repository validation fails due to missing .git directory, the VersioningUtility shall return structured error information indicating the directory is not a git repository.

**REQ-VERSION-011**: When repository validation fails due to directory access issues, the VersioningUtility shall return structured error information with filesystem error details.

### 3.3 Version History Requirements

**REQ-VERSION-005**: When repository history is requested, the VersioningUtility shall return a chronological list of commits with metadata including hash, author, timestamp, and message.

**REQ-VERSION-006**: When file history is requested for a specific file, the VersioningUtility shall return a chronological list of commits with metadata including hash, author, timestamp, and message specific to that file.

**REQ-VERSION-007**: When file differences are requested between two versions, the VersioningUtility shall return the differences between the specified versions.

## 4. Quality Attributes

**REQ-PERFORMANCE-001**: While the system is operational, repository operations shall complete within 5 seconds for repositories with up to 10,000 commits on a MacAir M4 under regular load.

**REQ-RELIABILITY-001**: If repository operations fail due to corruption or file system issues, then the VersioningUtility shall return structured error information without crashing the application.

**REQ-RELIABILITY-002**: If a merge conflict exists in the repository, then the VersioningUtility shall reject any modifying operation (staging and committing changes).

**REQ-USABILITY-001**: While the system is operational, the VersioningUtility shall accept both absolute and relative repository paths.

### 4.1 Performance Requirements

**REQ-PERFORMANCE-001**: While the system is operational, repository operations shall complete within 5 seconds for repositories with up to 10,000 commits on a MacAir M4 under regular load.

### 4.2 Reliability Requirements

**REQ-RELIABILITY-001**: If repository operations fail due to corruption or file system issues, then the VersioningUtility shall return structured error information without crashing the application.

**REQ-RELIABILITY-002**: If a merge conflict exists in the repository, then the VersioningUtility shall reject any modifying operation (staging and committing changes).

### 4.3 Usability Requirements

**REQ-USABILITY-001**: While the system is operational, the VersioningUtility shall accept both absolute and relative repository paths.

## 5. Service Contract Requirements

### 5.1 Interface Operations
The VersioningUtility shall provide the following operations:

#### Core Repository Operations
1. **InitializeRepository**: Initialize or open a repository at specified path
2. **GetRepositoryStatus**: Retrieve the current status of a repository
3. **StageChanges**: Stage all new and changed files for commit
4. **CommitChanges**: Commit all staged changes with message and author
5. **GetRepositoryHistory**: Retrieve the commit history of the repository
6. **GetFileHistory**: Retrieve the version history of a specific file
7. **GetFileDifferences**: Retrieve differences between two versions of a file

#### Repository Discovery and Validation Operations
8. **ValidateRepositoryAndPaths**: Verify directory contains valid git repository structure and optionally check existence of specified files and directories

Note: Raw file access is handled by standard file operations and is not provided by this interface. The interface intentionally provides a focused set of operations, with advanced operations like conflict resolution handled by external tools. Repository discovery operations maintain application-agnostic interfaces by accepting caller-specified file and directory requirements.

### 5.2 Data Contracts

#### Request Structures

**RepositoryValidationRequest Structure**:
- DirectoryPath: Absolute or relative path to directory to validate
- RequiredFiles: Optional list of file paths that must exist (empty list for repository-only validation)
- RequiredDirectories: Optional list of directory paths that must exist (empty list for repository-only validation)

#### Response Structures

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

**RepositoryValidationResult Structure**:
- RepositoryValid: Boolean indicating repository validation success
- ExistingPaths: List of files/directories that exist (empty if no paths specified)
- MissingPaths: List of files/directories that are missing (empty if no paths specified)
- ErrorMessage: Human-readable error description for failures
- TechnicalDetails: Detailed error information for debugging

### 5.3 Error Handling
All errors shall include:
- Error code classification
- Human-readable error message
- Technical details for debugging
- Suggested recovery actions where applicable
- Repository conflict status when applicable

## 6. Technical Constraints

### 6.1 Integration Requirements
**REQ-INTEGRATION-001**: The VersioningUtility shall be callable from ResourceAccess layer components.

**REQ-INTEGRATION-002**: The VersioningUtility shall use the LoggingUtility service for all operational logging.

**REQ-INTEGRATION-003**: The VersioningUtility shall not create dependencies on business logic components.

**REQ-INTEGRATION-004**: The VersioningUtility shall maintain application-agnostic interfaces for repository discovery operations, accepting caller-specified file and directory requirements.

### 6.2 Data Format Requirements
**REQ-FORMAT-001**: The VersioningUtility shall support local Git repositories for version control operations.

**REQ-FORMAT-002**: The VersioningUtility shall maintain commit metadata including author information, timestamps, and commit messages.

**REQ-FORMAT-003**: The VersioningUtility shall provide structured error reporting for all failure scenarios.

## 7. Acceptance Criteria

### 7.1 Functional Acceptance
- All interface operations work as specified in contract
- Repository initialization handles both new and existing repositories
- Version history and file differences are accurately retrieved
- Commit operations preserve author information and timestamps

### 7.2 Quality Acceptance
- Repository operations complete within specified performance limits
- Error conditions return structured information without system crashes
- Concurrent operations maintain data integrity
- File path flexibility supports both absolute and relative paths

### 7.3 Integration Acceptance
- BoardAccess and RulesAccess can successfully use versioning capabilities
- Repository discovery operations work with application-specific file/directory requirements
- Operations integrate smoothly with existing logging infrastructure
- Service follows iDesign utility service patterns
- Service maintains stateless operation regarding business data (repository state only)
- Service provides essential version control operations without complexity
- Service maintains data integrity under concurrent access
- Repository validation and path operations maintain application-agnostic interfaces

---

**Document Version**: 1.1
**Released**: 2025-09-07
**Updated**: 2025-09-20
**Changes**: Extended by repository structure validation
**Status**: Accepted