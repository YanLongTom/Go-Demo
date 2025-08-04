#!/bin/bash

go clean -cache -modcache -i -r
go build ./cmd/mytool
go build -toolexec=./mytool -o main