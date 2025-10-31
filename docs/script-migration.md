# Shell Scripts Audit and Migration Plan

**Audit Date:** October 31, 2025
**Phase:** 1 - Script Categorization for DDD Refactoring
**Total Scripts Found:** 54 files (50 .sh scripts, 1 .service file)

## Summary

This document categorizes all shell scripts in the `install/` directory for the transition from phase-based shell scripting to Domain-Driven Design (DDD) Go application.

### Categories

- **MANDATORY**: Scripts actively referenced in Go code - must be preserved and executed
- **INTEGRATION**: Scripts used by installation phases - will be wrapped by handlers
- **HELPER**: Utility scripts providing infrastructure functions - will be adapted to ports
- **CONFIG_TEMPLATE**: Configuration file templates - will be migrated to domain objects
- **DEPRECATED**: Legacy scripts for removal after migration

---

## Directory-by-Directory Analysis

### `install/bootstrap.sh` (Root Level)
- **Status**: MANDATORY
- **Purpose**: Main entry point for installation
- **Used By**: CLI entry point
- **Notes**: Will become dependency injection container

### `install/base/` (Base System Installation)
**Purpose**: Core Arch Linux system installation via pacstrap

| Script | Status | Used By | Notes |
|--------|--------|---------|-------|
| `all.sh` | INTEGRATION | BaseInstallationHandler | Orchestrates base installation |
| `pacstrap.sh` | INTEGRATION | all.sh | Runs pacstrap with package list |
| `kernel.sh` | INTEGRATION | all.sh | Selects and installs kernel variant |
| `amd-cpu.sh` | INTEGRATION | all.sh | Conditional AMD P-State setup |
| `enable-multilib.sh` | INTEGRATION | all.sh | Enables 32-bit support |
| `cachyos-repo.sh` | INTEGRATION | all.sh | Optional CachyOS repo setup |
| `pacman.sh` | INTEGRATION | all.sh | Pacman configuration |
| `fstab.sh` | INTEGRATION | all.sh | Generates /etc/fstab |

**Migration Strategy**: Extract domain logic to `domain/packages/` value objects, wrap script execution in `infrastructure/executor/script_executor.go`

---

### `install/boot/` (Bootloader Installation)
**Purpose**: EFI and bootloader setup

| Script | Status | Used By | Notes |
|--------|--------|---------|-------|
| `all.sh` | INTEGRATION | BootloaderHandler | Orchestrates bootloader setup |
| `limine.sh` | INTEGRATION | all.sh | Limine bootloader installation |

**Migration Strategy**: Extract bootloader configuration to `domain/bootloader/`, create `infrastructure/executor/bootloader_executor.go`

---

### `install/config/` (System Configuration)
**Purpose**: Hostname, locale, timezone, users, network

| Script | Status | Used By | Notes |
|--------|--------|---------|-------|
| `all.sh` | INTEGRATION | ConfigurationHandler | Orchestrates system config |
| `system.sh` | INTEGRATION | all.sh | Sets hostname, locale, timezone |
| `user.sh` | INTEGRATION | all.sh | Creates user, sets passwords |
| `network.sh` | INTEGRATION | all.sh | NetworkManager setup |
| `zram.sh` | INTEGRATION | all.sh | zRAM swap configuration |

**Migration Strategy**: Extract system config to `domain/system/config.go`, user management to `domain/user/`, create infrastructure adapters

---

### `install/configs/` (Configuration Templates)
**Purpose**: Template files for system configuration

| File | Status | Category | Destination |
|------|--------|----------|-------------|
| `limine.conf.template` | CONFIG_TEMPLATE | Bootloader | `domain/bootloader/limine_config.go` (embedded) |
| `chaotic-aur.conf` | CONFIG_TEMPLATE | Repository | `domain/packages/repository_config.go` (embedded) |
| `shell/starship.toml` | CONFIG_TEMPLATE | Shell | `infrastructure/persistence/shell_templates/` |
| `shell/shell` | CONFIG_TEMPLATE | Shell | `infrastructure/persistence/shell_templates/` |
| `shell/init` | CONFIG_TEMPLATE | Shell | `infrastructure/persistence/shell_templates/` |
| `shell/aliases` | CONFIG_TEMPLATE | Shell | `infrastructure/persistence/shell_templates/` |
| `shell/envs` | CONFIG_TEMPLATE | Shell | `infrastructure/persistence/shell_templates/` |
| `shell/rc` | CONFIG_TEMPLATE | Shell | `infrastructure/persistence/shell_templates/` |
| `shell/bashrc` | CONFIG_TEMPLATE | Shell | `infrastructure/persistence/shell_templates/` |

**Migration Strategy**: Load templates as embedded resources or at runtime, manage as value objects in domain layer

---

### `install/helpers/` (Helper Utilities)
**Purpose**: Utility functions used across scripts

| Script | Status | Used By | Purpose |
|--------|--------|---------|---------|
| `all.sh` | HELPER | Entry point | Loads all helpers |
| `logging.sh` | HELPER | All scripts | Logging functions |
| `errors.sh` | HELPER | All scripts | Error handling, cleanup |
| `config.sh` | HELPER | All scripts | Config file loading |
| `chroot.sh` | HELPER | Config, Boot, Repos | Chroot execution wrapper |
| `download.sh` | HELPER | Most scripts | HTTP download functions |
| `cleanup.sh` | HELPER | Error handling | System cleanup on failure |
| `multilib.sh` | HELPER | Config | Multilib detection/setup |
| `presentation.sh` | HELPER | All scripts | UI output (colors, formatting) |

**Migration Strategy**:
- `logging.sh` → `infrastructure/logger/shell_adapter.go` (wraps Go logger)
- `errors.sh` → `domain/*/errors.go` (domain error types)
- `config.sh` → `infrastructure/persistence/config_loader.go`
- `chroot.sh` → `infrastructure/executor/chroot_executor.go`
- `download.sh` → `infrastructure/http/client.go`
- `cleanup.sh` → `infrastructure/executor/cleanup_executor.go`
- `multilib.sh` → `domain/packages/multilib_detector.go`
- `presentation.sh` → `interfaces/tui/` (TUI layer)

---

### `install/partitioning/` (Disk Partitioning)
**Purpose**: Disk detection, partitioning, formatting, mounting

| Script | Status | Used By | Notes |
|--------|--------|---------|-------|
| `all.sh` | INTEGRATION | PartitioningHandler | Orchestrates partitioning workflow |
| `detect-disk.sh` | INTEGRATION | all.sh | Disk detection and validation |
| `partition.sh` | INTEGRATION | all.sh | Partition table creation |
| `format.sh` | INTEGRATION | all.sh | Filesystem creation (ext4, FAT) |
| `mount.sh` | INTEGRATION | all.sh | Mount partition hierarchy |

**Migration Strategy**: Extract disk logic to `domain/disk/`, create `infrastructure/executor/partitioning_executor.go`

---

### `install/post-boot/` (Post-Boot Setup)
**Status**: ✅ **MANDATORY** - ALL SCRIPTS

| Script | Status | Used By | Notes |
|--------|--------|---------|-------|
| `all.sh` | MANDATORY | Service | Main entry point |
| `snapper.sh` | MANDATORY | all.sh | Snapshot configuration |
| `ufw.sh` | MANDATORY | all.sh | Firewall setup |
| `ssh-keygen.sh` | MANDATORY | all.sh | SSH key generation |
| `archup-cli.sh` | MANDATORY | all.sh | CLI tool installation |
| `blesh.sh` | MANDATORY | all.sh | BLE shell integration |
| `archup-first-boot.service` | MANDATORY | Systemd | First-boot systemd service |

**Current References in Go Code**:
- `internal/config/config.go`: `PostBootScripts` array and `PostBootServiceTemplate` constant
- `internal/postboot/setup.go`: Downloads scripts and enables service

**Migration Strategy**:
- Create `install/mandatory/post-boot/` directory
- Keep scripts unchanged (shell-compatible)
- Create wrapper in `infrastructure/executor/postboot_executor.go`
- Validate scripts exist before execution
- Set environment variables in domain layer

---

### `install/post-install/` (Post-Installation Tasks)
**Purpose**: System finalization, themes, hooks, verification

| Script | Status | Used By | Notes |
|--------|--------|---------|-------|
| `all.sh` | INTEGRATION | PostInstallHandler | Orchestrates post-install tasks |
| `verify.sh` | INTEGRATION | all.sh | Installation verification |
| `unmount.sh` | INTEGRATION | all.sh | Unmount filesystems |
| `post-boot-setup.sh` | INTEGRATION | all.sh | Configures post-boot service |
| `boot-logo.sh` | INTEGRATION | all.sh | Boot splash image setup |
| `plymouth.sh` | INTEGRATION | all.sh | Plymouth theme setup |
| `snapper.sh` | INTEGRATION | all.sh | Snapshot initialization |
| `shell-config.sh` | INTEGRATION | all.sh | Shell initialization |
| `pacman.sh` | INTEGRATION | all.sh | Pacman hooks and config |
| `hooks.sh` | INTEGRATION | all.sh | Pacman hook installation |

**Migration Strategy**: Extract post-install logic to domain, wrap script execution in handler

---

### `install/preflight/` (Pre-Flight Checks)
**Purpose**: System validation before installation

| Script | Status | Used By | Notes |
|--------|--------|---------|-------|
| `all.sh` | INTEGRATION | PreflightHandler | Orchestrates all checks |
| `begin.sh` | INTEGRATION | all.sh | Entry point, setup |
| `guards.sh` | INTEGRATION | all.sh | Root/UEFI/disk validation |
| `detect-environment.sh` | INTEGRATION | all.sh | Environment detection |
| `identify.sh` | INTEGRATION | all.sh | System identification (CPU, GPU, etc.) |

**Migration Strategy**: Extract validation rules to `domain/installation/`, wrap in PreflightHandler

---

### `install/repos/` (Repository Configuration)
**Purpose**: AUR and additional repository setup

| Script | Status | Used By | Notes |
|--------|--------|---------|-------|
| `all.sh` | INTEGRATION | RepositoriesHandler | Orchestrates repo setup |
| `multilib.sh` | INTEGRATION | all.sh | Enables multilib repository |
| `aur.sh` | INTEGRATION | all.sh | AUR helper installation |
| `chaotic.sh` | INTEGRATION | all.sh | Chaotic-AUR setup |

**Migration Strategy**: Extract repo config to `domain/packages/repository.go`, wrap in RepositoriesHandler

---

## Script Classification Summary

### By Status

| Status | Count | Subdirectories | Action |
|--------|-------|-----------------|--------|
| **MANDATORY** | 7 | post-boot/ | Move to `install/mandatory/` - Keep unchanged |
| **INTEGRATION** | 44 | base/, boot/, config/, partitioning/, post-install/, preflight/, repos/ | Wrap with Go handlers |
| **HELPER** | 9 | helpers/ | Adapt logic to Go ports/adapters |
| **CONFIG_TEMPLATE** | 9+ | configs/ | Embed in Go or manage as resources |
| **DEPRECATED** | 0 | - | None identified for immediate removal |

### Total Impact
- **Total Shell Scripts**: 54 files
- **Lines to Preserve**: ~7 MANDATORY scripts
- **Lines to Refactor**: ~47 INTEGRATION + HELPER scripts
- **Expected Go LOC**: ~3000-4000 across domain, application, infrastructure layers

---

## Implementation Plan

### Phase 1.5: Directory Organization

1. ✅ Created `docs/script-migration.md` (this file)
2. Create `install/mandatory/` directory
3. Create `install/deprecated/` directory
4. Move post-boot scripts to `install/mandatory/post-boot/`
5. Update references in `internal/postboot/setup.go` to use new path
6. Commit changes

### Phase 2+: Script Wrapping

For each domain/phase:
1. Create Go domain entities with validation
2. Create command + handler in application layer
3. Create infrastructure adapter wrapping shell scripts
4. Write tests
5. Update TUI to use new handler
6. Update shell script paths if moved

---

## References

- DDD Architecture Plan: `/home/brice/dev/projects/archup/DDD_REFACTORING_PLAN.md`
- Current Config: `/home/brice/dev/projects/archup/internal/config/config.go`
- Post-Boot Setup: `/home/brice/dev/projects/archup/internal/postboot/setup.go`
- Original Project: `/home/brice/dev/clone/omarchy` (for reference)

---

## Notes

- All INTEGRATION scripts will be wrapped by Go handlers - no direct execution from Go code
- MANDATORY scripts remain unchanged to maintain compatibility with first-boot service
- HELPER scripts are utility functions - their logic will be adapted to Go but scripts can be deprecated
- CONFIG_TEMPLATE files will be embedded or loaded from `infrastructure/` layer
- No breaking changes to post-boot service execution
- Shell script config format must remain compatible (backward compatibility)

---

**Status**: ✅ Phase 1 Audit Complete
**Next Step**: Create directory structure and move mandatory scripts
