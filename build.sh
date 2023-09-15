#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o CLImgExport-linux-amd64
GOOS=windows GOARCH=amd64 go build -o CLImgExport-windows-amd64.exe
GOOS=darwin GOARCH=amd64 go build -o CLImgExport-osx-amd64
GOOS=darwin GOARCH=arm64  go build -o CLImgExport-osx-arm64
