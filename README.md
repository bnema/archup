# archup

**A fast, minimal Arch Linux auto-installer alternative to archinstall**

## Project Status

✅ **Phase 0: Core Infrastructure - COMPLETE**

### Completed
- ✅ Project structure created
- ✅ Helper utilities implemented (logging, errors, presentation, chroot)
- ✅ Main orchestrator (install.sh)
- ✅ Barebone package list
- ✅ Logo designed

### Project Structure
```
archup/
├── README.md
├── logo.txt
├── install.sh                 # Main orchestrator
├── install/
│   ├── helpers/
│   │   ├── all.sh            # Source all helpers
│   │   ├── logging.sh        # Log file + colorized output
│   │   ├── errors.sh         # Trap handlers, error display
│   │   ├── presentation.sh   # UI helpers (cursor, colors)
│   │   └── chroot.sh         # Chroot utilities
│   ├── preflight/            # (Phase 1 - TODO)
│   ├── partitioning/         # (Phase 2 - TODO)
│   ├── base/                 # (Phase 3 - TODO)
│   ├── config/               # (Phase 4 - TODO)
│   ├── security/             # (Phase 5 - TODO)
│   ├── bootloader/           # (Phase 6 - TODO)
│   ├── post-install/         # (Phase 7 - TODO)
│   └── presets/
│       └── barebone.packages # Minimal package list (~15 packages)
├── bin/                      # (Future utilities)
└── docs/                     # (Future documentation)
```

## Installation

### Quick Install (from Arch ISO)

```bash
# Download and run archup installer
wget -O- https://raw.githubusercontent.com/bnema/archup/main/install.sh | bash
```

Or download first, then run:

```bash
# Download the installer
wget https://raw.githubusercontent.com/bnema/archup/main/install.sh

# Make it executable
chmod +x install.sh

# Run the installer
./install.sh
```

**Note:** Currently in Phase 0 development. Full installation not yet available.

---

## Development

### Testing Syntax
```bash
# Check all shell scripts for syntax errors and linting issues
make check

# Check syntax only (bash -n)
make check-syntax

# Run shellcheck linting
make check-shellcheck
```

### Testing Installation (Phase 0)
Currently in development. To test Phase 0:

```bash
# Set the archup path
export ARCHUP_PATH=/path/to/archup

# Run the installer (currently just tests helpers)
./install.sh
```

## Environment Variables

- `ARCHUP_PATH` - Installation directory (default: `$HOME/.local/share/archup`)
- `ARCHUP_REPO_URL` - Repository URL for bug reports (default: `https://github.com/bnema/archup`)
- `ARCHUP_CHROOT_INSTALL` - Set to `1` for chroot mode

## Next Steps

**Phase 1: Preflight Validation** (Week 1)
- [ ] Guards (vanilla Arch check)
- [ ] UEFI check
- [ ] x86_64 architecture check
- [ ] Secure Boot disabled check
- [ ] Begin installation (display logo, start log)

## Reference

Inspired by:
- **Omarchy installer** - Modular architecture and security implementations
- **archinstall** - Official Arch installer
- **Arch philosophy** - Simplicity, minimalism, user choice

## License

TBD