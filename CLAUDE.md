# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Glossary
- SRS: Software Requirements Specifications
- STP: Software Test Plan

## Claude Guidelines

Please take the [README.md](README.md) into account, which contains public information relevant for human developers and for ML models/agents. The CLAUDE.md file is not shared, because it contains information I want to keep private.

Optionally, directory `.claude.d` contains project specific Claude instruction files, while this CLAUDE.md is generic and shared by multiple projects.

## Service Lifecycle Process

```
                      ┌───────────────────┐
                      │ Establish Context │
                      └─────────┬─────────┘
                                │
                                ▼
                      ┌───────────────────┐
                      │   Context Review  │
                      └─────────┬─────────┘
                                │
                                ▼
                             ┌─────┐
                             │ SRS │
                             └──┬──┘
                                │
                                ▼
                         ┌────────────┐
                         │ SRS Review │──────────────────────────┐
                         └──────┬─────┘                          │
                                │                                │
                                ▼                                ▼
                       ┌──────────────────┐                   ┌─────┐
                       │ Optionally:      │                   │ STP │
                       │ Design Prototype │                   └──┬──┘
                       └────────┬─────────┘                      │
                                │                                ▼
                                ▼                          ┌────────────┐
                        ┌───────────────┐                  │ STP Review │
                        │ Detailed      │                  └─────┬──────┘
                        │ Design        │                        │
                        └───────┬───────┘                        │
                                │                                │
                                ▼                                │
                        ┌───────────────┐                        │
                        │ Design Review │                        │
                        └───────┬───────┘                        │
                                │                                │
                                ▼                                │
      ┌─────────────────────────────────────────────────────┐    │
      │  ┌──────────────┐      ┌─────────────────────────┐  │    │
      │  │ Construction │─────▶│ Constructive Testing    │  │    │
      │  │              │◀─────│ of Internals            │  │    │
      │  └──────────────┘      └─────────────────────────┘  │    │
      └─────────────────────────┬───────────────────────────┘    │
                                │                                │
                                ▼                                │
                        ┌─────────────┐                          │
                        │ Code Review │                          │
                        └───────┬─────┘                          │
                                │                                │
                                ▼                                │
                      ┌──────────────────┐                       │
                      │ Integration      │                       │
                      │ Testing          │                       │
                      └─────────┬────────┘                       │
                                │                                │
                                ▼                                │
                   ┌────────────────────────┐                    │
                   │ Acceptance test demo   │◀───────────────────┘
                   │                        │
                   └────────┬───────────────┘
                            │
                            ▼
                   ┌────────────────────────┐
                   │ STR Report Creation    │
                   │                        │
                   └────────────────────────┘
```

### Mandatory Requirements

**CRITICAL**: Each step below MUST be followed exactly. No step may be skipped without explicit user approval.

#### Phase Requirements
1. **Establish Context**: Brainstorm possible functionality of the service
2. **Context Review**: Present to user, wait for explicit "Context approved"
3. **SRS Creation**: Based on feedback about Context create `doc/[services|client]/<ServiceName>_SRS.md` with EARS requirements with similar structure like other SRS documents, if existing, keeping requirements and interface contracts technology-agnostic except the requirements in section "Technical Constraints"
4. **SRS Review**: Present to user, wait for explicit "SRS approved"
5. **STP Creation**: Create `doc/[services|client]/<ServiceName>_STP.md` with destructive testing focus and similar structure like other STP documents, if existing, avoid redundancy between test cases in the same document
6. **STP Review**: Present to user, wait for explicit "STP approved"
7. **Design Prototype**: If desired by the user, interactively create design prototypes to explore the design space until the user is satisfied
8. **Detailed Design**: Present design options with enumeration, wait for user input on decisions and document in `doc/DDR.md`
9. **Design Review**: Present complete design, wait for "Design approved"
10. **Construction & Code Review**: Involve user in code review, wait for approval
11. **Integration Testing**: Required for components basing on other components, skip for leaf-compontents
12. **Acceptance Testing**: Demonstrate all STP tests pass, obtain user acceptance
13. **STR Creation**: Create `doc/[services|client]/<ServiceName>_STR.md` with actual test execution results and requirements verification

#### Interaction Requirements
- **MUST** ask explicit questions: "Please review this [SRS/STP/Design Prototype/Design]. Do you approve proceeding?"
- **MUST** wait for explicit approval: "approved", "proceed", "looks good"
- **MUST NOT** assume approval from silence or general comments
- **MUST** involve user as decision-maker, not just reviewer
- **MUST** re-scan files under review when user reports findings or approves before proceeding
- **SHOULD** update `CLAUDE.md` when user findings reveal preventable issues

#### Acceptance test demo
- **MUST** demonstrate all tests listed in the STP after implementation
- **Test Demonstration Methods** (in order of precedence):
  1. **Automated Tests**: Running actual test suites with passing results
  2. **Log Outputs**: Showing detailed execution logs that prove test coverage
  3. **Code Walkthrough**: Explaining test implementation and expected behavior
- **MUST** verify 100% STP test coverage through demonstration
- **MUST** document demonstration results in the STR document
- **MUST** update Requirements Verification Matrix in STR with actual test function names and results
- **MUST** obtain user acceptance after successful demonstration
- **MUST** mark STR document as "Accepted" with final status upon successful completion

## Documentation Standards
### Documentation Requirements
- **MUST** when creating a document, add current date as creation date to the footer and set version to 1.0
- **MUST** when changing a document (not when creating one), add current date as updated date to the footer and increase the version to the next minor or major version depending on the severity of the change
- **MUST** after approval of a document, the item IDs (e.g. requirement IDs or test IDs) must remain stable
- **MUST** after approval of a document, IDs of removed items must be listed as reseved with a comment "former <item>"

### File Locations and Formats
| Document Type | Location | Required Sections |
|---|---|---|
| SRS | `doc/[services|client]/<ServiceName>_SRS.md` | Purpose, Operations, Quality Attributes, Interface Requirements |
| STP | `doc/[services|client]/<ServiceName>_STP.md` | Test Strategy, Destructive Testing (NO Requirements Verification Matrix) |
| STR | `doc/[services|client]/<ServiceName>_STR.md` | Requirements Verification Matrix, Test Execution Results, Acceptance Status |
| Detail Design | `doc/DDR.md` | Decision, Context, Options, Rationale, User Approval ordered descending by date |
| Architecture | `doc/ADR.md` | Same format as DDR |

### EARS Requirements Syntax
All requirements must follow EARS format:
- **Template**: `<optional preconditions> <optional trigger> the <system name> shall <system response>`
- **Keywords**: Ubiquitous, Event-driven ("When"), Unwanted ("If...then"), State-driven ("While"), Optional ("Where")
- **Example**: `When a component calls LogError, the LoggingUtility shall capture stack trace information`

## Destructive Testing Requirements
STP must include:
- **API Contract Testing**: Invalid inputs, boundary values, null parameters
- **Resource Exhaustion**: Memory, disk, file handles, concurrent limits
- **Error Conditions**: External failures, corruption, configuration errors
- **Requirements Verification**: Each EARS requirement mapped to test cases
- **Graceful Degradation**: System behavior under adverse conditions

## Usage of 3rd party components
- **Avoid Duplication**: When using 3rd party components like libraries and frameworks, do not duplicate 3rd party component functionality in services

## Architectural Rules and Guidelines

### The Prime Directive
**Never design against the requirements.**

### Critical Directives
1. Avoid functional decomposition
2. Decompose based on volatility
3. Provide a composable design
4. Offer features as aspects of integration, not implementation
5. Design iteratively, build incrementally

### System Design Guidelines

#### Requirements Management
- Capture required behavior, not required functionality
- Describe required behavior with operations
- Document complex operations with activity diagrams
- Eliminate solutions masquerading as requirements
- Validate design supports all core operations

#### Component Architecture Rules

| Component Type | Cardinality Limits | Volatility | Reuse |
|---|---|---|---|
| Managers | ≤5 per system, ≤3 per subsystem | Decreases top-down | Increases top-down |
| Subsystems | Few per system | Medium | Medium |
| Engines | Golden ratio to Managers | Low | High |
| ResourceAccess | Can access multiple Resources | Low | High |

#### Layer Interaction Rules

| Interaction Type | Allowed | Forbidden |
|---|---|---|
| **Downward Calls** | All→Utilities, Managers/Engines→ResourceAccess, Managers→Engines | Calling >1 layer down |
| **Upward Calls** | None | All upward calls |
| **Sideways Calls** | Queued Manager→Manager | Direct sideways calls |
| **Event Publishing** | Managers only | Clients, Engines, ResourceAccess, Resources |
| **Event Subscription** | Managers, Clients | Engines, ResourceAccess, Resources |

#### Architecture Preferences
- **Symmetric Design**: Components should have balanced responsibilities
- **Manager Expendability**: Managers should be easily replaceable
- **Public Channels**: Never use for internal system interactions

### Service Contract Design
- **Operations per Contract**: 3-5 ideal, avoid 1 or >12, reject >20
- **Contract Limits**: 1-2 contracts per service
- **Avoid**: Property-like operations