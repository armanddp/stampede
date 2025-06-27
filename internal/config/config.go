package config

import (
	"flag"
	"time"
)

// Config holds all configuration for the load test
type Config struct {
	Users           int           `json:"users"`
	RPS             int           `json:"rps"`
	Duration        time.Duration `json:"duration"`
	ScriptPath      string        `json:"script_path"`
	LoginURL        string        `json:"login_url"`
	LoginHeader     string        `json:"login_header"`
	OutputFile      string        `json:"output_file"`
	Verbose         bool          `json:"verbose"`
	InsecureTLS     bool          `json:"insecure_tls"`
	CredentialsFile string        `json:"credentials_file"`
}

// Parse parses command line flags into config
func Parse() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.Users, "users", 10, "Number of concurrent users")
	flag.IntVar(&cfg.RPS, "rps", 1, "Requests per second per user")
	flag.DurationVar(&cfg.Duration, "duration", 30*time.Second, "Test duration")
	flag.StringVar(&cfg.ScriptPath, "script", "", "Path to test script (required)")
	flag.StringVar(&cfg.LoginURL, "login-url", "", "Optional login endpoint URL")
	flag.StringVar(&cfg.LoginHeader, "login-hdr", "", "Authentication header (format: key:value)")
	flag.StringVar(&cfg.OutputFile, "out", "", "Output file for JSON results")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Show live progress updates")
	flag.BoolVar(&cfg.InsecureTLS, "insecure-tls", false, "Skip TLS certificate verification")
	flag.StringVar(&cfg.CredentialsFile, "credentials", "", "Path to credentials file (format: username,password)")

	flag.Parse()

	return cfg
}
