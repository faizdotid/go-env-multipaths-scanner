package app

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Scanner probes URLs for exposed .env files.
type Scanner struct {
	client    *http.Client
	paths     []string
	writer    *ResultWriter
	foundURLs sync.Map
}

// NewScanner creates a new Scanner instance.
func NewScanner(paths []string, timeout time.Duration, writer *ResultWriter) *Scanner {
	return &Scanner{
		paths: paths,
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		writer: writer,
	}
}

// Run starts the worker pool and processes all targets.
func (s *Scanner) Run(ctx context.Context, urls []string, workers int) {
	jobs := make(chan string, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				s.scanURL(ctx, target)
			}
		}()
	}

	for _, u := range urls {
		jobs <- u
	}
	close(jobs)
	wg.Wait()
}

func (s *Scanner) scanURL(ctx context.Context, base string) {
	if !strings.Contains(base, "://") {
		base = "http://" + base
	}

	// Skip if this URL was already matched by another worker.
	if _, found := s.foundURLs.Load(base); found {
		return
	}

	for _, path := range s.paths {
		if _, found := s.foundURLs.Load(base); found {
			return
		}

		target, err := url.JoinPath(base, path)
		if err != nil {
			continue
		}

		if s.request(ctx, target, base) {
			s.foundURLs.Store(base, true)
			return
		}
	}
}

func (s *Scanner) request(ctx context.Context, target, base string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	bodyStr := string(body)
	if strings.Contains(bodyStr, "APP_KEY=base64") && !strings.Contains(bodyStr, "Laravel") {
		PrintOK(base)
		if s.writer != nil {
			_ = s.writer.Write(base)
		}
		return true
	}
	return false
}
