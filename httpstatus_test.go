/*
httpstatus - A CLI tool for looking up HTTP status codes in multiple formats.
Copyright (C) 2025  Adam Maltby

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

For questions, issues, or contributions, please visit:
https://github.com/yodanator/httpstatus
*/
package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"log"
	"os"
	"strconv"
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

// Test Partial Code Lookup
func TestPartialCodeLookup(t *testing.T) {
	// Test for all 4xx codes
	var matches []StatusCode
	for _, sc := range statusCodes {
		if strings.HasPrefix(strconv.Itoa(sc.Code), "4") {
			matches = append(matches, sc)
		}
	}
	if len(matches) == 0 {
		t.Error("Expected to find at least one 4xx code, found none")
	}
	for _, sc := range matches {
		if !strings.HasPrefix(strconv.Itoa(sc.Code), "4") {
			t.Errorf("Found code %d that does not start with 4", sc.Code)
		}
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

// Test printText output
func TestPrintText(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer

	printText(&buf, codes)
	output := buf.String()

	expected := []string{
		"Code: 200",
		"Type: Success",
		"Short: OK",
		"Long: All good",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain: %s\nGot: %s", exp, output)
		}
	}
}

// Test printJSON output
func TestPrintJSON(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer

	printJSON(&buf, codes, false)
	output := buf.String()

	// Parse output to verify valid JSON
	var decoded []StatusCode
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, output)
	}

	// Verify content
	if decoded[0].Code != 200 || *decoded[0].Short != "OK" {
		t.Errorf("Unexpected JSON content: %+v", decoded)
	}

	// Test pretty print
	buf.Reset()
	printJSON(&buf, codes, true)
	output = buf.String()
	if !strings.Contains(output, "  \"code\": 200") {
		t.Errorf("Pretty JSON missing expected indentation:\n%s", output)
	}
}

// Test printXML output
func TestPrintXML(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer

	printXML(&buf, codes, false)
	output := buf.String()

	// Parse output to verify valid XML
	var collection HTTPStatusCollection
	if err := xml.Unmarshal([]byte(output), &collection); err != nil {
		t.Fatalf("Invalid XML output: %v\nOutput: %s", err, output)
	}

	// Verify content
	if len(collection.Codes) != 1 || collection.Codes[0].Code != 200 {
		t.Errorf("Unexpected XML content: %+v", collection)
	}

	// Test pretty print
	buf.Reset()
	printXML(&buf, codes, true)
	output = buf.String()
	if !strings.Contains(output, "  <http_status>") {
		t.Errorf("Pretty XML missing expected indentation:\n%s", output)
	}
}

// Test printYAML output
func TestPrintYAML(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer

	printYAML(&buf, codes, false)
	output := buf.String()

	// Parse output to verify valid YAML
	var decoded []StatusCode
	if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("Invalid YAML output: %v\nOutput: %s", err, output)
	}

	// Verify content
	if decoded[0].Code != 200 || *decoded[0].Short != "OK" {
		t.Errorf("Unexpected YAML content: %+v", decoded)
	}

	// Test pretty with multiple items
	buf.Reset()
	codes = []StatusCode{
		{Code: 200, Type: "Success", Short: strPtr("OK")},
		{Code: 201, Type: "Success", Short: strPtr("Created")},
	}
	printYAML(&buf, codes, true)
	output = buf.String()
	if !strings.Contains(output, "---") {
		t.Errorf("Pretty YAML missing document separator:\n%s", output)
	}
}

// Test printTOML output
func TestPrintTOML(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer

	printTOML(&buf, codes)
	output := buf.String()

	expected := []string{
		"[200]",
		"type = \"Success\"",
		"short = \"OK\"",
		"long = \"All good\"",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected TOML output to contain: %s\nGot: %s", exp, output)
		}
	}
}

// Test printTable output
func TestPrintTable(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer

	printTable(&buf, codes)
	output := buf.String()

	expected := []string{
		"CODE\tTYPE\tSHORT\tLONG",
		"200\tSuccess\tOK\tAll good",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected table output to contain: %s\nGot: %s", exp, output)
		}
	}
}

// Test printMarkdown output
func TestPrintMarkdown(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer

	printMarkdown(&buf, codes)
	output := buf.String()

	expected := []string{
		"| Code | Type | Short | Long |",
		"|------|------|-------|------|",
		"| 200 | Success | OK | All good |",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected markdown output to contain: %s\nGot: %s", exp, output)
		}
	}
}

// Test printCSV output
func TestPrintCSV(t *testing.T) {
	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("All good")}}
	var buf bytes.Buffer

	printCSV(&buf, codes)
	output := buf.String()

	expected := []string{
		"Code,Type,Short,Long",
		"200,Success,OK,All good",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected CSV output to contain: %s\nGot: %s", exp, output)
		}
	}
}

// Test file output functionality
func TestWriteOutputToFiles(t *testing.T) {
	// Create temp directory for test files
	tempDir := t.TempDir()
	basePath := tempDir + "/output"

	formats := []struct {
		name    string
		enabled bool
	}{
		{"json", true},
		{"toml", true},
		{"csv", true},
	}

	codes := []StatusCode{{Code: 200, Type: "Success", Short: strPtr("OK")}}

	writeOutputToFiles(formats, codes, basePath)

	// Check that files were created
	expectedFiles := []string{
		basePath + ".json",
		basePath + ".toml",
		basePath + ".csv",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", file)
		}
	}
}

// Test output when no code or search is provided
func TestAllCodesOutput(t *testing.T) {
	// Simulate no code/search provided
	results := prepareOutputs(statusCodes, false, false)

	if len(results) != len(statusCodes) {
		t.Errorf("Expected %d codes, got %d", len(statusCodes), len(results))
	}
}

// Test file output with unknown format
func TestUnknownFormatFileOutput(t *testing.T) {
	tempDir := t.TempDir()
	basePath := tempDir + "/output"

	formats := []struct {
		name    string
		enabled bool
	}{
		{"unknown-format", true},
	}

	codes := []StatusCode{{Code: 200}}

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	writeOutputToFiles(formats, codes, basePath)

	if !strings.Contains(buf.String(), "Skipping unknown format") {
		t.Error("Expected warning about unknown format")
	}
}

// Test TOML escaping
func TestTOMLEscaping(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{`Hello "World"`, `Hello \"World\"`},
		{`Back\Slash`, `Back\\Slash`},
		{`No special chars`, `No special chars`},
	}

	for _, tc := range testCases {
		result := escapeTOMLString(tc.input)
		if result != tc.expected {
			t.Errorf("For input '%s', expected '%s', got '%s'", tc.input, tc.expected, result)
		}
	}
}

// Test prepareOutputs with empty long/short
func TestPrepareOutputsWithNil(t *testing.T) {
	// Create a test-specific status with nil descriptions
	testCode := StatusCode{Code: 999, Type: "Test", Short: nil, Long: nil}
	codes := []StatusCode{testCode}

	// Only short
	out := prepareOutputs(codes, false, false)
	if out[0].Short != nil {
		t.Error("Short should be nil for test code")
	}

	// Only long
	out = prepareOutputs(codes, true, false)
	if out[0].Long != nil {
		t.Error("Long should be nil for test code")
	}

	// Both
	out = prepareOutputs(codes, false, true)
	if out[0].Short != nil || out[0].Long != nil {
		t.Error("Both should be nil for test code")
	}
}

// Test printText with empty fields
func TestPrintTextWithNil(t *testing.T) {
	// Test code with nil descriptions
	testCode := StatusCode{Code: 999, Type: "Test", Short: nil, Long: nil}

	codes := []StatusCode{
		testCode,
		{Code: 404, Type: "Client Error", Short: strPtr("Not Found"), Long: strPtr("Resource not found")},
	}

	var buf bytes.Buffer
	printText(&buf, codes)
	output := buf.String()

	// Should contain only code and type for test code
	if !strings.Contains(output, "Code: 999") || !strings.Contains(output, "Type: Test") {
		t.Error("Output missing test code status")
	}

	// Should not contain "Short:" or "Long:" for test code
	if strings.Contains(output, "Short:") || strings.Contains(output, "Long:") {
		t.Error("Output should not contain short/long for test code")
	}

	// Should contain both for 404
	if !strings.Contains(output, "Short: Not Found") || !strings.Contains(output, "Long: Resource not found") {
		t.Error("Output missing 404 details")
	}
}
