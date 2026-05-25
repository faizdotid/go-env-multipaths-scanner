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
	Red    = "\033[31m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
	White  = "\033[1;37m"
	Reset  = "\033[0m"
	Yellow = "\033[33m"
)

var printMu sync.Mutex

// Options holds parsed CLI flags.
type Options struct {
	TargetsFile string
	PathsFile   string
	Workers     int
	Timeout     int
	Output      string
}

// ParseFlags parses and validates CLI flags.
func ParseFlags() Options {
	var opts Options
	flag.StringVar(&opts.TargetsFile, "f", "", "File containing target URLs")
	flag.StringVar(&opts.PathsFile, "p", "./config/paths.txt", "File containing path list")
	flag.IntVar(&opts.Workers, "t", 20, "Number of concurrent workers")
	flag.IntVar(&opts.Timeout, "timeout", 10, "HTTP timeout in seconds")
	flag.StringVar(&opts.Output, "o", "result.txt", "Output file")
	flag.Parse()

	if opts.TargetsFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return opts
}

// LoadPaths reads a file line-by-line and returns trimmed non-empty paths.
func LoadPaths(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var results []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			results = append(results, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no paths found in %s", path)
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
	fmt.Printf("%s[+]%s %s %s->%s %sOK%s\n", Green, Reset, url, Blue, Reset, Green, Reset)
}

// PrintErr prints an error message to stderr (thread-safe).
func PrintErr(format string, args ...interface{}) {
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Fprintf(os.Stderr, "%s[-]%s %s\n", Red, Reset, fmt.Sprintf(format, args...))
}
