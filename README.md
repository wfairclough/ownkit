# Ownkit

Personal development tools monorepo built with Bazel.

## Prerequisites

- [Bazelisk](https://github.com/bazelbuild/bazelisk) (recommended) or Bazel 8.4+
- Git

**Note**: All toolchains (Go 1.25.3, Shell tools, etc.) are managed by Bazel via bazel_env - no local installation required!

## Quick Start

### 1. Install Bazelisk (Bazel launcher)

#### macOS
```bash
brew install bazelisk
```

#### Linux
```bash
# Download from GitHub releases
wget https://github.com/bazelbuild/bazelisk/releases/latest/download/bazelisk-linux-amd64
chmod +x bazelisk-linux-amd64
sudo mv bazelisk-linux-amd64 /usr/local/bin/bazel
```

### 2. Verify Setup

```bash
bazel version
```

This will automatically download Bazel 8.4.0 (specified in `.bazelversion`).

### 3. Build Everything

```bash
bazel build //...
```

### 4. Run Tests

```bash
bazel test //...
```

## Repository Structure

```
ownkit/
├── cmd/              # Command-line applications
│   └── example/      # Example: bazel run //cmd/example
├── pkg/              # Public, reusable packages
│   └── utils/        # Example: importable by other packages
├── internal/         # Private, internal packages
│   └── helpers/      # Example: only importable within ownkit
└── tools/            # Development tools and scripts
```

## Common Commands

### Building

```bash
# Build everything
bazel build //...

# Build specific target
bazel build //cmd/mytool

# Build with debug info
bazel build --config=debug //cmd/mytool

# Build optimized release binary
bazel build --config=release //cmd/mytool
```

### Running

```bash
# Run a command-line tool
bazel run //cmd/mytool

# Run with arguments
bazel run //cmd/mytool -- --arg1 value1
```

### Testing

```bash
# Run all tests
bazel test //...

# Run specific test
bazel test //pkg/utils:utils_test

# Run tests with verbose output
bazel test --test_output=all //...
```

### Gazelle (BUILD File Generation)

```bash
# Generate/update BUILD.bazel files for Go code
bazel run //:gazelle

# Update Go module dependencies from go.mod
bazel run //:gazelle-update-repos
```

### Cleaning

```bash
# Clean build artifacts
bazel clean

# Deep clean (including external dependencies)
bazel clean --expunge
```

## Adding a New Go Tool

### 1. Create the directory structure

```bash
mkdir -p cmd/mytool
```

### 2. Create your Go file

```go
// cmd/mytool/main.go
package main

import "fmt"

func main() {
    fmt.Println("Hello from mytool!")
}
```

### 3. Initialize go.mod (if not already present)

```bash
go mod init github.com/wfairclough/ownkit
go mod tidy
```

### 4. Generate BUILD files

```bash
bazel run //:gazelle
```

### 5. Build and run

```bash
bazel run //cmd/mytool
```

## Adding Go Dependencies

### 1. Add dependency to your Go code

```go
import "github.com/spf13/cobra"
```

### 2. Update go.mod

```bash
go mod tidy
```

### 3. Update Bazel dependencies

```bash
bazel run //:gazelle-update-repos
bazel run //:gazelle
```

## Development Workflow

1. **Write Go code** in appropriate directory (`cmd/`, `pkg/`, `internal/`)
2. **Update dependencies** if needed:
   ```bash
   go mod tidy
   bazel run //:gazelle-update-repos
   ```
3. **Regenerate BUILD files**:
   ```bash
   bazel run //:gazelle
   ```
4. **Build and test**:
   ```bash
   bazel build //...
   bazel test //...
   ```
5. **Run your tool**:
   ```bash
   bazel run //cmd/yourtool
   ```

## Troubleshooting

### "Module not found" errors

Run Gazelle to update BUILD files:
```bash
bazel run //:gazelle
```

### Dependency issues

Sync Go dependencies:
```bash
go mod tidy
bazel run //:gazelle-update-repos
bazel run //:gazelle
```

### Clean build

If you encounter weird caching issues:
```bash
bazel clean --expunge
bazel build //...
```

### Bazel version mismatch

Ensure you're using Bazelisk, which respects `.bazelversion`:
```bash
which bazel
# Should show bazelisk installation path
```

## Configuration Files

- **`.bazelversion`**: Pins Bazel version (8.4.0)
- **`MODULE.bazel`**: Bzlmod dependency configuration with:
  - `rules_go@0.59.0`, `gazelle@0.47.0`, `rules_shell@0.6.1`
  - `bazel_env.bzl` for managing global toolchains
  - Go SDK configured to read version from `go.mod`
- **`.bazelrc`**: Bazel build configuration and performance flags
- **`BUILD.bazel`**: Root build file with:
  - Gazelle targets for BUILD file generation
  - `bazel_env` configuration for toolchain management
- **`go.mod`**: Go module definition with version (1.25.3)
  - Go version is the source of truth for the toolchain

## Toolchain Management with bazel_env

This monorepo uses **bazel_env** to manage global toolchain dependencies:

- **Go Toolchain**: Version is defined in `go.mod` (currently 1.25.3)
- **How it works**:
  - `MODULE.bazel` declares `bazel_env.bzl` as a development dependency
  - `go_sdk.from_file(go_mod = "//:go.mod")` reads the Go version from go.mod
  - Bazel automatically downloads and manages the Go toolchain
  - No need for local Go installation or version management

### Upgrading Go Version

To upgrade the Go version used by Bazel:

1. Update `go.mod`:
   ```bash
   go 1.25.3  # Change this to new version
   ```

2. Bazel will automatically pick up the new version on next build

3. Verify the new version:
   ```bash
   bazel run @go_sdk//:go -- version
   ```

## CI/CD

Example GitHub Actions workflow:

```yaml
name: CI
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Mount bazel cache
        uses: actions/cache@v4
        with:
          path: ~/.cache/bazel
          key: bazel-${{ runner.os }}-${{ hashFiles('.bazelversion') }}
      - name: Install Bazelisk
        run: |
          wget https://github.com/bazelbuild/bazelisk/releases/latest/download/bazelisk-linux-amd64
          chmod +x bazelisk-linux-amd64
          sudo mv bazelisk-linux-amd64 /usr/local/bin/bazel
      - name: Build
        run: bazel build --config=ci //...
      - name: Test
        run: bazel test --config=ci //...
```

## Resources

- [Bazel Documentation](https://bazel.build/docs)
- [rules_go Documentation](https://github.com/bazelbuild/rules_go)
- [Gazelle Documentation](https://github.com/bazelbuild/bazel-gazelle)
- [Bzlmod User Guide](https://bazel.build/external/module)

## License

[Your chosen license]
