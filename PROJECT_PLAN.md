# EisenKan project plan
## Tasks
### UI Components (Use manager + engines directly)
- Implement EntryView - Depends on
WorkflowManager + FormattingEngine
- Implement SubtaskExpansionView - Depends on
TaskWidget + LayoutEngine + BoardView (for column patterns)
- Implement TaskFormView - Depends on
CreateTaskDialog + ColumnWidget + TaskWidget + WorkflowManager +
FormattingEngine

### Application
- Implement Application Root
- Implement service EisenKan Admin Client