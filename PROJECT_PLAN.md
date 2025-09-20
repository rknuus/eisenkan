# EisenKan project plan
## Tasks
### UI Components (Use manager + engines directly)
- Implement BoardSelectionView - Depends on
TaskManager + FormattingEngine
- Implement SubtaskExpansionView - Depends on
TaskWidget + LayoutEngine + BoardView (for column patterns)
- Implement TaskFormView - Depends on
CreateTaskDialog + ColumnWidget + TaskWidget + WorkflowManager +
FormattingEngine

### Application
- Implement Application Root - Depends on BoardSelectionView + BoardView
- Implement service EisenKan Admin Client

## Extensions for Board Management

Following Service Lifecycle Process analysis for BoardSelectionView, several components require extension to support board discovery, validation, and management operations.

### Tasks

5. **BoardSelectionView Implementation** - UI Component
   - Implement BoardSelectionView using TaskManager board operations
   - Follow Service Lifecycle Process for UI component
   - *Dependencies*: TaskManager board operations fully tested and operational
   - *Service Lifecycle*: Context → SRS → STP → Design → Construction → Testing → STR

### Plan elaboration

**TaskManager Extension** ✅ *Completed*
- Added 5 new operations (OP-9 to OP-13) for board discovery and lifecycle management
- Added 10 new requirements (REQ-TASKMANAGER-023 to 032) for board validation, metadata extraction, and CRUD operations
- Extended to validate board directories, extract board metadata, and manage board lifecycle
- Version 1.1 SRS documented and accepted

**VersioningUtility Extension** ✅ *Completed*
- Added ValidateRepositoryAndPaths operation for repository validation and file/directory existence checking
- Added RepositoryValidationRequest and RepositoryValidationResult data structures
- Extended Repository interface with validation capabilities for board discovery operations
- Version 1.1 SRS documented and accepted

**RuleEngine Extension** ✅ *Completed*
- Added EvaluateBoardConfigurationChange operation for board configuration validation
- Added BoardConfigurationEvent and BoardConfiguration data structures for board validation context
- Implemented board title validation (non-empty, ≤100 chars, alphanumeric+spaces+hyphens only)
- Implemented board description validation (≤500 chars when provided)
- Implemented board configuration format validation (required fields, valid structure)
- Added "board_configuration" rule category for specialized validation logic
- Version 1.1 SRS and STP documented and accepted

**BoardAccess Extension** - Board Data Operations ✅ *Completed*
- Create IBoardDiscovery facet for directory validation and git repository detection
- Create IBoardMetadata facet for board metadata extraction and management
- Create IBoardLifecycle facet for board CRUD operations

**TaskManager Integration Testing** - Board Operations ✅ *Completed*
- Integration test TaskManager board operations with extended BoardAccess
- Verify all 5 new operations (OP-9 to OP-13) work correctly

**Required Component Extensions**:

1. **BoardAccess (Resource Access) - HIGH PRIORITY**
   - **Board Discovery Facet**: New interface `IBoardDiscovery` for directory validation, git repository detection, and board configuration file verification
   - **Board Metadata Facet**: New interface `IBoardMetadata` for extracting and managing board metadata (title, description, type, dates)
   - **Board Lifecycle Facet**: New interface `IBoardLifecycle` for board creation, updates, and deletion operations
   - *Rationale*: TaskManager coordinates board operations but delegates actual data access to BoardAccess

**Implementation Order**: VersioningUtility ✅ → RuleEngine ✅ → BoardAccess → BoardSelectionView implementation