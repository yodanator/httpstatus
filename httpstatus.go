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
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

// StatusCode represents an HTTP status code with metadata
type StatusCode struct {
	Code  int     `json:"code" xml:"code" yaml:"code"`
	Type  string  `json:"type" xml:"type" yaml:"type"`
	Short *string `json:"short,omitempty" xml:"short,omitempty" yaml:"short,omitempty"`
	Long  *string `json:"long,omitempty" xml:"long,omitempty" yaml:"long,omitempty"`
}

// HTTPStatusCollection wraps status codes for XML output
type HTTPStatusCollection struct {
	XMLName xml.Name     `xml:"http_statuses"`
	Codes   []StatusCode `xml:"http_status"`
}

// Application variables (set at build time)
var (
	AppName    = "httpstatus"
	AppVersion = "dev"
	GitHubURL  = "https://github.com/yodanator/httpstatus"
)

var statusCodes = []StatusCode{
	// 1xx Informational
	{Code: 100, Type: "Informational", Short: strPtr("Continue"), Long: strPtr("Server received request headers; client should proceed with body")},
	{Code: 101, Type: "Informational", Short: strPtr("Switching Protocols"), Long: strPtr("Server agrees to switch protocols as requested")},
	{Code: 102, Type: "Informational", Short: strPtr("Processing"), Long: strPtr("Server is processing request but no response available yet")},
	{Code: 103, Type: "Informational", Short: strPtr("Early Hints"), Long: strPtr("Suggests preloading resources while server prepares response")},

	// 2xx Success
	{Code: 200, Type: "Success", Short: strPtr("OK"), Long: strPtr("Standard response for successful HTTP requests")},
	{Code: 201, Type: "Success", Short: strPtr("Created"), Long: strPtr("New resource created as result of request")},
	{Code: 202, Type: "Success", Short: strPtr("Accepted"), Long: strPtr("Request accepted for processing but not completed")},
	{Code: 203, Type: "Success", Short: strPtr("Non-Authoritative Information"), Long: strPtr("Metadata not from origin server but local/third-party copy")},
	{Code: 204, Type: "Success", Short: strPtr("No Content"), Long: strPtr("Successfully processed but no content to return")},
	{Code: 205, Type: "Success", Short: strPtr("Reset Content"), Long: strPtr("Client should reset document view that caused request")},
	{Code: 206, Type: "Success", Short: strPtr("Partial Content"), Long: strPtr("Server delivering partial resource due to range header")},
	{Code: 207, Type: "Success", Short: strPtr("Multi-Status"), Long: strPtr("Conveys multiple response codes for sub-requests (WebDAV)")},
	{Code: 208, Type: "Success", Short: strPtr("Already Reported"), Long: strPtr("Prevents repeated enumeration of DAV binding members")},
	{Code: 226, Type: "Success", Short: strPtr("IM Used"), Long: strPtr("Response includes instance manipulations applied to resource")},

	// 3xx Redirection
	{Code: 300, Type: "Redirection", Short: strPtr("Multiple Choices"), Long: strPtr("Multiple options available for resource (agent-driven negotiation)")},
	{Code: 301, Type: "Redirection", Short: strPtr("Moved Permanently"), Long: strPtr("Resource permanently moved to new URI")},
	{Code: 302, Type: "Redirection", Short: strPtr("Found"), Long: strPtr("Resource temporarily available at different URI")},
	{Code: 303, Type: "Redirection", Short: strPtr("See Other"), Long: strPtr("Response can be found under another URI using GET")},
	{Code: 304, Type: "Redirection", Short: strPtr("Not Modified"), Long: strPtr("Resource not modified since version in request headers")},
	{Code: 305, Type: "Redirection", Short: strPtr("Use Proxy"), Long: strPtr("Resource must be accessed through proxy (deprecated)")},
	{Code: 306, Type: "Redirection", Short: strPtr("(Unused)"), Long: strPtr("Reserved status code, no longer used")},
	{Code: 307, Type: "Redirection", Short: strPtr("Temporary Redirect"), Long: strPtr("Request should be repeated with another URI")},
	{Code: 308, Type: "Redirection", Short: strPtr("Permanent Redirect"), Long: strPtr("Resource permanently moved with same HTTP method")},

	// 4xx Client Errors
	{Code: 400, Type: "Client Error", Short: strPtr("Bad Request"), Long: strPtr("Server cannot process request due to client error")},
	{Code: 401, Type: "Client Error", Short: strPtr("Unauthorized"), Long: strPtr("Authentication required and failed/not provided")},
	{Code: 402, Type: "Client Error", Short: strPtr("Payment Required"), Long: strPtr("Reserved for future digital payment systems")},
	{Code: 403, Type: "Client Error", Short: strPtr("Forbidden"), Long: strPtr("Client lacks permissions for requested resource")},
	{Code: 404, Type: "Client Error", Short: strPtr("Not Found"), Long: strPtr("Requested resource could not be found")},
	{Code: 405, Type: "Client Error", Short: strPtr("Method Not Allowed"), Long: strPtr("HTTP method not supported for this resource")},
	{Code: 406, Type: "Client Error", Short: strPtr("Not Acceptable"), Long: strPtr("No content matching Accept header criteria")},
	{Code: 407, Type: "Client Error", Short: strPtr("Proxy Authentication Required"), Long: strPtr("Client must authenticate with proxy first")},
	{Code: 408, Type: "Client Error", Short: strPtr("Request Timeout"), Long: strPtr("Server timed out waiting for request")},
	{Code: 409, Type: "Client Error", Short: strPtr("Conflict"), Long: strPtr("Request conflicts with current resource state")},
	{Code: 410, Type: "Client Error", Short: strPtr("Gone"), Long: strPtr("Resource permanently removed with no forwarding address")},
	{Code: 411, Type: "Client Error", Short: strPtr("Length Required"), Long: strPtr("Server requires Content-Length header")},
	{Code: 412, Type: "Client Error", Short: strPtr("Precondition Failed"), Long: strPtr("Server does not meet request preconditions")},
	{Code: 413, Type: "Client Error", Short: strPtr("Content Too Large"), Long: strPtr("Request exceeds server size limits")},
	{Code: 414, Type: "Client Error", Short: strPtr("URI Too Long"), Long: strPtr("Request URI exceeds server processing capacity")},
	{Code: 415, Type: "Client Error", Short: strPtr("Unsupported Media Type"), Long: strPtr("Media format not supported by server")},
	{Code: 416, Type: "Client Error", Short: strPtr("Range Not Satisfiable"), Long: strPtr("Cannot satisfy Range header request")},
	{Code: 417, Type: "Client Error", Short: strPtr("Expectation Failed"), Long: strPtr("Server cannot meet Expect header requirements")},
	{Code: 418, Type: "Client Error", Short: strPtr("I'm a teapot"), Long: strPtr("Server refuses to brew coffee (RFC 2324)")},
	{Code: 420, Type: "Client Error", Short: strPtr("Enhance Your Calm"), Long: strPtr("Client is being rate-limited (Twitter)")},
	{Code: 421, Type: "Client Error", Short: strPtr("Misdirected Request"), Long: strPtr("Request directed at non-responsive server")},
	{Code: 422, Type: "Client Error", Short: strPtr("Unprocessable Entity"), Long: strPtr("Well-formed request with semantic errors (WebDAV)")},
	{Code: 423, Type: "Client Error", Short: strPtr("Locked"), Long: strPtr("Resource is locked (WebDAV)")},
	{Code: 424, Type: "Client Error", Short: strPtr("Failed Dependency"), Long: strPtr("Request failed due to previous failure (WebDAV)")},
	{Code: 425, Type: "Client Error", Short: strPtr("Too Early"), Long: strPtr("Server unwilling to risk processing replay request")},
	{Code: 426, Type: "Client Error", Short: strPtr("Upgrade Required"), Long: strPtr("Client should switch to different protocol")},
	{Code: 428, Type: "Client Error", Short: strPtr("Precondition Required"), Long: strPtr("Origin server requires conditional request")},
	{Code: 429, Type: "Client Error", Short: strPtr("Too Many Requests"), Long: strPtr("Exceeded rate limit for requests")},
	{Code: 431, Type: "Client Error", Short: strPtr("Request Header Fields Too Large"), Long: strPtr("Header fields exceed server size limit")},
	{Code: 444, Type: "Client Error", Short: strPtr("No Response"), Long: strPtr("Server returns no information and closes connection (Nginx)")},
	{Code: 449, Type: "Client Error", Short: strPtr("Retry With"), Long: strPtr("Request should be retried after appropriate action (Microsoft)")},
	{Code: 450, Type: "Client Error", Short: strPtr("Blocked by Windows Parental Controls"), Long: strPtr("Access blocked by Windows Parental Controls (Microsoft)")},
	{Code: 451, Type: "Client Error", Short: strPtr("Unavailable For Legal Reasons"), Long: strPtr("Resource access denied for legal reasons")},
	{Code: 499, Type: "Client Error", Short: strPtr("Client Closed Request"), Long: strPtr("Connection closed by client during processing (Nginx)")},

	// 5xx Server Errors
	{Code: 500, Type: "Server Error", Short: strPtr("Internal Server Error"), Long: strPtr("Generic error when server encounters unexpected condition")},
	{Code: 501, Type: "Server Error", Short: strPtr("Not Implemented"), Long: strPtr("Server lacks ability to fulfill request")},
	{Code: 502, Type: "Server Error", Short: strPtr("Bad Gateway"), Long: strPtr("Invalid response from upstream server")},
	{Code: 503, Type: "Server Error", Short: strPtr("Service Unavailable"), Long: strPtr("Server temporarily overloaded or down")},
	{Code: 504, Type: "Server Error", Short: strPtr("Gateway Timeout"), Long: strPtr("Upstream server failed to respond in time")},
	{Code: 505, Type: "Server Error", Short: strPtr("HTTP Version Not Supported"), Long: strPtr("Server doesn't support HTTP protocol version")},
	{Code: 506, Type: "Server Error", Short: strPtr("Variant Also Negotiates"), Long: strPtr("Server configuration error in content negotiation")},
	{Code: 507, Type: "Server Error", Short: strPtr("Insufficient Storage"), Long: strPtr("Cannot store representation needed to complete request")},
	{Code: 508, Type: "Server Error", Short: strPtr("Loop Detected"), Long: strPtr("Infinite loop detected during processing")},
	{Code: 510, Type: "Server Error", Short: strPtr("Not Extended"), Long: strPtr("Further extensions required to fulfill request")},
	{Code: 511, Type: "Server Error", Short: strPtr("Network Authentication Required"), Long: strPtr("Client needs authentication for network access")},
}

// Package-level variables for flags
var (
	codeFlag       = flag.String("c", "", "HTTP status code(s) (comma-separated) (either this, search, or none for all codes)")
	searchFlag     = flag.String("search", "", "Search for HTTP status codes by keyword in short or long description")
	longFlag       = flag.Bool("l", false, "Output long description")
	allFlag        = flag.Bool("a", false, "Output both short and long descriptions")
	jsonOutput     = flag.Bool("json", false, "Output as JSON (raw)")
	jsonPretty     = flag.Bool("json-pretty", false, "Output as pretty JSON")
	xmlOutput      = flag.Bool("xml", false, "Output as XML (raw)")
	xmlPretty      = flag.Bool("xml-pretty", false, "Output as pretty XML")
	yamlOutput     = flag.Bool("yaml", false, "Output as YAML (raw)")
	yamlPretty     = flag.Bool("yaml-pretty", false, "Output as pretty YAML")
	tomlOutput     = flag.Bool("toml", false, "Output as TOML")
	tableOutput    = flag.Bool("table", false, "Output as text table")
	markdownOutput = flag.Bool("markdown", false, "Output as Markdown table")
	csvOutput      = flag.Bool("csv", false, "Output as CSV")
	toFileBase     = flag.String("to-file", "", "Save output to files with base name (automatic extensions)")
	helpFlag       = flag.Bool("help", false, "Show help information")
	versionFlag    = flag.Bool("version", false, "Show version information")
)

func main() {
	// Aliases for flags
	flag.StringVar(codeFlag, "code", "", "HTTP status code(s) (comma-separated) (either this, search, or none for all codes)")
	flag.StringVar(searchFlag, "s", "", "Search for HTTP status codes by keyword (shorthand)")
	flag.BoolVar(longFlag, "long", false, "Output long description")
	flag.BoolVar(allFlag, "all", false, "Output both short and long descriptions")

	flag.Parse()

	// Handle help flag
	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	// Handle version flag
	if *versionFlag {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		fmt.Printf("Source: %s\n", GitHubURL)
		os.Exit(0)
	}

	// Process inputs
	results := processInputs(*codeFlag, *searchFlag, flag.Args())

	// Prepare output based on flags
	outputs := prepareOutputs(results, *longFlag, *allFlag)

	// Handle multiple output formats
	outputFormats := []struct {
		name    string
		enabled bool
	}{
		{"json", *jsonOutput},
		{"json-pretty", *jsonPretty},
		{"xml", *xmlOutput},
		{"xml-pretty", *xmlPretty},
		{"yaml", *yamlOutput},
		{"yaml-pretty", *yamlPretty},
		{"toml", *tomlOutput},
		{"table", *tableOutput},
		{"markdown", *markdownOutput},
		{"csv", *csvOutput},
	}

	// Handle file output if requested
	if *toFileBase != "" {
		writeOutputToFiles(outputFormats, outputs, *toFileBase)
	} else {
		anyOutput := false
		for _, format := range outputFormats {
			if format.enabled {
				anyOutput = true
				switch format.name {
				case "json":
					printJSON(os.Stdout, outputs, false)
				case "json-pretty":
					printJSON(os.Stdout, outputs, true)
				case "xml":
					printXML(os.Stdout, outputs, false)
				case "xml-pretty":
					printXML(os.Stdout, outputs, true)
				case "yaml":
					printYAML(os.Stdout, outputs, false)
				case "yaml-pretty":
					printYAML(os.Stdout, outputs, true)
				case "toml":
					printTOML(os.Stdout, outputs)
				case "table":
					printTable(os.Stdout, outputs)
				case "markdown":
					printMarkdown(os.Stdout, outputs)
				case "csv":
					printCSV(os.Stdout, outputs)
				}
			}
		}

		// Default text output if no format specified
		if !anyOutput {
			printText(os.Stdout, outputs)
		}
	}
}

// processInputs handles the input processing and returns the status codes to display
func processInputs(codeStr, searchStr string, args []string) []StatusCode {
	var results []StatusCode
	seen := make(map[int]bool) // Track seen codes to prevent duplicates

	// Helper to add status code if not seen
	addIfNotSeen := func(sc StatusCode) {
		if !seen[sc.Code] {
			seen[sc.Code] = true
			results = append(results, sc)
		}
	}

	// Process code flag (comma-separated)
	if codeStr != "" {
		parts := strings.Split(codeStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			// Try to parse as exact code
			if codeInt, err := strconv.Atoi(part); err == nil {
				if sc, found := findStatusCode(codeInt); found {
					addIfNotSeen(sc)
					continue
				}
			}

			// Handle partial code match
			var matches []StatusCode
			for _, sc := range statusCodes {
				codeStr := strconv.Itoa(sc.Code)
				if strings.HasPrefix(codeStr, part) {
					matches = append(matches, sc)
				}
			}
			if len(matches) == 0 {
				log.Fatalf("No HTTP status codes found matching: '%s'", part)
			}
			for _, sc := range matches {
				addIfNotSeen(sc)
			}
		}
	}

	// Process positional arguments (comma-separated or single)
	if len(args) > 0 {
		for _, arg := range args {
			argParts := strings.Split(arg, ",")
			for _, part := range argParts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				// Try to parse as exact code
				if codeInt, err := strconv.Atoi(part); err == nil {
					if sc, found := findStatusCode(codeInt); found {
						addIfNotSeen(sc)
						continue
					}
				}

				// Handle partial code match
				var matches []StatusCode
				for _, sc := range statusCodes {
					codeStr := strconv.Itoa(sc.Code)
					if strings.HasPrefix(codeStr, part) {
						matches = append(matches, sc)
					}
				}
				if len(matches) == 0 {
					log.Fatalf("No HTTP status codes found matching: '%s'", part)
				}
				for _, sc := range matches {
					addIfNotSeen(sc)
				}
			}
		}
	}

	// Process search
	if searchStr != "" {
		searchResults := searchStatusCodes(searchStr)
		for _, sc := range searchResults {
			addIfNotSeen(sc)
		}
	}

	// If no filters applied, show all codes
	if codeStr == "" && len(args) == 0 && searchStr == "" {
		results = statusCodes
	} else if len(results) == 0 {
		log.Fatal("No HTTP status codes found matching your criteria")
	}

	return results
}

func printHelp() {
	fmt.Printf("%s %s\n\n", AppName, AppVersion)
	fmt.Println("A CLI tool for looking up HTTP status codes with multiple output formats")
	fmt.Printf("Source code and license: %s\n\n", GitHubURL)

	fmt.Println("USAGE:")
	fmt.Println("  httpstatus [flags] [status_code|partial_code]")
	fmt.Println("  httpstatus --search \"search term\"")
	fmt.Println("  httpstatus --code \"200,404\"")
	fmt.Println("  httpstatus \"4,5\" --json-pretty")
	fmt.Println("  httpstatus --to-file output --json --csv")
	fmt.Println("  httpstatus --table  # Show all codes in table format")
	fmt.Println("\nFLAGS:")
	fmt.Println("  -c, --code <codes>   HTTP status code(s) to look up (comma-separated)")
	fmt.Println("  -s, --search <term>  Search status codes by keyword")
	fmt.Println("  -l, --long           Show long description only")
	fmt.Println("  -a, --all            Show both short and long descriptions")
	fmt.Println("  --json               Output as JSON")
	fmt.Println("  --json-pretty        Output as formatted JSON")
	fmt.Println("  --xml                Output as XML")
	fmt.Println("  --xml-pretty         Output as formatted XML")
	fmt.Println("  --yaml               Output as YAML")
	fmt.Println("  --yaml-pretty        Output as formatted YAML")
	fmt.Println("  --toml               Output as TOML")
	fmt.Println("  --table              Output as text table")
	fmt.Println("  --markdown           Output as Markdown table")
	fmt.Println("  --csv                Output as CSV")
	fmt.Println("  --to-file <base>     Save output to files with base name (automatic extensions)")
	fmt.Println("  --help               Show this help message")
	fmt.Println("  --version            Show version information")

	fmt.Println("\nEXAMPLES:")
	fmt.Println("  Look up multiple status codes:")
	fmt.Println("      httpstatus -c \"200,404\"")
	fmt.Println("  Look up all 4xx and 5xx codes:")
	fmt.Println("      httpstatus \"4,5\"")
	fmt.Println("  Search for 'not found' and show 404:")
	fmt.Println("      httpstatus --search \"not found\" --code 404")
	fmt.Println("  Get status 200 and 201 in JSON format:")
	fmt.Println("      httpstatus 200,201 --json")
	fmt.Println("  Export all 2xx codes to CSV:")
	fmt.Println("      httpstatus 2 --csv --to-file success_codes")

	fmt.Println("\nPARTIAL CODE LOOKUP:")
	fmt.Println("  You can enter just the first digit (e.g., '4') or first two digits (e.g., '41')")
	fmt.Println("  to list all HTTP status codes in that set. This is separate from --search.")
	fmt.Println("  Multiple partial codes can be combined with commas: '4,5' shows all client and server errors")

	fmt.Println("\nFILE OUTPUT:")
	fmt.Println("  Use --to-file with a base filename to save output to files. The tool will automatically")
	fmt.Println("  add appropriate extensions based on the output format (.json, .yaml, .md, etc.).")
	fmt.Println("  Multiple formats can be saved simultaneously by specifying multiple output flags.")

	fmt.Println("\nLICENSE:")
	fmt.Println("  By using this application, you accept the license terms and warranty disclaimer")
	fmt.Println("  described in the LICENSE file at:")
	fmt.Println("    https://github.com/yodanator/httpstatus/blob/main/LICENSE")
	fmt.Println("  (This software is distributed under the GNU GPL v3. See LICENSE for details.)")

	fmt.Println("\nCONTACT:")
	fmt.Println("  For questions, issues, or contributions, please visit:")
	fmt.Println("    https://github.com/yodanator/httpstatus")
}

// searchStatusCodes finds status codes matching the search term
func searchStatusCodes(term string) []StatusCode {
	var results []StatusCode
	lowerTerm := strings.ToLower(term)

	for _, sc := range statusCodes {
		shortLower := ""
		if sc.Short != nil {
			shortLower = strings.ToLower(*sc.Short)
		}
		longLower := ""
		if sc.Long != nil {
			longLower = strings.ToLower(*sc.Long)
		}

		if strings.Contains(shortLower, lowerTerm) ||
			strings.Contains(longLower, lowerTerm) {
			results = append(results, sc)
		}
	}
	return results
}

// findStatusCode looks up a specific status code
func findStatusCode(code int) (StatusCode, bool) {
	for _, sc := range statusCodes {
		if sc.Code == code {
			return sc, true
		}
	}
	return StatusCode{}, false
}

// prepareOutputs creates output structures based on flags
func prepareOutputs(codes []StatusCode, long, all bool) []StatusCode {
	var outputs []StatusCode

	for _, sc := range codes {
		output := sc
		if all {
			// Keep both short and long
		} else if long {
			output.Short = nil // Omit short when only long is requested
		} else {
			output.Long = nil // Omit long when only short is requested
		}
		outputs = append(outputs, output)
	}
	return outputs
}

// printText outputs human-readable text
func printText(w io.Writer, codes []StatusCode) {
	for i, sc := range codes {
		if i > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "---")
		}
		fmt.Fprintf(w, "Code: %d\nType: %s\n", sc.Code, sc.Type)
		if sc.Short != nil && sc.Long != nil {
			fmt.Fprintf(w, "Short: %s\nLong: %s\n", *sc.Short, *sc.Long)
		} else if sc.Long != nil {
			fmt.Fprintf(w, "Long: %s\n", *sc.Long)
		} else if sc.Short != nil {
			fmt.Fprintf(w, "Short: %s\n", *sc.Short)
		}
	}
}

// printJSON outputs JSON format
func printJSON(w io.Writer, codes []StatusCode, pretty bool) {
	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(codes, "", "  ")
	} else {
		data, err = json.Marshal(codes)
	}

	if err != nil {
		log.Fatalf("JSON error: %v", err)
	}
	fmt.Fprintln(w, string(data))
}

// printXML outputs XML format
func printXML(w io.Writer, codes []StatusCode, pretty bool) {
	// Wrap in a root element for valid XML
	collection := HTTPStatusCollection{Codes: codes}

	var data []byte
	var err error

	if pretty {
		data, err = xml.MarshalIndent(collection, "", "  ")
	} else {
		data, err = xml.Marshal(collection)
	}

	if err != nil {
		log.Fatalf("XML error: %v", err)
	}

	// Add XML header
	fmt.Fprint(w, xml.Header+string(data))
}

// printYAML outputs YAML format
func printYAML(w io.Writer, codes []StatusCode, pretty bool) {
	for i, sc := range codes {
		if pretty && i > 0 {
			fmt.Fprintln(w, "---")
		}
		data, err := yaml.Marshal(sc)
		if err != nil {
			log.Fatalf("YAML error: %v", err)
		}
		fmt.Fprintln(w, string(data))
	}
}

// printTOML outputs TOML format
func printTOML(w io.Writer, codes []StatusCode) {
	for i, sc := range codes {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "[%d]\n", sc.Code)
		fmt.Fprintf(w, "type = \"%s\"\n", sc.Type)

		if sc.Short != nil {
			fmt.Fprintf(w, "short = \"%s\"\n", escapeTOMLString(*sc.Short))
		}

		if sc.Long != nil {
			fmt.Fprintf(w, "long = \"%s\"\n", escapeTOMLString(*sc.Long))
		}
	}
}

// escapeTOMLString escapes special characters in TOML strings
func escapeTOMLString(s string) string {
	// TOML requires escaping backslashes and quotes
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "\"", "\\\"")
}

// printTable outputs tabular text format
func printTable(w io.Writer, codes []StatusCode) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	// Header
	fmt.Fprintln(tw, "CODE\tTYPE\tSHORT\tLONG")

	for _, sc := range codes {
		short := ""
		if sc.Short != nil {
			short = *sc.Short
		}

		long := ""
		if sc.Long != nil {
			long = *sc.Long
		}

		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", sc.Code, sc.Type, short, long)
	}
}

// printMarkdown outputs Markdown table format
func printMarkdown(w io.Writer, codes []StatusCode) {
	// Table header
	fmt.Fprintln(w, "| Code | Type | Short | Long |")
	fmt.Fprintln(w, "|------|------|-------|------|")

	for _, sc := range codes {
		short := ""
		if sc.Short != nil {
			short = *sc.Short
		}

		long := ""
		if sc.Long != nil {
			long = *sc.Long
		}

		fmt.Fprintf(w, "| %d | %s | %s | %s |\n", sc.Code, sc.Type, short, long)
	}
}

// printCSV outputs CSV format
func printCSV(w io.Writer, codes []StatusCode) {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Write header
	cw.Write([]string{"Code", "Type", "Short", "Long"})

	for _, sc := range codes {
		short := ""
		if sc.Short != nil {
			short = *sc.Short
		}

		long := ""
		if sc.Long != nil {
			long = *sc.Long
		}

		cw.Write([]string{
			strconv.Itoa(sc.Code),
			sc.Type,
			short,
			long,
		})
	}
}

// writeOutputToFiles saves output to files based on format
func writeOutputToFiles(formats []struct {
	name    string
	enabled bool
}, codes []StatusCode, basePath string) {
	extMap := map[string]string{
		"json":        ".json",
		"json-pretty": ".json",
		"xml":         ".xml",
		"xml-pretty":  ".xml",
		"yaml":        ".yaml",
		"yaml-pretty": ".yaml",
		"toml":        ".toml",
		"table":       ".txt",
		"markdown":    ".md",
		"csv":         ".csv",
	}

	for _, format := range formats {
		if !format.enabled {
			continue
		}

		ext, ok := extMap[format.name]
		if !ok {
			log.Printf("Skipping unknown format: %s", format.name)
			continue
		}

		filename := basePath + ext
		file, err := os.Create(filename)
		if err != nil {
			log.Printf("Error creating %s: %v", filename, err)
			continue
		}
		defer file.Close()

		switch format.name {
		case "json":
			printJSON(file, codes, false)
		case "json-pretty":
			printJSON(file, codes, true)
		case "xml":
			printXML(file, codes, false)
		case "xml-pretty":
			printXML(file, codes, true)
		case "yaml":
			printYAML(file, codes, false)
		case "yaml-pretty":
			printYAML(file, codes, true)
		case "toml":
			printTOML(file, codes)
		case "table":
			printTable(file, codes)
		case "markdown":
			printMarkdown(file, codes)
		case "csv":
			printCSV(file, codes)
		}
		log.Printf("Output saved to %s", filename)
	}
}
