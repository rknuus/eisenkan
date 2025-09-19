# EisenKan project plan
## Tasks
### UI Components (Use manager + engines directly)
- Implement CreateTaskDialog - Depends on
TaskWidget + FormValidationEngine + LayoutEngine
- Implement SubtaskExpansionView - Depends on
TaskWidget + LayoutEngine
- Implement TaskFormView - Depends on
CreateTaskDialog + ColumnWidget + TaskWidget + WorkflowManager +
FormattingEngine
- Implement EntryView - Depends on
WorkflowManager + FormattingEngine
- Implement BoardView - Depends on
WorkflowManager + ColumnWidget + TaskWidget +
SubtaskExpansionView + FormValidationEngine

### Application
- Implement Application Root
- Implement service EisenKan Admin Client