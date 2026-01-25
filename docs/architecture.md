# ArchUp Architecture - Domain-Driven Design (DDD)

## Overview

ArchUp has been restructured using Domain-Driven Design (DDD) principles to achieve:
- Better testability (domain logic tested without infrastructure)
- Clear separation of concerns (business logic isolated from infrastructure)
- Improved maintainability (easy to extend without modifying existing code)
- Reduced coupling (dependencies point inward)

## Core Principles

1. **Business logic ONLY in domain** - NO external dependencies (no OS, HTTP, DB, etc.)
2. **Domain defines interfaces (Ports)** - Infrastructure implements them (Adapters)
3. **Dependencies point INWARD** - Everything depends on domain, domain depends on nothing

## Dependency Flow

```
┌─────────────────────────────────────┐
│      User Interface (TUI/CLI)       │  ← Entry points
├─────────────────────────────────────┤
│      Application Layer              │  ← Use cases (commands, handlers, services)
├─────────────────────────────────────┤
│         Domain Layer                │  ← Business logic (entities, value objects, rules)
├─────────────────────────────────────┤
│     Infrastructure Layer            │  ← Adapters implementing domain ports
└─────────────────────────────────────┘
```

Dependencies: UI → App → Domain ← Infrastructure

## Directory Structure

### `/internal/domain/` - Pure Business Logic

**NO external dependencies allowed** (except: `context`, `errors`, `fmt`, `time`, `strings`, `github.com/google/uuid`)

#### `domain/installation/`
- **installation.go** - Installation aggregate root (entity) managing lifecycle
- **state.go** - State enum and transition validation
- **events.go** - Domain events (InstallationStarted, PhaseCompleted, etc.)
- **errors.go** - Domain-specific error types

#### `domain/system/`
- **config.go** - SystemConfig value object (hostname, locale, timezone, keymap)
- **validation.go** - System validation rules
- **cpu.go** - CPU detection value object (vendor, microcode)

#### `domain/disk/`
- **disk.go** - Disk entity (device, size, partitions)
- **partition.go** - Partition value object (device, size, filesystem, mount point)
- **encryption.go** - Encryption value object (none/LUKS/LUKS-LVM, password requirements)
- **layout.go** - Partition layout rules and validation

#### `domain/bootloader/`
- **bootloader.go** - Bootloader value object (type, configuration)
- **limine.go** - Limine-specific configuration and rules

#### `domain/user/`
- **user.go** - User entity (username, groups, shell)
- **credentials.go** - Credentials value object (password hashing, validation)
- **validation.go** - User validation rules (username format, password strength)

#### `domain/packages/`
- **package.go** - Package value object
- **repository.go** - Repository configuration (Arch Linux, AUR, Chaotic-AUR)
- **kernel.go** - Kernel choice value object (linux, linux-zen, linux-lts, etc.)

#### `domain/ports/` - Hexagonal Architecture Ports
**ALL infrastructure interactions via these interfaces**

- **filesystem.go** - FileSystem port (Read, Write, Chmod, MkdirAll, etc.)
- **executor.go** - CommandExecutor, ChrootExecutor, ScriptExecutor ports
- **repository.go** - InstallationRepository port (Save, Load)
- **downloader.go** - HTTPClient port (Get, Post)
- **logger.go** - Logger port (Info, Warn, Error, Debug)

### `/internal/application/` - Use Case Orchestration

Glues together domain logic with infrastructure adapters. **Imports domain/** but not **infrastructure/**.

#### `application/commands/`
Plain data structures (DTOs) representing user intents:
- **preflight.go** - PreflightCommand
- **partition.go** - PartitionDiskCommand
- **install_base.go** - InstallBaseCommand
- **configure_system.go** - ConfigureSystemCommand
- **install_bootloader.go** - InstallBootloaderCommand
- **setup_repos.go** - SetupRepositoriesCommand
- **post_install.go** - PostInstallCommand

#### `application/handlers/`
Command handlers implementing use cases. Each handler:
1. Receives a command (DTO)
2. Uses ports to gather system info
3. Creates domain objects with validation
4. Calls domain methods
5. Returns result (DTO)

- **preflight_handler.go** - Handles PreflightCommand
- **partition_handler.go** - Handles PartitionDiskCommand
- **install_base_handler.go** - Handles InstallBaseCommand
- **configure_system_handler.go** - Handles ConfigureSystemCommand
- **bootloader_handler.go** - Handles InstallBootloaderCommand
- **repos_handler.go** - Handles SetupRepositoriesCommand
- **post_install_handler.go** - Handles PostInstallCommand

#### `application/services/`
Application services coordinating the overall installation:
- **installation_service.go** - Main orchestrator managing all handlers and state
- **progress_tracker.go** - Progress tracking with event publishing

#### `application/dto/`
Data Transfer Objects for communication across boundaries:
- **installation_status.go** - InstallationStatus, PhaseStatus
- **progress_update.go** - ProgressUpdate for UI subscribers

### `/internal/infrastructure/` - Adapters (External System Implementation)

Implements domain ports. **Imports domain/** but NOT **application/**.

#### `infrastructure/filesystem/`
Filesystem port adapters:
- **local.go** - LocalFileSystem (implements domain.FileSystem)
- **mock.go** - MockFileSystem (for testing)

#### `infrastructure/executor/`
Command execution adapters:
- **shell.go** - ShellExecutor (implements domain.CommandExecutor)
- **chroot.go** - ChrootExecutor (implements domain.ChrootExecutor)
- **script.go** - ScriptExecutor for mandatory shell scripts
- **mock.go** - MockExecutor (for testing)

#### `infrastructure/persistence/`
Data persistence adapters:
- **file_repository.go** - FileRepository (implements domain.InstallationRepository)
- **config_file.go** - Config file loader/saver (shell-compatible format)

#### `infrastructure/http/`
HTTP client adapters:
- **client.go** - HTTPClient (implements domain.Downloader)
- **mock.go** - MockHTTPClient (for testing)

#### `infrastructure/logger/`
Logging adapters:
- **slog_adapter.go** - SlogAdapter (implements domain.Logger, wraps Go's slog)

### `/internal/interfaces/` - Entry Points (Thin Layer)

UI/CLI layer. **Imports application/** but minimal domain access.

#### `interfaces/tui/`
BubbleTea terminal UI application:
- **app.go** - Main TUI application
 - **wizard/** - Post-boot wizard TUI (desktop setup)

#### `interfaces/tui/models/`
BubbleTea models (UI state only, no business logic):
- **form.go** - Form input model
- **installation.go** - Installation progress model
- **progress.go** - Progress display model

#### `interfaces/tui/views/`
Pure presentation layer (rendering):
- **form_view.go** - Form rendering
- **progress_view.go** - Progress rendering
- **summary_view.go** - Summary rendering

#### `interfaces/tui/handlers/`
UI event handlers (delegate to application services):
- **form_handler.go** - Form submission handler
- **install_handler.go** - Installation event handler

#### `interfaces/cli/` (Future)
CLI commands (e.g., using Cobra):
- **root.go** - Root command
- **install.go** - Install command

### `/cmd/archup/`
Application entry point with dependency injection:
- **main.go** - Wire all dependencies (adapters, handlers, services, TUI)

### Wizard Flow (post-boot)

- Welcome → Compositor → SDDM → Optional → Confirm → Install → Monitors → Apply Config → Complete
- Config outputs:
  - Hyprland: `~/.config/hypr/archup.conf`, `~/.config/hypr/archup-monitors.conf`
  - Niri: `~/.config/niri/archup.kdl`, `~/.config/niri/archup-monitors.kdl`
  - Waybar: `~/.config/waybar/config`, `~/.config/waybar/style.css`
  - Locks: `~/.config/hypr/hyprlock.conf`, `~/.config/hypr/hypridle.conf`

## Architecture Patterns

### 1. Aggregate Root (Installation)

The `Installation` aggregate manages the entire installation lifecycle:
```
Installation (aggregate root)
├── state: InstallationState
├── config: InstallationConfig (value object)
├── system: SystemConfig (value object)
├── disk: Disk (entity with partitions)
├── user: User (entity)
└── ...
```

### 2. Value Objects

Immutable, validated objects representing concepts:
- `SystemConfig` - hostname, timezone, locale
- `Encryption` - encryption type and password requirements
- `Credentials` - user password (hashed)
- `Kernel` - kernel choice

### 3. Ports (Interfaces)

Domain defines what external systems must provide:
```go
// domain/ports/filesystem.go
type FileSystem interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    Chmod(path string, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error
}
```

### 4. Adapters

Infrastructure implements ports:
```go
// infrastructure/filesystem/local.go
type LocalFileSystem struct{}

func (fs *LocalFileSystem) ReadFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}
// ... implements domain.FileSystem
```

### 5. Command Handler Pattern

Each use case has:
1. **Command** (DTO): Input data
2. **Handler**: Orchestrates domain + infrastructure
3. **Result** (DTO): Output data

```go
// Define command
cmd := commands.PreflightCommand{
    MountPoint: "/mnt",
}

// Execute via handler
result, err := handler.Handle(ctx, cmd)
```

## Testing Strategy

### Domain Tests
- **Where**: `internal/domain/*/` with `_test.go` files
- **What**: Pure unit tests of business rules
- **How**: No mocks, no external dependencies
- **Speed**: Fast (milliseconds)

```go
func TestSystemConfig_Validation(t *testing.T) {
    cfg, err := domain.NewSystemConfig("hostname", "UTC", "en_US.UTF-8", "us")
    assert.NoError(t, err)
    assert.Equal(t, "hostname", cfg.Hostname())
}
```

### Application Tests
- **Where**: `internal/application/*/` with `_test.go` files
- **What**: Handler logic with mocked ports
- **How**: Use mock implementations (infrastructure/*/mock.go)
- **Speed**: Fast (milliseconds)

```go
func TestPreflightHandler_Handle(t *testing.T) {
    mockFS := &filesystem.MockFileSystem{}
    mockExec := &executor.MockExecutor{}
    handler := handlers.NewPreflightHandler(mockFS, mockExec, logger)

    result, err := handler.Handle(ctx, cmd)
    assert.NoError(t, err)
}
```

### Integration Tests
- **Where**: `test/integration/` or similar
- **What**: Full end-to-end scenarios
- **How**: Real adapters in VM environment
- **Speed**: Slow (minutes), run before commit

```go
func TestFullInstallation(t *testing.T) {
    // End-to-end test in VM
    // Uses real filesystem, chroot, pacstrap, etc.
}
```

## Migration Checklist

For each phase migrated from shell scripts:

1. **Extract business rules** → domain layer
2. **Create value objects** with validation
3. **Define port interfaces** needed
4. **Create command DTO** for input
5. **Implement handler** in application layer
6. **Create adapters** implementing ports
7. **Write domain tests** (pure unit tests)
8. **Write application tests** (with mocks)
9. **Write integration tests** (in VM)
10. **Update TUI** to use handler
11. **Test in VM** to verify functionality
12. **Document changes**

## Development Workflow

### Adding a New Feature

1. Define domain rules in `domain/*/`
2. Create command in `application/commands/`
3. Create handler in `application/handlers/`
4. Implement adapters in `infrastructure/*/`
5. Add UI in `interfaces/tui/`
6. Wire in `cmd/archup/main.go`

### Fixing a Bug

1. Write failing test at appropriate layer
2. Fix in domain (if business logic) or adapter (if infrastructure)
3. Verify tests pass
4. Test in VM if infrastructure-related

### Changing Infrastructure

1. Create new adapter implementing port
2. Update `cmd/archup/main.go` wiring
3. Existing domain and application code unchanged
4. No need to modify business logic

## Important Files

- **DDD_REFACTORING_PLAN.md** - Complete refactoring plan with detailed phases
- **docs/script-migration.md** - Shell scripts categorization and migration strategy
- **internal/domain/ports/** - All port interfaces (critical!)
- **cmd/archup/main.go** - Dependency injection wiring

## References

- Eric Evans - Domain-Driven Design (DDD book)
- Hexagonal Architecture (Ports & Adapters)
- Clean Architecture by Robert C. Martin
- Project DDD Plan: `DDD_REFACTORING_PLAN.md`
- Shell Scripts Audit: `docs/script-migration.md`

## Key Rules

### Domain Layer
- ✅ Pure business logic
- ✅ Validation of invariants
- ✅ Value objects (immutable)
- ✅ Entities with identity
- ❌ NO external dependencies
- ❌ NO file system access
- ❌ NO HTTP calls
- ❌ NO database access

### Application Layer
- ✅ Orchestrates domain + infrastructure
- ✅ Implements use cases
- ✅ Delegates to ports
- ✅ Returns DTOs
- ❌ NO business logic
- ❌ NO direct infrastructure calls
- ❌ NO direct UI logic

### Infrastructure Layer
- ✅ Implements ports
- ✅ Handles external systems
- ✅ Contains adapters
- ❌ NO business logic
- ❌ NO use case orchestration

### Interfaces Layer
- ✅ Handles user input/output
- ✅ Delegates to application services
- ✅ Pure presentation
- ❌ NO business logic
- ❌ NO infrastructure details

---

**Status**: Phase 2 Complete - Directory structure established
**Next**: Phase 3 - Define Domain Layer with entities and value objects
