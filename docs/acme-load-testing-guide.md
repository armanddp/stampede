# Rails Application Load Testing Guide

This guide shows how to use Stampede Shooter to load test Rails applications that use Devise authentication with CSRF protection.

## 🎯 **What We're Testing**

The example Rails application allows users to:
1. **Login** via Rails Devise authentication
2. **Browse events** and view event listings
3. **Watch events** (streaming video content)
4. **Access videos** and other authenticated content

## 🔧 **Enhanced Features for Rails**

Stampede Shooter now includes:

- ✅ **Automatic CSRF Token Extraction** from HTML responses
- ✅ **Rails Devise Session Management** with cookie persistence
- ✅ **Realistic Browser Headers** matching Chrome/Firefox
- ✅ **Template Variable Support** for unique user credentials
- ✅ **Session State Persistence** across requests
- ✅ **Credentials File Support** with round-robin assignment

## 📋 **Step-by-Step Setup**

### Step 1: Create Credentials File

Create a text file with your test credentials in `username,password` format:

```bash
# examples/credentials.txt
user1@example.com,password123
user2@example.com,password123
user3@example.com,password123
test-user-1@example.com,secure_password_123
test-user-2@example.com,secure_password_123
```

**File Format:**
- One credential pair per line: `username,password`
- Lines starting with `#` are comments and ignored
- Empty lines are ignored
- Usernames can be email addresses for Rails apps

### Step 2: Analyze Your Browser Session

```bash
# Save your browser recording as browser-recording.json
make analyze-headers FILE=browser-recording.json
```

This will show you:
- Required headers for your application
- CSRF token requirements
- Session management needs

### Step 3: Test Manual Authentication

Before running load tests, verify the authentication flow works:

```bash
# Test basic connectivity
curl -v -H "User-Agent: Stampede-Shooter/1.0" https://your-app.com/

# Test login page access
curl -c cookies.txt -v \
  -H "User-Agent: Stampede-Shooter/1.0" \
  https://your-app.com/users/sign_in

# Check for CSRF token in response
grep -o 'csrf-token.*content="[^"]*"' response.html
```

### Step 4: Configure Your Test Script

Use the provided `examples/acme-rails-auth.yml` as a template. The script now uses credential placeholders:

```yaml
- name: LoginSubmit
  method: POST
  url: https://your-app.com/users/sign_in
  body: |
    user%5Bemail%5D={{username}}&user%5Bpassword%5D={{password}}&authenticity_token=CSRF_TOKEN_PLACEHOLDER
```

**Credential Placeholders:**
- `{{username}}` - Replaced with username from credentials file
- `{{password}}` - Replaced with password from credentials file
- `{{email}}` - Also replaced with username (for Rails apps)

## 🚀 **Running Load Tests**

### Quick Test with Credentials (3 users, 10 seconds)
```bash
./build/stampede-shooter \
  --users 3 \
  --rps 1 \
  --duration 10s \
  --script examples/acme-demo.yml \
  --credentials examples/credentials.txt \
  --verbose
```

### Real Application Test with Credentials (5 users, 1 minute)
```bash
./build/stampede-shooter \
  --users 5 \
  --rps 2 \
  --duration 60s \
  --script examples/acme-rails-auth.yml \
  --credentials examples/credentials.txt \
  --out test-results.json \
  --verbose
```

### Production Load Test (20 users, 5 minutes)
```bash
./build/stampede-shooter \
  --users 20 \
  --rps 3 \
  --duration 300s \
  --script examples/acme-rails-auth.yml \
  --credentials examples/credentials.txt \
  --out production-test.json
```

### Using Makefile Targets
```bash
# Test credentials feature
make test-credentials

# Test real application with credentials
make test-acme-with-creds
```

## 🔄 **Round-Robin Credential Assignment**

The tool automatically assigns credentials to users in round-robin fashion:

- **User 1** gets credentials from line 1
- **User 2** gets credentials from line 2
- **User 3** gets credentials from line 3
- **User 4** gets credentials from line 1 (wraps around)
- And so on...

This ensures:
- ✅ Each user has unique credentials per test run
- ✅ Credentials are reused efficiently
- ✅ No hardcoded credentials in scripts
- ✅ Easy credential management

## 📊 **Monitoring Test Results**

### Good Session Management with Credentials
```
Action        OK   ERR   p50   p90   p99   RPS
──────────── ──── ──── ───── ───── ───── ────
LoginSubmit   50    0   45ms  89ms  156ms  10
Dashboard     50    0   67ms  145ms 289ms  10
ViewEvent     50    0   78ms  167ms 334ms  10
WatchEvent    50    0   123ms 245ms 456ms  10
```

### Credential Issues (High Error Rates)
```
Action        OK   ERR   p50   p90   p99   RPS
──────────── ──── ──── ───── ───── ───── ────
LoginSubmit    25   25   45ms  89ms  156ms   5  # 50% login failures
Dashboard       0   50    0ms   0ms   0ms   0  # 100% dashboard failures
```

## 🔍 **Key Headers for Rails Applications**

### Required Headers
```yaml
headers:
  User-Agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
  Accept: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
  Accept-Language: "en-US,en;q=0.9"
  Accept-Encoding: "gzip, deflate, br"
  DNT: "1"
  Connection: "keep-alive"
  Upgrade-Insecure-Requests: "1"
```

### Session Headers (Auto-managed)
- `Cookie` - Session cookies (automatic via cookie jar)
- `X-CSRF-Token` - Rails CSRF protection (auto-extracted)
- `Referer` - Page navigation flow (set per request)

## 🚨 **Common Issues & Solutions**

### Issue: "CSRF Token Mismatch" (422 errors)
**Symptoms:** High error rate on login requests
**Solution:**
1. Ensure login page is fetched first
2. Check CSRF token extraction is working
3. Verify token is included in login form data

### Issue: "Session Expired" (401/403 errors)
**Symptoms:** Dashboard/events returning auth errors
**Solution:**
1. Check cookie persistence is working
2. Verify session timeout settings
3. Monitor for session invalidation

### Issue: "Invalid Credentials" (401 errors)
**Symptoms:** Login requests returning 401
**Solution:**
1. Verify credentials in file are correct
2. Check username/password format
3. Ensure no extra spaces in credentials file

### Issue: "Too Many Failed Logins"
**Symptoms:** Login requests returning 422/429
**Solution:**
1. Use unique credentials per user in file
2. Reduce RPS for login requests
3. Add delays between login attempts

### Issue: "Event Not Found" (404 errors)
**Symptoms:** Event pages returning 404
**Solution:**
1. Verify event IDs exist in the system
2. Check user permissions for events
3. Use dynamic event selection: `{{randInt 1 100}}`

## 🎯 **Best Practices for Rails Applications**

### 1. **Start Small**
```bash
# Begin with 1-2 users to verify authentication
./build/stampede-shooter --users 1 --script examples/acme-demo.yml --credentials examples/credentials.txt --duration 30s --verbose
```

### 2. **Use Realistic Credentials**
```bash
# In your credentials.txt file:
test-user-1@yourdomain.com,secure_password_123
test-user-2@yourdomain.com,secure_password_123
test-user-3@yourdomain.com,secure_password_123
```

### 3. **Monitor Resource Usage**
- Watch CPU/memory on both client and server
- Monitor database connection pools
- Check video streaming bandwidth

### 4. **Test Different Scenarios**
```bash
# Test login flow only
./build/stampede-shooter --users 10 --script login-only.yml --credentials examples/credentials.txt --duration 60s

# Test event watching
./build/stampede-shooter --users 5 --script event-watching.yml --credentials examples/credentials.txt --duration 120s

# Test mixed workload
./build/stampede-shooter --users 20 --script mixed-workload.yml --credentials examples/credentials.txt --duration 300s
```

## 📈 **Scaling Guidelines**

### Conservative Scaling
- **Start:** 1-2 users, 1 RPS per user
- **Small:** 5-10 users, 2 RPS per user  
- **Medium:** 20-50 users, 3 RPS per user
- **Large:** 100+ users, 5 RPS per user

### Monitoring Points
- **Login Success Rate:** Should be >95%
- **Session Persistence:** Dashboard access should work
- **Event Loading:** Video pages should load successfully
- **Response Times:** p95 should be <2s for most requests

## 🔧 **Advanced Configuration**

### Custom CSRF Token Extraction
If your Rails app uses a different CSRF token format, modify `internal/worker/worker.go`:

```go
// Add custom extraction patterns
customPattern := regexp.MustCompile(`your-custom-pattern`)
if matches := customPattern.FindStringSubmatch(htmlContent); len(matches) > 1 {
    w.csrfToken = matches[1]
}
```

### Session Cleanup
Add logout actions to clean up server sessions:

```yaml
- name: Logout
  method: DELETE
  url: https://your-app.com/users/sign_out
  headers:
    X-CSRF-Token: CSRF_TOKEN_PLACEHOLDER
  expect_status: 302
```

### Multiple Credential Files
For different test scenarios, create multiple credential files:

```bash
# production-users.txt
prod-user-1@yourdomain.com,prod_password_123
prod-user-2@yourdomain.com,prod_password_123

# staging-users.txt
stage-user-1@yourdomain.com,stage_password_123
stage-user-2@yourdomain.com,stage_password_123
```

## 📚 **Troubleshooting**

### Debug Mode
Enable verbose output to see detailed request/response info:
```bash
./build/stampede-shooter --users 1 --script debug.yml --credentials examples/credentials.txt --duration 10s --verbose
```

### Check JSON Results
Examine the detailed JSON output for per-request analysis:
```bash
cat test-results.json | jq '.actions.LoginSubmit'
```

### Monitor Server Logs
Watch Rails logs for authentication and session issues:
```bash
tail -f log/production.log | grep -E "(authentication|session|csrf)"
```

### Validate Credentials File
Check your credentials file format:
```bash
# Check for syntax errors
cat examples/credentials.txt | grep -v '^#' | grep -v '^$' | while IFS=',' read -r user pass; do
  echo "Username: $user, Password: $pass"
done
```

## 🎉 **Success Metrics**

Your load test is successful when:
- ✅ Login success rate >95%
- ✅ Session persistence works across requests
- ✅ Event pages load successfully
- ✅ Response times are acceptable (p95 <2s)
- ✅ No 422 CSRF errors
- ✅ No 401/403 authentication errors
- ✅ Credentials are properly assigned to users

This setup will give you realistic load testing of Rails applications with proper session management, authentication handling, and credential management! 