#!/bin/bash

rm -rf dist
mkdir -p dist/ci-utility-amd64
CGO_ENABLED=0 go build -ldflags="-s -w" -o ./dist/ci-utility-amd64/ci-utility ./cmd/ci-utility
