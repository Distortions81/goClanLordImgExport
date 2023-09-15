#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o CLImgExport
GOOS=windows GOARCH=amd64 go build -o CLImgExport.exe
