package main

import (
	"context"
	"fmt"
	"go-env-multipaths-scanner/app"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	opts := app.ParseFlags()

	paths, err := app.LoadPaths(opts.PathsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading paths: %v\n", err)
		os.Exit(1)
	}

	urls, err := app.LoadPaths(opts.TargetsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading targets: %v\n", err)
		os.Exit(1)
	}

	writer, err := app.NewResultWriter(opts.Output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating writer: %v\n", err)
		os.Exit(1)
	}
	defer writer.Close()

	scanner := app.NewScanner(paths, time.Duration(opts.Timeout)*time.Second, writer, opts.Insecure)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("Scanning %d targets with %d workers (timeout: %ds)\n\n",
		len(urls), opts.Workers, opts.Timeout)
	scanner.Run(ctx, urls, opts.Workers)
}
