#!/usr/bin/env bash
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -buildvcs=false -o tdev
