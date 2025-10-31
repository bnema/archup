# Dependency Injection Architecture (Phase 7)

## Overview

Phase 7 implements complete dependency injection wiring for the DDD architecture, enabling proper inversion of control and testability across all layers.

## Architecture

### Dependency Flow (Correct Inversion)

```
UI Layer (interfaces/tui)
  ↓ depends on
Application Layer (commands, handlers, services)
  ↓ depends on
Domain Layer (aggregates, ports, value objects)
  ↑ no dependencies
Infrastructure Layer (adapters) implements domain ports
```

## Wiring in cmd/archup-installer/main.go

### 1. Infrastructure Adapters (Lowest Layer)

All adapters are created first and implement domain ports:

```go
// Filesystem operations
fsAdapter := &filesystem.LocalFileSystem{}

// Command execution
shellExec := executor.NewShellExecutor(slogAdapter)
chrootExec := executor.NewChrootExecutor(slogAdapter)
scriptExec := executor.NewScriptExecutor(fsAdapter, shellExec, config.DefaultInstallDir)

// Persistence
repoAdapter, err := persistence.NewFileRepository(config.DefaultConfigPath)

// Logging
slogAdapter := infralogger.NewSlogAdapter(oldLog.Slog())
```

### 2. Command Handlers (Application Layer)

Handlers are injected with port dependencies only (never concrete implementations):

```go
preflightHandler := apphandlers.NewPreflightHandler(fsAdapter, shellExec, slogAdapter)
partitionHandler := apphandlers.NewPartitionHandler(shellExec, slogAdapter)
baseHandler := apphandlers.NewInstallBaseHandler(fsAdapter, shellExec, chrootExec, slogAdapter)
configHandler := apphandlers.NewConfigureSystemHandler(chrootExec, slogAdapter)
bootloaderHandler := apphandlers.NewBootloaderHandler(chrootExec, slogAdapter)
reposHandler := apphandlers.NewReposHandler(chrootExec, slogAdapter)
postInstallHandler := apphandlers.NewPostInstallHandler(chrootExec, scriptExec, slogAdapter)
```

### 3. Installation Service

The main orchestrator receives all handlers and coordinates the installation:

```go
installService := services.NewInstallationService(
    repoAdapter,
    slogAdapter,
    preflightHandler,
    partitionHandler,
    baseHandler,
    configHandler,
    bootloaderHandler,
    reposHandler,
    postInstallHandler,
)
```

### 4. TUI Application

The user interface receives the service and tracker, with no direct access to domain/infrastructure:

```go
tuiApp := tui.NewApp(installService, installService.Tracker(), slogAdapter)
```

## Constructor Signatures (Phase 7 Requirements)

### ShellExecutor
```go
func NewShellExecutor(logger ports.Logger) *ShellExecutor
```
- Requires: Logger for command execution logging

### ChrootExecutor
```go
func NewChrootExecutor(logger ports.Logger) *ChrootExecutor
```
- Requires: Logger for chroot operation logging

### ScriptExecutor
```go
func NewScriptExecutor(fs ports.FileSystem, cmdExec ports.CommandExecutor, scriptDir string) *ScriptExecutor
```
- Requires: FileSystem for script validation
- Requires: CommandExecutor for optional command execution
- Requires: Script directory path

### Logger Adapter
```go
func NewSlogAdapter(logger *slog.Logger) *SlogAdapter
```
- Requires: Pre-configured `*slog.Logger` instance
- No error return (file operations handled by caller)

### FileRepository
```go
func NewFileRepository(basePath string) (*FileRepository, error)
```
- Requires: Base path for state persistence
- Returns: error if directory creation fails

## Port Implementations

All infrastructure adapters implement domain ports exactly:

| Port | Adapter | Location |
|------|---------|----------|
| `FileSystem` | `LocalFileSystem` | `internal/infrastructure/filesystem/local.go` |
| `CommandExecutor` | `ShellExecutor` | `internal/infrastructure/executor/shell.go` |
| `ChrootExecutor` | `ChrootExecutor` | `internal/infrastructure/executor/chroot.go` |
| `ScriptExecutor` | `ScriptExecutor` | `internal/infrastructure/executor/script.go` |
| `Logger` | `SlogAdapter` | `internal/infrastructure/logger/slog_adapter.go` |
| `InstallationRepository` | `FileRepository` | `internal/infrastructure/persistence/file_repository.go` |

## Testing with Mocks

All ports have auto-generated mocks via mockgen:

```bash
go generate ./internal/domain/ports
```

Mocks enable unit testing of handlers without infrastructure:

```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()
mockFS := mocks.NewMockFileSystem(ctrl)
mockExec := mocks.NewMockCommandExecutor(ctrl)

handler := NewPreflightHandler(mockFS, mockExec, mockLogger)
// Test with mocks
```

## Deferred to Phase 8

HTTP client support for timezone API and file downloads will be added during Phase 8 (Incremental Migration) when migrating the system configuration and post-installation phases.

## Key Principles

1. **Dependency Inversion**: Handlers depend on abstractions (ports), not concrete types
2. **Single Responsibility**: Each adapter implements exactly one port interface
3. **Constructor Injection**: All dependencies passed via constructors, never globals
4. **Port Segregation**: Handlers receive only the ports they need
5. **Testability**: Mocks enable testing without infrastructure dependencies
