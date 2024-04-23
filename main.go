package main

import (
	"go-env-multipath-scan/app"
)

func main() {
	defer app.RecoverIfPanic()
	paths, err := app.LoadPathsFromFile("./config/paths.txt")
	if err != nil {
		panic(err)
	}
	parseFlag := app.ParseFlag()
	urls, err := app.LoadPathsFromFile(parseFlag.Filename)
	if err != nil {
		panic(err)
	}
	envScanner := app.NewEnvScanner(paths)
	envScanner.Runner(urls, parseFlag.Thread)
}
