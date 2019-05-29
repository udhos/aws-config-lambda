package main_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/udhos/aws-config-lambda"
)

func TestHandler(t *testing.T) {

	invokeEvent := map[string]string{}
	invokeEvent["foo"] = "bar"
	i, errJson := json.Marshal(invokeEvent)
	if errJson != nil {
		t.Errorf("internal json error: %v", errJson)
	}
	invoke := string(i)

	tests := []struct {
		request events.ConfigEvent
		expect  string
		err     bool
	}{
		{
			request: events.ConfigEvent{InvokingEvent: invoke, ConfigRuleName: "non-empty"},
			expect:  "ok",
			err:     false,
		},
		{
			request: events.ConfigEvent{InvokingEvent: invoke},
			expect:  "custom error: empty config rule name",
			err:     true,
		},
	}

	for _, test := range tests {
		ctx := context.Background()
		response, err := main.Handler(ctx, test.request)
		if response.Str != test.expect {
			t.Errorf("response request=%v expected=[%s] got=[%s]", test.request, test.expect, response.Str)
		}
		if (err != nil) != test.err {
			t.Errorf("error request=%v expected=%v got=%v", test.request, test.err, err)
		}
	}

}
