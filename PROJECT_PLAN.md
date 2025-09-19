# EisenKan project plan
## Tasks
### UI Components (Use manager + engines directly)
- Implement ColumnWidget - Depends on
WorkflowManager + DragDropEngine +
LayoutEngine + FyneUtility
- Implement NewStoryWidget - Depends on
WorkflowManager + FormValidationEngine
- Implement EisenhowerMatrixDialog - Depends on
TaskWidget + NewStoryWidget + FormValidationEngine + LayoutEngine
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