#!/bin/bash

export GOOS=linux
export GOARCH=amd64

go get github.com/aws/aws-lambda-go/lambda
go get github.com/aws/aws-sdk-go

go build -o main
