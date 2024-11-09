#!/bin/bash

go build -ldflags="-s -w -X main.buildat=$(date '+%Y%m%d%H%M%S%3N') -X main.commit=$(git rev-parse --short HEAD)" -o usoftirqueue .
