#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o CLImgExport
zip CLImgExport-linux-amd64.zip CLImgExport 
GOOS=windows GOARCH=amd64 go build -o CLImgExport.exe
zip CLImgExport-windows-amd64.zip CLImgExport.exe
GOOS=darwin GOARCH=amd64 go build -o CLImgExport
zip CLImgExport-mac-amd64.zip CLImgExport 
GOOS=darwin GOARCH=arm64  go build -o CLImgExport
zip CLImgExport-mac-m1.zip CLImgExport
rm CLImgExport
rm CLImgExport.exe
