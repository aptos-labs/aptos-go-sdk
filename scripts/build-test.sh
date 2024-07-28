#!/bin/sh

# Change to the root of the git repository
cd "$(git rev-parse --show-toplevel)"

go test -c ./...