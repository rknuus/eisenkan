# EisenKan project plan
## Tasks
### Single Manager (Orchestrates all engines)
- Implement WorkflowManager - Depends on All
Engines + ResourceAccess

### UI Components (Use manager + engines directly)
- Implement TaskWidget - Depends on
WorkflowManager + FormattingEngine
- Implement ColumnWidget - Depends on
WorkflowManager + DragDropEngine +
LayoutEngine + FyneUtility
- Implement NewStoryFormArea - Depends on
WorkflowManager + FormValidationEngine
- Implement EisenhowerMatrixDialog - Depends on
TaskWidget + FormValidationEngine + LayoutEngine
- Implement SubtaskExpansionView - Depends on
TaskWidget + LayoutEngine
- Implement TaskFormView - Depends on
EisenhowerMatrixDialog + WorkflowManager +
FormattingEngine
- Implement EntryView - Depends on
WorkflowManager + FormattingEngine
- Implement BoardView - Depends on
WorkflowManager + ColumnWidget + TaskWidget +
SubtaskExpansionView + FormValidationEngine

### Application
- Implement Application Root
- Implement service EisenKan Admin Client