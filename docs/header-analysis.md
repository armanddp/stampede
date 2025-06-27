# Header Analysis Tool

This tool analyzes browser recordings to identify important headers for session management and authentication.

## Overview

The header analysis tool helps you:
- Extract required headers from browser recordings
- Identify session cookies and CSRF tokens
- Generate YAML templates for load testing
- Understand authentication flows

## Usage

### Basic Usage

```bash
# Analyze a browser recording
make analyze-headers FILE=your-recording.json

# Or run directly
go run tools/analyze-headers.go your-recording.json
```

### Example Output

```
Analyzing browser recording: your-recording.json

Headers to persist across requests:
- Cookie: _session_id=abc123...
- X-CSRF-Token: xyz789...
- Authorization: Bearer token123...

Generated YAML template:
- name: LoginPage
  method: GET
  url: https://example.com/users/sign_in
  headers:
    User-Agent: "Mozilla/5.0..."
    Accept: "text/html,application/xhtml+xml..."
```

## Browser Recording Format

The tool expects a JSON file with browser network requests in this format:

```json
[
  {
    "method": "GET",
    "url": "https://example.com/login",
    "headers": {
      "User-Agent": "Mozilla/5.0...",
      "Accept": "text/html...",
      "Cookie": "_session_id=abc123..."
    },
    "response": {
      "status": 200,
      "headers": {
        "Set-Cookie": "_session_id=def456...",
        "X-CSRF-Token": "xyz789..."
      }
    }
  }
]
```

## Manual Testing

Before running load tests, verify the authentication flow manually:

```bash
# Test basic connectivity
curl -v -H "User-Agent: Stampede-Shooter/1.0" https://example.com/

# Test login page access
curl -c cookies.txt -b cookies.txt https://example.com/users/sign_in

# Check for CSRF token in response
grep -o 'csrf-token.*content="[^"]*"' response.html

# Test authenticated request
curl -b cookies.txt https://example.com/dashboard
```

## Common Headers to Persist

### Session Cookies
- `_session_id` - Rails session cookie
- `_csrf_token` - CSRF protection token
- `remember_token` - Remember me functionality

### Authentication Headers
- `Authorization` - Bearer tokens or API keys
- `X-CSRF-Token` - Rails CSRF protection
- `X-API-Key` - API authentication

### Browser Headers
- `User-Agent` - Browser identification
- `Accept` - Content type preferences
- `Accept-Language` - Language preferences

## Troubleshooting

### No Headers Found
- Ensure the browser recording includes network requests
- Check that authentication requests are captured
- Verify the JSON format is correct

### Missing CSRF Tokens
- Look for `X-CSRF-Token` in response headers
- Check for `csrf-token` meta tags in HTML
- Verify form inputs with `authenticity_token`

### Session Issues
- Check for `Set-Cookie` headers in responses
- Verify cookie domains and paths
- Ensure cookies are being sent in subsequent requests 