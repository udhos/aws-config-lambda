package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Out struct {
	Str string
}

// https://github.com/aws/aws-lambda-go/blob/master/events/README_Config.md
// https://github.com/aws/aws-lambda-go/blob/master/events/config.go

func Handler(ctx context.Context, configEvent events.ConfigEvent) (Out, error) {

	fmt.Printf("fmt: logging from handler: event: %v", configEvent)
	log.Printf("log: logging from handler: event: %v", configEvent)

	fmt.Printf("AWS Config rule: %s\n", configEvent.ConfigRuleName)
	fmt.Printf("Invoking event JSON: %s\n", configEvent.InvokingEvent)
	fmt.Printf("Event version: %s\n", configEvent.Version)

	var err error

	if configEvent.ConfigRuleName == "" {
		err = fmt.Errorf("custom error: empty config rule name")
	}

	return Out{"Hello lambda!"}, err
}

func main() {
	lambda.Start(Handler)
}
