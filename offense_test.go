package main

import (
	"encoding/json"
	"testing"
)

func TestOffenseMap(t *testing.T) {

	tests := []struct {
		target  map[string]interface{}
		item    map[string]interface{}
		offense bool
	}{
		{
			target:  map[string]interface{}{"empty": ""},
			item:    map[string]interface{}{"empty": ""},
			offense: false,
		},
		{
			target:  map[string]interface{}{"tags": map[string]interface{}{"key1": "value1"}},
			item:    map[string]interface{}{"tags": map[string]interface{}{"key1": "value1"}},
			offense: false,
		},
		{
			target:  map[string]interface{}{"tags": map[string]interface{}{"key1": "value1"}},
			item:    map[string]interface{}{"tags": map[string]interface{}{"key1": "value2"}},
			offense: true,
		},
	}

	for _, test := range tests {
		o, annotation := findOffenseMap("", test.item, test.target)
		if o != test.offense {
			t.Errorf("offenseExpected=%v offenseFound=%v annotation=%s target=%v item=%v", test.offense, o, annotation, test.target, test.item)
		}
	}

}

func TestOffenseJson(t *testing.T) {

	tests := []struct {
		target  string
		item    string
		offense bool
	}{
		{
			target:  `{"empty":""}`,
			item:    `{"empty":""}`,
			offense: false,
		},
		{
			target:  `{"tags":{"key1":"value1"}}`,
			item:    `{"tags":{"key1":"value1"}}`,
			offense: false,
		},
		{
			target:  `{"tags":{"key1":"value1"}}`,
			item:    `{"tags":{"key1":"value2"}}`,
			offense: true,
		},
		{
			target:  `{"tags":{"key1":"123"}}`,
			item:    `{"tags":{"key1":"123"}}`,
			offense: false,
		},
		{
			target:  `{"tags":{"key1":"123"}}`,
			item:    `{"tags":{"key1":"1234"}}`,
			offense: true,
		},
	}

	for _, test := range tests {
		tm := map[string]interface{}{}
		if err := json.Unmarshal([]byte(test.target), &tm); err != nil {
			t.Errorf("bad json target=%v %v", test.target, err)
		}
		im := map[string]interface{}{}
		if err := json.Unmarshal([]byte(test.item), &im); err != nil {
			t.Errorf("bad json item=%v %v", test.item, err)
		}
		o, annotation := findOffenseMap("", im, tm)
		if o != test.offense {
			t.Errorf("offenseExpected=%v offenseFound=%v annotation=%s target=%v item=%v", test.offense, o, annotation, test.target, test.item)
		}
	}

}
