package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
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
	var bucket string
	restrictResourceTypes := map[string]struct{}{}

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

			if types, found := ruleParameters["ResourceTypes"]; found {
				for _, s := range strings.Split(types, ",") {
					restrictResourceTypes[s] = struct{}{}
				}
			}

			bucket = ruleParameters["Bucket"]
		}
	}

	clientConf := getConfig()
	if clientConf == nil {
		err = fmt.Errorf("could not get aws client - aborting")
		out.Str = err.Error()
		fmt.Println(out.Str)
		return
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

	// Decode configuration item

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
	compliance := configservice.ComplianceTypeNotApplicable
	annotation := ""

	isApplicable := (status == "OK" || status == "ResourceDiscovered") && !configEvent.EventLeftScope

	if isApplicable && len(restrictResourceTypes) > 0 {
		if _, found := restrictResourceTypes[resourceType]; !found {
			fmt.Printf("resourceType=%s missing from parameter ResourceTypes\n", resourceType)
			isApplicable = false
		}
	}

	if isApplicable {
		compliance, annotation = eval(clientConf.s3, configItem, bucket, resourceId, dumpConfigItem)
		if annotation != "" {
			fmt.Println(annotation)
		}
	}

	// Send evaluation result

	sendEval(clientConf.config, configEvent.ResultToken, resourceType, resourceId, t, compliance, annotation)

	return
}

// eval: compare item against target
func eval(s3Client *s3.Client, configItem map[string]interface{}, bucket, resourceId string, dump bool) (configservice.ComplianceType, string) {

	// Fetch target configuration

	target, errTarget := fetch(s3Client, bucket, resourceId)
	if errTarget != nil {
		return configservice.ComplianceTypeNonCompliant, fmt.Sprintf("fetch: bucket=%s key=%s %v", bucket, resourceId, errTarget)
	}

	if dump {
		logItem("dump config item target: ", target)
	}

	if offense, annotation := findOffenseMap("", configItem, target); offense {
		return configservice.ComplianceTypeNonCompliant, annotation
	}

	return configservice.ComplianceTypeCompliant, ""
}

func findOffenseMap(path string, item, target map[string]interface{}) (bool, string) {

	for tk, tv := range target {
		iv, foundKey := item[tk]
		if !foundKey {
			return true, fmt.Sprintf("path=[%s] key=%s missing key on item", path, tk)
		}

		child := path + "." + tk

		// encoded?
		tvj, tvString := tv.(string)
		if tvString {
			if isJSON(tvj) {
				var j interface{}
				if errJson := json.Unmarshal([]byte(tvj), &j); errJson != nil {
					return true, fmt.Sprintf("path=[%s] key=%s target bad json: %v", path, tk, errJson)
				}
				if offense, annotation := findOffense(child, iv, j); offense {
					return offense, annotation
				}
			} else {
				// scalar?
				if offense, annotation := findOffenseScalar(child, iv, tvj); offense {
					return true, annotation
				}
			}
			continue // no offense found
		}

		// map?
		tvm, tvMap := tv.(map[string]interface{})
		if tvMap {
			ivm, ivMap := iv.(map[string]interface{})
			if !ivMap {
				return true, fmt.Sprintf("path=[%s] key=%s item non-map value: %v", path, tk, iv)
			}
			return findOffenseMap(child, ivm, tvm)
		}

		// slice?
		tvSlice, tvIsSlice := tv.([]interface{})
		if tvIsSlice {
			ivSlice, ivIsSlice := iv.([]interface{})
			if !ivIsSlice {
				return true, fmt.Sprintf("path=[%s] key=%s item non-slice value: %v", path, tk, iv)
			}
			return findOffenseSlice(child, ivSlice, tvSlice)
		}

		// scalar?
		if offense, annotation := findOffenseScalar(child, iv, tv); offense {
			return true, annotation
		}
	}

	return false, ""
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func findOffenseScalar(path string, item, target interface{}) (bool, string) {
	tvs, errTv := scalarString(target)
	if errTv != nil {
		return true, fmt.Sprintf("path=[%s] target value: %v", path, errTv)
	}
	ivs, errIv := scalarString(item)
	if errIv != nil {
		return true, fmt.Sprintf("path=[%s] item value: %v", path, errIv)
	}
	if tvs != ivs {
		if matchNumber(path, tvs, ivs) {
			return false, ""
		}
		if matchTime(path, tvs, ivs) {
			return false, ""
		}
		return true, fmt.Sprintf("path=[%s] value mismatch: targetValue=%s itemValue=%s", path, tvs, ivs)
	}

	return false, ""
}

func matchNumber(path string, s1, s2 string) bool {
	f1, errFloat1 := strconv.ParseFloat(s1, 64)
	if errFloat1 != nil {
		return false
	}
	f2, errFloat2 := strconv.ParseFloat(s2, 64)
	if errFloat2 != nil {
		return false
	}
	return f1 == f2
}

func matchTime(path string, s1, s2 string) bool {
	return timeAndUnix(path, s1, s2) || timeAndUnix(path, s2, s1)
}

func timeAndUnix(path string, s1, s2 string) bool {
	fmt.Printf("path=[%s] timeAndUnix: %s x %s\n", path, s1, s2)
	t1, errTime := time.Parse(time.RFC3339, s1)
	if errTime != nil {
		fmt.Printf("path=[%s] timeAndUnix: %s x %s: %v\n", path, s1, s2, errTime)
		return false
	}
	f, errFloat := strconv.ParseFloat(s2, 64)
	if errFloat != nil {
		fmt.Printf("path=[%s] timeAndUnix: %s x %s: %v\n", path, s1, s2, errFloat)
		return false
	}
	t2 := time.Unix(int64(f), 0)
	result := t1.Equal(t2)
	fmt.Printf("path=[%s] timeAndUnix: %s x %s: %v x %v: %v\n", path, s1, s2, t1, t2, result)
	return result

}

func findOffenseSlice(path string, item, target []interface{}) (bool, string) {
	if len(item) != len(target) {
		return true, fmt.Sprintf("path=[%s] slice size mismatch: target=%d item=%d", path, len(target), len(item))
	}
	for i, t := range target {
		it := item[i]
		child := path + "." + fmt.Sprint(i)
		offense, annotation := findOffense(child, it, t)
		if offense {
			return true, annotation
		}
	}
	return false, ""
}

func findOffense(path string, item, target interface{}) (bool, string) {
	tm, tMap := target.(map[string]interface{})
	if tMap {
		im, iMap := item.(map[string]interface{})
		if !iMap {
			return true, fmt.Sprintf("path=[%s] target is map, item is not", path)
		}
		return findOffenseMap(path, im, tm)
	}

	ts, tSlice := target.([]interface{})
	if tSlice {
		is, iSlice := item.([]interface{})
		if !iSlice {
			return true, fmt.Sprintf("path=[%s] target is slice, item is not", path)
		}
		return findOffenseSlice(path, is, ts)
	}

	return findOffenseScalar(path, item, target)
}

func scalarString(v interface{}) (string, error) {
	s, str := v.(string)
	if str {
		return s, nil
	}
	i64, isInt64 := v.(int64)
	if isInt64 {
		return fmt.Sprint(i64), nil
	}
	f32, isF32 := v.(float32)
	if isF32 {
		return fmt.Sprint(f32), nil
	}
	f64, isF64 := v.(float64)
	if isF64 {
		return fmt.Sprint(f64), nil
	}
	return "", fmt.Errorf("non-string/int/float: %v", v)
}

func sendEval(config *configservice.Client, resultToken, resourceType, resourceId string, timestamp time.Time, compliance configservice.ComplianceType, annotation string) {

	fmt.Printf("configuration item compliance: %s\n", compliance)

	var ann *string
	if annotation != "" {
		if len(annotation) > 255 {
			fmt.Printf("truncating annotation to 255-char: %s\n", annotation)
			annotation = annotation[:255]
		}
		ann = &annotation
	}

	eval := configservice.Evaluation{
		Annotation:             ann,
		ComplianceResourceType: &resourceType,
		ComplianceResourceId:   &resourceId,
		ComplianceType:         compliance,
		OrderingTimestamp:      &timestamp,
	}
	report := configservice.PutEvaluationsInput{
		ResultToken: &resultToken,
		Evaluations: []configservice.Evaluation{eval},
	}
	req := config.PutEvaluationsRequest(&report)
	resp, errPut := req.Send(context.TODO())
	if errPut == nil {
		fmt.Println("PutEvaluations ok: ", resp)
	} else {
		fmt.Println("PutEvaluations error: ", errPut)
	}
}

func fetch(client *s3.Client, bucket, resourceId string) (map[string]interface{}, error) {

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),     // Required
		Key:    aws.String(resourceId), // Required
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
