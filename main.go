package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gosom/google-maps-scraper/scraper"
)

const (
	defaultConcurrency = 3
	defaultDepth       = 20
	defaultLang        = "en"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var cfg scraper.Config

	flag.StringVar(&cfg.InputFile, "input", "", "path to input file with search queries (one per line)")
	flag.StringVar(&cfg.OutputFile, "output", "output.csv", "path to output CSV file")
	flag.IntVar(&cfg.Concurrency, "concurrency", defaultConcurrency, "number of concurrent scrapers")
	flag.IntVar(&cfg.Depth, "depth", defaultDepth, "maximum number of results per query")
	flag.StringVar(&cfg.Lang, "lang", defaultLang, "language code for Google Maps results")
	flag.BoolVar(&cfg.Debug, "debug", false, "enable debug logging")
	flag.BoolVar(&cfg.JSON, "json", false, "output results as JSON instead of CSV")
	flag.Parse()

	// If no input file provided, check for positional arguments
	if cfg.InputFile == "" && flag.NArg() > 0 {
		cfg.InputFile = flag.Arg(0)
	}

	if cfg.InputFile == "" {
		flag.Usage()
		return fmt.Errorf("input file is required")
	}

	// Set up context with cancellation on OS signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			fmt.Fprintf(os.Stderr, "received signal %s, shutting down...\n", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	s, err := scraper.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create scraper: %w", err)
	}
	defer s.Close()

	if err := s.Run(ctx); err != nil {
		return fmt.Errorf("scraper run failed: %w", err)
	}

	return nil
}
