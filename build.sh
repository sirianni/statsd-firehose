#!/usr/bin/env bash
set -euo pipefail

CGO_ENABLED=0 GOOS=linux go build -v -o build/linux-x86-64/statsd-firehose .
CGO_ENABLED=0 GOOS=darwin go build -v -o build/darwin-amd64/statsd-firehose .
