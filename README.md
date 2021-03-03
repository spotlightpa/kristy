# kristy [![GoDoc](https://godoc.org/github.com/spotlightpa/kristy?status.svg)](https://godoc.org/github.com/spotlightpa/kristy) [![Go Report Card](https://goreportcard.com/badge/github.com/spotlightpa/kristy)](https://goreportcard.com/report/github.com/spotlightpa/kristy)

Kristy is a baby-sitter for your processes. It reports errors to [HealthCheck.io](https://HealthCheck.io) and Slack as needed to ensure no cron job falls through the cracks.

## Installation

First install [Go](http://golang.org).

If you just want to install the binary to your current directory and don't care about the source code, run

```bash
GOBIN=$(pwd) go install github.com/spotlightpa/kristy@latest
```

## Screenshots

```
$ kristy -h
kristy - a baby-sitter for your cron jobs

Kristy tells HealthCheck.io how your cronjobs are doing. If it can't reach
HealthCheck.io, it falls back to warning Slack that something went wrong.

Usage:

        kristy [options] <command to babysit>

Options may be also passed as environmental variables prefixed with KRISTY_.

Options:
  -healthcheck UUID
        UUID for HealthCheck.io job
  -silent
        don't log debug output
  -slack URL
        Slack hook URL
  -timeout duration
        timeout for HTTP requests (default 10s)

$ KRISTY_HEALTHCHECK='x' KRISTY_SLACK='https://none.example/' kristy gronk
kristy 2021/03/03 12:33:45 starting
kristy 2021/03/03 12:33:46 done
Error: 4 errors:
        error 1: problem sending start signal: problem connecting to HealthCheck.io: unexpected status: 404 Not Found
        error 2: problem sending status: problem connecting to HealthCheck.io: unexpected status: 404 Not Found
        error 3: Post "https://none.example/": dial tcp: lookup none.example: no such host
        error 4: could not start process: exec: "gronk": executable file not found in $PATH

$ KRISTY_HEALTHCHECK='x' KRISTY_SLACK='https://none.example/' kristy ls
kristy 2021/03/03 12:34:03 starting
LICENSE
README.md
go.mod
go.sum
healthcheckio
main.go
sitter
kristy 2021/03/03 12:34:04 done
Error: 3 errors:
        error 1: problem sending start signal: problem connecting to HealthCheck.io: unexpected status: 404 Not Found
        error 2: problem sending status: problem connecting to HealthCheck.io: unexpected status: 404 Not Found
        error 3: Post "https://none.example/": dial tcp: lookup none.example: no such host
```
