#!/bin/bash

gofmt -s -w .
go fix .
go vet -vettool="$(which shadow)" .
go test .
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -v .

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

zip main.zip main
