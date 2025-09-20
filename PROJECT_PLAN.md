# EisenKan project plan
## Tasks
### UI Components (Use manager + engines directly)
- Implement BoardView - Depends on
WorkflowManager + ColumnWidget + TaskWidget + FormValidationEngine
- Implement SubtaskExpansionView - Depends on
TaskWidget + LayoutEngine + BoardView (for column patterns)
- Implement TaskFormView - Depends on
CreateTaskDialog + ColumnWidget + TaskWidget + WorkflowManager +
FormattingEngine

### Application
- Implement Application Root - Depends on BoardSelectionView + BoardView
- Implement service EisenKan Admin Client
