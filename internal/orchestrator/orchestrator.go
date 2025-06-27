package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"

	"stampede-shooter/internal/config"
	"stampede-shooter/internal/metrics"
	"stampede-shooter/internal/reporter"
	"stampede-shooter/internal/script"
	"stampede-shooter/internal/util"
	"stampede-shooter/internal/worker"
)

// Orchestrator coordinates the load test execution
type Orchestrator struct {
	cfg         config.Config
	script      *script.Script
	collector   *metrics.Collector
	reporter    *reporter.Reporter
	credentials *util.CredentialsManager
}

// New creates a new orchestrator
func New(cfg config.Config) (*Orchestrator, error) {
	// Load test script
	script, err := script.LoadScript(cfg.ScriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load script: %w", err)
	}

	// Load credentials if provided
	var credentials *util.CredentialsManager
	if cfg.CredentialsFile != "" {
		credentials, err = util.LoadCredentials(cfg.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w", err)
		}

		// Validate we have enough credentials
		if err := credentials.Validate(cfg.Users); err != nil {
			log.Printf("Warning: %v", err)
		} else {
			log.Printf("Loaded %d credentials for %d users", credentials.Count(), cfg.Users)
		}
	}

	// Create metrics collector
	collector := metrics.NewCollector()

	// Create reporter
	reporter := reporter.New(collector, cfg.Verbose)

	return &Orchestrator{
		cfg:         cfg,
		script:      script,
		collector:   collector,
		reporter:    reporter,
		credentials: credentials,
	}, nil
}

// Run executes the load test
func (o *Orchestrator) Run() error {
	log.Printf("Starting load test with %d users for %v...", o.cfg.Users, o.cfg.Duration)
	log.Printf("Loaded script with %d actions", len(o.script.Actions))

	if o.credentials != nil {
		log.Printf("Using credentials from: %s (%d available)", o.cfg.CredentialsFile, o.credentials.Count())
	}

	// Start metrics collector
	o.collector.Start()
	defer o.collector.Stop()

	// Start live reporter
	o.reporter.StartLiveReporting()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), o.cfg.Duration)
	defer cancel()

	// Start workers
	log.Printf("Starting %d workers...", o.cfg.Users)

	var wg sync.WaitGroup
	for i := 0; i < o.cfg.Users; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Create worker with credentials
			w := worker.New(userID, o.cfg, o.script, o.collector, o.credentials)

			// Run worker
			if err := w.Run(ctx, o.cfg.LoginURL); err != nil {
				log.Printf("Worker %d error: %v", userID, err)
			}
		}(i + 1) // User IDs start from 1
	}

	// Wait for test duration or context cancellation
	<-ctx.Done()

	log.Println("Test completed, waiting for workers to finish...")

	// Wait for all workers to finish
	wg.Wait()

	// Generate final report
	o.reporter.PrintFinalReport()

	// Save results if output file specified
	if o.cfg.OutputFile != "" {
		if err := o.reporter.SaveReport(o.cfg.OutputFile); err != nil {
			return fmt.Errorf("failed to save results: %w", err)
		}
		log.Printf("Results saved to: %s", o.cfg.OutputFile)
	}

	return nil
}
