package app

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	Green = "\033[32m"
	Blue  = "\033[34m"
	Reset = "\033[0m"
)

var printMu sync.Mutex

// Options holds parsed CLI flags.
type Options struct {
	TargetsFile string
	PathsFile   string
	Workers     int
	Timeout     int
	Output      string
	Insecure    bool
}

// ParseFlags parses and validates CLI flags.
func ParseFlags() Options {
	var opts Options
	flag.StringVar(&opts.TargetsFile, "f", "", "File containing target URLs")
	flag.StringVar(&opts.PathsFile, "p", "./config/paths.txt", "File containing path list")
	flag.IntVar(&opts.Workers, "t", 20, "Number of concurrent workers")
	flag.IntVar(&opts.Timeout, "timeout", 10, "HTTP timeout in seconds")
	flag.StringVar(&opts.Output, "o", "result.txt", "Output file")
	flag.BoolVar(&opts.Insecure, "insecure", true, "Skip TLS certificate verification")
	flag.Parse()

	if opts.TargetsFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if opts.Workers < 1 {
		fmt.Fprintf(os.Stderr, "Error: -t (workers) must be >= 1\n")
		os.Exit(1)
	}
	if opts.Timeout < 1 {
		fmt.Fprintf(os.Stderr, "Error: -timeout must be >= 1 second\n")
		os.Exit(1)
	}
	return opts
}

// LoadPaths reads a file line-by-line and returns deduplicated, trimmed,
// non-empty entries preserving their original order.
func LoadPaths(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	seen := make(map[string]struct{})
	var results []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		results = append(results, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no entries found in %s", path)
	}
	return results, nil
}

// ResultWriter handles buffered, thread-safe writing to a result file.
type ResultWriter struct {
	mu     sync.Mutex
	writer *bufio.Writer
	file   *os.File
}

// NewResultWriter creates a buffered result writer.
func NewResultWriter(path string) (*ResultWriter, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return &ResultWriter{
		file:   f,
		writer: bufio.NewWriter(f),
	}, nil
}

// Write appends a line atomically to the buffer.
func (rw *ResultWriter) Write(line string) error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	_, err := fmt.Fprintln(rw.writer, line)
	return err
}

// Close flushes the buffer and closes the file.
func (rw *ResultWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if err := rw.writer.Flush(); err != nil {
		_ = rw.file.Close()
		return err
	}
	return rw.file.Close()
}

// PrintOK prints a success match to stdout (thread-safe).
func PrintOK(url string) {
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Printf("%s[+]%s %s%sOK%s\n", Green, Reset, url, Green, Reset)
}
