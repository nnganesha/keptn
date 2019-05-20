#!/bin/bash

GIT_SHA=$1
TYPE=$2
NUMBER=$3
DATE=$4

FOLDER="${TYPE}-${NUMBER}-${DATE}"

# MAC
env GOOS=darwin GOARCH=amd64 go get ./...
env GOOS=darwin GOARCH=amd64 go build -o keptn
zip keptn-macOS.zip keptn
tar -zcvf keptn-macOS.tar.gz keptn
rm keptn

# Linux
env GOOS=linux GOARCH=amd64 go get ./...
env GOOS=linux GOARCH=amd64 go build -o keptn
zip keptn-linux.zip keptn
tar -zcvf keptn-linux.tar.gz keptn
rm keptn

gsutil cp keptn-linux.zip gs://keptn-cli/${FOLDER}/

# Windows
env GOOS=windows GOARCH=amd64 go get ./...
env GOOS=windows GOARCH=amd64 go build -o keptn.exe
zip keptn-windows.zip keptn.exe
tar -zcvf keptn-windows.tar.gz keptn.exe
rm keptn.exe

gsutil cp keptn-windows.zip gs://keptn-cli/${FOLDER}/

ls -lsa
