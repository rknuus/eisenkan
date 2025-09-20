# EisenKan TODO List

## Cross-platform support for Recently Used Documents
- **Status**: TODO
- **Priority**: Medium
- **Description**: Extend BoardSelectionView to support "recently used" board persistence on Linux and Windows platforms
- **Current Status**: Only macOS support implemented using NSDocumentController recent documents
- **Required Work**:
  - Linux: Implement using XDG Recent Files specification (`~/.local/share/recently-used.xbel`)
  - Windows: Implement using Windows Registry recent documents or Jump List API
  - Create abstraction layer for cross-platform recent documents management
- **Dependencies**: BoardSelectionView implementation completion

## EisenKan settings
- **Status**: TODO
- **Priority**: Medium
- **Description**: Support customization of EisenKan settings like the number of entries in the recent list etc.
- **Current Status**: Not supported
- **Required Work**:
  - Storage of configuration in OS-specific application data directory
  - Change recent board limit
  - Change keyboard shortcuts
  - Change board theme and styling
  - Optionally enable screen reader and keyboard navigation
- **Dependencies**: UX improvements

## Filter board repos
- **Status**: TODO
- **Priority**: Low
- **Description**: When browsing for board repos, optionally support listing only valid board repos
- **Current Status**: All subdirectories are listed
- **Required Work**:
  - Extend OS specific browsing dialogs by filters "boards only" and "all"
  - The challenge: filtering depends on directory content, not on file types
- **Dependencies**: BoardSelectionView UX improvement
