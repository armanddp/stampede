# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Stampede is a high-performance load testing tool written in Go, optimized for testing Rails applications with authentication. It uses goroutines for concurrent request execution and supports YAML-based test scripts with template variables.

## Essential Commands

### Building
- `make build` - Build binary for current platform
- `make build-all` - Build for Linux, macOS (Intel/ARM), Windows
- `make install` - Install to /usr/local/bin

### Testing
- `make smoke-test` - Quick test against httpbin.org
- `make test-acme-demo` - Test Rails authentication flow
- `make test-credentials` - Test credential management

### Development
- `make deps` - Download and verify dependencies
- `make clean` - Clean build artifacts
- `make analyze-headers FILE=recording.json` - Analyze browser recordings

## Architecture

The codebase follows a pipeline architecture:

```
CLI → Config → Orchestrator → Workers → Metrics → Reporter
```

Key directories:
- `/cmd/shooter/` - CLI entry point
- `/internal/orchestrator/` - Manages worker lifecycle and test execution
- `/internal/worker/` - HTTP request execution with Rails-specific features
- `/internal/metrics/` - HDR histogram-based performance tracking
- `/internal/script/` - YAML script parsing and template processing

## Rails-Specific Features

When working on Rails/Devise authentication features:
1. CSRF tokens are automatically extracted from HTML responses containing `csrf-token`
2. Session cookies are managed via cookie jar
3. Headers like X-CSRF-Token persist across requests
4. Look at `/internal/worker/worker.go` for authentication logic

## Template Variables

The system supports these template variables in YAML scripts:
- `{{userId}}` - Current user ID (1-based)
- `{{username}}`, `{{password}}` - From credentials file
- `{{randInt min max}}` - Random integer
- `{{epochms}}` - Current timestamp

Template processing happens in `/internal/script/template.go`.

## Testing Approach

The project currently lacks unit tests. When adding tests:
1. Follow Go testing conventions
2. Place tests alongside source files (*_test.go)
3. Focus on testing the worker, metrics, and script parsing components
4. Use the existing smoke tests as integration test examples

## Common Development Tasks

To add a new template variable:
1. Edit `/internal/script/template.go`
2. Add the variable to `replaceVariables()` function
3. Update documentation in README.md

To modify authentication behavior:
1. Edit `/internal/worker/worker.go`
2. Look for `extractCSRFToken()` and cookie handling
3. Test with `make test-acme-demo`

To add new metrics:
1. Edit `/internal/metrics/collector.go`
2. Update `/internal/reporter/reporter.go` for output
3. Consider performance impact on high-concurrency scenarios