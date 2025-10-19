.PHONY: check check-syntax check-shellcheck clean help

# Default target
help:
	@echo "archup - Makefile targets:"
	@echo ""
	@echo "  make check          - Run all checks (syntax + shellcheck)"
	@echo "  make check-syntax   - Check shell script syntax with bash -n"
	@echo "  make check-shellcheck - Run shellcheck linting"
	@echo "  make clean          - Remove generated files"
	@echo ""

# Run all checks
check: check-syntax check-shellcheck
	@echo ""
	@echo "✓ All checks passed!"

# Check syntax with bash -n
check-syntax:
	@echo "========================================="
	@echo "Checking shell script syntax..."
	@echo "========================================="
	@for script in install.sh install/helpers/*.sh install/preflight/*.sh install/partitioning/*.sh install/base/*.sh install/config/*.sh install/boot/*.sh; do \
		if [ -f "$$script" ]; then \
			echo "Checking: $$script"; \
			bash -n "$$script" || exit 1; \
		fi \
	done
	@echo "✓ Syntax check passed"

# Run shellcheck
check-shellcheck:
	@echo ""
	@echo "========================================="
	@echo "Running shellcheck..."
	@echo "========================================="
	@if ! command -v shellcheck >/dev/null 2>&1; then \
		echo "⚠ shellcheck not found, skipping..."; \
		echo "Install with: sudo pacman -S shellcheck"; \
		exit 0; \
	fi
	@for script in install.sh install/helpers/*.sh install/preflight/*.sh install/partitioning/*.sh install/base/*.sh install/config/*.sh install/boot/*.sh; do \
		if [ -f "$$script" ]; then \
			echo ""; \
			echo "Checking: $$script"; \
			shellcheck --exclude=SC1091,SC2086,SC2004,SC2059,SC2129,SC2155 "$$script" || exit 1; \
		fi \
	done
	@echo ""
	@echo "✓ shellcheck passed"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	@rm -f /var/log/archup-install.log
	@echo "✓ Clean complete"
