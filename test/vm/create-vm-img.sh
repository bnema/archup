#!/bin/bash
# Create QEMU disk image for Arch testing

set -e

# Get test directory
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default values
IMAGE_NAME="${1:-arch-test.qcow2}"
IMAGE_SIZE="${2:-20G}"
IMAGE_DIR="${3:-$TEST_DIR}"

# Create directory if it doesn't exist
mkdir -p "$IMAGE_DIR"

IMAGE_PATH="$IMAGE_DIR/$IMAGE_NAME"

# Check if image already exists
if [[ -f "$IMAGE_PATH" ]]; then
    echo "Warning: Image '$IMAGE_PATH' already exists."
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 0
    fi
    rm -f "$IMAGE_PATH"
fi

# Create the disk image
echo "Creating disk image: $IMAGE_PATH (${IMAGE_SIZE})"
qemu-img create -f qcow2 "$IMAGE_PATH" "$IMAGE_SIZE"

echo "Disk image created successfully at: $IMAGE_PATH"
