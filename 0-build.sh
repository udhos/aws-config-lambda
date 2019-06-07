#!/bin/bash

gofmt -s -w .
go fix .
go vet .

if ! hash shadow; then
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	tidy=1
fi

go vet -vettool="$(which shadow)" .

if ! hash gosec; then
	go install github.com/securego/gosec/cmd/gosec
	tidy=1
fi

gosec .

go test .
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -v .

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

zip main.zip main

[ -n "$tidy" ] && go mod tidy
