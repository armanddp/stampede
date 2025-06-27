# Session Management Guide

This guide explains how to handle session management and authentication in load testing with Stampede Shooter.

## Overview

Proper session management is crucial for realistic load testing of authenticated applications. This guide covers:

- Identifying session headers to persist
- Testing authentication flows manually
- Creating YAML scripts with session handling
- Running load tests with session management

## Identifying Session Headers

### 1. Browser Developer Tools Method

1. **Open Chrome DevTools** (F12)
2. **Go to Network tab**
3. **Clear existing requests**
4. **Perform your login flow**
5. **Look for these patterns:**

#### Key Headers to Check:

**Request Headers:**
- `Cookie` - Session cookies, CSRF tokens
- `X-CSRF-Token` - Rails CSRF protection
- `X-Requested-With` - AJAX requests
- `Authorization` - Bearer tokens, API keys

**Response Headers:**
- `Set-Cookie` - New session cookies
- `X-CSRF-Token` - Updated CSRF tokens
- `Location` - Redirect URLs (for 302 responses)

### 2. Common Session Patterns

#### Rails Applications:
```
Cookie: _session_id=abc123; remember_user_token=xyz789
X-CSRF-Token: def456
```

#### JWT-based APIs:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### PHP Applications:
```
Cookie: PHPSESSID=abc123; csrf_token=def456
```

### 3. Login Flow Analysis

1. **GET /login** - Extract CSRF token from form
2. **POST /login** - Submit credentials + CSRF token
3. **Response** - Get session cookie from Set-Cookie header
4. **Subsequent requests** - Include session cookie

### 4. Testing Header Persistence

Use curl to test which headers are required:

```bash
# Step 1: Get login page and extract CSRF token
curl -c cookies.txt -b cookies.txt https://example.com/users/sign_in

# Step 2: Login with credentials and CSRF token
curl -c cookies.txt -b cookies.txt -X POST \
  -d "user[email]=test@example.com&user[password]=password&authenticity_token=TOKEN" \
  https://example.com/users/sign_in

# Step 3: Test authenticated request
curl -b cookies.txt https://example.com/dashboard
```

### 5. Common Issues

- **CSRF Token Mismatch** - Need to extract and use current token
- **Session Timeout** - Cookies expire, need to re-authenticate
- **IP Restrictions** - Some sites bind sessions to IP addresses
- **User-Agent Validation** - Must use consistent User-Agent string

## Creating YAML Scripts with Session Management

### Basic Rails Authentication Flow

```yaml
# 1. Get login page (extracts CSRF token automatically)
- name: LoginPage
  method: GET
  url: https://example.com/
  headers:
    User-Agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
    Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
  expect_status: 200

# 2. Submit login (uses extracted token)
- name: LoginSubmit
  method: POST
  url: https://example.com/users/sign_in
  headers:
    Content-Type: "application/x-www-form-urlencoded"
    Referer: "https://example.com/"
    Origin: "https://example.com"
  body: |
    user%5Bemail%5D={{username}}&user%5Bpassword%5D={{password}}&authenticity_token=CSRF_TOKEN_PLACEHOLDER
  expect_status: 302

# 3. Access authenticated content
- name: Dashboard
  method: GET
  url: https://example.com/dashboard
  headers:
    Referer: "https://example.com/users/sign_in"
  expect_status: 200
```

### Key Features

- **Automatic CSRF Token Extraction** - Tool extracts tokens from HTML responses
- **Cookie Persistence** - Session cookies are automatically maintained
- **Header Persistence** - Important headers are carried forward
- **Template Variables** - Use `{{username}}` and `{{password}}` for credentials

## Running Load Tests

### Basic Test

```bash
./build/stampede-shooter \
  --users 5 \
  --rps 2 \
  --duration 60s \
  --script auth-test.yml \
  --credentials users.txt \
  --verbose
```

### Advanced Test with Session Management

```bash
./build/stampede-shooter \
  --users 10 \
  --rps 3 \
  --duration 300s \
  --script rails-auth.yml \
  --credentials production-users.txt \
  --out session-test-results.json \
  --verbose
```

## Monitoring Session Health

### Good Session Management
```
Action        OK   ERR   p50   p90   p99   RPS
──────────── ──── ──── ───── ───── ───── ────
LoginSubmit   100    0   45ms  89ms  156ms  10
Dashboard     100    0   67ms  145ms 289ms  10
ViewEvent     100    0   78ms  167ms 334ms  10
```

### Session Issues
```
Action        OK   ERR   p50   p90   p99   RPS
──────────── ──── ──── ───── ───── ───── ────
LoginSubmit    50   50   45ms  89ms  156ms   5  # 50% login failures
Dashboard       0  100    0ms   0ms   0ms   0  # 100% dashboard failures
```

## Troubleshooting

### Session Expired Errors (401/403)
- Check cookie persistence is working
- Verify session timeout settings
- Monitor for session invalidation

### CSRF Token Errors (422)
- Ensure login page is fetched first
- Check token extraction is working
- Verify token is included in form data

### High Error Rates
- Start with fewer users
- Reduce RPS per user
- Check server capacity

## Best Practices

1. **Start Small** - Begin with 1-2 users to verify authentication
2. **Monitor Sessions** - Watch for session expiration and errors
3. **Use Realistic Credentials** - Test with actual user accounts
4. **Check Server Logs** - Monitor for authentication issues
5. **Validate Headers** - Ensure required headers are being sent

## Advanced Configuration

### Custom Session Headers

If your application requires custom session headers, add them to your YAML:

```yaml
- name: AuthenticatedRequest
  method: GET
  url: https://example.com/protected-resource
  headers:
    X-Custom-Session: "your-session-value"
    X-API-Version: "v1"
  expect_status: 200
```

### Session Cleanup

Add logout actions to clean up server sessions:

```yaml
- name: Logout
  method: DELETE
  url: https://example.com/users/sign_out
  headers:
    X-CSRF-Token: CSRF_TOKEN_PLACEHOLDER
  expect_status: 302
```

This setup ensures realistic session management for your load tests! 