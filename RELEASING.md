# Release Process

This document describes how to create releases for ArchUp using goreleaser.

## Prerequisites

Install goreleaser:
```bash
# On Arch Linux
sudo pacman -S goreleaser

# Or with Go
go install github.com/goreleaser/goreleaser/v2@latest
```

## Release Types

### Development/Pre-release Builds

Development builds are tagged with suffixes like `-dev`, `-rc1`, `-alpha`, `-beta`.

**Examples:**
- `v0.15.3-dev`
- `v0.16.0-rc1`
- `v0.16.0-alpha`

These are automatically marked as pre-releases on GitHub and can be installed with:
```bash
curl -fsSL https://archup.run/install/bin | bash -s -- --dev
```

### Stable Releases

Stable releases use semantic versioning without suffixes.

**Examples:**
- `v0.15.3`
- `v1.0.0`

These are marked as the latest stable release and installed by default:
```bash
curl -fsSL https://archup.run/install/bin | bash
```

## Creating a Dev Release

1. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat(installer): add new feature"
   ```

2. **Tag the dev version:**
   ```bash
   git tag v0.15.3-dev
   ```

3. **Push the tag:**
   ```bash
   git push origin v0.15.3-dev
   ```

4. **Run goreleaser:**
   ```bash
   # Ensure you have GITHUB_TOKEN set
   export GITHUB_TOKEN="your_github_token"

   # Create the release
   goreleaser release --clean
   ```

The release will be created on GitHub as a pre-release.

## Creating a Stable Release

Same process as dev release, but use a clean version tag:

```bash
git tag v0.15.3
git push origin v0.15.3
goreleaser release --clean
```

## Testing Releases Locally

Before pushing, test the release process locally:

```bash
# Test without publishing
goreleaser release --snapshot --skip=publish --clean

# Check the artifacts in dist/
ls -lh dist/
```

## Testing Installation from VM

After creating a dev release, test the installation:

```bash
# On your Arch Linux VM/ISO
curl -fsSL https://archup.run/install/bin | bash -s -- --dev
```

Or test a specific version:
```bash
curl -fsSL https://archup.run/install/bin | bash -s -- --version v0.15.3-dev
```

## Artifacts Generated

Each release creates:
- `archup-installer` - Standalone binary
- `archup_VERSION_linux_x86_64.tar.gz` - Archive with binary + install scripts
- `checksums.txt` - SHA256 checksums for verification

## Troubleshooting

**goreleaser not found:**
```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

**GITHUB_TOKEN not set:**
Create a personal access token at https://github.com/settings/tokens with `repo` scope:
```bash
export GITHUB_TOKEN="ghp_your_token_here"
# Or add to ~/.bashrc for persistence
```

**Release failed:**
Check the goreleaser output for errors. Common issues:
- Tag already exists (delete with `git tag -d TAG` and `git push origin :refs/tags/TAG`)
- Invalid version format
- Missing permissions on GitHub token

## CI/CD (Future)

Consider setting up GitHub Actions to automate releases:
```yaml
# .github/workflows/release.yml
on:
  push:
    tags:
      - 'v*'
```
