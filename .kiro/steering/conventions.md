# Development Guidelines for AI Agents

This document provides comprehensive development guidelines for AI agents working on the SysPkg project, consolidating guidance from project-specific and universal development principles.

## 🚨 Critical Safety Rules

### Docker Safety (MANDATORY)
- **ALWAYS use Docker** for package manager operations
- **NEVER run package operations on development system**
- Use `make test-docker-*` for testing package managers
- All fixtures are generated in Docker containers for safety
- Integration tests MUST run in containers, never on host system

### Fixture Integrity (SACRED)
- **NEVER modify fixture files** in `testing/fixtures/`
- Fixtures contain authentic package manager outputs
- They are excluded from formatters to preserve exact output
- Use fixtures as primary test data, not inline mocks
- Generate new fixtures using Docker entrypoints when needed

### Security Requirements (MANDATORY)
- **Input validation is mandatory** before any command execution
- Package names MUST be validated using `manager/security.go` helpers
- Command injection is a real and present danger
- Use parameterized command construction, never shell interpretation
- All user inputs must be sanitized before reaching system commands

## 🏗️ Architecture Patterns

### Plugin Development
- **Embed BaseManager** for new package manager plugins
- Implement only **2-3 required methods** per plugin (IsAvailable, Search, specific operations)
- Use **CommandRunner interface** for all command execution (enables testing)
- Follow established plugin registration pattern with `init()` functions
- Inherit common functionality from BaseManager (90% of needs covered)

### Interface Compliance
- All package managers MUST implement the unified `PackageManager` interface
- Support all 11 operations: Search, List, Install, Remove, GetInfo, Refresh, Upgrade, Clean, AutoRemove, Verify, Status
- Return `ErrOperationNotSupported` for unsupported operations, don't panic
- Use consistent `PackageInfo` structure with flexible `Metadata` field

### CommandRunner Pattern (REQUIRED)
- All package managers must use CommandRunner interface
- Enables testing without system calls via MockCommandRunner
- See APT/YUM implementations as reference patterns
- Proper context handling for timeouts and cancellation

## 📋 Code Style & Conventions

### Go Standards
- Follow Go standard formatting (`gofmt`, `goimports`)
- Use meaningful variable names, avoid abbreviations
- Write comprehensive comments for public APIs
- Follow Go naming conventions (PascalCase for exported, camelCase for private)
- Use `goimports` with local imports: `github.com/bluet/syspkg`

### Project-Specific Patterns
- Package manager categories: `system`, `language`, `app`, `container`, `game`, etc.
- Status constants: `StatusInstalled`, `StatusAvailable`, `StatusUpgradable`, `StatusUnknown`
- Error handling: Use `ReturnStatus` enumeration for consistent exit codes
- Concurrent operations: Use `*AllConcurrent` methods for multi-manager operations

### File Organization
- Plugin structure: `manager/[manager-name]/plugin.go`
- Tests: `plugin_test.go` in same directory
- Utilities: `utils.go` for parsing functions
- Exit codes: `EXIT_CODES.md` documenting actual behavior (not assumptions)

## 🧪 Testing Requirements

### Testing Hierarchy (3-Layer)
1. **Unit Tests**: Use authentic fixtures (safe, fast, realistic)
2. **Integration Tests**: Run in Docker containers (safe system testing)
3. **System Tests**: Only in CI or dedicated environments

### Fixture-Based Testing (PRIMARY)
- Use real fixture files from `testing/fixtures/` in unit tests
- Fixtures contain full raw outputs from actual package managers
- Inline mocks only for quick tests and edge cases
- Follow naming convention: `{operation}.{mode}.{state}.{distro}-{version}.txt`

### Test Coverage Standards
- All new plugins need **comprehensive test coverage** (20+ test functions)
- Test all 11 PackageManager interface operations
- Include error condition testing and edge cases
- Security testing for all injection prevention patterns
- Performance testing for concurrent operations

### Docker Testing Requirements
- Use Docker for integration testing, fixture generation, cross-PM testing
- Never run package installation/removal on development system
- Snap testing limitation: Requires native runners (doesn't work in Docker)
- Use `make test-docker-ubuntu`, `make test-docker-rocky`, etc.

### Docker Fixture Generation (CRITICAL)
- **Always use Docker** for generating test fixtures safely
- **Use entrypoint scripts** for comprehensive fixture generation
- **Monitor Docker operations** - check `docker ps` and `docker logs` for long-running containers
- **Fix file ownership** - always `chown 1000:1000` fixtures in Docker containers
- **Never generate fixtures on host system** - package operations must be containerized

## 🔒 Security Guidelines

### Input Validation (MANDATORY)
- Always validate package names with regex patterns before execution
- Use `ValidatePackageNames()` function from `manager/security.go`
- Test all malicious input patterns (command injection, path traversal)
- Never execute shell commands with unvalidated user input

### Command Construction
- Use parameterized command building, never string concatenation
- Prevent shell interpretation of user input
- Implement timeout enforcement using `context.WithTimeout`
- No privilege escalation attempts - clear error messages when sudo needed

### Security Testing Patterns
- Test 18+ malicious input patterns for each package manager
- Validate regex patterns prevent injection attacks
- Ensure error messages don't leak sensitive information
- Test timeout and cancellation behavior

## 🚀 Development Workflow

### Branch Strategy
- Feature branches from main branch
- Small, focused commits with clear messages

### Essential Commands
```bash
make test              # Smart OS-aware testing (always run first)
make check             # Code quality (run before commits)
make test-docker-*     # Package manager testing in containers
make build             # Build for current platform
make format            # Fix formatting issues
```

### Docker vs Native Testing
- **Native (`make test`)**: Parser tests, core logic, daily development
- **Docker (REQUIRED)**: Integration tests, fixture generation, package operations
- **Never native**: Package installation/removal operations

### Commit Guidelines
- Prefix: `fix:`, `feat:`, `docs:`, `test:`, `refactor:`
- Scope: `(apt)`, `(yum)`, `(core)`, etc.
- Example: `fix(apt): correct exit code handling for package not found`
- Always run `make check` before committing

## 📚 Documentation Standards

### Code Documentation
- Write clear Go documentation for all public interfaces
- Include usage examples in documentation comments
- Document error conditions and return values
- Explain complex parsing logic with comments

### Project Documentation
- Keep specs focused and avoid duplication
- Reference existing docs rather than copying content
- Update documentation as implementation progresses
- Verify examples actually compile and work

### Documentation Verification Protocol
Before claiming documentation consistency:
1. Run verification commands and show output
2. Report: "Commands: [X], Files: [Y], Issues: [Z]"
3. Test that examples actually work
4. Never say "I checked all docs" - show specific verification results

## ⚠️ Common Pitfalls to Avoid

### Interface Mistakes
- `GetType()` doesn't exist - use `GetName()` or `GetCategory()`
- `NewPackageInfo(..., CategorySystem)` - 4th param is manager name like "apt"
- Interface definitions must match `manager/interfaces.go` exactly

### Testing Anti-Patterns
- Don't modify fixture files to make tests pass
- Don't run package operations on development system
- Don't assume exit codes work the same across package managers
- Don't skip security validation tests

### Architecture Violations
- Don't bypass CommandRunner interface
- Don't implement package managers without BaseManager
- Don't ignore the unified interface requirements
- Don't assume package manager availability without checking

## 🎯 Project Philosophy
- **Tool-focused, not OS-focused**: If apt works in a container or on macOS, we support it
- **Unified interface**: 11 package operations work consistently across all managers
- **Safety first**: All destructive operations run in Docker containers
- **Authentic testing**: Use real package manager outputs as test data

## 📖 Quick References

- **Architecture**: `docs/ARCHITECTURE.md`
- **Testing Guide**: `docs/TESTING.md`
- **Exit Codes**: `docs/EXIT_CODES.md`
- **Contributing**: `CONTRIBUTING.md`
- **Plugin Development**: `docs/PLUGIN_DEVELOPMENT.md`

## 🧠 Investigation Protocol

### Before Making Changes
1. **Read existing code** to understand what it ACTUALLY does
2. **Use grep/search tools** to verify assumptions
3. **Understand the context** and execution paths
4. **Test your understanding** before implementing

### Verification Steps
- State assumptions explicitly and verify with tools
- Read actual error messages, don't interpret
- Check what tests ACTUALLY test by reading exact lines
- Trace code execution paths before proposing solutions

### Red Flags Requiring Extra Caution
- Questions about your reasoning → Re-investigate
- Unexpected test results → Investigate why
- Modifying tests → Verify what they actually do
- CI inconsistencies → Check actual configuration

## 🎯 Work Management & Quality Assurance

### Core Development Principles
- **Think Carefully, Plan in Detail, Design the Best Solution, and Execute Precisely**
- **Don't pretend you've done something** - You must do it exactly, carefully, and well
- **Don't assume/pretend** - Provide actual facts and results
- **Always review your own work** like an experienced domain expert for solution correctness and code quality

### Session Management & Continuity
- **Never skip current job** and pretend it's done after session compacting
- **Always recheck/continue** work after compacting to ensure completion
- **Update todos frequently** with detailed context to keep tracking easier
- **Write down multiple proposals** so you can retrieve knowledge later when showing options
- **Keep detailed investigation notes** - knowledge might be lost after session compacting

### Work Completion Standards
- **Show summary of what you did** and results at the end after finishing all assigned work
- **Reveal critical details/information** to help with review
- **Complete all assigned tasks** (excluding what was explicitly deferred)
- **Track and fix identified issues** - don't just investigate without resolving

### File Operations
- **Always read existing files** before writing to preserve important information
- **Use Edit instead of Write** when updating files to avoid data loss
- **Check official docs or search online** if stuck on problems

### Docker Operations & Monitoring
- **Use Docker for safety** for full tests, testing package manager operations, and generating fixtures
- **Always chown fixtures in Docker** to ensure proper file ownership
- **Monitor long-running Docker operations** - check `docker ps` and `docker logs`
- **Generate fixtures safely** using Docker entrypoints with proper file permissions

### Documentation Quality
- **Write clear, human-friendly, scannable documentation**
- **Choose right format for content**: headers for major sections, lists for features/steps, tables for comparisons, code blocks for examples
- **Let content flow naturally** - don't over-section or over-emphasize
- **Use bold text only for key terms**

Remember: **Think Carefully, Plan in Detail, Design the Best Solution, and Execute Precisely.**
