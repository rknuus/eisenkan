# Design Decision Records (DDR)

## [2025-09-20] - BoardAccess: IBoard Facet Implementation Design

**Decision**: Implement IBoard as single comprehensive interface with fail-fast error handling and direct service integration

**Context**: BoardAccess Extension requires IBoard facet implementation to support board discovery, metadata management, lifecycle operations, and configuration management. Design must balance simplicity with functionality while integrating with VersioningUtility and RuleEngine services.

**Options Considered**:

### Design Decision 1: IBoard Facet Implementation Strategy
**Options**:
A. **Single Comprehensive Interface**: Implement IBoard as one large interface with all operations
B. **Composed Sub-Interfaces**: Split IBoard into smaller interfaces and compose them
C. **Hybrid Approach**: Keep IBoard as main interface but implement using internal sub-components

**Chosen**: Option A - Single Comprehensive Interface for simple dependency injection and unified error handling

### Design Decision 4: Error Handling Strategy
**Options**:
A. **Fail-Fast Approach**: Return errors immediately when any operation fails
B. **Graceful Degradation**: Continue with reduced functionality when some operations fail
C. **Retry with Circuit Breaker**: Retry failed operations with exponential backoff

**Chosen**: Option A - Fail-Fast Approach with enhancement for directory structure checks to report all issues together rather than stopping at first failure

### Design Decision 5: Integration with Existing Services
**Options**:
A. **Direct Integration**: BoardAccess directly calls VersioningUtility and RuleEngine
B. **Adapter Pattern**: Create adapters for external service integration
C. **Event-Based Integration**: Use events for communication with external services

**Chosen**: Option A - Direct Integration (previously decided in architecture)

**Rationale**: Single comprehensive interface provides simplicity for initial implementation. Fail-fast error handling ensures clear error semantics while directory validation enhancement provides better user experience by reporting multiple issues at once. Direct integration maintains consistency with existing service patterns.

**User Approval**: Approved 2025-09-20

---

## [2025-09-20] - RuleEngine: Board Configuration Validation Extension Design

**Decision**: Extend RuleEngine with new `EvaluateBoardConfigurationChange` operation using dedicated event type and reusing existing rule evaluation infrastructure for board validation rules

**Context**: RuleEngine requires extension to support board configuration validation operations for TaskManager board operations. The extension must maintain consistency with existing rule evaluation patterns while adding specialized validation for board metadata (title, description, format) without knowledge of directory structure.

**Options Considered**:

### Design Decision 1: Operation Integration Approach
**Options**:
A. **Unified Operation**: Extend existing `EvaluateTaskChange` with board configuration event types
B. **Separate Operation**: Create dedicated `EvaluateBoardConfigurationChange` operation
C. **Rule Category Extension**: Add board validation as new rule category within existing operation

**Chosen**: Option B - Separate Operation (`EvaluateBoardConfigurationChange`) to maintain clear operation boundaries, provide type-safe event handling, and support distinct validation contexts while reusing rule evaluation infrastructure

### Design Decision 2: Event Structure Design
**Options**:
A. **Reuse TaskEvent**: Extend TaskEvent with board configuration fields
B. **New Event Type**: Create `BoardConfigurationEvent` with board-specific fields
C. **Generic Event**: Use generic `ValidationEvent` for both task and board validation

**Chosen**: Option B - New Event Type to provide type safety, clear separation of concerns, and dedicated validation context while avoiding pollution of task-focused data structures

### Design Decision 3: Rule Evaluation Reuse Strategy
**Options**:
A. **Complete Duplication**: Separate board rule evaluation pipeline
B. **Infrastructure Reuse**: Reuse `evaluateRules()`, `filterApplicableRules()` with board rules
C. **Hybrid Approach**: Board-specific evaluation with shared violation aggregation

**Chosen**: Option B - Infrastructure Reuse to leverage existing rule filtering, evaluation sequencing, violation aggregation, and priority sorting while maintaining consistency

### Design Decision 4: Board Rule Category Integration
**Options**:
A. **New Rule Category**: Add "board_configuration" rule category
B. **Existing Categories**: Use "validation" category with board-specific conditions
C. **Separate Rule System**: Independent board rule definitions

**Chosen**: Option A - New Rule Category to provide clear separation, specialized board validation logic, and distinct rule identification while integrating with existing rule infrastructure

### Design Decision 5: Context Enrichment Approach
**Options**:
A. **No Context Enrichment**: Board validation uses event data only
B. **Minimal Context**: Board metadata only from BoardAccess
C. **Full Context Reuse**: Reuse EnrichedContext with board metadata

**Chosen**: Option A - No Context Enrichment since board configuration validation operates on submitted data without requiring existing board state, avoiding unnecessary BoardAccess integration

**Rationale**:
- Maintains clear operation boundaries and type safety
- Reuses proven rule evaluation infrastructure for consistency
- Provides specialized board validation without task context pollution
- Integrates seamlessly with existing rule category system
- Avoids BoardAccess dependency for configuration validation
- Supports future board validation rule expansion

**User Approval**: Approved (with consistent naming: `EvaluateBoardConfigurationChange`)

## [2025-09-20] - VersioningUtility: Repository Validation Extension Design

**Decision**: Extend VersioningUtility with single combined repository validation operation using optional path lists and unified validation result structure

**Context**: VersioningUtility requires extension to support board discovery operations for TaskManager, specifically repository validation and file/directory existence checking. The extension must maintain application-agnostic design while providing infrastructure for BoardAccess to validate board repository structures.

**Options Considered**:

### Design Decision 1: Validation Operation Granularity
**Options**:
A. **Separate Operations**: `ValidateRepository()` and `ValidateRequiredPaths()` as independent operations
B. **Combined Operation**: Single `ValidateRepositoryAndPaths()` with optional path parameters
C. **Hierarchical Operations**: Repository validation required before path validation

**Chosen**: Option B - Combined Operation to reduce API surface area, eliminate redundant repository access, and simplify caller code while supporting both repository-only and repository-with-paths validation scenarios

### Design Decision 2: Request Structure Design
**Options**:
A. **Separate Request Types**: `RepositoryValidationRequest` and `PathValidationRequest` structures
B. **Unified Request Type**: Single request with optional file/directory lists
C. **Polymorphic Request**: Interface-based request with multiple implementations

**Chosen**: Option B - Unified Request Type with optional path lists to match the combined operation design, reducing complexity and API surface while maintaining type safety

### Design Decision 3: Application-Specific Knowledge Separation
**Options**:
A. **Board-Aware Validation**: VersioningUtility knows board file requirements
B. **Generic Path Validation**: Caller specifies required files/directories
C. **Configuration-Driven**: External configuration defines validation rules

**Chosen**: Option B - Generic Path Validation to maintain utility reusability and application-agnostic design, allowing BoardAccess to specify board-specific requirements while keeping VersioningUtility generic

### Design Decision 4: Directory Creation Responsibility
**Options**:
A. **VersioningUtility Creates**: Extend utility with `EnsurePathsExist()` operation
B. **Caller Creates**: ResourceAccess components handle directory creation
C. **Hybrid Approach**: Optional creation parameter in validation requests

**Chosen**: Option B - Caller Creates to maintain clear separation of concerns, avoid filesystem modification in validation operations, and leverage existing staging/commit capabilities for repository initialization

**Rationale**:
- Maintains utility layer principles and reusability
- Simplifies API surface area with minimal operations
- Preserves application-agnostic design for maximum reuse
- Leverages existing repository initialization patterns
- Supports both repository-only and file/directory validation use cases

**User Approval**: Approved

## [2025-09-19] - BoardView: Component Architecture and Integration Pattern

**Decision**: Implement BoardView using Widget Pattern with 4-Column Eisenhower Matrix Layout, Direct Engine Integration, Immutable State Management, and Task Ownership Transfer Pattern

**Context**: BoardView requires implementation as the primary kanban board interface with 4-column Eisenhower Matrix layout for comprehensive task management and workflow coordination. The component must coordinate multiple ColumnWidget instances, integrate TaskWidget components for task display, and orchestrate task workflow management across priority columns through drag-drop operations and business rule validation.

**Options Considered**:

### Design Decision 1: Widget Architecture Pattern
**Options**:
A. **Custom Widget with Renderer Pattern**: Extend `widget.BaseWidget` with custom `fyne.WidgetRenderer`
B. **Container Composition Pattern**: Use `container.NewBorder` with 4 column containers
C. **Custom Layout Manager**: Implement EisenhowerMatrixLayout with ColumnWidget embedding

**Chosen**: Option A - Custom Widget with Renderer Pattern for complete control over board layout and rendering, follows Fyne best practices established by ColumnWidget and TaskWidget patterns

### Design Decision 2: Column Organization Strategy
**Options**:
A. **Fixed 4-Column Grid**: Use `container.NewGridWithColumns(2)` in 2x2 arrangement for Eisenhower Matrix
B. **Dynamic Column Management**: Use `container.NewHBox` with variable column count
C. **Responsive Column Layout**: Adaptive column arrangement based on screen size

**Chosen**: Option B - Dynamic Column Management because the columns will be configurable in the future, providing flexibility for different board configurations beyond the initial Eisenhower Matrix

### Design Decision 3: Dependency Integration Pattern
**Options**:
A. **Direct Engine Dependencies**: Accept WorkflowManager, FormValidationEngine as constructor parameters
B. **Service Locator Pattern**: Runtime dependency discovery and binding
C. **Dependency Injection Container**: Central container managing all dependencies

**Chosen**: Option A - Direct Engine Dependencies for architectural compliance with ColumnWidget and TaskWidget patterns, straightforward testing, and minimal abstraction overhead

### Design Decision 4: State Management Architecture
**Options**:
A. **Immutable State with Channels**: Single BoardState with state channels for updates
B. **Distributed State per Column**: Independent ColumnWidget state management
C. **Event-Driven State Machine**: Formal state machine for board lifecycle

**Chosen**: Option A - Immutable State with Channels for consistency with ColumnWidget patterns, centralized coordination, and proven state management approach from existing widgets

### Design Decision 5: Task Display Coordination Strategy
**Options**:
A. **Direct TaskWidget Creation**: BoardView creates TaskWidget instances and passes to ColumnWidget
B. **Delegation to ColumnWidget**: ColumnWidget manages TaskWidget creation and lifecycle
C. **Shared TaskWidget Pool**: Centralized widget pool shared across columns

**Chosen**: Option B - Delegation to ColumnWidget for clear separation of concerns, leveraging existing ColumnWidget capabilities, and avoiding widget lifecycle complexity

### Design Decision 6: Drag-Drop Workflow Integration
**Options**:
A. **Board-Level Drag Coordination**: BoardView handles all drag-drop events and coordinates with WorkflowManager
B. **Column-Level Drag Delegation**: ColumnWidget handles drag events, BoardView coordinates WorkflowManager
C. **Task-Level Drag Events**: TaskWidget handles drag events, bubbles up to BoardView

**Chosen**: Option A - Board-Level Drag Coordination to avoid complexity of mixed responsibility (originally considered A+B mix where ColumnWidget handles intra-column drags and BoardView handles inter-column drags, but this would be complicated). May remove no longer required drag code from ColumnWidget in the future.

### Design Decision 7: WorkflowManager Integration Pattern
**Options**:
A. **Direct WorkflowManager Calls**: BoardView calls WorkflowManager.Task() and .Drag() directly
B. **Workflow Coordinator Layer**: Intermediate coordinator between BoardView and WorkflowManager
C. **Event-Driven Workflow Integration**: Workflow operations triggered through event system

**Chosen**: Option A - Direct WorkflowManager Calls for architectural compliance, direct manager access pattern established in ColumnWidget, and minimal abstraction overhead

**Final Architecture**:
```
BoardView (extends widget.BaseWidget)
├── Custom Renderer (Dynamic Column Layout)
├── Dependencies (Constructor Injection)
│   ├── WorkflowManager (Task querying, movement coordination)
│   ├── FormValidationEngine (Operation validation)
│   ├── ColumnWidget×N (Configurable columns, initially 4 for Eisenhower)
│   └── TaskWidget factory pattern via ColumnWidget delegation
├── State Management (Immutable BoardState with Channels)
├── Drag-Drop Coordination (Board-level with WorkflowManager)
├── Event Handling (Task movement, board refresh, validation)
└── Column Coordination (State synchronization, WIP limits)
```

**Key Design Principles**:
- **Configurable Column Layout**: Dynamic column management for future configurability (initially Eisenhower Matrix)
- **ColumnWidget Delegation**: Leverage existing ColumnWidget capabilities for task management
- **Board-Level Drag Coordination**: Centralized drag-drop handling to avoid complexity
- **Immutable State**: Thread-safe state management with channel-based updates
- **Direct Engine Integration**: Direct WorkflowManager and FormValidationEngine access
- **Performance Optimization**: Efficient rendering within 300ms requirement (BV-REQ-041)
- **Scalability Support**: Handle up to 1000 tasks across all columns (BV-REQ-046)

**Integration Pattern**:
- BoardView coordinates configurable ColumnWidget instances (initially 4 for Eisenhower quadrants)
- WorkflowManager.Task() provides task querying and data management
- WorkflowManager.Drag() handles task movement workflow coordination
- FormValidationEngine validates task operations before workflow submission
- ColumnWidget manages TaskWidget lifecycle with DisplayMode for embedded tasks
- Board-level drag-drop coordinates all task movement (may remove ColumnWidget drag code later)

**Rationale**: This design provides flexibility for future column configurability while leveraging existing architectural patterns. Custom widget with renderer provides control over layout. Dynamic column management supports evolution beyond Eisenhower Matrix. Delegation to ColumnWidget simplifies task management while board-level drag coordination avoids mixed responsibility complexity. Immutable state management follows proven patterns while maintaining thread safety.

**User Approval**: Approved

## [2025-09-19] - CreateTaskDialog: Eisenhower Matrix Dialog Architecture

**Decision**: Implement CreateTaskDialog with custom Fyne dialog architecture, fixed grid layout, direct TaskWidget integration, deferred WorkflowManager coordination, centralized state management, and direct engine dependencies

**Context**: CreateTaskDialog requires a modal interface for task creation within Eisenhower Matrix context, supporting task creation in "non-urgent non-important" quadrant, display of existing tasks in other quadrants, and drag-drop task organization with ownership transfer to ColumnWidget.

**Options Considered**:

### Design Decision 1: Dialog Architecture Pattern
**Options**:
A. **Custom Fyne Dialog with Embedded Container**: Extend fyne.Dialog with custom content container
B. **Custom Widget implementing fyne.Widget Interface**: Standalone widget with Show() method
C. **Composite Pattern with Dialog Wrapper**: EisenhowerMatrixWidget wrapped with DialogManager

**Chosen**: Option A for framework compliance, modal behavior management, and straightforward Fyne integration

### Design Decision 2: Eisenhower Matrix Layout Strategy
**Options**:
A. **Fixed Grid Layout with 4 Containers**: Use container.NewGridWithColumns(2) for 2x2 layout
B. **Dynamic Layout with LayoutEngine Integration**: Delegate layout decisions to LayoutEngine
C. **Custom Layout Manager**: Implement custom EisenhowerMatrixLayout

**Chosen**: Option A for simplicity, predictable layout behavior, and clear quadrant separation

### Design Decision 3: Task Widget Integration Pattern
**Options**:
A. **Direct TaskWidget Instantiation**: Create TaskWidget instances directly for each task
B. **TaskWidget Factory Pattern**: Use factory for consistent TaskWidget configuration
C. **Delegation Pattern with Widget Manager**: WidgetManager handles TaskWidget operations

**Chosen**: Option A with ownership transfer constraint: when dialog closes, tasks in 3 Eisenhower quadrants transfer ownership to Todo ColumnWidget, enabling seamless integration with existing kanban workflow

### Design Decision 4: Drag-Drop Coordination Strategy
**Options**:
A. **Direct DragDropEngine Integration**: Direct connection with manual WorkflowManager coordination
B. **Drag-Drop Mediator Pattern**: DragDropMediator coordinates between engines
C. **Event-Driven Drag-Drop System**: Event-based drag-drop coordination

**Chosen**: Option A with deferred WorkflowManager pattern: DragDropEngine handles spatial mechanics immediately, WorkflowManager coordination deferred until dialog close for performance and batch operations

**Consequences of Deferred WorkflowManager Integration**:
- **Positive**: Immediate visual feedback, batch operations, simple rollback, offline capability
- **Negative**: Temporary state management required, validation at close-time, deferred error handling
- **Mitigation Strategies**: TaskMovementTracker utility, Command pattern for operations, WorkflowManager batch API, abstract DragDropCoordinator interface

### Design Decision 5: State Management Approach
**Options**:
A. **Centralized State with Immutable Updates**: Single DialogState with state channels
B. **Distributed State per Quadrant**: Independent quadrant state management
C. **State Machine Pattern**: Formal state machine for dialog lifecycle

**Chosen**: Option A for consistency with TaskWidget patterns, centralized coordination, and proven state management approach

### Design Decision 6: Engine Integration Coordination
**Options**:
A. **Direct Engine Dependencies**: Accept engines as constructor parameters with direct calls
B. **Engine Coordinator Pattern**: EngineCoordinator manages engine interactions
C. **Service Locator Pattern**: Runtime engine discovery and binding

**Chosen**: Option A for architectural compliance, straightforward testing, and minimal abstraction overhead

**Rationale**: These decisions prioritize simplicity, performance, and architectural consistency while supporting the specific requirements of Eisenhower Matrix task creation with seamless ColumnWidget integration.

**User Approval**: Approved 2025-01-23

## [2025-09-19] - TaskWidget Enhancement: Task Creation and Inline Editing Support

**Decision**: Enhance TaskWidget with unified renderer approach supporting task creation and inline editing modes with FormValidationEngine integration

**Context**: TaskWidget requires enhancement to support task creation and inline editing capabilities for CreateTaskDialog integration. The component must maintain backward compatibility while adding form-based creation mode (nil TaskData) and inline editing mode with real-time validation using FormValidationEngine.

**Options Considered**:

### Design Decision 1: Renderer Architecture Pattern
**Options**:
A. **Unified Form Renderer**: Extend existing TaskWidgetRenderer to handle both display and form modes
B. **Separate Form Component**: Create separate form components swapped during mode changes
C. **Mode-Specific Renderers**: Use different renderer implementations for display vs edit/create modes

**Chosen**: Option A for simplicity, easier state transitions, and shared validation logic

### Design Decision 2: Constructor API Design
**Options**:
A. **Overloaded Constructors**: `NewTaskWidget()` and `NewTaskWidgetForCreation()`
B. **Mode Parameter**: `NewTaskWidget(wm, fe, fve, taskData, mode)`
C. **Factory Methods**: Separate `NewDisplayTaskWidget()`, `NewCreationTaskWidget()`, `NewEditTaskWidget()`

**Chosen**: Option B with convenience wrappers for clean API and explicit mode specification

### Design Decision 3: State Management Strategy
**Options**:
A. **Extend Existing State**: Add Mode, FormData, IsFormDirty fields to TaskWidgetState
B. **Separate State Objects**: DisplayState and FormState with CurrentMode coordinator
C. **State Machine Pattern**: Discrete state objects with transition validation

**Chosen**: Option A for consistency with existing patterns and simpler state management

### Design Decision 4: FormValidationEngine Integration Pattern
**Options**:
A. **Direct Integration**: Call `validationEngine.ValidateFormInputs()` directly from TaskWidget
B. **Validation Service Layer**: Intermediate TaskValidationService wrapping FormValidationEngine
C. **Event-Driven Validation**: Validation triggered through event system

**Chosen**: Option A to maintain architectural compliance with direct Engine layer access

### Design Decision 5: Form Field Implementation Approach
**Options**:
A. **Direct Fyne Widgets**: Use `widget.Entry`, `widget.Select` directly with custom handlers
B. **Custom Form Builder**: Fluent API for form construction with integrated validation
C. **Form Component Library**: Reusable form field components with built-in validation

**Chosen**: Option A for framework consistency and direct control over form behavior

**Rationale**: This design maintains backward compatibility while cleanly extending TaskWidget with creation and editing capabilities. The unified renderer approach keeps complexity manageable while the mode parameter provides explicit control over widget behavior. Direct FormValidationEngine integration follows architectural guidelines for Engine layer access.

**User Approval**: Approved

---

## [2025-09-19] - TaskWidget Design: Custom Widget Architecture with Event Delegation

**Decision**: Implement TaskWidget using Custom Widget + Renderer Pattern with Event Delegation to WorkflowManager and Layered Error Handling

**Context**: TaskWidget requires implementation as a Fyne UI widget that displays individual task information in the EisenKan kanban board system. The component must integrate with WorkflowManager for task operations and FormattingEngine for text presentation while maintaining proper architectural layer separation and providing clean interface for parent containers.

**Options Considered**:

### Design Decision 1: Widget Architecture Pattern
**Options**:
A. **Custom Widget + Renderer Pattern**: Extend `widget.BaseWidget` with custom `fyne.WidgetRenderer`
B. **Container Composition Pattern**: Use `container.NewBorder` with standard Fyne widgets
C. **Rich Text Widget Extension**: Extend `widget.RichText` for task content display

**Chosen**: Option A for complete control over layout and rendering, follows Fyne best practices

### Design Decision 2: Dependency Management Strategy
**Options**:
A. **Constructor Injection**: Pass dependencies via `NewTaskWidget(wm, fe, taskData)`
B. **Setter Injection**: Create widget then set dependencies via methods
C. **Service Locator**: Widget queries global service registry for dependencies

**Chosen**: Option A with nil checks for graceful degradation

### Design Decision 3: State Management Approach
**Options**:
A. **Immutable State with Channels**: Use Go channels for state updates with immutable data structures
B. **Mutex-Protected Mutable State**: Use `sync.RWMutex` to protect mutable state fields
C. **Event-Driven State Machine**: State changes triggered only by discrete events

**Chosen**: Option A for thread-safe state management and clear state transitions

### Design Decision 4: User Interaction Event Handling
**Options**:
A. **Direct Fyne Event Handlers**: Implement `fyne.Tappable`, `fyne.Draggable`, etc. interfaces directly
B. **Event Delegation Pattern**: UI events trigger internal methods that delegate to WorkflowManager
C. **Command Pattern**: Convert UI events to command objects processed by WorkflowManager

**Chosen**: Option C (Event Delegation) because managing workflow related events is the WorkflowManager's responsibility

### Design Decision 5: Drag-Drop Implementation Strategy
**Options**:
A. **Native Fyne Drag-Drop**: Use `fyne.Draggable` and `fyne.DropTarget` interfaces
B. **Custom Drag-Drop System**: Implement custom drag detection and visual feedback
C. **Hybrid Approach**: Use Fyne drag events but custom visual feedback and workflow coordination

**Chosen**: Option C for maximum flexibility with Fyne integration

### Design Decision 6: Error Handling and Recovery Strategy
**Options**:
Based on error scenario analysis, selected **Local → Parent → Fail Early** approach:

**Local Widget-Level Handling**:
- Input sanitization using FormValidationEngine before sending to backend
- Visual error indicators (red borders, tooltips) for field-level validation errors
- Basic UI state management (focus, selection, etc.)

**Parent Context Escalation**:
- Backend validation errors (show error dialog with full backend message)
- WorkflowManager operation failures
- Data synchronization conflicts requiring user decision

**Fail Early and Loud (Outermost Context)**:
- Backend/WorkflowManager unavailability → Application-level failure
- FormattingEngine unavailability → Application-level failure
- Critical system resource exhaustion → Application-level failure

**Final Architecture**:
```
TaskWidget (extends widget.BaseWidget)
├── Custom Renderer (fyne.WidgetRenderer)
├── Dependencies (Constructor Injection)
│   ├── WorkflowManager (ITask, IDrag facets)
│   └── FormattingEngine (Text, Metadata facets)
├── State Management (Immutable with Channels)
├── Event Handling (Delegation to WorkflowManager)
├── Drag-Drop (Hybrid Fyne + Custom)
└── Error Handling (Layered: Local → Parent → Fail Early)
```

**Key Design Principles**:
- **Fyne Integration**: Native widget extending BaseWidget with custom renderer
- **Clean Dependencies**: Constructor injection with graceful degradation for missing dependencies
- **Event Delegation**: WorkflowManager handles all workflow-related event processing
- **Input Validation**: FormValidationEngine sanitizes inputs before workflow submission
- **Visual Error Feedback**: Red borders, tooltips for field errors, error dialogs for backend errors
- **Fail Early**: Clean application failure when core dependencies unavailable
- **Thread Safety**: Immutable state with channel-based updates for concurrent operations

**Integration Pattern**:
- TaskWidget processes Fyne UI events and forwards workflow events to WorkflowManager
- FormattingEngine handles text and metadata presentation formatting
- Input sanitization through FormValidationEngine before workflow submission
- Visual error indicators for validation failures, error dialogs for backend failures
- Parent containers handle workflow failures and data conflicts requiring user decisions

**Rationale**: This design provides optimal balance between native Fyne integration, clean architectural separation, and robust error handling. Custom widget with renderer provides maximum control over task display while event delegation maintains proper separation between UI concerns and workflow logic. The layered error handling approach allows local recovery where possible while escalating appropriately to parent contexts or failing early when dependencies are unavailable.

**User Approval**: Approved

## [2025-09-19] - WorkflowManager Design: Two-Facet Architecture with Workflow State Tracking

**Decision**: Implement WorkflowManager using Two-Facet Architecture with ITask and IDrag facets, integrated validation, and workflow state tracking

**Context**: WorkflowManager requires implementation as a Client Manager layer component that orchestrates client-side task workflow operations by coordinating UI engines with backend task management services. The manager must provide clear separation between task operations and drag-drop operations while maintaining workflow state for complex multi-engine coordination.

**Options Considered**:

### Option 1: Workflow-Specific Manager
- **Approach**: Single interface organized around workflow types (CreateTaskWorkflow, UpdateTaskWorkflow, etc.)
- **Benefits**: Simpler interface, follows established Manager patterns
- **Drawbacks**: Less modular, mixing task and drag concerns

### Option 2: Three-Facet Architecture
- **Approach**: ITask, IDrag, IValidation facets with separate validation concerns
- **Benefits**: Clear separation of all concerns
- **Drawbacks**: Over-engineering for Manager layer, validation split from domain operations

### Option 3: Two-Facet Architecture (CHOSEN)
- **Approach**: ITask and IDrag facets with integrated validation operations
- **Benefits**: Clear domain separation, validation integrated with domain concerns, appropriate complexity for Manager layer
- **Drawbacks**: Slightly more complex than single interface

**Final Architecture**:
```
WorkflowManager
├── ITask        // CreateTaskWorkflow, UpdateTaskWorkflow, DeleteTaskWorkflow, QueryTasksWorkflow
│                // ValidateTaskData, ValidateStatusTransition
└── IDrag        // ProcessDragDropWorkflow, ValidateMovement
```

**Key Design Principles**:
- **Domain-Focused Facets**: Task operations vs Drag operations with their respective validations
- **Workflow State Tracking**: Essential workflow context management for multi-engine coordination
- **Engine Integration**: Direct engine injection (FormValidationEngine, FormattingEngine, DragDropEngine)
- **Backend Integration**: Direct TaskManagerAccess integration for consistent patterns
- **Error Handling**: Simple error wrapping for consistency with established patterns
- **Manager Layer Compliance**: Proper orchestration responsibilities without Engine layer violations

**Workflow State Management**:
- Active workflow tracking with WorkflowState for complex operations
- Multi-engine coordination context management
- Progress tracking for long-running workflow operations
- Thread-safe concurrent workflow support

**Integration Pattern**:
- ITask facet coordinates FormValidationEngine and FormattingEngine for task operations
- IDrag facet coordinates DragDropEngine with task movement through backend
- Workflow state tracks multi-step operations across engine boundaries
- Error aggregation from engines provides unified user feedback

**Rationale**: This design provides optimal balance between domain separation and implementation simplicity for Manager layer responsibilities. Two facets clearly separate task and drag concerns while keeping validation integrated with domain operations. Workflow state tracking enables proper coordination of complex multi-engine operations essential for client-side workflow management.

**User Approval**: Approved

## [2025-09-19] - DragDropEngine Design: Focused Faceted Architecture

**Decision**: Implement DragDropEngine using Focused Faceted Architecture with IDrag, IDrop, and IVisualize facets

**Context**: DragDropEngine requires implementation as a stateless Engine layer component providing drag-drop coordination for kanban-style task management interfaces. The engine must focus purely on drag-drop mechanics without TaskWorkflowManager integration, allowing components above to handle task movement coordination.

**Options Considered**:

### Option 1: Faceted Architecture with Integration Coordination
- **Approach**: Four facets including IntegrationCoordinator for TaskWorkflowManager coordination
- **Facets**: DragCoordinator, DropZoneManager, DragVisualizer, IntegrationCoordinator
- **Benefits**: Complete drag-drop workflow including task movement
- **Drawbacks**: Engine layer handling Manager layer responsibilities, violates separation of concerns

### Option 2: Monolithic Interface
- **Approach**: Single comprehensive interface handling all drag-drop operations
- **Benefits**: Simpler interface, fewer components
- **Drawbacks**: Less modular, violates domain separation principles

### Option 3: Focused Faceted Architecture (CHOSEN)
- **Approach**: Three focused facets handling pure drag-drop mechanics only
- **Facets**: IDrag, IDrop, IVisualize
- **Benefits**: Clear separation of concerns, Engine focuses on drag-drop mechanics only, components above handle task coordination
- **Drawbacks**: Requires coordination layer above Engine

**Final Architecture**:
```
DragDropEngine
├── IDrag        // StartDrag, UpdateDragPosition, CompleteDrag, CancelDrag
├── IDrop        // RegisterDropZone, UnregisterDropZone, ValidateDropTarget, GetActiveZones
└── IVisualize   // CreateDragIndicator, UpdateIndicatorPosition, ShowDropFeedback, CleanupVisuals
```

**Key Design Principles**:
- **Engine Focus**: Pure drag-drop mechanics without task movement logic
- **Clean Separation**: Components above handle TaskWorkflowManager coordination and Fyne event processing
- **Immutable State**: Functional updates for safer concurrent operations
- **Temporary Fyne Objects**: Leverage framework capabilities for drag indicators
- **Geometric Bounds Checking**: Precise spatial detection without widget hierarchy dependencies
- **Architectural Compliance**: No Manager layer integration, maintains Engine layer boundaries

**Integration Pattern**:
- Engine provides drag-drop mechanics through focused facets
- UI components above process Fyne events and forward to Engine
- Task movement coordination handled by components that can access TaskWorkflowManager
- Engine remains stateless and focused on spatial drag-drop operations

**Rationale**: This design maintains proper architectural layer separation where Engine focuses purely on drag-drop mechanics. Components above handle integration concerns like Fyne events and TaskWorkflowManager coordination, following the principle that engines should be focused, stateless services without manager-level responsibilities.

**User Approval**: Approved

## [2025-09-18] - LayoutEngine Design: Layered Architecture with Domain-Specific Facets

**Decision**: Implement LayoutEngine using Layered Architecture with Domain-Specific Facets (Option 3) where ISpatialMath is internal implementation

**Context**: LayoutEngine requires implementation as a stateless Engine layer component providing sophisticated spatial calculations for UI components. The engine must integrate with Fyne Utility dependency and support layout calculation, responsive adaptation, animation support, Kanban optimization, and drag-and-drop spatial operations with 2ms performance requirements and 0.1 pixel mathematical precision.

**Options Considered**:

### Option 1: Pure Mathematical Engine
- **Approach**: Focus purely on mathematical calculations without UI framework knowledge
- **Benefits**: Maximum reusability, platform independence, pure mathematical precision
- **Drawbacks**: Requires additional integration layer, less optimization for Fyne-specific patterns

### Option 2: Fyne-Integrated Layout Engine
- **Approach**: Tight integration with Fyne container system and coordinate models
- **Benefits**: Optimal performance for Fyne workflows, direct container manipulation
- **Drawbacks**: Tighter coupling to Fyne, less portable across UI frameworks

### Option 3: Layered Architecture with Domain-Specific Facets (CHOSEN)
- **Approach**: Core spatial math layer with domain-specific facets for layout concerns
- **Benefits**: Balanced approach, testable layers, clear domain separation, extensible architecture
- **Drawbacks**: Slightly more complex than pure approaches

### Option 4: Domain-Specific Modules
- **Approach**: Separate modules for different layout concerns with inter-module coordination
- **Benefits**: Clear separation of concerns, focused optimization per domain
- **Drawbacks**: Potential code duplication, complex inter-module coordination

**Final Architecture**:
```
LayoutEngine
├── ISpatialMath      // INTERNAL: Core coordinate/bounds calculations, constraint solving
├── IResponsive       // PUBLIC: Breakpoint management, adaptive layouts
├── IAnimation        // PUBLIC: Layout state capture, interpolation, transitions
├── IKanban          // PUBLIC: Board layouts, card arrangements, scrolling optimization
├── IDragDrop        // PUBLIC: Drop validation, snap points, spatial feedback
└── IOptimization    // INTERNAL: Caching, performance optimization, memory management
```

**Key Design Principles**:
- **Internal Facets**: ISpatialMath and IOptimization provide foundational capabilities
- **Public Facets**: Domain-specific interfaces that UI components interact with
- **Layered Approach**: Core spatial math supports all domain-specific operations
- **Clean Abstractions**: UI components receive layout results, not raw mathematical calculations
- **Performance Focus**: Internal optimization facet ensures 2ms calculation requirements
- **Mathematical Precision**: Core spatial math maintains 0.1 pixel tolerance

**Rationale**: This design provides the optimal balance between mathematical precision, performance optimization, domain separation, and architectural cleanliness. The layered approach with internal spatial math ensures that complex coordinate calculations remain encapsulated while domain-specific facets provide clean, purpose-built interfaces for different layout concerns. This supports the 40 SRS requirements while maintaining Engine layer architectural compliance.

**User Approval**: Approved

---

## [2025-09-18] - FormattingEngine Design: Domain-Specific Facets

**Decision**: Implement FormattingEngine using Domain-Specific Facets architecture

**Context**: FormattingEngine requires implementation as a stateless Engine layer component providing comprehensive formatting capabilities for UI components. The engine must integrate with Format Utility dependency and support text, numeric, temporal, structure, template, and locale formatting operations.

**Options Considered**:

### Option 1: Domain-Specific Formatters
- **Approach**: Separate specialized formatters for each domain with shared coordination
- **Benefits**: Clear separation of domains, easy extension, domain specialization
- **Drawbacks**: Code duplication, complex coordination, inconsistent caching

### Option 2: Operation-Centric Engine
- **Approach**: Single engine with operation-based methods using Format Utility
- **Benefits**: Unified interface, centralized management, simple maintenance
- **Drawbacks**: Large component, less specialization, potential bottlenecks

### Option 3: Domain-Specific Facets (CHOSEN)
- **Approach**: FormattingEngine with specialized facets for each formatting domain using interface naming
- **Benefits**: Clean domain separation, focused responsibilities, easy extension, maintainable, testable
- **Drawbacks**: Slightly more complex structure than single engine

**Final Architecture**:
```
FormattingEngine
├── IText         // FormatText, FormatLabel, FormatMessage, FormatError
├── INumber       // FormatNumber, FormatPercentage, FormatFileSize, FormatCurrency
├── ITime         // FormatDateTime, FormatDuration, FormatRelativeTime, FormatTimeRange
├── IDatastructure // FormatList, FormatKeyValue, FormatJSON, FormatHierarchy
├── ITemplate     // ProcessTemplate, ValidateTemplate, CacheTemplate, GetTemplateMetadata
└── ILocale       // SetLocale, SetNumberFormat, SetDateFormat, SetCurrencyFormat
```

**Rationale**: Domain-Specific Facets provides the optimal balance of separation of concerns, maintainability, and performance. Each facet can specialize in its formatting domain while maintaining clean integration with Format Utility. The interface naming convention (IText, INumber, etc.) provides clear, concise identification of each facet's responsibility.

**User Approval**: Approved - "I like the domains. But I want one facet per domain." Updated naming: "But name the facets IText, INumber, ITime, IDatastructure etc."

## [2025-09-18] - FormValidationEngine Design: Function-Based Validation Engine

**Decision**: Implement FormValidationEngine using Function-Based Validation Engine (Option 2)

**Context**: FormValidationEngine requires implementation as a stateless Engine layer component providing pure form input validation patterns. Three design approaches were considered: Rule-Based, Function-Based, and Pipeline-Based validation engines.

**Options Considered**:

### Option 1: Rule-Based Validation Engine
- **Approach**: Configurable rule objects composed and applied to form data
- **Benefits**: Highly configurable, extensible, clean separation of concerns
- **Drawbacks**: Complex implementation, performance overhead, rule dependency management

### Option 2: Function-Based Validation Engine (CHOSEN)
- **Approach**: Specialized validation functions for each type with coordination
- **Benefits**: Simple implementation, better performance, easier maintenance, clear boundaries
- **Drawbacks**: Less flexible for dynamic scenarios, potential code duplication

### Option 3: Pipeline-Based Validation Engine
- **Approach**: Processing stages pipeline (Sanitization → Format → Pattern → Structure)
- **Benefits**: Clear processing flow, easy to extend, early termination performance
- **Drawbacks**: Overkill for simple cases, pipeline overhead, complex error handling

**Rationale**:
1. **Engine Pattern Alignment**: Function-based approach fits iDesign stateless engine pattern
2. **Performance**: Direct function calls without rule interpretation overhead
3. **Simplicity**: Easier to implement, test, and maintain
4. **Dependency Compliance**: Clean integration with ValidationUtility only
5. **Client Suitability**: Engine components should be focused and simple

**Implementation Architecture**:
- Core validation functions (ValidateFormInputs, ValidateFieldFormats, ValidateStructure, SanitizeInputs)
- Specialized validators (Text, Numeric, Date, Email, URL, Pattern)
- Cross-field validation components (Dependency, Conditional)
- ValidationUtility integration for basic patterns and error formatting

**User Approval**: Approved

---

## [2025-09-18] - UIStateAccess Removal: Integration into Client Components

**Decision**: Remove UIStateAccess from project plan and integrate its functionality directly into client components

**Context**: UIStateAccess was originally planned as a ResourceAccess layer component providing UI state persistence (window geometry, user preferences, view states, session data). However, after implementing TaskManager IContext interface with git-repo based storage and considering architectural constraints, UIStateAccess no longer provides sufficient unique value.

**Options Considered**:

### Option 1: Keep UIStateAccess as Originally Planned
- **Approach**: Implement UIStateAccess with full SRS specification (25+ operations)
- **Storage**: Use TaskManager IContext for git-repo persistence
- **Problems**:
  - ResourceAccess cannot call ResourceAccess (architectural constraint)
  - Without platform-native storage, reduces to data types + validation only
  - Adds unnecessary layer complexity for minimal value
  - Client would need Manager-to-ResourceAccess-to-Manager calls

### Option 2: Integrate UIStateAccess Functionality into Client Components (CHOSEN)
- **Approach**: Define UI state types and validation directly in client managers
- **Storage**: Client managers call TaskManager IContext directly
- **Benefits**:
  - Follows architectural constraints (Manager-to-Manager calls allowed)
  - Leverages existing git-repo storage through TaskManager
  - Eliminates unnecessary layer and reduces complexity
  - Components own their specific UI state concerns
  - Direct access to persistence without intermediate layers

**Rationale**:
1. **Architectural Compliance**: Manager-to-Manager calls are allowed, ResourceAccess-to-ResourceAccess are not
2. **Storage Solution Exists**: TaskManager IContext provides git-repo persistence that was requested
3. **Reduced Complexity**: Eliminates intermediate layer with minimal functionality
4. **Better Ownership**: WindowManager owns window state, NavigationManager owns navigation state
5. **Previous Storage Failure**: Platform-native storage approach already failed, git-repo approach succeeded

**Consequences**:
- Define WindowState, ViewState, etc. types in respective client managers
- Implement validation/processing functions directly in components that use them
- Client managers call TaskManager.IContext for UI state persistence

**User Approval**: Approved

---

## [2025-09-17] - IContext and IConfiguration Facets Design Decision: Implementation Architecture

**Decision**: Option 1 - Separate Interface Extensions with Embedded Structs

**Context**: Need to implement git-based JSON configuration management by extending existing TaskManager and BoardAccess services with facet interfaces. After UIStateAccess build failure, user directed pivot to simpler approach using git repository for configuration storage instead of OS-specific directories.

**Options Considered**:

### Option 1: Separate Interface Extensions with Embedded Structs
- **Approach**: Create separate interface definitions for IContext and IConfiguration
- **Implementation**: Use Go interface embedding and type assertion
- **Storage**: Direct git operations through VersioningUtility
- **Data Format**: Separate JSON files for context vs configuration
- **Integration**: Services contain facet instances as fields

### Option 2: Facet Pattern with Service Composition
- **Approach**: Implement facets as composable service components
- **Implementation**: Services contain facet instances as fields
- **Storage**: Unified storage abstraction with type-specific handlers
- **Data Format**: Single JSON structure with facet-specific sections
- **Integration**: Shared facet container accessed by multiple services

### Option 3: Method Extensions with Context Parameters
- **Approach**: Extend existing services with context-aware methods
- **Implementation**: Add context parameter to existing operations
- **Storage**: Context-driven storage routing
- **Data Format**: Context-determined JSON schema
- **Integration**: Additional methods on existing service interfaces

**Rationale**: Choose Option 1 for clear separation of concerns between context and configuration data, follows Go interface composition patterns, maintains architectural layer integrity, and supports independent testing and evolution. This approach provides optimal git diff behavior for configuration changes and enables clean delegation from Manager layer to Resource Access layer.

**Consequences**:
- **TaskManager IContext Facet**: `Load(contextType string) (ContextData, error)` and `Store(contextType string, data ContextData) error` for UI context management
- **BoardAccess IConfiguration Facet**: `Load(configType, identifier string) (ConfigurationData, error)` and `Store(configType, identifier string, data ConfigurationData) error` for board configuration management
- **Git Repository Structure**: `.eisenkan/context/` for UI state/preferences/sessions, `.eisenkan/config/boards/` for board configurations, `.eisenkan/config/workflows/` for workflow rules
- **Data Contracts**: ContextData and ConfigurationData structs with type, version, metadata, and JSON-serializable content
- **Error Handling**: Structured error responses with recovery suggestions following established patterns
- **Validation**: JSON schema validation for both context and configuration data
- **Layer Integration**: TaskManager delegates to git storage, BoardAccess coordinates with VersioningUtility
- **Requirements Coverage**: Extends existing SRS documents REQ-TASKMANAGER-019 through REQ-TASKMANAGER-022 and REQ-BOARDACCESS-025 through REQ-BOARDACCESS-029

**User Approval**: Approved on [2025-09-17]

## [2025-09-17] - UIStateAccess Design Decision: Implementation Architecture

**Decision**: Option C - Hybrid Platform-Optimized Approach

**Context**: Need to determine the implementation approach for UIStateAccess while providing cross-platform UI state persistence that leverages native OS storage mechanisms. The service requires optimal platform integration, excellent performance (<10ms state access), and robust error recovery while maintaining data integrity across Windows, macOS, and Linux platforms.

**Options Considered**:
- **Option A: Simple File-Based Approach** - Direct JSON file storage with basic error handling, simple implementation but limited performance optimization and platform differences exposed
- **Option B: Database-Backed Approach** - Embedded SQLite database with structured schema, excellent concurrency and ACID transactions but additional dependency and complexity overhead
- **Option C: Hybrid Platform-Optimized Approach** - Platform-specific storage backends with common interface, optimal platform integration and performance but more complex implementation
- **Option D: Layered Cache Architecture** - In-memory cache with persistent backend, excellent performance but memory usage concerns and cache coherency complexity

**Rationale**: Choose Option C to provide optimal balance of platform integration, performance, and maintainability. This approach leverages native OS storage mechanisms (Windows Registry + JSON files, macOS Preferences + plist, Linux XDG + JSON) for best user experience while maintaining cross-platform interface consistency. Platform-specific optimizations provide superior performance and reliability compared to generic approaches.

**Consequences**:
- Platform-specific storage implementations: Windows (Registry + AppData JSON), macOS (Preferences API + plist), Linux (XDG directories + JSON)
- Common interface abstracts platform differences through Strategy pattern
- Core components: StateManager (main interface), PlatformStorage (platform backends), StateValidator (data validation), BackupManager (multi-tier recovery), CacheLayer (performance optimization), LoggingIntegration (comprehensive monitoring)
- Key design patterns: Strategy (platform storage), Template Method (common validation), Observer (state change notifications), Command (atomic operations), Factory (backend creation)
- Multi-layer backup strategy with platform-specific optimizations
- Excellent cross-platform compatibility while leveraging each platform's strengths
- Performance targets met through platform-native optimizations and intelligent caching

**User Approval**: Approved on [2025-09-17]

## [2025-09-16] - FyneUtility Design Decision: Implementation Architecture

**Decision**: Option C - Hybrid Approach with Smart Defaults

**Context**: Need to determine the implementation approach for FyneUtility while providing foundational Fyne framework abstraction that enables consistent widget creation, theme management, and resource handling across all client UI components. The service requires cross-platform compatibility, high performance (widget creation <1ms), and seamless integration with ValidationUtility and FormatUtility.

**Options Considered**:
- **Option A: Functional Library with Typed Configuration** - Pure functional approach with strongly-typed configuration structs, explicit parameters, complete type safety but complex function signatures
- **Option B: Builder Pattern with Fluent Interface** - Builder pattern for complex configurations with method chaining, flexible but potentially stateful and complex state management
- **Option C: Hybrid Approach with Smart Defaults** - Simple functions with smart defaults plus functional options pattern, clean API for common cases with power when needed

**Rationale**: Choose Option C to provide excellent developer experience with simple API for common cases (`CreateButton("OK")`) while maintaining flexibility through functional options for advanced scenarios. This approach maintains stateless functional design for thread safety, enables easy extension without breaking changes, and provides self-documenting configuration through option function names. The pattern works well with ValidationUtility and FormatUtility integration requirements.

**Consequences**:
- Simple API for common cases: `CreateButton("Save")`
- Advanced configuration through functional options: `CreateButton("Save", WithButtonStyle(PrimaryButton), WithButtonIcon(icon))`
- 30+ core functions across 10 functional areas (widget creation, layout management, theme operations, etc.)
- Maintains stateless design for thread safety
- Easy extension with new option functions without breaking existing code
- Integration points with ValidationUtility (`WithValidation()`) and FormatUtility (`WithFormatter()`)
- Performance optimizations through resource caching, widget factories, and lazy loading
- Cross-platform compatibility with graceful platform-specific fallbacks

**User Approval**: Approved on [2025-09-16]

## [2025-09-16] - ValidationUtility Design Decision: Implementation Architecture

**Decision**: Option A - Simple Functional Approach

**Context**: Need to determine the implementation approach for ValidationUtility while maintaining consistency with FormatUtility's successful functional design and supporting all 13 SRS interface operations with proper error handling and validation result structures.

**Options Considered**:
- **Option A: Simple Functional Approach** - Direct functions, validation result structs, minimal types, standard error handling
- **Option B: Interface-Based Service Pattern** - IValidationUtility interface, struct implementation, rich error handling, consistent with other services
- **Option C: Hybrid Approach** - Core interface with functional helpers, mixed complexity

**Rationale**: Choose Option A to maintain consistency with FormatUtility's proven functional design. ValidationUtility operations are stateless validation functions that benefit from direct function calls without interface overhead. The functional approach aligns with the universal utility pattern and provides optimal performance for validation operations.

**Consequences**:
- Direct functions for all 13 operations (ValidateString, ValidateNumber, etc.)
- Rich data contracts (ValidationResult, ValidationRule, StringConstraints, NumericConstraints structs)
- Standard Go error handling with structured validation results
- No interface overhead - direct function calls
- Consistent with FormatUtility design patterns
- Thread-safe by design (stateless functions)
- Easy to use without service instantiation

**User Approval**: Accepted

## [2025-09-16] - FormatUtility Design Decision: Implementation Architecture

**Decision**: Option A - Simple Functional Approach

**Context**: Need to determine the implementation approach for FormatUtility while maintaining consistency with existing utility patterns (LoggingUtility, CacheUtility) and supporting all 11 SRS interface operations with proper error handling and extensibility.

**Options Considered**:
- **Option A: Simple Functional Approach** - Direct functions, no state, minimal types, simple error handling
- **Option B: Interface-Based Service Pattern** - IFormatUtility interface, struct implementation, rich error handling, consistent with existing utilities
- **Option C: Hybrid Approach** - Core interface with functional helpers, mixed complexity

**Rationale**: Choose Option A because FormatUtility operations are purely functional/stateless with no need for mocking in tests. Direct functions provide simpler implementation, easier testing, and optimal performance without interface overhead. The stateless nature makes interfaces unnecessary for abstraction.

**Consequences**:
- Direct functions for all 11 operations (TrimText, ConvertCase, etc.)
- Minimal type definitions (TextCaseType, FileSizeUnit, ValidationRule enums/structs)
- Standard Go error handling with contextual information
- No interface overhead - direct function calls
- Simpler implementation and testing
- Thread-safe by design (stateless functions)
- Easy to use without service instantiation

**User Approval**: Approved on [2025-09-16]

## [2025-09-14] - TaskManagerAccess: Implementation Architecture

**Decision**: Option A - Simple Channel-based Implementation

**Context**: Need to determine the internal implementation approach for TaskManagerAccess async operations while maintaining interface contract requirements and ensuring proper error handling, data transformation, and cache coordination.

**Options Considered**:
- **Option A: Simple Channel-based Implementation** - Direct channel returns from all async methods, minimal internal state management, direct TaskManager service calls with error wrapping
- **Option B: Worker Pool with Request Queue** - Internal worker goroutines handling service calls, request queuing for batching and optimization, more complex but potentially better performance
- **Option C: Hybrid Approach with Smart Caching** - Channel-based interface with internal caching logic, request deduplication and batching where beneficial, balance between simplicity and performance

**Rationale**: Choose Option A to start with simplicity and clean interface implementation. Direct channel-based approach provides straightforward async operations without internal complexity. Performance optimizations can be added later without changing the interface contract. This approach aligns with iDesign principles of starting simple and adding complexity only when needed.

**Consequences**:
- Direct async method implementation with immediate channel returns
- Minimal internal state and complexity
- Straightforward error translation and data transformation logic
- Simple cache coordination without internal queuing
- Easy to test and debug
- Performance optimization opportunities preserved for future enhancement
- Clear separation between interface contract and implementation details

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - Native GUI Framework Selection: Fyne

**Decision**: Fyne for native GUI implementation

**Context**: User prefers native GUIs over web UIs and needs drag-and-drop support for task management. Evaluated Go native GUI frameworks for cross-platform compatibility, drag-and-drop capabilities, and development simplicity.

**Options Considered**:
- **Fyne**: Built-in drag-and-drop via `fyne.Draggable` interface, cross-platform, Material Design-inspired, active development
- **Wails v2**: HTML5 drag/drop API support, web frontend with native packaging, modern architecture
- **Walk**: Native Windows drag/drop, Windows-only, lightweight, true native look
- **Gio**: Gesture-based drag support, GPU-accelerated immediate mode GUI, steeper learning curve
- **GTK (go-gtk)**: Full GTK drag/drop capabilities, Linux-native, comprehensive widgets
- **Qt (therecipe/qt)**: Complete Qt drag/drop framework, professional feature set, complex setup

**Rationale**: Fyne provides the optimal balance of simplicity, cross-platform support, and built-in drag-and-drop capabilities for a task management application. Material Design-inspired interface aligns well with modern user expectations, and the framework has active development with good documentation.

**Consequences**:
- Cross-platform native GUI (Windows, macOS, Linux, mobile)
- Built-in drag-and-drop support via `fyne.Draggable` interface
- Material Design-inspired UI components
- Single dependency with good Go ecosystem integration
- Active community and development support
- Simple deployment and distribution

**User Approval**: Approved on [2025-09-14]

## [2025-09-15] - CacheUtility: Implementation Architecture

**Decision**: Option A - Map-Based with RWMutex

**Context**: Need to determine the internal implementation approach for CacheUtility thread-safe caching operations while maintaining performance requirements (1ms Get, 5ms Set) and supporting TTL management, pattern-based invalidation, and LRU eviction.

**Options Considered**:
- **Option A: Map-Based with RWMutex** - Single map with read-write mutex protection, LRU tracking using doubly-linked list, simple implementation
- **Option B: Segmented Cache with Fine-Grained Locking** - Multiple cache segments with separate locks, reduced contention but more complex
- **Option C: Channel-Based Actor Model** - Single goroutine handling all operations via channels, serialized operations but channel overhead
- **Option D: Sync.Map with Custom TTL Management** - Go's built-in concurrent map with separate TTL tracking, good read performance but complex TTL management

**Rationale**: Choose Option A to align with the "start simple" approach used for TaskManagerAccess. Map-based implementation with RWMutex provides straightforward thread safety, meets performance requirements for expected cache sizes, and allows for clear LRU tracking and pattern invalidation implementation. Background cleanup goroutine handles TTL expiration efficiently.

**Consequences**:
- Simple implementation with clear thread safety guarantees
- LRU tracking using doubly-linked list with map pointers for O(1) access
- Background cleanup goroutine for automated expired entry removal
- Pattern matching using filepath.Match() for wildcard support
- Atomic operations for statistics to avoid mutex overhead during reads
- Easy to test and debug with predictable behavior
- Performance optimization opportunities preserved for future enhancement

**User Approval**: Approved on [2025-09-15]

## [2025-09-14] - TaskManager: Interface Contract Design

**Decision**: Option A - Single Comprehensive Interface

**Context**: Need to determine how TaskManager should organize its operations and data contracts while maintaining interface consistency with SRS requirements.

**Options Considered**:
- **Option A: Single Comprehensive Interface** - All 7 operations in one interface with rich data contracts
- **Option B: Separated CRUD and Workflow Interfaces** - Split between TaskCRUD and TaskWorkflow interfaces  
- **Option C: Operation-Grouped Interfaces** - Three focused interfaces (TaskManagement, TaskQuery, TaskWorkflow)

**Rationale**: Choose Option A because all operations concern the same "facet" of the manager.

**Consequences**:
- Single TaskManager interface with 7 operations as specified in SRS
- Rich data contracts: TaskRequest, TaskResponse, WorkflowStatus, ValidationResult entities
- Direct alignment with SRS service contract requirements
- Simplified client integration with single interface
- Future interface segregation possible if needed

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - TaskManager: Internal Architecture Structure

**Decision**: Option A - Simple Manager with Direct Dependencies

**Context**: Need to determine internal component organization for TaskManager while maintaining Manager layer architectural constraints.

**Options Considered**:
- **Option A: Simple Manager with Direct Dependencies** - TaskManager directly coordinates with BoardAccess, RuleEngine, LoggingUtility
- **Option B: Manager with Internal Workflow Orchestrator** - Additional WorkflowOrchestrator internal component
- **Option C: Manager with Specialized Internal Services** - Multiple internal components (TaskValidator, WorkflowOrchestrator, CascadeHandler)

**Rationale**: Choose Option A because the TaskManager with its dependencies builds a subsystem accordign to iDesign.

**Consequences**:
- Single TaskManager service implementation
- Direct calls to BoardAccess for data persistence
- Direct calls to RuleEngine for business rule validation
- Direct calls to LoggingUtility for operational logging
- Embedded workflow orchestration logic within TaskManager
- Simple implementation and maintenance approach

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - TaskManager: Subtask Workflow Coupling Strategy

**Decision**: TaskManager-Orchestrated with RuleEngine Compliance Verification

**Context**: Need to determine responsibility allocation between TaskManager and RuleEngine for subtask workflow coupling rules implementation.

**User Specification**: "The RuleEngine shall verify compliance with the rules as specified in the RuleEngine SRS and the RuleEngine code. All the remaining responsibilities shall be covered by the TaskManager."

**Rationale**: Clear separation of concerns where RuleEngine validates business rule compliance (as per its SRS) while TaskManager implements the actual workflow coupling orchestration. This maintains the Manager layer's orchestration responsibilities while leveraging RuleEngine for rule validation.

**Consequences**:
- TaskManager implements subtask workflow coupling logic (REQ-TASKMANAGER-016, REQ-TASKMANAGER-017)
- RuleEngine validates rule compliance before TaskManager applies changes
- TaskManager coordinates parent-child status transitions based on active policies
- TaskManager handles cascade operations for task deletion/archival
- Clear responsibility boundaries between rule validation and workflow orchestration

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - TaskManager: Error Handling Strategy

**Decision**: Option B - Error Wrapping with Context

**Context**: Need to choose error handling approach that provides structured information per SRS requirements while maintaining consistency with existing DDR patterns.

**Options Considered**:
- **Option A: Structured Error Types** - Custom error types with code, message, details, suggestions
- **Option B: Error Wrapping with Context** - Go standard error wrapping following BoardAccess DDR pattern
- **Option C: Hybrid Domain-Specific Errors** - Structured for workflow violations, wrapped for system failures

**Rationale**: Choose Option B for consistency with existing BoardAccess DDR decision and Go idioms. Error wrapping with contextual information provides sufficient debugging capability while maintaining implementation simplicity and architectural consistency across ResourceAccess components.

**Consequences**:
- Go standard error wrapping with contextual information
- Error chain preservation from dependencies (BoardAccess, RuleEngine)
- Contextual annotations for TaskManager operations
- Consistency with established DDR patterns
- Good error debugging capability without over-engineering

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - TaskManager: Concurrency and Thread Safety Strategy

**Decision**: Option A - Service-Level Mutex

**Context**: Need to ensure thread-safe operations for concurrent task workflow orchestration while maintaining data consistency per SRS performance and reliability requirements.

**Options Considered**:
- **Option A: Service-Level Mutex** - Single RWMutex following BoardAccess DDR pattern
- **Option B: Operation-Level Locking** - Fine-grained locking by operation type
- **Option C: Stateless with Dependency Coordination** - Rely on dependency locking

**Rationale**: Choose Option A for consistency with BoardAccess DDR decision and guaranteed data consistency. Service-level RWMutex ensures safe concurrent access to workflow orchestration operations while maintaining architectural consistency with other Manager components.

**Consequences**:
- Single RWMutex protecting all TaskManager operations
- Multiple concurrent readers, exclusive writers
- Guaranteed workflow consistency for parent-child operations
- Consistency with established BoardAccess concurrency pattern
- Simple implementation with reliable thread safety

**User Approval**: Approved on [2025-09-14]

## [2025-09-14] - Subtask Position Storage Decision: Position in Filename Not Content

**Decision**: Store task and subtask position information in the filename prefix, not in the JSON content

**Context**: Need to determine where to store position information for tasks and subtasks within columns/sections for proper ordering while maintaining optimal git diff behavior and data consistency.

**Options Considered**:

### Option A: Position in JSON Content
- **Structure**: Task JSON contains position field: `{"id": "12345", "position": 1, "title": "..."}`
- **Filename**: Static names like `task-12345.json`, `subtask-67890.json`
- **Advantages**:
  - Position data travels with task content
  - Simpler filename management
  - No filename changes for position updates
- **Disadvantages**:
  - Position changes require JSON file content modification
  - Larger git diffs for position updates
  - Risk of position data inconsistency with file location
  - Complex validation between JSON content and directory location

### Option B: Position in Filename Prefix
- **Structure**: Position encoded in filename prefix
- **Task Filename**: `<position>-task-<id>.json` (e.g., `001-task-12345.json`)
- **Subtask Filename**: `<position>-subtask-<id>.json` (e.g., `001-subtask-67890.json`)
- **Directory Names**: No position prefix (e.g., `task-12345/` for subtask container)
- **Advantages**:
  - Optimal git diffs for position changes (file rename vs content change)
  - Position immediately visible in directory listings
  - Natural sorting by filename gives correct position order
  - No risk of position data inconsistency
  - JSON content focuses purely on task data

**Rationale**: Choose Option B to optimize for REQ-FORMAT-002 (minimal git diffs for common operations). Position changes are common operations in Kanban boards (reordering tasks within columns/sections). Using filename prefixes makes position changes into git file renames rather than content modifications, resulting in cleaner version history and better merge conflict resolution.

**Consequences**:
- Task files: `001-task-12345.json`, `002-task-67890.json` etc.
- Subtask files: `001-subtask-11111.json`, `002-subtask-22222.json` etc.  
- Subtask directories: `task-12345/` (no position prefix to minimize directory moves)
- Position changes become file rename operations (optimal for git)
- Directory listings naturally sort by position
- JSON content remains focused on task attributes only
- File management operations must handle position prefix updates

**User Approval**: Approved on [2025-09-14]

## [2025-09-13] - ValidationEngine Service Decision: Service Not Required

**Decision**: Do not implement ValidationEngine service for EisenKan

**Context**: After analyzing the ValidationEngine scope and architectural requirements, determined that the validity of requests is already sufficiently verified by existing components in the system. The planned ValidationEngine would primarily duplicate validation logic already present in BoardAccess (data integrity validation) and RuleEngine (business rule validation).

**Options Considered**:
- **Option A: Implement ValidationEngine** - Add new Engine component for orchestrated validation scenarios
- **Option B: Enhance existing components** - Extend BoardAccess and RuleEngine validation capabilities as needed
- **Option C: No ValidationEngine** - Leave validation distributed across appropriate components (BoardAccess for data integrity, RuleEngine for business rules)

**Rationale**: Existing architecture already provides comprehensive validation:
- BoardAccess handles data integrity validation (required fields, formats, consistency)
- RuleEngine handles business rule validation (WIP limits, workflow transitions, dependencies)
- Cross-task dependency validation can be implemented in BoardAccess when needed
- External system integration validation can be added to specific components as requirements emerge
- No compelling use cases identified that require orchestrated validation beyond existing capabilities

**Consequences**:
- Simplified architecture with fewer components
- Validation logic remains close to domain responsibility (data validation in data layer, business rules in rule engine)
- Future validation needs addressed incrementally in appropriate components
- No validation orchestration layer - complex validations handled by extending existing components
- Reduced system complexity and maintenance overhead

**User Approval**: Approved on [2025-09-13]

## [2025-09-13] - RuleEngine Design Decision: Rule Context Data Access Strategy

**Decision**: Option C - Rule Engine with ResourceAccess Integration

**Context**: RuleEngine needs broader board context for Kanban rules (WIP limits, subtask status, column timestamps, other tasks' priorities) beyond single task event context defined in current SRS.

**Options Considered**:
- **Option A: Minimal Context** - Single task only, simple but cannot implement complex Kanban rules
- **Option B: Rich Context** - Manager provides board context, keeps Engine pure but increases Manager complexity
- **Option C: ResourceAccess Integration** - RuleEngine calls ResourceAccess directly for board data when needed

**Rationale**: Option C provides complete rule evaluation capabilities while maintaining architectural compliance (Engines can access ResourceAccess components per iDesign). This enables all identified Kanban rule types without overcomplicating the Manager layer.

**Consequences**:
- RuleEngine can implement WIP limits, age limits, subtask dependencies, and priority flow rules
- Direct access to BoardAccess for task counts and timestamps
- Direct access to task hierarchy information for subtask rules  
- RuleEngine becomes more capable but retains stateless operation
- Cleaner separation between rule logic (Engine) and workflow orchestration (Manager)

**User Approval**: Approved on [2025-09-13]

## [2025-09-13] - RuleEngine Design Decision: Rule Evaluation Architecture

**Decision**: Option B - Complete Sequential Processor

**Context**: Need to meet REQ-RULEENGINE-002 requirement to report all violations in one evaluation.

**Options Considered**:
- **Option A: Eager Sequential Processor** - Simple but cannot report all violations at once, because it stops at first violation
- **Option B: Complete Sequential Processor** - Simple and can report all violations at once, because it aggregates all violations
- **Option C: Parallel Evaluator** - Evaluate all rules in parallel, aggregate results to report all violations
- **Option D: Hybrid Priority Groups** - Complex implementation with partial violation reporting

**Rationale**: Option B meets REQ-RULEENGINE-002 requirement to evaluate all matching rules and report all violations. Parallel evaluation of option C) would provide better performance for large rule sets, but that's an implementation detail not considered critical and could be changed later on without breaking the interface.

**Consequences**:
- Can report all rule violations in single evaluation (meets SRS requirement)
- Stateless evaluation per individual rule
- Requires result aggregation
- All applicable rules evaluated regardless of priority (violations sorted by priority in results)

**User Approval**: Approved on [2025-09-13]

## [2025-09-13] - Fractal Design Decision

**Decision**: Option B - Integrate subtasks directly into existing interfaces

**Context**: Whether to use fractal design (treating parent tasks as boards with separate instances) or integrate subtasks into existing interfaces.

**Options Considered**:
- **Option A: Fractal design** - Separate system instances per task with subtasks, elegant API but complex routing and UI integration
- **Option B: Direct integration** - Single system instance managing both tasks and subtasks, simpler architecture

**Rationale**: Could not solve routing/instance issues cleanly in fractal approach while maintaining iDesign principles.

**Consequences**:
- Interfaces are moderately more complicated and explicit
- Subtask special cases mean the implementation is more complicated
- Can still use the "task as a board" concept as mental model

**User Approval**: Approved on [2025-09-13]

## [2025-09-13] - Subtasks Design Decision

**Decision**: Option A - Support subtasks

**Context**: Need to organize related tasks within boards. Previous workarounds (extra columns, tags) were tedious for tracking task groups with same goal.

**Options Considered**:
- **Option A: Support subtasks** - Fulfills user need but increases complexity
- **Option B: No subtasks** - Simpler but doesn't meet user need

**Rationale**: User need trumps implementation effort.

**Consequences**:
- More complex interfaces requiring subtask-aware operations
- Implementation complexity in storage, querying, and state management
- Enhanced data model with parent-child relationships
- Additional validation logic for subtask hierarchies
- UI complexity for nested task visualization

**User Approval**: Approved on [2025-09-13]

## [2025-09-12] - RulesAccess Design Decision: Concurrent Access Strategy

**Decision**: Option C - VersioningUtility-Level Coordination

**Context**: RulesAccess must handle concurrent read/write operations safely while maintaining data consistency and performance.

**Options Considered**:
- **Option A: File-Level Locking**
  - Use file locks (flock) for rule file access
  - OS-level coordination across processes
  - Simple implementation
  - Potential performance bottleneck
  
- **Option B: In-Memory Mutex with Caching**
  - Mutex per directory path
  - Cache rule sets in memory with TTL
  - Better performance for read-heavy workloads
  - Memory usage and cache consistency concerns
  
- **Option C: VersioningUtility-Level Coordination**
  - Rely on VersioningUtility for concurrency control
  - Atomic commit operations handle conflicts
  - Consistent with other ResourceAccess components
  - Version control overhead for all operations

**Rationale**: Choose Option C for architectural consistency and leveraging existing infrastructure. VersioningUtility already provides atomic operations and conflict detection. This approach maintains consistency with BoardAccess and other ResourceAccess components while providing proper concurrency control through version control mechanisms.

**Consequences**:
- Consistent with other ResourceAccess layer components
- Leverages existing VersioningUtility concurrency control
- Atomic operations and conflict detection built-in
- Version control overhead for all operations (acceptable trade-off)
- Simplified RulesAccess implementation by delegating concurrency to VersioningUtility

**User Approval**: Approved on [2025-09-12]

## [2025-09-12] - RulesAccess Design Decision: Rule Validation Architecture

**Decision**: Option A - Embedded Schema Validation

**Context**: RulesAccess must validate rule syntax, semantics, dependencies, and conflicts. Need to determine validation architecture and extensibility approach.

**Options Considered**:
- **Option A: Embedded Schema Validation**
  - JSON schema validation built into RulesAccess
  - Schema defined as Go structs with validation tags
  - Simple implementation, fast validation
  - Schema changes require code changes
  
- **Option B: External Schema File**
  - JSON schema stored as separate file (rules-schema.json)
  - Runtime schema loading and validation
  - Schema updates without code changes
  - More complex validation logic
  
- **Option C: Plugin-Based Validation**
  - Extensible validation interface for different rule types
  - Support for custom validators per workflow methodology
  - Maximum flexibility for future extensions
  - Complex implementation and testing

**Rationale**: Choose Option A for simplicity and performance. Embedded schema validation using Go structs with validation tags provides fast, compile-time safety and straightforward implementation. Schema changes requiring code changes is acceptable trade-off for initial implementation.

**Consequences**:
- Fast validation performance with compile-time safety
- Simple implementation and testing
- Schema changes require code updates and recompilation
- Good starting point that can be enhanced later if needed
- Direct integration with Go type system

**User Approval**: Approved on [2025-09-12]

## [2025-09-12] - RulesAccess Design Decision: Rule Storage Structure

**Decision**: Option A - Single rules.json File

**Context**: RulesAccess needs to store rule sets for board directories in JSON format with version control. Need to determine the file organization and naming strategy within directories.

**Options Considered**:
- **Option A: Single rules.json File**
  - Store entire rule set in one `rules.json` file per directory
  - Simple atomic replacement for rule set changes
  - Easy to read/write complete rule set
  - Version control tracks entire rule set changes
  
- **Option B: Multiple Rule Category Files**
  - Separate files: `validation-rules.json`, `workflow-rules.json`, `automation-rules.json`, `notification-rules.json`
  - Granular version control per rule category
  - Smaller files for specific rule types
  - More complex atomic replacement logic
  
- **Option C: Individual Rule Files**
  - One file per rule: `rule-{id}.json`
  - Maximum granular version control
  - Complex rule set assembly and validation
  - Contradicts SRS requirement for atomic rule set operations

**Rationale**: Choose Option A to avoid coordination conflicts with BoardAccess. Originally considered storing rules in board.json, but that would require BoardAccess and RulesAccess to coordinate file access. A separate rules.json file provides clean separation of concerns while maintaining atomic rule set operations required by SRS.

**Consequences**:
- Clean separation between board configuration and rule data
- No coordination required with BoardAccess for file access
- Simple atomic rule set replacement implementation
- Version control tracks complete rule set changes as single units
- Easy to implement and maintain

**User Approval**: Approved on [2025-09-12]

## [2025-09-11] - BoardAccess Design Decision: Concurrency and Thread Safety Strategy

**Decision**: Service-Level Mutex Protection (Option A)  

**Context**: SRS requires concurrent operations without data corruption (REQ-PERFORMANCE-002) and data consistency under simultaneous operations (REQ-RELIABILITY-002). VersioningUtility provides repository-level locking, but BoardAccess needs operation-level coordination.

**Options Considered**:

### Option A: Service-Level Mutex Protection
- **Strategy**: Single mutex protecting all TaskAccess operations
- **Implementation**: RWMutex allowing multiple readers, exclusive writers
- **Advantages**:
  - Simple implementation
  - Guaranteed data consistency
  - No deadlock potential
- **Disadvantages**:
  - Limited concurrency (serializes all operations)
  - Suboptimal performance for read-heavy workloads
  - Doesn't leverage VersioningUtility's repository-level locking

### Option B: Operation-Level Locking
- **Strategy**: Different locks for read vs. write operations, with task-level granularity
- **Implementation**: Map of task ID mutexes for fine-grained locking
- **Advantages**:
  - Maximum concurrency for independent tasks
  - Optimal read/write separation
  - Better performance characteristics
- **Disadvantages**:
  - Complex lock management
  - Potential deadlock scenarios
  - Memory overhead for lock map

### Option C: Repository Delegation with Atomic Operations
- **Strategy**: Rely on VersioningUtility repository locking, make TaskAccess operations atomic
- **Implementation**: Each operation completes entirely within VersioningUtility transaction boundaries
- **Advantages**:
  - Leverages existing VersioningUtility thread safety
  - Consistent with architectural layering
  - No additional locking complexity
- **Disadvantages**:
  - Limited by VersioningUtility locking granularity
  - May not optimize for TaskAccess-specific access patterns
  - Potential performance bottleneck for bulk operations

**Rationale**: Option A chosen for guaranteed data consistency and simple implementation. Service-level RWMutex ensures file consistency under concurrent requests, which is critical for directory-structure-per-column approach. Simpler than complex lock management while providing reliable concurrent access.

**User Approval**: Option A approved on [2025-09-11]

## [2025-09-11] - BoardAccess Design Decision: Error Handling and Recovery Strategy  

**Decision**: Error Wrapping with Context (Option B)

**Context**: SRS requires structured error information (REQ-RELIABILITY-001) and graceful failure handling when VersioningUtility unavailable (REQ-RELIABILITY-003). Need consistent error response format and recovery mechanisms.

**Options Considered**:

### Option A: Structured Error Types with Recovery Actions
- **Error Structure**: Custom error types implementing structured format
  ```go
  type TaskAccessError struct {
      Code        string            // ERROR_TASK_NOT_FOUND, ERROR_STORAGE_FAILED
      Message     string            // Human-readable description
      Details     map[string]interface{} // Technical debugging info
      Suggestions []string          // Recovery action suggestions
      Cause       error            // Underlying error if any
  }
  ```
- **Recovery Strategy**: Return specific suggestions per error type
- **Advantages**:
  - Meets SRS structured error requirements precisely
  - Clear recovery guidance for callers
  - Rich debugging information
- **Disadvantages**:
  - More complex error handling implementation
  - Potential over-engineering for simple errors

### Option B: Error Wrapping with Context
- **Error Structure**: Go standard error wrapping with context
- **Strategy**: Use fmt.Errorf with contextual information
- **Advantages**:
  - Follows Go idioms
  - Simple implementation
  - Good error chain preservation
- **Disadvantages**:
  - Less structured than SRS requirements
  - Limited recovery action guidance

### Option C: Hybrid Approach - Structured for Domain Errors, Simple for System Errors
- **Strategy**: Structured errors for business logic, wrapped errors for system failures
- **Implementation**: Custom types for task-specific errors, standard wrapping for storage/logging errors
- **Advantages**:
  - Meets SRS requirements for important cases
  - Simpler handling for infrastructure errors
  - Balanced complexity
- **Disadvantages**:
  - Inconsistent error handling patterns
  - Callers need to handle multiple error types

**Rationale**: Option B chosen for simpler implementation following Go idioms while still providing good error chain preservation. Standard error wrapping with contextual information provides sufficient debugging capability without over-engineering.

**User Approval**: Option B approved on [2025-09-11]

## [2025-09-11] - BoardAccess Design Decision: Data Storage and File Organization Strategy

**Decision**: Directory Structure with Board Configuration (User-Specified Approach)

**Context**: BoardAccess requires efficient storage of board data, column configuration, and task data with JSON format (REQ-FORMAT-001), version control integration (REQ-INTEGRATION-001), and separate active/archived task organization (REQ-FORMAT-003). Need to optimize for minimal git diffs during common operations like priority changes (REQ-FORMAT-002).

**Options Considered**:

### Option A: Single Task Per File Approach
- **Structure**: Each task stored in separate JSON file (e.g., `tasks/active/task-12345.json`, `tasks/archived/task-12345.json`)
- **File Organization**: 
  ```
  tasks/
  ├── active/
  │   ├── task-12345.json
  │   └── task-67890.json
  └── archived/
      └── task-11111.json
  ```
- **Advantages**:
  - Optimal git diffs: only affected task file changes
  - Easy conflict resolution during merges
  - Simple task archiving (move file between directories)
  - Natural task history per file through VersioningUtility
  - No need for complex JSON manipulation
- **Disadvantages**:
  - More files to manage
  - Bulk queries require multiple file reads
  - Directory operations for task enumeration

### Option B: Priority-Grouped JSON Files
- **Structure**: Tasks grouped by Eisenhower matrix quadrant in separate files
- **File Organization**:
  ```
  tasks/
  ├── active/
  │   ├── urgent-important.json
  │   ├── urgent-not-important.json
  │   ├── not-urgent-important.json
  │   └── not-urgent-not-important.json
  └── archived/
      └── archived-tasks.json
  ```
- **JSON Format**: Array of tasks per priority level
- **Advantages**:
  - Fewer files to manage
  - Fast priority-based queries (single file read)
  - Natural grouping matches domain model
- **Disadvantages**:
  - Large git diffs when moving tasks between priorities
  - Complex JSON array manipulation
  - Potential merge conflicts on same file
  - Archive operations require JSON modification

### Option C: Hybrid Single Index + Individual Files
- **Structure**: Master index file with task metadata + individual task files
- **File Organization**:
  ```
  tasks/
  ├── index.json          # Master index: {id, priority, status, archived}
  ├── active/
  │   ├── task-12345.json
  │   └── task-67890.json
  └── archived/
      └── task-11111.json
  ```
- **Advantages**:
  - Fast bulk queries via index
  - Minimal diffs for priority changes (only index)
  - Individual task history preserved
- **Disadvantages**:
  - Index consistency challenges
  - Two-stage operations (index + task file)
  - Complex recovery if index corrupts

### Option D: Two Aggregate Files (Active + Archived)
- **Structure**: All active tasks in single file, all archived tasks in separate file
- **File Organization**:
  ```
  tasks/
  ├── active-tasks.json    # All active tasks as JSON array/object
  └── archived-tasks.json  # All archived tasks as JSON array/object
  ```
- **JSON Format**: Either array of tasks or object with task IDs as keys
- **Advantages**:
  - Minimal files to manage (only 2 files)
  - Simple bulk operations (single file read/write)
  - Easy backup and synchronization
  - Fast enumeration of all tasks
  - Directly meets REQ-FORMAT-003 (separate active/archived)
- **Disadvantages**:
  - Large git diffs for any task change
  - Potential merge conflicts on same file
  - No individual task history tracking
  - Entire file rewrite for single task changes
  - Poor performance for large task sets
  - Complex JSON manipulation for individual operations
  - File locking issues under high concurrency

### **User-Specified Approach: Directory Structure with Board Configuration**

**Structure**: 
- **Board Configuration**: `board.json` - contains column definitions and Eisenhower sections
- **Active Tasks**: `<column>[/<section>]/task-<id>.json` - tasks organized by column/section directory structure  
- **Archived Tasks**: `archived/task-<id>.json` - archived tasks in dedicated directory

**File Organization**:
```
board.json                           # Board and column configuration
todo/
├── urgent-important/
│   ├── task-12345.json
│   └── task-67890.json
├── urgent-not-important/
│   └── task-11111.json
├── not-urgent-important/
│   └── task-22222.json
└── not-urgent-not-important/
    └── task-33333.json
doing/
├── task-44444.json
└── task-55555.json
done/
├── task-66666.json
└── task-77777.json
archived/
├── task-99999.json
└── task-88888.json
```

**Advantages**:
- **Optimal git diffs**: Moving between sections = file move operation, minimal diff
- **Board configuration centralized**: Column definitions, Eisenhower setup in `board.json`
- **Natural directory queries**: List files in directory for column/section queries
- **Simple archiving**: Move file to `archive/` directory
- **Column context implicit**: Directory structure provides column/section information
- **Clean separation**: Board structure vs task content clearly separated

**Implementation Details**:
- `board.json` contains column definitions, section mappings, workflow rules
- Directory structure mirrors logical board organization
- Task files contain only task-specific data (no column redundancy)
- Archive operations are simple file moves
- Section queries become directory listings

**Rationale**: User-specified approach provides optimal git diff behavior for common operations (priority/column moves), centralizes board configuration, and uses directory structure as natural organizational mechanism.

**User Approval**: **APPROVED** - User specified this exact approach

---

**Design Review Status**: Complete design approved by user on [2025-09-11] - Ready for implementation

## [2025-09-07] - VersioningUtility Design Decision: Per-Repository Mutex Locking

**Decision**: Per-Repository Mutex Locking (Option A)

**Context**: Need thread-safe operations for concurrent access to repositories, as go-git repositories may not be inherently thread-safe.

**Options Considered**:
- Option A: Per-Repository Mutex Locking - Fine-grained locking by path
- Option B: Operation-Level Locking - Coarser locking approach
- Option C: go-git Native Concurrency - Rely on library thread safety

**Rationale**: Per-repository mutex locking provides optimal concurrency by allowing operations on different repositories to proceed simultaneously while protecting individual repositories from concurrent modifications.

**Consequences**:
- Maximum concurrency for multi-repository scenarios
- Path-based mutex map requires memory management
- Deadlock prevention through consistent lock ordering
- Lock cleanup needed for unused repositories
- Thread-safe access to repository operations

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-07] - VersioningUtility Design Decision: Lazy Loading with Limits Plus Streaming

**Decision**: Combined Lazy Loading with Limits and Streaming Results (Hybrid Option A+C)

**Context**: REQ-PERFORMANCE-001 requires 5-second completion for repositories with 10,000 commits, need scalable approach for large repositories.

**Options Considered**:
- Option A: Lazy Loading with Limits - Good for bounded results
- Option B: Caching with Background Updates - Complex state management
- Option C: Streaming Results - Good for large result sets
- Hybrid: Combine A+C for optimal flexibility

**Rationale**: Combining lazy loading with streaming provides both immediate responsiveness for small requests and scalability for large ones. Interface can provide both synchronous (limited) and asynchronous (streaming) access patterns.

**Consequences**:
- Synchronous methods with limits for simple use cases
- Streaming methods for large result sets
- Optimal memory usage for different scenarios
- Dual interface approach requires careful design
- Performance meets requirements under various loads

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-07] - VersioningUtility Design Decision: Selective Staging with Patterns

**Decision**: Selective Staging with Patterns (Option B)

**Context**: REQ-VERSION-003 requires staging "all modifications" but need flexibility for different staging scenarios.

**Options Considered**:
- Option A: Stage All Changes - Simple but inflexible
- Option B: Selective Staging with Patterns - Granular control with patterns
- Option C: Smart Staging with Conflict Detection - Complex logic

**Rationale**: Selective staging with patterns provides flexibility for different use cases while maintaining simplicity. Default pattern can stage all files, but callers can specify patterns for selective staging when needed.

**Consequences**:
- Interface supports both "stage all" and selective staging
- Pattern-based approach familiar to git users
- More flexible than simple stage-all approach
- Requires pattern validation and error handling
- Default behavior stages all changes for simplicity

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-07] - VersioningUtility Design Decision: Direct Error Passthrough with Context

**Decision**: Direct go-git Error Passthrough with Annotations (Option C)

**Context**: Need structured error information (REQ-RELIABILITY-001) while integrating with go-git error handling.

**Options Considered**:
- Option A: Error Wrapping with Context - Custom error structures
- Option B: Error Translation to Domain Errors - Abstract away go-git errors
- Option C: Direct go-git Error Passthrough with Annotations - Preserve original errors with context

**Rationale**: Direct passthrough preserves all go-git error information while adding necessary context. This provides maximum debugging information and maintains compatibility with go-git error handling patterns.

**Consequences**:
- Rich error information preserved from go-git
- Context annotations provide operation and path information
- Callers can handle specific go-git error types if needed
- Error messages include full chain of context
- Maintains compatibility with Go error handling idioms

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-07] - VersioningUtility Design Decision: Repository Handle Pattern

**Decision**: Repository Handle Pattern (Option C)

**Context**: Need to choose repository object management approach for go-git integration, balancing performance and simplicity for 7 interface operations.

**Options Considered**:
- Option A: Repository Instance Per Operation - Simple but high overhead
- Option B: Repository Caching with Lifecycle Management - Complex state management
- Option C: Repository Handle Pattern - Explicit lifecycle control

**Rationale**: Repository Handle Pattern provides optimal performance for multi-operation workflows while giving callers explicit control over repository lifecycle. This aligns with the performance requirements (REQ-PERFORMANCE-001) and supports efficient resource management.

**Consequences**:
- Interface returns handles for multi-operation scenarios
- Callers manage repository lifecycle explicitly
- Better performance for batch operations
- Slight increase in interface complexity
- Requires careful handle cleanup in error scenarios

**User Approval**: [User] approved on [2025-09-07]

## [2025-09-06] - LoggingUtility Design Decision: Interface Design Revision

**Decision**: 3-Operation Interface with Extended Log Method

**Context**: Original SRS suggestion specified 4 operations including separate LogWithStructuredData. User wants to simplify to 3 operations.

**Original Proposed Interface**:
- LogWithStructuredData(level, context, data)
- Log(level, component, message) 
- LogError(level, component, error, context)
- IsLevelEnabled(level)

**Revised Interface Decision**:
- Log(level, component, message, data interface{}) - Extended with data parameter
- LogError(component, error, data interface{}) - Removed level parameter, always logs at Error level
- IsLevelEnabled(level) - Unchanged

**Rationale**: User wants to keep it simple with only 3 operations. Extending Log() with data parameter allows arbitrary data logging without separate method. This removes the need for StructuredLogContext structure.

**Consequences**:
- Simpler interface with fewer methods
- Single Log method handles both simple and structured logging
- No need for separate StructuredLogContext structure
- May need to handle nil data parameter gracefully
- **SRS Update Required**: Interface contract section needs revision

**User Approval**: [User] approved on [2025-09-06]

## [2025-09-06] - LoggingUtility Design Decision: Error Handling Strategy

**Decision**: Panic on Internal Failures

**Context**: Need to determine how to handle internal logging failures (file system errors, configuration issues, etc.).

**Options Considered**:
- Option A: Silent Failure with Internal Error Tracking
- Option B: Best-Effort with Fallback
- Option C: Error Return with Graceful Degradation
- Option D: Panic on Internal Failures

**Rationale**: User decision to use fail-fast behavior. Panicking on internal failures provides clear failure indication and simpler implementation. Logging failures are typically configuration/environment issues that should be addressed immediately. This is in accordance to **REQ-RELIABILITY-001**: If log output fails, then the LoggingUtility shall crash the application.

**Consequences**:
- Simpler error handling implementation
- Clear failure signals for debugging
- Callers must handle potential panics or fix logging configuration
- Removes requirement for complex fallback logic

**User Approval**: [User] approved on [2025-09-06]

## [2025-09-06] - LoggingUtility Design Decision: Circular Reference Handling

**Decision**: Depth Limiting Only (Option B)

**Context**: STP requires handling of self-referencing structures without infinite loops. SRS specifies 5-level depth limit (REQ-FORMAT-003).

**Options Considered**:
- Option A: Visited Map with Pointer Tracking
- Option B: Depth Limiting Only
- Option C: Visited Map + Depth Limiting

**Rationale**: Choose Option B to keep implementation simple and directly meet the "5 levels" requirement from SRS. Depth limiting prevents infinite loops effectively while maintaining simple implementation.

**Consequences**:
- Simple implementation without pointer tracking overhead
- Directly satisfies SRS depth requirement
- May not detect circular references at shallow depths (acceptable trade-off)
- Deterministic truncation behavior

**User Approval**: [User] approved on [2025-09-06]

## [2025-09-06] - LoggingUtility Design Decision: Structured Data Serialization Strategy

**Decision**: Type Switch + Interface-based Approach (Option B)

**Context**: The SRS requires support for arbitrary Go types with preserved type information and human-readable narratives. Need to choose between reflection-based, type-switch-based, or hybrid approaches.

**Options Considered**: 
- Option A: Reflection-based JSON + Custom Formatting
- Option B: Type Switch + Interface-based Approach  
- Option C: Hybrid Approach

**Rationale**: Choose Option B to keep implementation simple as a starting point. Type switches provide better performance for common types, and interface-based approach allows extensibility. Fallback to fmt.Sprintf ensures all types are handled, even if not optimally.

**Consequences**: 
- Better performance for common logging scenarios
- Simpler implementation and maintenance
- May require future enhancement for complex edge cases
- Interface adoption needed for optimal custom type logging

**User Approval**: [User] approved on [2025-09-06]