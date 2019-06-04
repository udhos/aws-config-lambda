package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
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

	dump := false

	for _, test := range tests {
		o, annotation := findOffenseMap("", test.item, test.target, dump)
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

	dump := false

	for _, test := range tests {
		tm := map[string]interface{}{}
		if err := json.Unmarshal([]byte(test.target), &tm); err != nil {
			t.Errorf("bad json target=%v %v", test.target, err)
		}
		im := map[string]interface{}{}
		if err := json.Unmarshal([]byte(test.item), &im); err != nil {
			t.Errorf("bad json item=%v %v", test.item, err)
		}
		o, annotation := findOffenseMap("", im, tm, dump)
		if o != test.offense {
			t.Errorf("offenseExpected=%v offenseFound=%v annotation=%s target=%v item=%v", test.offense, o, annotation, test.target, test.item)
		}
	}

}

func TestOffenseData(t *testing.T) {
	root := "testdata"
	dirItem := filepath.Join(root, "item")     // testdata/item/resource-id   = item
	dirTarget := filepath.Join(root, "target") // testdata/target/resource-id = target
	dirResult := filepath.Join(root, "result") // testdata/result/resource-id = annotation = empty "" means no error

	files, errDir := ioutil.ReadDir(dirItem)
	if errDir != nil {
		t.Errorf("missing dir testdata: %v", errDir)
		return
	}

	for _, f := range files {
		pathItem := filepath.Join(dirItem, f.Name())
		bufItem, errItem := ioutil.ReadFile(pathItem)
		if errItem != nil {
			t.Errorf("item file %s: %v", f.Name(), errItem)
			continue
		}
		pathTarget := filepath.Join(dirTarget, f.Name())
		bufTarget, errTarget := ioutil.ReadFile(pathTarget)
		if errTarget != nil {
			t.Errorf("target file %s: %v", f.Name(), errTarget)
			continue
		}
		pathResult := filepath.Join(dirResult, f.Name())
		bufResult, _ := ioutil.ReadFile(pathResult)
		expectOffense := len(bufResult) == 0
		im := map[string]interface{}{}
		if err := json.Unmarshal(bufItem, &im); err != nil {
			t.Errorf("bad json item %s: %v", f.Name(), err)
		}
		tm := map[string]interface{}{}
		if err := json.Unmarshal(bufTarget, &tm); err != nil {
			t.Errorf("bad json target %s: %v", f.Name(), err)
		}
		dump := false
		o, annotation := findOffenseMap("", im, tm, dump)
		if o != expectOffense {
			t.Errorf("%s offenseExpected=%v offenseFound=%v", f.Name(), expectOffense, o)
		}
		if annotation != string(bufResult) {
			t.Errorf("%s annotationExpected=%s annotationFound=%s", f.Name(), string(bufResult), annotation)
		}
	}
}
