#!/bin/sh

# Run coverage
go test ./... -coverprofile=c.out; go tool cover -html="c.out" 
