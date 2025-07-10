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

	// Test single item
	printYAML(&buf, codes, false)
	output := buf.String()

	// Parse output to verify valid YAML
	var decoded StatusCode
	if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("Invalid YAML output: %v\nOutput: %s", err, output)
	}

	// Verify content
	if decoded.Code != 200 || *decoded.Short != "OK" {
		t.Errorf("Unexpected YAML content: %+v", decoded)
	}

	// Test multiple items with pretty output
	buf.Reset()
	codes = []StatusCode{
		{Code: 200, Type: "Success", Short: strPtr("OK")},
		{Code: 201, Type: "Success", Short: strPtr("Created")},
	}
	printYAML(&buf, codes, true)
	output = buf.String()

	// Split documents for multi-item output
	documents := strings.Split(strings.TrimSpace(output), "\n---\n")
	if len(documents) != 2 {
		t.Fatalf("Expected 2 YAML documents, got %d", len(documents))
	}

	// Parse first document
	var item1 StatusCode
	if err := yaml.Unmarshal([]byte(documents[0]), &item1); err != nil {
		t.Fatalf("Invalid YAML document 1: %v\n%v", err, documents[0])
	}
	if item1.Code != 200 || *item1.Short != "OK" {
		t.Errorf("Unexpected first item: %+v", item1)
	}

	// Parse second document
	var item2 StatusCode
	if err := yaml.Unmarshal([]byte(documents[1]), &item2); err != nil {
		t.Fatalf("Invalid YAML document 2: %v\n%v", err, documents[1])
	}
	if item2.Code != 201 || *item2.Short != "Created" {
		t.Errorf("Unexpected second item: %+v", item2)
	}

	// Verify document separator
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

	// Split into lines and trim space
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("Not enough lines in output: %d", len(lines))
	}

	// Check headers
	expectedHeaders := []string{"CODE", "TYPE", "SHORT", "LONG"}
	for _, header := range expectedHeaders {
		if !strings.Contains(lines[0], header) {
			t.Errorf("Header missing %q in: %s", header, lines[0])
		}
	}

	// Check data row
	expectedData := []string{"200", "Success", "OK", "All good"}
	for _, data := range expectedData {
		if !strings.Contains(lines[1], data) {
			t.Errorf("Data missing %q in: %s", data, lines[1])
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

	// Split output by code sections
	sections := strings.Split(output, "---")
	if len(sections) < 2 {
		t.Fatalf("Expected at least 2 sections, got %d", len(sections))
	}

	// Section 1: Test code (999)
	section1 := sections[0]
	if !strings.Contains(section1, "Code: 999") || !strings.Contains(section1, "Type: Test") {
		t.Error("Output missing test code status")
	}
	if strings.Contains(section1, "Short:") || strings.Contains(section1, "Long:") {
		t.Error("Test code section should not contain short/long headers")
	}

	// Section 2: 404 code
	section2 := sections[1]
	if !strings.Contains(section2, "Short: Not Found") || !strings.Contains(section2, "Long: Resource not found") {
		t.Error("404 section missing expected details")
	}
}

// Test multi-code input
func TestMultiCodeInput(t *testing.T) {
	results, err := processInputs("200,404", "", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 codes, got %d", len(results))
	}
	found200 := false
	found404 := false
	for _, r := range results {
		if r.Code == 200 {
			found200 = true
		}
		if r.Code == 404 {
			found404 = true
		}
	}
	if !found200 || !found404 {
		t.Errorf("Missing expected codes: found200=%v, found404=%v", found200, found404)
	}
}

// Test combined search and codes
func TestCombinedSearchAndCodes(t *testing.T) {
	results, err := processInputs("404", "teapot", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) < 2 {
		t.Fatalf("Expected at least 2 codes, got %d", len(results))
	}
	found404 := false
	found418 := false
	for _, r := range results {
		if r.Code == 404 {
			found404 = true
		}
		if r.Code == 418 {
			found418 = true
		}
	}
	if !found404 || !found418 {
		t.Errorf("Missing expected codes: found404=%v, found418=%v", found404, found418)
	}
}

// Test partial code input
func TestPartialCodeInput(t *testing.T) {
	results, err := processInputs("4,5", "", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Count 4xx and 5xx codes
	count4xx := 0
	count5xx := 0
	for _, r := range results {
		if r.Code >= 400 && r.Code < 500 {
			count4xx++
		} else if r.Code >= 500 && r.Code < 600 {
			count5xx++
		}
	}

	if count4xx == 0 || count5xx == 0 {
		t.Errorf("Expected both 4xx and 5xx codes, got 4xx=%d, 5xx=%d", count4xx, count5xx)
	}
}

// Test duplicate prevention
func TestDuplicatePrevention(t *testing.T) {
	results, err := processInputs("404,404,4", "", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify no duplicates
	codes := make(map[int]bool)
	for _, r := range results {
		if codes[r.Code] {
			t.Errorf("Duplicate found for code %d", r.Code)
		}
		codes[r.Code] = true
	}
}

// Test invalid code input
func TestInvalidCodeInput(t *testing.T) {
	_, err := processInputs("abc", "", nil)
	if err == nil {
		t.Error("Expected error for invalid code input")
	} else {
		expected := "invalid status code: 'abc' - must be numeric"
		if err.Error() != expected {
			t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
		}
	}
}

// Test empty input
func TestEmptyInput(t *testing.T) {
	results, err := processInputs("", "", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != len(statusCodes) {
		t.Errorf("Expected all codes, got %d instead of %d", len(results), len(statusCodes))
	}
}
