package script

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Action represents a single HTTP action in the test script
type Action struct {
	Name         string            `yaml:"name"`
	Method       string            `yaml:"method"`
	URL          string            `yaml:"url"`
	JSONBody     string            `yaml:"json_body"`
	Body         string            `yaml:"body"`
	Headers      map[string]string `yaml:"headers"`
	ExpectStatus int               `yaml:"expect_status"`
	Timeout      string            `yaml:"timeout"`
	Delay        string            `yaml:"delay"`     // Fixed delay (e.g., "2s", "500ms")
	DelayMin     string            `yaml:"delay_min"` // Minimum random delay
	DelayMax     string            `yaml:"delay_max"` // Maximum random delay
}

// Script holds the parsed test script
type Script struct {
	Actions []Action
}

// LoadScript loads and parses a YAML script file
func LoadScript(filename string) (*Script, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read script file: %w", err)
	}

	var actions []Action
	if err := yaml.Unmarshal(data, &actions); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &Script{Actions: actions}, nil
}

// ExpandTemplates replaces template variables in the action
func (a *Action) ExpandTemplates(userID int) Action {
	expanded := *a

	// Replace template variables in URL
	expanded.URL = expandString(a.URL, userID)

	// Replace template variables in JSON body
	expanded.JSONBody = expandString(a.JSONBody, userID)

	// Replace template variables in body
	expanded.Body = expandString(a.Body, userID)

	// Replace template variables in headers
	expanded.Headers = make(map[string]string)
	for key, value := range a.Headers {
		expanded.Headers[key] = expandString(value, userID)
	}

	return expanded
}

// expandString processes template variables in a string
func expandString(s string, userID int) string {
	result := s

	// Replace {{userId}} with the actual user ID
	result = strings.ReplaceAll(result, "{{userId}}", strconv.Itoa(userID))

	// Replace {{epochms}} with current timestamp in milliseconds
	result = strings.ReplaceAll(result, "{{epochms}}", strconv.FormatInt(time.Now().UnixMilli(), 10))

	// Handle {{randInt min max}} - find and replace all occurrences
	for strings.Contains(result, "{{randInt") {
		start := strings.Index(result, "{{randInt")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2

		// Extract the randInt expression
		expr := result[start:end]
		parts := strings.Fields(expr[9 : len(expr)-2]) // Remove {{randInt and }}

		if len(parts) == 2 {
			min, err1 := strconv.Atoi(parts[0])
			max, err2 := strconv.Atoi(parts[1])

			if err1 == nil && err2 == nil && max > min {
				randVal := rand.Intn(max-min+1) + min
				result = result[:start] + strconv.Itoa(randVal) + result[end:]
			} else {
				// If parsing fails, just remove the template
				result = result[:start] + "1" + result[end:]
			}
		} else {
			// Invalid format, replace with 1
			result = result[:start] + "1" + result[end:]
		}
	}

	// Handle {{pick movies}} - simple implementation that picks from a predefined list
	movieList := []string{"movie1", "movie2", "movie3", "movie4", "movie5"}
	if strings.Contains(result, "{{pick movies}}") {
		picked := movieList[rand.Intn(len(movieList))]
		result = strings.ReplaceAll(result, "{{pick movies}}", picked)
	}

	// Handle {{randDelay min max}} - generates random delay in milliseconds
	for strings.Contains(result, "{{randDelay") {
		start := strings.Index(result, "{{randDelay")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2

		// Extract the randDelay expression
		expr := result[start:end]
		parts := strings.Fields(expr[11 : len(expr)-2]) // Remove {{randDelay and }}

		if len(parts) == 2 {
			min, err1 := strconv.Atoi(parts[0])
			max, err2 := strconv.Atoi(parts[1])

			if err1 == nil && err2 == nil && max > min {
				randVal := rand.Intn(max-min+1) + min
				result = result[:start] + strconv.Itoa(randVal) + result[end:]
			} else {
				// If parsing fails, just remove the template
				result = result[:start] + "1000" + result[end:]
			}
		} else {
			// Invalid format, replace with 1000ms
			result = result[:start] + "1000" + result[end:]
		}
	}

	return result
}

// GetDelay calculates the delay duration for this action
func (a *Action) GetDelay() time.Duration {
	// If fixed delay is specified, use it
	if a.Delay != "" {
		if delay, err := time.ParseDuration(a.Delay); err == nil {
			return delay
		}
	}

	// If random delay range is specified, pick a random value
	if a.DelayMin != "" && a.DelayMax != "" {
		minDelay, err1 := time.ParseDuration(a.DelayMin)
		maxDelay, err2 := time.ParseDuration(a.DelayMax)

		if err1 == nil && err2 == nil && maxDelay > minDelay {
			// Convert to nanoseconds for random calculation
			minNanos := minDelay.Nanoseconds()
			maxNanos := maxDelay.Nanoseconds()
			randomNanos := rand.Int63n(maxNanos-minNanos+1) + minNanos
			return time.Duration(randomNanos)
		}
	}

	// No delay specified
	return 0
}
