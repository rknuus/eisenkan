# iDesign Project Template Sync

A collection of tools to keep your iDesign-based projects synchronized with the latest template updates and best practices.

## Overview

This repository provides utilities to maintain consistency across iDesign projects by synchronizing common files and configurations from a central template
repository. The tools help ensure that architectural guidelines, development practices, and documentation standards remain up-to-date across all your iDesign
projects.

## Tools

### `update-claude-md.sh`

Updates the `CLAUDE.md` file in your project with the latest version from the iDesign project template repository.

Note: you have to create your own https://github.com/your-org/idesign_project_template_sync.git repo.

**Features:**
- ✅ Automatic backup creation before updates
- ✅ GitHub CLI integration for secure authentication
- ✅ Diff display showing changes made
- ✅ Interactive cleanup options
- ✅ Comprehensive error handling

**Prerequisites:**
- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated
- Git repository (run from project root)

**Usage:**
```bash
# First time installation via Git Subtree
git subtree add --prefix=3rd-party/sync https://github.com/your-org/idesign_project_template_sync.git main --squash

# Update to latest
git subtree pull --prefix=3rd-party/sync https://github.com/your-org/idesign_project_template_sync.git main --squash

# First time setup (authenticate with GitHub)
gh auth login --web

# Update CLAUDE.md in your project
./3rd-party/sync/scripts/update-claude-md.sh your-org idesign_project_template main

# Update with custom file paths
./3rd-party/sync/scripts/update-claude-md.sh your-org idesign_project_template main docs/CLAUDE.md CLAUDE.md

# Show help
./3rd-party/sync/scripts/update-claude-md.sh --help
```

## Repository Structure

```
├── README.md                 # This file
├── scripts/                  # iDesign Project Template support scripts
└── LICENSE                   # MIT License
```

## Configuration

The `update-claude-md.sh` script accepts the following parameters:

**Required Parameters:**
- `repo_user` - GitHub repository owner (e.g., "your-org")
- `repo_name` - GitHub repository name (e.g., "idesign_project_template")
- `branch` - Branch to sync from (e.g., "main")

**Optional Parameters:**
- `file_path` - Path to file in source repo (default: "CLAUDE.md")
- `local_file` - Local destination file (default: "CLAUDE.md")

**Examples:**
```bash
# Basic usage with defaults
./scripts/update-claude-md.sh your-org idesign_project_template main

# Custom file paths
./scripts/update-claude-md.sh your-org my_template develop docs/CLAUDE.md CLAUDE.md

# Different branch
./scripts/update-claude-md.sh your-org idesign_project_template feature-branch
```

**Integration with Makefile:**
For easier management, projects can define these parameters in their Makefile:
```makefile
TEMPLATE_USER := your-org
TEMPLATE_REPO := idesign_project_template  
TEMPLATE_BRANCH := main

sync-claude:
	./3rd-party/sync/scripts/update-claude-md.sh $(TEMPLATE_USER) $(TEMPLATE_REPO) $(TEMPLATE_BRANCH)
```

## Development Guidelines

- Follow shell scripting best practices (set -euo pipefail)
- Provide comprehensive error handling and user feedback
- Include usage examples and help text
- Test with both public and private repositories
- Maintain backward compatibility when possible

## License

GPLv3 License - see [LICENSE](LICENSE) file for details.

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License version 3 as published by the Free Software Foundation.