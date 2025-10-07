# Integration Test Taskfile Usage

## Overview

The `Taskfile.yml` in `tests/integration/` provides convenient commands for running integration tests with the correct build tags and flags.

## Prerequisites

```bash
# Install go-task (if not already installed)
go install github.com/go-task/task/v3/cmd/task@latest

# Or on most Linux distros
sudo pacman -S go-task  # Arch
sudo apt install go-task  # Debian/Ubuntu
```

## Quick Start

```bash
# From project root
cd /path/to/launchpad/tests/integration

# List all available tasks
go-task --list

# Run a specific task
go-task chain
go-task get-templates
go-task repo:user
```
## Key Features

### ✅ Automatic Build Tags

All tasks automatically include `-tags=integration`:

```yaml
vars:
  TEST_FLAGS: -v -count=1 -tags=integration
```

You don't need to remember to add the tag manually!

### ✅ Environment Variables

Tasks automatically set:
- `BASE_URL`: http://localhost:3000
- `TEST_USER_ID`: 550e8400-e29b-41d4-a716-446655440000

## Examples

### Running Chain Tests

```bash
$ go-task chain
=== RUN   TestCreateChain
    chain_test.go:45: Created test template: Integration Test Template 1759729083152878446
    chain_test.go:101: Successfully created chain with ID: 4b5d4041-affc-4135-b46b-19f681aa3d78
--- PASS: TestCreateChain (0.07s)
=== RUN   TestCreateChainWithoutTemplate
--- PASS: TestCreateChainWithoutTemplate (0.04s)
=== RUN   TestCreateChainValidation
--- PASS: TestCreateChainValidation (0.00s)
PASS
ok  	github.com/enielson/launchpad/tests/integration	0.129s
```

### Running Repository Tests (Fast!)

```bash
$ go-task repo:user
=== RUN   TestUserFixture_CreateAndRetrieve
    repository_user_test.go:35: Created user with ID: f6ae6ae3-2775-47bf-88c5-7baf91b5bc92
--- PASS: TestUserFixture_CreateAndRetrieve (0.01s)
PASS
ok  	github.com/enielson/launchpad/tests/integration	0.019s
```

Notice: **0.019s** for repository tests vs **0.129s** for HTTP tests!

### Generate Coverage Report

```bash
$ go-task coverage
=== RUN   TestCreateChain
...
PASS
coverage: 45.2% of statements
ok  	github.com/enielson/launchpad/tests/integration	0.135s	coverage: 45.2% of statements
Coverage report generated at coverage.html
```

Then open `tests/integration/coverage.html` in your browser.

### Watch Mode for Development

```bash
$ go-task watch
Watching for changes in tests/integration/...
# Edit a test file
# Tests automatically re-run!
```

## Task Configuration

### Custom Test Flags

You can override test flags:

```bash
# Run with custom flags
go-task chain TEST_FLAGS="-v -count=5 -tags=integration"

# Run with timeout
go-task chain TEST_FLAGS="-v -timeout=30s -tags=integration"
```

### Custom Base URL

```bash
# Test against different environment
go-task chain BASE_URL=http://staging.example.com:3000
```

## Comparison: Task vs Manual

**Before (manual):**
```bash
# Easy to forget the tag!
go test -v -run TestCreateChain ./tests/integration

# Error: no tests to run (forgot -tags=integration)
```

**After (using task):**
```bash
# Tag automatically included!
go-task chain

# Always works correctly
```

## Running from Different Directories

### From `tests/integration/`
```bash
cd tests/integration
go-task chain
```

### From Project Root
```bash
# Use -t flag to specify Taskfile location
go-task -t tests/integration/Taskfile.yml chain

# Or change directory first
cd tests/integration && go-task chain
```

## CI/CD Integration

In your CI pipeline:

```yaml
# .github/workflows/test.yml
- name: Run Integration Tests
  run: |
    cd tests/integration
    go-task all
```

Or:

```yaml
- name: Run Specific Tests
  run: |
    cd tests/integration
    go-task chain
    go-task auth
    go-task repo:all
```

## Tips & Best Practices

### 1. Use Repository Tests During Development

Repository tests are **10x faster** than HTTP tests:

```bash
# Fast feedback loop
go-task repo:all        # 0.020s
go-task repo:user       # 0.019s
```

Use these while developing business logic, then run full HTTP tests before commit.

### 2. Run Specific Test Patterns

You can still use go test directly with custom patterns:

```bash
cd tests/integration
go test -v -tags=integration -run "TestCreate.*Validation"
```

But tasks are more convenient for common patterns!

### 3. Clean Cache Between Test Runs

If tests behave unexpectedly:

```bash
go-task clean
go-task chain  # Fresh run
```

### 4. Check Server Before Running

```bash
go-task health-check
# If fails, start server:
# cd ../.. && set -a && source .env && set +a && go run main.go
```

## Troubleshooting

### Tests Fail with "command not found: task"

Install go-task:
```bash
go install github.com/go-task/task/v3/cmd/task@latest
# Then use 'task' or 'go-task' depending on your PATH
```

### Tests Fail with Connection Refused

Server isn't running:
```bash
# From project root
set -a && source .env && set +a && go run main.go

# Or use docker-compose
docker-compose up
```

### Tests Fail with Duplicate Key Error

Template names conflict. This is fixed by using timestamps in fixture names:
```go
templateName := fmt.Sprintf("Test Template %d", time.Now().UnixNano())
```

## Summary

The Taskfile provides:

✅ **Correct build tags** - Always includes `-tags=integration`
✅ **Convenient commands** - `go-task chain` vs manual `go test -v -tags=integration -run TestCreateChain`
✅ **Organization** - Logical grouping of related tests
✅ **Documentation** - `go-task --list` shows all available tests
✅ **Consistency** - Same flags and environment for all tests

Use it for faster, more reliable test execution!
