package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
)

type Out struct {
	Str string
}

var count int

func getConfig() *configservice.ConfigService {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Printf("getConfig: %v", err)
		return nil
	}

	cfg.Region = endpoints.SaEast1RegionID

	config := configservice.New(cfg)

	return config
}

// https://github.com/aws/aws-lambda-go/blob/master/events/README_Config.md
// https://github.com/aws/aws-lambda-go/blob/master/events/config.go

func Handler(ctx context.Context, configEvent events.ConfigEvent) (Out, error) {

	var err error
	out := Out{"ok"}

	count++

	fmt.Printf("fmt: logging from handler: event: %v\n", configEvent)
	log.Printf("log: logging from handler: event: %v", configEvent)

	log.Printf("count=%d", count)

	fmt.Printf("AWS Config rule: %s\n", configEvent.ConfigRuleName)
	fmt.Printf("Invoking event JSON: %s\n", configEvent.InvokingEvent)
	fmt.Printf("Event version: %s\n", configEvent.Version)

	config := getConfig()
	if config == nil {
		log.Printf("could not get config service")
	}

	if configEvent.ConfigRuleName == "" {
		out.Str = "custom error: empty config rule name"
		err = fmt.Errorf(out.Str)
	}

	return out, err
}

func main() {
	lambda.Start(Handler)
}
