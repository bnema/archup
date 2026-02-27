#!/bin/bash
# Start installed Arch system with desktop hardware profile
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$TEST_DIR/start-qemu-installed.sh" --profile desktop "$@"
