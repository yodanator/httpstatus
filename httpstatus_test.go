package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test findStatusCode returns correct struct and not found
func TestFindStatusCode(t *testing.T) {
	code, found := findStatusCode(200)
	if !found {
		t.Fatal("Expected to find code 200")
	}
	if code.Type != "Success" || code.Short == nil || *code.Short != "OK" {
		t.Errorf("Unexpected code struct: %+v", code)
	}

	_, found = findStatusCode(999)
	if found {
		t.Error("Should not find code 999")
	}
}

// Test searchStatusCodes finds by short and long description
func TestSearchStatusCodes(t *testing.T) {
	results := searchStatusCodes("teapot")
	if len(results) != 1 || results[0].Code != 418 {
		t.Errorf("Expected to find code 418, got %+v", results)
	}

	results = searchStatusCodes("not found")
	found := false
	for _, r := range results {
		if r.Code == 404 {
			found = true
		}
	}
	if !found {
		t.Error("Expected to find code 404 in search for 'not found'")
	}
}

// Test prepareOutputs respects flags
func TestPrepareOutputs(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}

	// Only short
	out := prepareOutputs(codes, false, false)
	if out[0].Long != nil {
		t.Error("Long should be nil when only short requested")
	}

	// Only long
	out = prepareOutputs(codes, true, false)
	if out[0].Short != nil {
		t.Error("Short should be nil when only long requested")
	}

	// Both
	out = prepareOutputs(codes, false, true)
	if out[0].Short == nil || out[0].Long == nil {
		t.Error("Both short and long should be present when all requested")
	}
}

// Test printText output (checks for panic)
func TestPrintText(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printText panicked: %v", r)
		}
	}()
	printText(codes)
}

// Test printJSON output
func TestPrintJSON(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(codes)
	if err != nil {
		t.Errorf("printJSON failed: %v", err)
	}
	if !strings.Contains(buf.String(), "OK") {
		t.Error("printJSON output missing expected content")
	}
}

// Test printXML output
func TestPrintXML(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	err := enc.Encode(codes)
	if err != nil {
		t.Errorf("printXML failed: %v", err)
	}
	if !strings.Contains(buf.String(), "OK") {
		t.Error("printXML output missing expected content")
	}
}

// Test printYAML output
func TestPrintYAML(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	err := enc.Encode(codes)
	if err != nil {
		t.Errorf("printYAML failed: %v", err)
	}
	if !strings.Contains(buf.String(), "OK") {
		t.Error("printYAML output missing expected content")
	}
}
