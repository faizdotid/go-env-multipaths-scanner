package app

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"errors"
)

const (
	Red   = "\033[31m"
	Green = "\033[32m"
	Reset = "\033[0m"
	Blue  = "\033[34m"
	White = "\033[1;37m"
)

type ParseFlagStruct struct {
	Filename string
	Thread   int
}

func LoadPathsFromFile(path string) ([]string, error) {
	var results []string
	fileBuffer, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	stringsSplit := strings.Split(
		string(fileBuffer),
		"\n",
	)
	if len(stringsSplit) == 0 {
		return nil, errors.New("no paths found in paths.txt")
	}
	for _, path := range stringsSplit {
		if path != "" {
			results = append(results, strings.TrimSpace(path))
		}
	}
	return results, nil
}


func LogError(err error) {
	fmt.Printf("%sErr %s-> %s%s%s\n", Red, Blue, White, err.Error(), Reset)
}

func RecoverIfPanic() {
	if r := recover(); r != nil {
		LogError(r.(error))
	}
}

func WriteResultToFile(result string) {
	file, err := os.OpenFile("result.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		LogError(err)
		return
	}
	defer file.Close()
	file.WriteString(result + "\n")
}

func MergeUrlAndPath(url, path string) string {
	if strings.HasSuffix(url, "/") {
		return url + path
	}
	return url + "/" + path
}

func ParseFlag() ParseFlagStruct {
	var parseFlagStruct ParseFlagStruct
	flag.StringVar(&parseFlagStruct.Filename, "f", "", "Filename")
	flag.IntVar(&parseFlagStruct.Thread, "t", 1, "Thread")
	flag.Parse()
	if parseFlagStruct.Filename == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}
	return parseFlagStruct
}
