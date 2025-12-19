# Ownkit Bazel Monorepo - Development Guide

This document explains how the ownkit monorepo is structured and how to work with it effectively.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Build System](#build-system)
4. [Toolchain Management](#toolchain-management)
5. [Directory Structure](#directory-structure)
6. [Development Workflow](#development-workflow)
7. [Common Commands](#common-commands)
8. [Troubleshooting](#troubleshooting)

## Project Overview

**Ownkit** is a personal development tools monorepo built with Bazel. It provides a centralized, reproducible build system for creating Go-based CLI tools and libraries.

### Key Features

- **Bazel 8.4+** with bzlmod for modern, reproducible builds
- **Go 1.25.3** toolchain managed by Bazel (no local Go installation needed)
- **bazel_env** for managing global toolchain dependencies
- **Gazelle** for auto-generating BUILD files from Go code
- **Flat monorepo structure** optimized for personal tool development

## Architecture

### Build System: Bazel with bzlmod

This repo uses **Bazel 8.4+** with **bzlmod** (Module Bazel) for dependency management:

- **No WORKSPACE file** - We use modern bzlmod exclusively
- **MODULE.bazel** - All external dependencies declared here
- **go.mod** - Standard Go module file, used by Bazel for SDK and dependency management
- **.bazelrc** - Optimized build configuration with proven performance flags

### Why Bazel?

Bazel provides several key benefits for this monorepo:

1. **Reproducibility** - Identical builds across all machines and environments
2. **Incremental Builds** - Only rebuild what changed
3. **Parallelization** - Builds run in parallel across all cores
4. **Toolchain Management** - Go, compilers, and tools are Bazel-managed
5. **Cache-friendly** - Remote caching support for team/CI scenarios

## Build System

### Critical Configuration Files

#### `.bazelversion`
```
8.4.2
```
Pins Bazel version for reproducibility. When using Bazelisk, this file ensures everyone uses the same Bazel version.

#### `MODULE.bazel`
Declares all external dependencies and configures the build system:

```python
# Core build rules
bazel_dep(name = "rules_go", version = "0.59.0", repo_name = "io_bazel_rules_go")
bazel_dep(name = "gazelle", version = "0.47.0", repo_name = "bazel_gazelle")
bazel_dep(name = "rules_shell", version = "0.6.1")

# Development dependencies
bazel_dep(name = "bazel_env.bzl", version = "0.5.0", dev_dependency = True)

# Go SDK Configuration
go_sdk = use_extension("@io_bazel_rules_go//go:extensions.bzl", "go_sdk")
go_sdk.download(name = "go_sdk", version = "1.25.3")
use_repo(go_sdk, "go_sdk")
```

**Key Points:**
- `repo_name` parameter maps module names to their actual repository names
- Go SDK version is explicitly pinned (must match go.mod)
- bazel_env is included for toolchain management
- No hardcoded Go version in build config - it's in go.mod

#### `.bazelrc`
Optimized build configuration with:

```bash
# Bzlmod (required)
common --enable_bzlmod

# Performance optimizations
build --jobs=40
build --disk_cache=
build --experimental_inmemory_dotd_files
build --experimental_inmemory_jdeps_files
build --grpc_keepalive_time=30s

# Go-specific
build --@io_bazel_rules_go//go/config:static
```

**Key Points:**
- `--jobs=40` enables parallel builds
- `--disk_cache=` disables disk cache (avoids lock contention)
- `--grpc_keepalive_time=30s` prevents gRPC timeouts
- Repository names use explicit `@io_bazel_rules_go` (not `@rules_go`)

#### `go.mod`
Standard Go module file:

```
module github.com/wfairclough/ownkit
go 1.25.3
```

**Key Points:**
- Go version here is the source of truth for Bazel
- MODULE.bazel go_sdk version must match this
- Additional dependencies are added here as you import packages

#### `BUILD.bazel` (root)
Defines build targets and bazel_env configuration:

```python
bazel_env(
    name = "bazel_env",
    toolchains = {},
    tools = {
        "go": "@io_bazel_rules_go//go",
    },
)

gazelle(
    name = "gazelle",
    prefix = "github.com/wfairclough/ownkit",
)

gazelle(
    name = "gazelle-update-repos",
    args = ["-from_file=go.mod", "-to_macro=go_deps.bzl%go_dependencies"],
    command = "update-repos",
)
```

**Key Points:**
- `bazel_env` target manages global toolchains
- Two Gazelle targets: one for BUILD files, one for dependencies

## Toolchain Management

### bazel_env

**bazel_env** is a Bazel extension that manages global toolchain dependencies. In ownkit, it's configured to make the Go toolchain available through `bazel run :bazel_env setup`.

### Go Toolchain

The Go toolchain is managed entirely by Bazel:

1. **Version Source** - `go.mod` (single source of truth)
2. **Download** - Bazel downloads Go 1.25.3 on first build
3. **Caching** - Cached locally in `~/.cache/bazel`
4. **No Local Installation** - No local Go installation needed

### Upgrading Go

To upgrade the Go version:

1. **Update go.mod**:
   ```bash
   go 1.25.3  # Change to new version
   ```

2. **Update MODULE.bazel** (must match go.mod):
   ```python
   go_sdk.download(name = "go_sdk", version = "1.25.3")  # Change version
   ```

3. **Verify**:
   ```bash
   bazel run @go_sdk//:go -- version
   ```

## Directory Structure

```
ownkit/
├── cmd/                    # Command-line applications
│   └── hello/              # Example: bazel run //cmd/hello
│       ├── main.go
│       └── BUILD.bazel     # Auto-generated by Gazelle
│
├── pkg/                    # Public, reusable packages
│   └── BUILD.bazel         # Placeholder
│
├── internal/               # Private, internal packages
│   └── BUILD.bazel         # Placeholder
│
├── tools/                  # Development tools
│   └── BUILD.bazel         # Placeholder
│
├── MODULE.bazel            # Bzlmod dependency declaration
├── .bazelversion           # Bazel version pin (8.4.2)
├── .bazelrc                # Build configuration
├── BUILD.bazel             # Root build file (Gazelle targets, bazel_env)
├── go.mod                  # Go module (version 1.25.3)
├── go_deps.bzl             # Generated by Gazelle (initially empty)
├── .gitignore              # Exclude Bazel artifacts
├── README.md               # User guide
└── CLAUDE.md               # This file
```

### Directory Conventions

- **cmd/** - Standalone CLI tools. Each subdirectory is a separate tool.
  - Example: `cmd/mytool/main.go` → `bazel run //cmd/mytool`

- **pkg/** - Public libraries. Importable by other packages and external projects.
  - Example: `pkg/utils/utils.go` can be imported by any package

- **internal/** - Private libraries. Only importable within ownkit.
  - Example: `internal/helpers/` can only be imported from ownkit packages

- **tools/** - Development/build tools and scripts.
  - Example: build helpers, code generators, etc.

## Development Workflow

### 1. Creating a New CLI Tool

```bash
# 1. Create directory
mkdir -p cmd/mytool

# 2. Create main.go
cat > cmd/mytool/main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("My tool works!")
}
EOF

# 3. Generate BUILD file
bazel run //:gazelle

# 4. Build
bazel build //cmd/mytool

# 5. Run
bazel run //cmd/mytool
```

### 2. Adding a Go Dependency

```bash
# 1. Add import to your code
# import "github.com/spf13/cobra"

# 2. Run go mod tidy (if needed)
go mod tidy

# 3. Generate/update BUILD files
bazel run //:gazelle

# 4. Build
bazel build //...
```

### 3. Creating a Shared Package

```bash
# 1. Create directory
mkdir -p pkg/mylib

# 2. Create files
cat > pkg/mylib/lib.go << 'EOF'
package mylib

func MyFunction() string {
    return "Hello from mylib"
}
EOF

# 3. Generate BUILD file
bazel run //:gazelle

# 4. Import from cmd
# import "github.com/wfairclough/ownkit/pkg/mylib"
```

### 4. Development Cycle

Typical development workflow:

```bash
# Edit code
vim cmd/mytool/main.go

# Auto-generate BUILD files (after changes)
bazel run //:gazelle

# Build and test
bazel build //...
bazel test //...

# Run your tool
bazel run //cmd/mytool -- arg1 arg2

# Or with environment setup
bazel run :bazel_env setup
```

## Common Commands

### Building

```bash
# Build everything
bazel build //...

# Build specific target
bazel build //cmd/hello

# Build with optimizations
bazel build --config=release //cmd/hello

# Build with debug symbols
bazel build --config=debug //cmd/hello
```

### Running

```bash
# Run a tool
bazel run //cmd/hello

# Run with arguments
bazel run //cmd/hello -- arg1 arg2

# Run with environment
bazel run :bazel_env setup
```

### Testing

```bash
# Run all tests
bazel test //...

# Run specific test
bazel test //pkg/mylib:mylib_test

# Verbose test output
bazel test --test_output=all //...
```

### Gazelle (BUILD File Management)

```bash
# Generate/update BUILD.bazel files
bazel run //:gazelle

# Update Go dependencies (if you edit go.mod directly)
bazel run //:gazelle-update-repos

# Check without modifying (useful for CI)
bazel run //:gazelle-update-repos -- -mode diff
```

### Cleaning

```bash
# Clean build artifacts
bazel clean

# Deep clean (including downloads)
bazel clean --expunge

# Remove specific cache
rm -rf ~/.cache/bazel
```

### Debugging

```bash
# Verbose output
bazel build --verbose_failures //cmd/hello

# Show what's being built
bazel build --announce_rc //cmd/hello

# See build graph
bazel query 'deps(//cmd/hello)'

# See reverse dependencies
bazel query 'rdeps(//cmd/hello, //:all)'
```

## Troubleshooting

### Issue: "error loading package" or repository not found

**Cause:** Gazelle hasn't generated BUILD files yet.

**Solution:**
```bash
bazel run //:gazelle
```

### Issue: "Module not found" in Go imports

**Cause:** Go dependencies not synced with Bazel.

**Solution:**
```bash
go mod tidy
bazel run //:gazelle-update-repos
bazel run //:gazelle
```

### Issue: Build hangs or times out

**Cause:** Usually due to gRPC timeouts or disk cache contention.

**Solution:** Already handled in `.bazelrc` with:
- `--grpc_keepalive_time=30s`
- `--disk_cache=`

If still hanging:
```bash
bazel clean --expunge
bazel build //...
```

### Issue: "No repository visible as '@go_sdk'"

**Cause:** MODULE.bazel go_sdk not properly registered.

**Solution:** Ensure MODULE.bazel has:
```python
go_sdk.download(name = "go_sdk", version = "1.25.3")
use_repo(go_sdk, "go_sdk")
```

### Issue: Gazelle not finding packages

**Cause:** BUILD.bazel files missing or outdated.

**Solution:**
```bash
bazel run //:gazelle
# Or clean and regenerate
bazel clean
bazel run //:gazelle
```

### Issue: "incompatible with dynamic configurations" errors

**Cause:** Flag compatibility issue with .bazelrc.

**Solution:** Check .bazelrc uses correct repository names:
```bash
# WRONG: --@rules_go//go/config:static
# RIGHT: --@io_bazel_rules_go//go/config:static
```

### Issue: Go version mismatch between Bazel and local

**Cause:** Different versions in go.mod and MODULE.bazel.

**Solution:** Ensure they match:
```bash
# go.mod
go 1.25.3

# MODULE.bazel
go_sdk.download(name = "go_sdk", version = "1.25.3")
```

## Performance Tips

1. **Use `--config=release` for final binaries** - Enables optimizations
2. **Keep go.mod clean** - Remove unused dependencies
3. **Use Gazelle in watch mode for rapid iteration** - (when supported)
4. **Enable remote caching for team** - Share build artifacts
5. **Run `bazel clean --expunge` periodically** - Clear stale caches

## References

- [Bazel Documentation](https://bazel.build/docs)
- [rules_go](https://github.com/bazelbuild/rules_go)
- [Gazelle](https://github.com/bazelbuild/bazel-gazelle)
- [bzlmod User Guide](https://bazel.build/external/module)
- [bazel_env](https://github.com/bazel-contrib/bazel_env)

## Contributing

When adding new features or tools:

1. **Follow the directory structure** - Tools go in `cmd/`, libraries in `pkg/`
2. **Write standard Go code** - Bazel handles compilation
3. **Let Gazelle generate BUILD files** - Don't write them manually
4. **Keep go.mod updated** - Add dependencies as you use them
5. **Document new tools** - Add comments explaining purpose

## Key Principles

1. **Single Source of Truth** - Go version in go.mod, dependencies in go.mod
2. **Automatic BUILD Files** - Gazelle generates them, don't edit manually
3. **Reproducible Builds** - Same input = same output, always
4. **Zero Local Setup** - Bazel downloads everything needed
5. **Team-Friendly** - Same .bazelversion ensures everyone builds the same way
