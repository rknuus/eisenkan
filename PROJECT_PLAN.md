# EisenKan project plan
## Tasks
### Single Manager (Orchestrates all engines)
- Implement TaskWorkflowManager - Depends on All
Engines + ResourceAccess

### UI Components (Use manager + engines directly)
- Implement TaskWidget - Depends on
TaskWorkflowManager + FormattingEngine
- Implement ColumnWidget - Depends on
TaskWorkflowManager + DragDropEngine +
LayoutEngine + FyneUtility
- Implement NewStoryFormArea - Depends on
TaskWorkflowManager + FormValidationEngine
- Implement EisenhowerMatrixDialog - Depends on
TaskWidget + FormValidationEngine + LayoutEngine
- Implement SubtaskExpansionView - Depends on
TaskWidget + LayoutEngine
- Implement TaskFormView - Depends on
EisenhowerMatrixDialog + TaskWorkflowManager +
FormattingEngine
- Implement EntryView - Depends on
TaskWorkflowManager + FormattingEngine
- Implement BoardView - Depends on
TaskWorkflowManager + ColumnWidget + TaskWidget +
SubtaskExpansionView + FormValidationEngine

### Application
- Implement Application Root
- Implement service EisenKan Admin Client