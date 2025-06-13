#!/bin/bash

mkdir -p bin/x86_64 bin/aarch64

# x86_64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.buildat=$(date '+%Y%m%d%H%M%S%3N') -X main.commit=$(git rev-parse --short HEAD)" -o bin/x86_64/usoftirqueue .

# aarch64
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.buildat=$(date '+%Y%m%d%H%M%S%3N') -X main.commit=$(git rev-parse --short HEAD)" -o bin/aarch64/usoftirqueue .
