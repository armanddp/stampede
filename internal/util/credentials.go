package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Credentials represents a username/password pair
type Credentials struct {
	Username string
	Password string
}

// CredentialsManager handles loading and round-robin assignment of credentials
type CredentialsManager struct {
	credentials []Credentials
	mu          sync.Mutex
	current     int
}

// LoadCredentials loads credentials from a file
func LoadCredentials(filepath string) (*CredentialsManager, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open credentials file: %w", err)
	}
	defer file.Close()

	var credentials []Credentials
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse username,password format
		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid credentials format on line %d: expected 'username,password', got '%s'", lineNum, line)
		}

		username := strings.TrimSpace(parts[0])
		password := strings.TrimSpace(parts[1])

		if username == "" || password == "" {
			return nil, fmt.Errorf("empty username or password on line %d", lineNum)
		}

		credentials = append(credentials, Credentials{
			Username: username,
			Password: password,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading credentials file: %w", err)
	}

	if len(credentials) == 0 {
		return nil, fmt.Errorf("no valid credentials found in file")
	}

	return &CredentialsManager{
		credentials: credentials,
		current:     0,
	}, nil
}

// GetCredentials returns the next credentials in round-robin fashion
func (cm *CredentialsManager) GetCredentials() Credentials {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	creds := cm.credentials[cm.current]
	cm.current = (cm.current + 1) % len(cm.credentials)
	return creds
}

// GetCredentialsForUser returns credentials for a specific user ID
func (cm *CredentialsManager) GetCredentialsForUser(userID int) Credentials {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	index := userID % len(cm.credentials)
	return cm.credentials[index]
}

// Count returns the number of available credentials
func (cm *CredentialsManager) Count() int {
	return len(cm.credentials)
}

// Validate checks if we have enough credentials for the requested number of users
func (cm *CredentialsManager) Validate(userCount int) error {
	if userCount > len(cm.credentials) {
		return fmt.Errorf("requested %d users but only %d credentials available. Users will be assigned credentials in round-robin fashion", userCount, len(cm.credentials))
	}
	return nil
}
