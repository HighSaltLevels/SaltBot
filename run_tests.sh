#!/usr/bin/env bash

mkdir -p /tmp/saltbot
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o /tmp/saltbot/result.html
