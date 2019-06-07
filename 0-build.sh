#!/bin/bash

gofmt -s -w .
go fix .
go vet -vettool="$(which shadow)" .

if ! hash gosec; then
	go_get_gosec=1
	go get github.com/securego/gosec/cmd/gosec
fi

gosec .

go test .
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -v .

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main main.go

zip main.zip main

[ -n "$go_get_gosec" ] && go mod tidy
