package main_test

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/udhos/aws-config-lambda"
)

func TestHandler(t *testing.T) {

	tests := []struct {
		request events.ConfigEvent
		expect  string
		err     bool
	}{
		{
			request: events.ConfigEvent{ConfigRuleName: "non-empty"},
			expect:  "Hello lambda!",
			err:     false,
		},
		{
			request: events.ConfigEvent{},
			expect:  "Hello lambda!",
			err:     true,
		},
	}

	for _, test := range tests {
		ctx := context.Background()
		response, err := main.Handler(ctx, test.request)
		if response.Str != test.expect {
			t.Errorf("response expected=[%s] got=[%s]", test.expect, response.Str)
		}
		if (err != nil) != test.err {
			t.Errorf("error expected=%v got=%v", test.err, err)
		}
	}

}
