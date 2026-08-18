# Go Env Multipaths Scanner

Fast concurrent scanner to find exposed `.env` files across multiple paths.

## Features

- Worker pool concurrency with adjustable goroutines
- Context-aware HTTP requests
- Buffered, thread-safe result writing
- Skips redundant scans once a target is matched
- TLS verification disabled for testing environments

## Installation

```bash
go build -o go-env-multipaths-scanner main.go
```

## Usage

```bash
./go-env-multipaths-scanner -f targets.txt -t 100
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-f` | *(required)* | File containing target URLs |
| `-p` | `./config/paths.txt` | File containing path list |
| `-t` | `20` | Number of concurrent workers |
| `-timeout` | `10` | HTTP timeout in seconds |
| `-o` | `result.txt` | Output file for findings |
| `-insecure` | `true` | Skip TLS certificate verification |

### Examples

Scan with custom paths and 50 workers:

```bash
./go-env-multipaths-scanner -f targets.txt -p ./config/paths.txt -t 50 -o out.txt
```

Scan with longer timeout:

```bash
./go-env-multipaths-scanner -f targets.txt -t 100 -timeout 20
```

## Disclaimer

This tool is intended for authorized security testing and research only. Always obtain proper permission before testing systems you do not own.
