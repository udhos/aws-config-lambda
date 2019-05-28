package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

type in struct {
	str string
}

type out struct {
	str string
}

func handler(c context.Context, event in) (out, error) {
	return out{"Hello ƛ! - " + event.str}, nil
}

func main() {
	lambda.Start(handler)
}
