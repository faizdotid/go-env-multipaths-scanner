package app

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type EnvScanner struct {
	Paths  []string
	Client *http.Client
}

func NewEnvScanner(paths []string) *EnvScanner {
	return &EnvScanner{
		Paths: paths,
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

func (e *EnvScanner) Runner(urls []string, thread int) {
	var threadChan = make(chan struct{}, thread)
	var wg sync.WaitGroup
	for _, url := range urls {
		threadChan <- struct{}{}
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			e.Scan(url)
			<-threadChan
		}(url)
	}
	wg.Wait()
}

func (e *EnvScanner) Scan(url string) {
	if !strings.Contains(url, "://") {
		url = "http://" + url
	}
	for _, path := range e.Paths {
		returnValue := e.Request(MergeUrlAndPath(url, path))
		if returnValue {
			break
		}
	}
}

func (e *EnvScanner) Request(url string) bool {
	RecoverIfPanic()
	resp, err := e.Client.Get(url)
	if err != nil {
		LogError(err)
		return false
	}
	defer resp.Body.Close()
	bodyBuffer, err := io.ReadAll(resp.Body)
	if err != nil {
		LogError(err)
		return false
	}
	bodyString := string(bodyBuffer)
	if (strings.Contains(bodyString, "APP_KEY=base64")) && !strings.Contains(bodyString, "Laravel") {
		fmt.Printf("%s%s %s%s-> %sOK%s\n", White, url, Reset, Blue, Green, Reset)
		WriteResultToFile(url)
		return true
	} else {
		fmt.Printf("%s%s %s%s-> %sBAD%s\n", White, url, Reset, Blue, Red, Reset)
		return false
	}
}
