# Stampede Shooter - High-Performance Load Testing Tool

A simple, highly efficient load testing service for streaming web platforms. Built in Go with support for configurable user counts, authentication, scripted requests, and comprehensive metrics tracking.

## 🚀 **Key Features**

- **High Performance**: Written in Go for maximum efficiency
- **Rails Devise Support**: Automatic CSRF token extraction and session management
- **Realistic User Simulation**: Cookie persistence, header management, rate limiting
- **Flexible Scripting**: YAML-based test scripts with template variables
- **Comprehensive Metrics**: HDR histograms, response times, error tracking
- **Live Reporting**: Real-time test progress and final JSON reports
- **Credentials Management**: Round-robin credential assignment from text files

## 🎯 **Perfect for Rails Applications**

Stampede Shooter is specifically enhanced for Rails applications that use:
- **Devise authentication** with CSRF protection
- **Session-based authentication** for web interfaces
- **Cookie management** and header persistence
- **Realistic browser simulation**

## 📦 **Quick Start**

### Installation
```bash
git clone <repository>
cd stampede
make build
```

### Basic Usage
```bash
# Simple smoke test
./build/stampede-shooter --users 5 --script examples/smoke.yml --duration 30s

# Rails authentication demo
make test-acme-demo

# Test with credentials file
make test-credentials

# Real Rails application test
make test-acme-real
```

## 🔧 **Configuration**

### Command Line Options
```bash
./build/stampede-shooter \
  --users 10 \           # Number of concurrent users
  --rps 5 \              # Requests per second per user
  --duration 60s \       # Test duration
  --script test.yml \    # Test script file
  --credentials creds.txt \ # Credentials file (username,password)
  --out results.json \   # Output file
  --verbose \            # Detailed logging
  --insecure-tls         # Skip TLS verification
```

### Test Script Format (YAML)
```yaml
- name: Login
  method: POST
  url: https://app.com/login
  headers:
    Content-Type: application/x-www-form-urlencoded
  body: |
    email={{username}}&password={{password}}&authenticity_token=CSRF_TOKEN_PLACEHOLDER
  expect_status: 302

- name: Dashboard
  method: GET
  url: https://app.com/dashboard
  expect_status: 200
```

### Credentials File Format
```bash
# credentials.txt
user1@example.com,password123
user2@example.com,password123
user3@example.com,password123
# Lines starting with # are comments
```

## 🎯 **Rails Devise Authentication**

### Automatic Features
- ✅ **CSRF Token Extraction** from HTML responses
- ✅ **Session Cookie Management** via cookie jar
- ✅ **Header Persistence** across requests
- ✅ **Template Variable Support** for unique users
- ✅ **Credential Placeholders** for secure authentication

### Example Rails Flow
```yaml
# 1. Get login page (extracts CSRF token)
- name: LoginPage
  method: GET
  url: https://app.com/users/sign_in

# 2. Submit login (uses extracted token and credentials)
- name: LoginSubmit
  method: POST
  url: https://app.com/users/sign_in
  body: |
    user[email]={{username}}&user[password]={{password}}&authenticity_token=CSRF_TOKEN_PLACEHOLDER

# 3. Access authenticated content
- name: Dashboard
  method: GET
  url: https://app.com/dashboard
```

## 📊 **Template Variables**

Use dynamic values in your scripts:
- `{{userId}}` - Current user ID (1, 2, 3...)
- `{{randInt 1 100}}` - Random integer between 1-100
- `{{epochms}}` - Current timestamp in milliseconds
- `{{username}}` - Username from credentials file
- `{{password}}` - Password from credentials file

## 🔄 **Round-Robin Credential Assignment**

The tool automatically assigns credentials to users:
- **User 1** gets credentials from line 1
- **User 2** gets credentials from line 2
- **User 3** gets credentials from line 3
- **User 4** gets credentials from line 1 (wraps around)

## 📈 **Metrics & Reporting**

### Live Metrics
```
Elapsed: 30s | Requests: 150 | Errors: 2 | Success: 98.7% | RPS: 5.0
```

### Final Report
```
Action        OK   ERR   p50   p90   p99   RPS
──────────── ──── ──── ───── ───── ───── ────
Login         50    0   45ms  89ms  156ms  1.7
Dashboard     50    0   67ms  145ms 289ms  1.7
ViewEvent     50    0   78ms  167ms 334ms  1.7
```

### JSON Output
```json
{
  "summary": {
    "total_requests": 150,
    "success_rate": 0.987,
    "duration": "30s",
    "avg_rps": 5.0
  },
  "actions": {
    "Login": {
      "count": 50,
      "errors": 0,
      "p50": 45,
      "p90": 89,
      "p99": 156
    }
  }
}
```

## 🎯 **Examples**

### Quick Demo
```bash
# Test Rails authentication flow with httpbin.org
make test-acme-demo
```

### Test with Credentials
```bash
# Test credentials feature
make test-credentials

# Test actual Rails application (requires valid credentials)
make test-acme-with-creds
```

### Analyze Browser Recording
```bash
# Extract headers and create test script from browser recording
make analyze-acme FILE=your-recording.json
```

## 📚 **Documentation**

- **[Rails Load Testing Guide](docs/acme-load-testing-guide.md)** - Complete guide for Rails applications
- **[Session Management Guide](docs/session-management-guide.md)** - Authentication and session handling
- **[Header Analysis Tool](docs/header-analysis.md)** - Browser recording analysis

## 🔧 **Development**

### Build
```bash
make build          # Build binary
make build-all      # Build for all platforms
make clean          # Clean build artifacts
```

### Test
```bash
make test           # Run all tests
make test-smoke     # Quick smoke test
make test-benchmark # Performance benchmark
```

### Examples
```bash
make test-acme-demo    # Rails authentication demo
make test-credentials   # Test credentials feature
make test-acme-real    # Real Rails application test
make analyze-headers    # Analyze browser recordings
```

## 🚨 **Troubleshooting**

### Common Issues

**CSRF Token Errors (422)**
- Ensure login page is fetched first
- Check token extraction is working
- Verify token is included in form data

**Session Expired (401/403)**
- Check cookie persistence
- Verify session timeout settings
- Monitor for session invalidation

**Invalid Credentials (401)**
- Verify credentials in file are correct
- Check username/password format
- Ensure no extra spaces in credentials file

**High Error Rates**
- Start with fewer users
- Reduce RPS per user
- Check server capacity

### Debug Mode
```bash
./build/stampede-shooter --users 1 --script debug.yml --credentials examples/credentials.txt --duration 10s --verbose
```

## 📄 **License**

MIT License - see LICENSE file for details.

## 🤝 **Contributing**

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

---

**Ready to load test your Rails application?** Start with `make test-credentials` to see Stampede Shooter in action! 