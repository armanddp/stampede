package main

import (
	"log"

	"stampede-shooter/internal/config"
	"stampede-shooter/internal/orchestrator"
)

func main() {
	// Parse configuration
	cfg := config.Parse()

	// Validate required parameters
	if cfg.ScriptPath == "" {
		log.Fatal("--script parameter is required")
	}

	// Create and run orchestrator
	o, err := orchestrator.New(*cfg)
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
	}

	if err := o.Run(); err != nil {
		log.Fatalf("Test failed: %v", err)
	}
}
