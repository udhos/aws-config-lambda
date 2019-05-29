package main

import (
	"context"
	"encoding/json"
	"fmt"

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
		fmt.Printf("getConfig: %v\n", err)
		return nil
	}

	cfg.Region = endpoints.SaEast1RegionID

	config := configservice.New(cfg)

	return config
}

// https://github.com/aws/aws-lambda-go/blob/master/events/README_Config.md
// https://github.com/aws/aws-lambda-go/blob/master/events/config.go

func Handler(ctx context.Context, configEvent events.ConfigEvent) (out Out, err error) {

	out = Out{"ok"}

	count++

	fmt.Printf("count=%d\n", count)
	fmt.Printf("AWS Config rule: %s\n", configEvent.ConfigRuleName)
	//fmt.Printf("Invoking event JSON: %s\n", configEvent.InvokingEvent)

	config := getConfig()
	if config == nil {
		fmt.Printf("could not get config service\n")
	}

	if configEvent.ConfigRuleName == "" {
		out.Str = "custom error: empty config rule name"
		err = fmt.Errorf(out.Str)
		return
	}

	// InvokingEvent:
	// If the event is published in response to a resource configuration change, this value contains a JSON configuration item
	// https://github.com/aws/aws-lambda-go/blob/master/events/config.go
	invokingEvent := map[string]interface{}{}
	errJson := json.Unmarshal([]byte(configEvent.InvokingEvent), &invokingEvent)
	if errJson != nil {
		err = fmt.Errorf("InvokingEvent: %v", errJson)
		out.Str = err.Error()
		fmt.Println(out.Str)
		return
	}

	// ComplianceType
	// https://godoc.org/github.com/aws/aws-sdk-go-v2/service/configservice#ComplianceType
	compliance := configservice.ComplianceTypeNotApplicable

	eval := configservice.Evaluation{
		//ComplianceResourceType: configurationItem.resourceType,
		//ComplianceResourceId: configurationItem.resourceId,
		ComplianceType: compliance,
		//OrderingTimestamp: configurationItem.configurationItemCaptureTime,
	}
	report := configservice.PutEvaluationsInput{
		ResultToken: &configEvent.ResultToken,
		Evaluations: []configservice.Evaluation{eval},
	}
	req := config.PutEvaluationsRequest(&report)
	resp, errPut := req.Send(context.TODO())
	if errPut == nil {
		fmt.Println(resp)
	} else {
		fmt.Println(errPut)
	}

	return
}

func main() {
	lambda.Start(Handler)
}
