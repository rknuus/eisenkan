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

1. **VersioningUtility Extension** - Repository Discovery & Validation
   - Extend VersioningUtility with repository discovery methods
   - Add git repository structure validation capabilities
   - Add board-specific repository initialization methods
   - *Dependencies*: None (leaf utility component)
   - *Service Lifecycle*: Context → SRS → STP → Design → Construction → Testing → STR

2. **RuleEngine Extension** - Board Validation Rules
   - Extend RuleEngine with board configuration validation rules
   - Add board operation constraint validation
   - *Dependencies*: None (independent engine component)
   - *Service Lifecycle*: Context → SRS → STP → Design → Construction → Testing → STR

3. **BoardAccess Extension** - Board Data Operations
   - Create IBoardDiscovery facet for directory validation and git repository detection
   - Create IBoardMetadata facet for board metadata extraction and management
   - Create IBoardLifecycle facet for board CRUD operations
   - *Dependencies*: VersioningUtility extension (for repository operations), RuleEngine extension (for validation)
   - *Service Lifecycle*: Context → SRS → STP → Design → Construction → Testing → STR

4. **TaskManager Integration Testing** - Board Operations
   - Integration test TaskManager board operations with extended BoardAccess
   - Verify all 5 new operations (OP-9 to OP-13) work correctly
   - *Dependencies*: BoardAccess extension completed
   - *Testing Phase*: Integration testing only

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

**Required Component Extensions**:

1. **BoardAccess (Resource Access) - HIGH PRIORITY**
   - **Board Discovery Facet**: New interface `IBoardDiscovery` for directory validation, git repository detection, and board configuration file verification
   - **Board Metadata Facet**: New interface `IBoardMetadata` for extracting and managing board metadata (title, description, type, dates)
   - **Board Lifecycle Facet**: New interface `IBoardLifecycle` for board creation, updates, and deletion operations
   - *Rationale*: TaskManager coordinates board operations but delegates actual data access to BoardAccess

2. **VersioningUtility (Utilities) - MEDIUM PRIORITY**
   - **Repository Discovery**: Methods to detect valid git repositories in directories
   - **Repository Validation**: Methods to check repository structure and required files
   - **Repository Initialization**: Enhanced methods for creating board-specific repository structures
   - *Rationale*: Board validation requires git repository detection and validation capabilities

3. **RuleEngine (Engines) - LOW PRIORITY**
   - **Board Validation Rules**: Business rules for valid board configurations
   - **Board Operation Rules**: Validation for board creation/modification constraints
   - *Rationale*: Board operations should be validated against business rules for architectural consistency

**Implementation Order**: BoardAccess → VersioningUtility → RuleEngine → BoardSelectionView implementation