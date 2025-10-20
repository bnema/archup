#!/bin/bash
# Boot to ISO and check LUKS setup

echo "Checking LUKS volume..."

# Try to test if LUKS volume exists
cryptsetup luksDump /dev/vda2

echo ""
echo "Trying to open with 'test' password..."
echo "test" | cryptsetup open --test-passphrase /dev/vda2 && echo "Password 'test' works!" || echo "Password 'test' does NOT work!"
