package main

import (
	"fmt"
	"log"
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

	fmt.Printf("fmt: logging from handler: event: %v", event)
	log.Printf("log: logging from handler: event: %v", event)

	var err error

	if event.str == "" {
		err = fmt.Errorf("custom error: empty input string")
	}

	return out{"Hello ƛ! - " + event.str}, err
}

func main() {
	lambda.Start(handler)
}
