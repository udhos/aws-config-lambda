package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	lambda.Start(Handler)
}

type Out struct {
	Str string
}

const version = "0.1"

var count int

// https://github.com/aws/aws-lambda-go/blob/master/events/README_Config.md
// https://github.com/aws/aws-lambda-go/blob/master/events/config.go

func Handler(ctx context.Context, configEvent events.ConfigEvent) (out Out, err error) {

	out = Out{"ok"}

	fmt.Printf("version=%s runtime=%s GOMAXPROCS=%d OS=%s ARCH=%s\n", version, runtime.Version(), runtime.GOMAXPROCS(0), runtime.GOOS, runtime.GOARCH)

	count++

	fmt.Printf("count=%d\n", count)
	fmt.Printf("AWS Config rule: %s\n", configEvent.ConfigRuleName)
	fmt.Printf("EventLeftScope=%v\n", configEvent.EventLeftScope)

	var dumpConfigItem bool
	var bucket, key string

	if params := configEvent.RuleParameters; params != "" {
		ruleParameters := map[string]string{}
		if errParam := json.Unmarshal([]byte(params), &ruleParameters); errParam != nil {
			fmt.Printf("RuleParameters: %v\n", errParam)
		} else {
			for k, v := range ruleParameters {
				fmt.Printf("RuleParameters: %s=%s\n", k, v)
			}

			if dump, found := ruleParameters["Dump"]; found {
				if dump == "ConfigItem" {
					dumpConfigItem = true
				}
			}

			bucket = ruleParameters["Bucket"]
			key = ruleParameters["Key"]
		}
	}

	clientConf := getConfig()
	config := clientConf.config
	if config == nil {
		fmt.Printf("could not get config service\n")
	}

	// InvokingEvent:
	// If the event is published in response to a resource configuration change, this value contains a JSON configuration item
	// https://github.com/aws/aws-lambda-go/blob/master/events/config.go
	invokingEvent := map[string]interface{}{}
	if errJson := json.Unmarshal([]byte(configEvent.InvokingEvent), &invokingEvent); errJson != nil {
		err = fmt.Errorf("InvokingEvent: %v", errJson)
		out.Str = err.Error()
		fmt.Println(out.Str)
		return
	}

	// invokingEvent:
	//   configurationItem: map
	//   messageType: ConfigurationItemChangeNotification

	item, foundItem := invokingEvent["configurationItem"]
	if !foundItem {
		err = fmt.Errorf("configurationItem not found in InvokingEvent")
		out.Str = err.Error()
		fmt.Println(out.Str)
		return
	}

	configItem, itemMap := item.(map[string]interface{})
	if !itemMap {
		err = fmt.Errorf("configurationItem not a map")
		out.Str = err.Error()
		fmt.Println(out.Str)
		return
	}

	if dumpConfigItem {
		logItem("dump config item: ", configItem)
	}

	status := mapString(configItem, "configurationItemStatus")
	resourceType := mapString(configItem, "resourceType")
	resourceId := mapString(configItem, "resourceId")
	timestamp := mapString(configItem, "configurationItemCaptureTime")
	t, errTime := time.Parse(time.RFC3339, timestamp)
	if errTime != nil {
		fmt.Printf("parse time: '%s': %v\n", timestamp, errTime)
	}

	fmt.Printf("configuration item status: %s\n", status)
	fmt.Printf("configuration item type: %s\n", resourceType)
	fmt.Printf("configuration item id: %s\n", resourceId)

	// ComplianceType
	// https://godoc.org/github.com/aws/aws-sdk-go-v2/service/configservice#ComplianceType
	compliance := configservice.ComplianceTypeCompliant

	snapshot, errSnap := fetch(clientConf.s3, bucket, key, resourceId)
	if errSnap != nil {
		fmt.Printf("snapshot: %v\n", errSnap)
		compliance = configservice.ComplianceTypeInsufficientData
	}

	if dumpConfigItem {
		logItem("dump config item snapshot: ", snapshot)
	}

	fmt.Printf("configuration item compliance: %s\n", compliance)

	eval := configservice.Evaluation{
		ComplianceResourceType: &resourceType,
		ComplianceResourceId:   &resourceId,
		ComplianceType:         compliance,
		OrderingTimestamp:      &t,
	}
	report := configservice.PutEvaluationsInput{
		ResultToken: &configEvent.ResultToken,
		Evaluations: []configservice.Evaluation{eval},
	}
	req := config.PutEvaluationsRequest(&report)
	resp, errPut := req.Send(context.TODO())
	if errPut == nil {
		fmt.Println("PutEvaluations ok: ", resp)
	} else {
		fmt.Println("PutEvaluations error: ", errPut)
	}

	return
}

func fetch(client *s3.Client, bucket, key string, resourceId string) (map[string]interface{}, error) {

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(key),    // Required
	}

	req := client.GetObjectRequest(params)
	resp, errSend := req.Send(context.TODO())
	if errSend != nil {
		return nil, errSend
	}

	buf, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return nil, errRead
	}

	item := map[string]interface{}{}
	if errJson := json.Unmarshal(buf, &item); errJson != nil {
		return nil, errJson
	}

	return item, nil
}

func logItem(prefix string, configItem map[string]interface{}) {
	for k, v := range configItem {
		fmt.Printf("%s %s = %v\n", prefix, k, v)
	}
}

func mapString(m map[string]interface{}, key string) string {
	v, found := m[key]
	if found {
		s, isStr := v.(string)
		if isStr {
			return s
		}
	}
	return ""
}

type conf struct {
	cfg    aws.Config
	config *configservice.Client
	s3     *s3.Client
}

func getConfig() *conf {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		fmt.Printf("getConfig: %v\n", err)
		return nil
	}

	cfg.Region = endpoints.SaEast1RegionID

	c := conf{
		cfg:    cfg,
		config: configservice.New(cfg),
		s3:     s3.New(cfg),
	}

	return &c
}
