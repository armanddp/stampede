package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// BrowserRecording represents the structure of a Chrome DevTools recording
type BrowserRecording struct {
	Title string `json:"title"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Type     string                 `json:"type"`
	URL      string                 `json:"url,omitempty"`
	Method   string                 `json:"method,omitempty"`
	Headers  map[string]string      `json:"headers,omitempty"`
	PostData map[string]interface{} `json:"postData,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run analyze-headers.go <recording.json>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Read the recording file
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var recording BrowserRecording
	if err := json.Unmarshal(data, &recording); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Printf("Analyzing recording: %s\n", recording.Title)
	fmt.Println("=" + strings.Repeat("=", len(recording.Title)+18))

	// Track important headers
	sessionHeaders := make(map[string]bool)
	authHeaders := make(map[string]bool)
	csrfHeaders := make(map[string]bool)

	// Analyze navigation steps
	var navigationSteps []Step
	for _, step := range recording.Steps {
		if step.Type == "navigate" {
			navigationSteps = append(navigationSteps, step)
		}
	}

	fmt.Printf("\n🌐 Navigation Flow (%d steps):\n", len(navigationSteps))
	for i, step := range navigationSteps {
		fmt.Printf("  %d. %s\n", i+1, step.URL)
	}

	// Provide recommendations
	fmt.Println("\n🔍 Header Analysis Recommendations:")
	fmt.Println("=====================================")

	// Check for common authentication patterns
	hasLogin := false
	for _, step := range navigationSteps {
		if strings.Contains(strings.ToLower(step.URL), "login") ||
			strings.Contains(strings.ToLower(step.URL), "sign_in") ||
			strings.Contains(strings.ToLower(step.URL), "auth") {
			hasLogin = true
			break
		}
	}

	if hasLogin {
		fmt.Println("✅ Login flow detected - Session management required")
		fmt.Println("   Headers to monitor:")
		fmt.Println("   - Cookie (session cookies)")
		fmt.Println("   - X-CSRF-Token (Rails CSRF protection)")
		fmt.Println("   - Set-Cookie (response headers)")
		sessionHeaders["Cookie"] = true
		csrfHeaders["X-CSRF-Token"] = true
	}

	// Check for API-like URLs
	hasAPI := false
	for _, step := range navigationSteps {
		if strings.Contains(step.URL, "/api/") ||
			strings.Contains(step.URL, ".json") {
			hasAPI = true
			break
		}
	}

	if hasAPI {
		fmt.Println("✅ API endpoints detected")
		fmt.Println("   Headers to monitor:")
		fmt.Println("   - Authorization (Bearer tokens)")
		fmt.Println("   - Content-Type (application/json)")
		authHeaders["Authorization"] = true
	}

	// Provide curl command template
	fmt.Println("\n🧪 Testing Commands:")
	fmt.Println("====================")

	if len(navigationSteps) > 0 {
		firstURL := navigationSteps[0].URL
		fmt.Printf("# Test basic connectivity:\n")
		fmt.Printf("curl -v -H \"User-Agent: Stampede-Shooter/1.0\" \"%s\"\n\n", firstURL)

		if hasLogin {
			fmt.Printf("# Test with session persistence:\n")
			fmt.Printf("curl -c cookies.txt -b cookies.txt -H \"User-Agent: Stampede-Shooter/1.0\" \"%s\"\n\n", firstURL)
		}
	}

	// Generate YAML template
	fmt.Println("📝 YAML Template:")
	fmt.Println("=================")

	for i, step := range navigationSteps {
		actionName := fmt.Sprintf("Step%d", i+1)
		if strings.Contains(step.URL, "login") || strings.Contains(step.URL, "sign_in") {
			actionName = "Login"
		} else if strings.Contains(step.URL, "dashboard") {
			actionName = "Dashboard"
		} else if i == 0 {
			actionName = "HomePage"
		}

		fmt.Printf("- name: %s\n", actionName)
		fmt.Printf("  method: GET\n")
		fmt.Printf("  url: %s\n", step.URL)
		fmt.Printf("  headers:\n")
		fmt.Printf("    User-Agent: \"Mozilla/5.0 (compatible; Stampede-Shooter/1.0)\"\n")
		fmt.Printf("    Accept: \"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\"\n")

		if i > 0 {
			fmt.Printf("    Referer: \"%s\"\n", navigationSteps[i-1].URL)
		}

		fmt.Printf("  expect_status: 200\n\n")
	}

	// Summary
	fmt.Println("💡 Key Recommendations:")
	fmt.Println("=======================")
	fmt.Println("1. Enable cookie persistence in Stampede Shooter")
	fmt.Println("2. Monitor response headers for CSRF tokens")
	fmt.Println("3. Test with low user count first (1-2 users)")
	fmt.Println("4. Use consistent User-Agent strings")
	if hasLogin {
		fmt.Println("5. Handle login failures gracefully")
		fmt.Println("6. Consider session timeout handling")
	}
}
