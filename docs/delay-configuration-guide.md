# Delay Configuration Guide

This guide explains how to add realistic delays between requests in your load testing scripts to simulate real user behavior.

## Overview

Real users don't make requests instantly - they take time to read pages, think, and interact with the interface. Adding delays between requests makes your load tests more realistic and helps identify performance issues that only occur under realistic usage patterns.

## Delay Options

### 1. Fixed Delays

Use a `delay` field to add a consistent delay after each action:

```yaml
- name: HomePage
  method: GET
  url: https://example.com/
  expect_status: 200
  delay: "2s"  # Wait 2 seconds after this request
```

**Supported time units:**
- `ms` - milliseconds (e.g., "500ms")
- `s` - seconds (e.g., "2s", "1.5s")
- `m` - minutes (e.g., "1m")

### 2. Random Delays

Use `delay_min` and `delay_max` to add random delays within a range:

```yaml
- name: BrowseProducts
  method: GET
  url: https://example.com/products?page={{randInt 1 5}}
  expect_status: 200
  delay_min: "1s"    # Minimum delay
  delay_max: "5s"    # Maximum delay
```

This will wait between 1-5 seconds randomly after the request.

### 3. Template Variable Delays

Use the `{{randDelay min max}}` template variable for inline random delays:

```yaml
- name: ViewProduct
  method: GET
  url: https://example.com/products/{{randInt 1 100}}
  expect_status: 200
  delay: "{{randDelay 1000 3000}}ms"  # Random 1-3 second delay
```

## Realistic User Patterns

### E-commerce User Journey

```yaml
- name: HomePage
  method: GET
  url: https://shop.example.com/
  expect_status: 200
  delay: "3s"  # Time to read homepage

- name: BrowseCategory
  method: GET
  url: https://shop.example.com/category/electronics
  expect_status: 200
  delay_min: "5s"    # Time to browse products
  delay_max: "15s"

- name: ViewProduct
  method: GET
  url: https://shop.example.com/product/{{randInt 1 100}}
  expect_status: 200
  delay_min: "10s"   # Time to read product details
  delay_max: "30s"

- name: AddToCart
  method: POST
  url: https://shop.example.com/cart/add
  json_body: |
    {
      "product_id": {{randInt 1 100}},
      "quantity": {{randInt 1 3}}
    }
  expect_status: 200
  delay: "2s"  # Brief delay after adding to cart

- name: Checkout
  method: GET
  url: https://shop.example.com/checkout
  expect_status: 200
  delay_min: "20s"   # Time to fill checkout form
  delay_max: "60s"
```

### Content Browsing User

```yaml
- name: Login
  method: POST
  url: https://content.example.com/login
  body: "username={{username}}&password={{password}}"
  expect_status: 200
  delay: "1s"

- name: Dashboard
  method: GET
  url: https://content.example.com/dashboard
  expect_status: 200
  delay_min: "2s"    # Time to scan dashboard
  delay_max: "8s"

- name: ReadArticle
  method: GET
  url: https://content.example.com/articles/{{randInt 1 50}}
  expect_status: 200
  delay_min: "30s"   # Time to read article
  delay_max: "120s"

- name: Comment
  method: POST
  url: https://content.example.com/comments
  json_body: |
    {
      "article_id": {{randInt 1 50}},
      "comment": "Great article!"
    }
  expect_status: 200
  delay_min: "5s"    # Time to write comment
  delay_max: "15s"
```

## Best Practices

### 1. Match Real User Behavior

- **Homepage**: 2-5 seconds (reading content)
- **Product browsing**: 5-15 seconds per page
- **Product details**: 10-30 seconds (reading specs, reviews)
- **Form filling**: 20-60 seconds (checkout, registration)
- **Content reading**: 30-120 seconds (articles, documentation)

### 2. Vary Delays by Action Type

```yaml
# Quick actions
- name: ClickButton
  method: POST
  url: https://example.com/api/click
  delay: "500ms"

# Reading actions
- name: ViewPage
  method: GET
  url: https://example.com/page
  delay_min: "5s"
  delay_max: "20s"

# Complex actions
- name: SubmitForm
  method: POST
  url: https://example.com/submit
  delay_min: "10s"
  delay_max: "45s"
```

### 3. Consider Page Complexity

- **Simple pages** (login, error pages): 1-3 seconds
- **Medium pages** (product lists, search results): 3-10 seconds
- **Complex pages** (product details, forms): 10-30 seconds
- **Content-heavy pages** (articles, documentation): 30+ seconds

### 4. Account for User Experience

```yaml
# First-time user (slower)
- name: FirstVisit
  method: GET
  url: https://example.com/
  delay_min: "10s"
  delay_max: "30s"

# Returning user (faster)
- name: ReturnVisit
  method: GET
  url: https://example.com/
  delay_min: "2s"
  delay_max: "8s"
```

## Testing Different Scenarios

### High-Traffic Scenario

```yaml
# Fast-paced browsing for stress testing
- name: QuickBrowse
  method: GET
  url: https://example.com/products
  delay: "1s"

- name: QuickView
  method: GET
  url: https://example.com/product/{{randInt 1 100}}
  delay: "2s"
```

### Realistic Load Scenario

```yaml
# Normal user behavior for capacity testing
- name: NormalBrowse
  method: GET
  url: https://example.com/products
  delay_min: "5s"
  delay_max: "15s"

- name: NormalView
  method: GET
  url: https://example.com/product/{{randInt 1 100}}
  delay_min: "10s"
  delay_max: "25s"
```

## Monitoring Delay Impact

### Check Response Times

With delays, you'll see more realistic response time patterns:

```
Action        OK   ERR   p50   p90   p99   RPS
──────────── ──── ──── ───── ───── ───── ────
HomePage      100    0   45ms  89ms  156ms   0.2  # Lower RPS due to delays
BrowseProducts 100    0   67ms  145ms 289ms  0.1
ViewProduct   100    0   78ms  167ms 334ms  0.05
```

### Server Resource Usage

Delays help identify:
- **Memory leaks** - Resources not released between requests
- **Connection pooling** - How well connections are reused
- **Database connection limits** - Impact of longer sessions
- **Cache effectiveness** - How caching performs with realistic timing

## Troubleshooting

### Delays Too Long

If your test is taking too long:

```yaml
# Reduce delay ranges
delay_min: "1s"    # Instead of "5s"
delay_max: "3s"    # Instead of "15s"
```

### Delays Too Short

If you need more realistic timing:

```yaml
# Increase delay ranges
delay_min: "10s"   # More realistic reading time
delay_max: "30s"   # Account for user thinking time
```

### Mixed Delay Strategies

You can combine different delay approaches:

```yaml
- name: QuickAction
  method: GET
  url: https://example.com/api/status
  delay: "100ms"  # Fixed quick delay

- name: UserAction
  method: POST
  url: https://example.com/api/submit
  delay_min: "5s"   # Random realistic delay
  delay_max: "15s"

- name: TemplateDelay
  method: GET
  url: https://example.com/page
  delay: "{{randDelay 2000 8000}}ms"  # Template variable delay
```

## Advanced Usage

### Conditional Delays

Use different delays based on response status:

```yaml
- name: SearchProducts
  method: GET
  url: https://example.com/search?q={{randInt 1 100}}
  expect_status: 200
  delay: "3s"  # Normal delay for successful search

# If search fails, user might try again quickly
- name: RetrySearch
  method: GET
  url: https://example.com/search?q={{randInt 1 100}}
  expect_status: 200
  delay: "1s"  # Quick retry delay
```

### Session-Based Delays

Simulate different user types:

```yaml
# New user (slower)
- name: NewUserHomepage
  method: GET
  url: https://example.com/
  delay_min: "10s"
  delay_max: "25s"

# Experienced user (faster)
- name: ExperiencedUserHomepage
  method: GET
  url: https://example.com/
  delay_min: "2s"
  delay_max: "8s"
```

This delay configuration system allows you to create highly realistic load tests that accurately simulate real user behavior! 