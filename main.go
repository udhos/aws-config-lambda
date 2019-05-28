package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type out struct {
	str string
}

// https://github.com/aws/aws-lambda-go/blob/master/events/README_Config.md
// https://github.com/aws/aws-lambda-go/blob/master/events/config.go

func handler(ctx context.Context, configEvent events.ConfigEvent) (out, error) {

	fmt.Printf("fmt: logging from handler: event: %v", configEvent)
	log.Printf("log: logging from handler: event: %v", configEvent)

	fmt.Printf("AWS Config rule: %s\n", configEvent.ConfigRuleName)
	fmt.Printf("Invoking event JSON: %s\n", configEvent.InvokingEvent)
	fmt.Printf("Event version: %s\n", configEvent.Version)

	var err error

	if configEvent.ConfigRuleName == "" {
		err = fmt.Errorf("custom error: empty config rule name")
	}

	return out{"Hello lambda!"}, err
}

func main() {
	lambda.Start(handler)
}
