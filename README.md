# README.md

EisenKan is a Kanban board application with a twist: when creating new tasks you have to assign them to a quadrant of the Eisenhower matrix. The quadrant decides, where in the `todo` column the new task appears.

## Project Overview

EisenKan is a Kanban board application written in Go. The project uses Protocol Buffers for API definitions and follows a standard Go project layout.

This project follows the **iDesign Method** (volatility-based design) for creating robust, adaptable software architectures that can evolve with changing requirements.

### Getting Started with iDesign
If you're new to the iDesign method:
1. Read about volatility-based decomposition principles (see [#References])
2. Understand the five-layer architecture pattern
3. Study the use case validation approach
4. Practice identifying volatility areas in requirements
5. Learn to avoid functional decomposition anti-patterns

### Core iDesign Principles
- **No Functional Decomposition**: Never map requirements directly to services/modules
- **Volatility-Based Decomposition**: Organize system around areas of change, not functional requirements
- **Encapsulation of Change**: Each volatile area should be isolated in its own architectural component
- **Components are Services**: Each component is a service and hence the architecture is service-based
- **Use Case Driven**: Focus on business use cases, not technical functions

## Domain overview

### Domain Context
See [SAD](./doc/SAD.md)

### Technology Constraints
See [SAD](./doc/SAD.md)

## Architecture Overview

### System Decomposition (The Method Components)
Our architecture follows the standard iDesign layered approach:

**Clients Layer**
- User interfaces, external system interfaces
- Presentation logic only
- No business logic

**Managers Layer** 
- Business workflow orchestration
- Use case implementation
- Stateless coordination between engines

**Engines Layer**
- Pure business logic
- Domain-specific calculations and rules
- Stateless and side-effect free

**Resource Access Layer**
- Encapsulation of resources
- Provides API in form of "atomic business verbs"
- Resource management

**Resource Layer**
- Data persistency
- Filesystem
- External services

**Utilities Layer**
- Cross-cutting concerns (logging, security, etc.)
- Reusable infrastructure components

### Volatility Areas Identified
See [SAD](./doc/SAD.md)

### Core Use Cases
See [SAD](./doc/SAD.md)

### Architecture Validation
See [SAD](./doc/SAD.md)

## Development Guidelines

### Approach
- Follow volatility-based thinking in all design decisions
- Validate new features against core use cases
- Maintain architectural documentation in `/doc/`
- Update this CLAUDE.md when architectural decisions change

### Code Organization
Follow iDesign namespace conventions:
```
<Company>.<Concept>.[<Product>].<Subsystem>
```

**Concepts:**
- `Manager` - Business workflow components  
- `Engine` - Business logic components
- `DataAccess` - Data persistence components
- `Proxy` - External service adapters
- `Test` - All test implementations

**Standard Subsystems:**
- `Common` - Shared components across subsystems
- `Utilities` - Cross-cutting infrastructure
- `iFX` - Framework and hosting infrastructure

### Service Design
- **Managers**: Focus on workflow orchestration, delegate to engines
- **Engines**: Implement pure business logic, no side effects
- **Data Access**: Encapsulate resources

### Error Handling
- Validate inputs at service boundaries
- Use consistent error propagation patterns
- Log at appropriate architectural layers

### Design Validation Rules
When implementing features, always validate against:

1. **Use Case Coverage**: Does this support our core use cases?
2. **Volatility Isolation**: Is each change area properly encapsulated?
3. **Layer Responsibility**: Is each component in the correct architectural layer?
4. **Dependency Direction**: Do dependencies flow downward through layers?
5. **Statelessness**: Are managers and engines stateless?

### Anti-Patterns to Avoid
- **Functional Decomposition**: Don't create services that mirror requirements
- **Client Orchestration**: Clients should not coordinate multiple managers
- **Cross-Layer Dependencies**: Don't skip layers or create circular dependencies
- **Stateful Engines**: Business logic should be pure and stateless
- **Manager-to-Manager Calls**: Managers should not call other managers directly

## Project Structure

- `cmd/eisenkan/` - Main application entry point
- `api/` - Protocol Buffer definitions for the Kanban API
  - `kanban.proto` - Core Kanban data models (Board, Column, Card, etc.)
  - `tasks-repo.proto` - Task repository service definitions
- `doc/` - Documentation including Software Architecture Document (SAD.md)
- `bin/` - Build artifacts directory

### Documentation Structure
```
/doc/
├── SAD.md                      # Software Architecture Document
├── ADR.md                      # Architectural Decision Records
├── DDD.md                      # Detail Design Document
├── DDR.md                      # Design Decision Records
```

## Development Commands

### Building
```bash
make build                    # Build the application
make build VERSION=1.0.0     # Build with specific version
```

### Running
```bash
make run                      # Build and run the application
./bin/eisenkan               # Run built binary directly
```

### Testing
```bash
make test                     # Run tests (currently untested per FIXME comment)
go test ./...                # Run tests directly
```

### Other Commands
```bash
make clean                    # Remove build artifacts
make install                  # Install to $GOBIN
```

### Git Subtree Synchronization

This project uses git subtrees to integrate external tooling and templates. The subtrees are managed through Makefile targets for consistency.

#### Setup Git Subtrees (One-time)
```bash
make sync-setup               # Add iDesign project template sync tools
```

This command adds the [idesign-project-template-sync](https://github.com/rknuus/idesign_project_template_sync) repository as a subtree in `3rd-party/sync/`. This repo supports to keep the files stemming from idesign-project-template in sync. Note that the idesign-project-template repo is private and that `CLAUDE.md` is not under version-control for security reasons.

#### Update Git Subtrees
```bash
make sync-update              # Pull latest changes from sync tools repository
```

Use this command periodically to update the sync tools with the latest improvements and new features.

#### Update CLAUDE.md from Template
```bash
make sync-claude              # Update CLAUDE.md from template repository (includes sync-update)
```

This command:
1. Updates the sync tools to the latest version
2. Downloads the latest CLAUDE.md from the template repository  
3. Creates a backup of your current CLAUDE.md
4. Shows differences between old and new versions

## Dependencies

- Go 1.24.3
- Local dependency: `github.com/rknuus/versionaccess` (replaced with `../versionaccess`)

## Architecture Notes

- API is defined using Protocol Buffers with comprehensive Kanban domain models
- The workspace includes both the main project and a separate versionaccess dependency
- Version information is injected at build time via ldflags

## Claude Code Integration Guidelines

### When changing code
1. **Identify Volatility**: Determine which volatility area the change affects
2. **Validate Use Cases**: Ensure the change supports core use cases
3. **Choose Layer**: Place components in appropriate architectural layers
4. **Design Contracts**: Define stable interfaces first
5. **Implement Bottom-Up**: Start with engines, then managers, then clients

### Code Review Focus Areas
- Architectural layer compliance
- Volatility encapsulation effectiveness
- Use case coverage and validation
- Interface design and stability
- Dependency direction enforcement

### Refactoring Guidelines
- Always refactor within volatility boundaries first
- Validate use case support before and after changes
- Maintain interface stability during refactoring
- Use architecture tests to prevent regression

## References
- [SAD](./doc/SAD.md)
- [Righting Software](https://www.amazon.com/Righting-Software-Juval-L%C3%B6wy/dp/0136524036) by Juwal Löwy
- [iDesign web casts](https://idesign.net/Resources/Webcasts)

---

*This project follows the iDesign Method as developed by Juval Löwy and IDesign Inc. For more information, visit [IDesign.net](https://www.idesign.net/) or read "Righting Software" by Juval Löwy.*