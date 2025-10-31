# Old TUI Implementation - DEPRECATED

## Status: Deprecated as of Phase 6 (2025-10-31)

This directory contains the legacy BubbleTea TUI implementation used before the Domain-Driven Design refactoring.

## Migration Status

**New Implementation:** `internal/interfaces/tui/`

The new TUI layer has been completely refactored to:
- ✅ Separate concerns (Models → State Only, Views → Rendering, Handlers → Events)
- ✅ Use application layer services instead of phase-based logic
- ✅ Integrate with progress tracking from ProgressTracker
- ✅ Follow the MVC pattern properly

## Files in This Directory (Legacy)

```
internal/ui/
├── assets/
│   └── assets.go           # Asset loading (DEPRECATED)
├── components/
│   ├── forms.go            # Form components (DEPRECATED)
│   ├── keybinds.go         # Key bindings (DEPRECATED)
│   └── output.go           # Output formatting (DEPRECATED)
├── model/
│   └── model.go            # Model structure (DEPRECATED)
├── styles/
│   ├── containers.go       # Container styling (DEPRECATED)
│   └── theme.go            # Theme definition (DEPRECATED)
├── views/
│   ├── confirmation.go     # Confirmation view (DEPRECATED)
│   ├── execution.go        # Execution view (DEPRECATED)
│   ├── forms.go            # Forms view (DEPRECATED)
│   ├── helpers.go          # View helpers (DEPRECATED)
│   ├── network.go          # Network view (DEPRECATED)
│   ├── results.go          # Results view (DEPRECATED)
│   └── welcome.go          # Welcome view (DEPRECATED)
├── installer_forms.go      # Installer form definitions (DEPRECATED)
├── model.go                # Main model file (DEPRECATED)
└── sections.go             # Section definitions (DEPRECATED)
```

## Why Deprecated?

The old UI implementation was tightly coupled with business logic and phases:

1. **Mixed Concerns**: Views, models, and business logic were intertwined
2. **Phase Coupling**: Direct dependencies on `internal/phases/`
3. **Non-testable**: Hard to test views without running business logic
4. **Inconsistent Patterns**: Multiple ways to handle state and events

## Migration Path

The new implementation at `internal/interfaces/tui/` provides:

### New Structure
```
internal/interfaces/tui/
├── app.go                  # Main application coordinator
├── models/
│   ├── form.go            # Form state (inputs, focus) - STATE ONLY
│   ├── installation.go    # Installation state - STATE ONLY
│   └── progress.go        # Progress state - STATE ONLY
├── views/
│   ├── form_view.go       # Form rendering - RENDERING ONLY
│   ├── progress_view.go   # Progress rendering - RENDERING ONLY
│   └── summary_view.go    # Summary rendering - RENDERING ONLY
├── handlers/
│   ├── form_handler.go    # Form event handling - EVENTS ONLY
│   └── install_handler.go # Installation flow handling - EVENTS ONLY
├── interfaces.go          # Interface definitions
├── messages.go            # Message types for events
└── errors.go              # TUI-specific errors
```

### Key Improvements

1. **Clear Separation of Concerns**
   - Models: Pure state data structures
   - Views: Pure rendering functions
   - Handlers: Pure event processing functions

2. **Service Integration**
   - Forms submit to application layer commands
   - Application layer handlers are invoked
   - Progress comes from ProgressTracker, not direct phase polling

3. **Testability**
   - View functions can be tested with mock models
   - Handler functions can be tested with mock commands
   - Models are simple DTOs, easy to test

4. **Maintainability**
   - Easy to add new views (just create new rendering function)
   - Easy to add new handlers (just create new event handler)
   - Easy to modify business logic (doesn't affect views)

## Deletion Timeline

- **Now (2025-10-31)**: Mark as deprecated
- **After Phase 6 Complete**: Keep for reference
- **After 4 Weeks (2025-11-28)**: Remove from main branch

## What To Do If You Find Bugs

If you find bugs in the OLD UI:
1. File an issue with "legacy-ui" label
2. If it's critical, it may be fixed in the deprecated code
3. Prefer fixing in the new UI at `internal/interfaces/tui/`

If you find bugs in the NEW UI:
1. File an issue with "tui" label
2. Create a PR fixing it in `internal/interfaces/tui/`
3. This is the primary implementation going forward

## References

- **New TUI**: `/home/brice/dev/projects/archup/internal/interfaces/tui/`
- **DDD Architecture**: `/home/brice/dev/projects/archup/docs/architecture.md`
- **Reference Document**: `/home/brice/dev/projects/archup/DDD_REFACTORING_PLAN.md`

## Questions?

See the architecture documentation or contact the team about the DDD migration.
