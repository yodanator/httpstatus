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
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

// StatusCode represents an HTTP status code with metadata
type StatusCode struct {
	Code  int     `json:"code" xml:"code"`
	Type  string  `json:"type" xml:"type"`
	Short *string `json:"short,omitempty" xml:"short,omitempty"`
	Long  *string `json:"long,omitempty" xml:"long,omitempty"`
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
	GitHubURL  = "https://github.com/yourusername/httpstatus"
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

func main() {
	// Command-line flags
	code := flag.Int("c", 0, "HTTP status code (either this or search is required)")
	search := flag.String("search", "", "Search for HTTP status codes by keyword in short or long description")
	long := flag.Bool("l", false, "Output long description")
	all := flag.Bool("a", false, "Output both short and long descriptions")
	jsonOutput := flag.Bool("json", false, "Output as JSON (raw)")
	jsonPretty := flag.Bool("json-pretty", false, "Output as pretty JSON")
	xmlOutput := flag.Bool("xml", false, "Output as XML (raw)")
	xmlPretty := flag.Bool("xml-pretty", false, "Output as pretty XML")
	yamlOutput := flag.Bool("yaml", false, "Output as YAML (raw)")
	yamlPretty := flag.Bool("yaml-pretty", false, "Output as pretty YAML")

	// Aliases for flags
	flag.IntVar(code, "code", 0, "HTTP status code (either this or search is required)")
	flag.StringVar(search, "s", "", "Search for HTTP status codes by keyword (shorthand)")
	flag.BoolVar(long, "long", false, "Output long description")
	flag.BoolVar(all, "all", false, "Output both short and long descriptions")

	// Help and version flags
	help := flag.Bool("help", false, "Show help information")
	version := flag.Bool("version", false, "Show version information")

	flag.Parse()

	// Handle help flag
	if *help {
		printHelp()
		os.Exit(0)
	}

	// Handle version flag
	if *version {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		fmt.Printf("Source: %s\n", GitHubURL)
		os.Exit(0)
	}

	// Get positional arguments
	args := flag.Args()

	// Validate input: either code or search must be provided
	if *code == 0 && *search == "" && len(args) == 0 {
		printHelp()
		log.Fatal("\nError: Either HTTP code (-c) or search term (-search) must be specified")
	}

	// Check for conflicting options
	if *code != 0 && *search != "" {
		log.Fatal("Error: Cannot specify both -c and -search simultaneously")
	}
	if *all && *long {
		fmt.Fprintln(os.Stderr, "Note: --long ignored because --all was specified")
		*long = false
	}

	var results []StatusCode

	// Handle search mode
	if *search != "" {
		results = searchStatusCodes(*search)
		if len(results) == 0 {
			log.Fatalf("No HTTP status codes found matching search: '%s'", *search)
		}
	} else {
		var resultsSet bool
		var httpCode int
		if *code != 0 {
			httpCode = *code
		} else if len(args) > 0 {
			codeArg := args[0]
			if len(codeArg) < 3 {
				var matches []StatusCode
				for _, sc := range statusCodes {
					codeStr := strconv.Itoa(sc.Code)
					if strings.HasPrefix(codeStr, codeArg) {
						matches = append(matches, sc)
					}
				}
				if len(matches) == 0 {
					log.Fatalf("No HTTP status codes found starting with '%s'", codeArg)
				}
				results = matches
				resultsSet = true
			} else {
				parsedCode, err := strconv.Atoi(codeArg)
				if err != nil {
					log.Fatalf("Error: Invalid HTTP code '%s' - must be a number", codeArg)
				}
				httpCode = parsedCode
			}
		} else {
			log.Fatal("Error: HTTP code must be specified with -c/--code or as a positional argument")
		}

		if !resultsSet {
			result, found := findStatusCode(httpCode)
			if !found {
				log.Fatalf("Error: HTTP status code %d not found", httpCode)
			}
			results = []StatusCode{result}
		}
	}

	// Prepare output based on flags
	outputs := prepareOutputs(results, *long, *all)

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
	}

	anyOutput := false
	for _, format := range outputFormats {
		if format.enabled {
			anyOutput = true
			switch format.name {
			case "json":
				printJSON(outputs, false)
			case "json-pretty":
				printJSON(outputs, true)
			case "xml":
				printXML(outputs, false)
			case "xml-pretty":
				printXML(outputs, true)
			case "yaml":
				printYAML(outputs, false)
			case "yaml-pretty":
				printYAML(outputs, true)
			}
		}
	}

	// Default text output if no format specified
	if !anyOutput {
		printText(outputs)
	}
}

func printHelp() {
	fmt.Printf("%s v%s\n\n", AppName, AppVersion)
	fmt.Println("A CLI tool for looking up HTTP status codes with multiple output formats")
	fmt.Printf("Source code and license: %s\n\n", GitHubURL)

	fmt.Println("USAGE:")
	fmt.Println("  httpstatus [flags] [status_code|partial_code]")
	fmt.Println("  httpstatus --search \"search term\"")
	fmt.Println("  httpstatus --code 404")
	fmt.Println("  httpstatus 200 --json-pretty")
	fmt.Println("\nFLAGS:")
	fmt.Println("  -c, --code <number>     HTTP status code to look up")
	fmt.Println("  -s, --search <term>     Search status codes by keyword")
	fmt.Println("  -l, --long              Show long description only")
	fmt.Println("  -a, --all               Show both short and long descriptions")
	fmt.Println("  --json                  Output as JSON")
	fmt.Println("  --json-pretty           Output as formatted JSON")
	fmt.Println("  --xml                   Output as XML")
	fmt.Println("  --xml-pretty            Output as formatted XML")
	fmt.Println("  --yaml                  Output as YAML")
	fmt.Println("  --yaml-pretty           Output as formatted YAML")
	fmt.Println("  --help                  Show this help message")
	fmt.Println("  --version               Show version information")

	fmt.Println("\nEXAMPLES:")
	fmt.Println("  Look up status code 404:")
	fmt.Println("      httpstatus -c 404")
	fmt.Println("  Look up all 4xx codes (client errors):")
	fmt.Println("      httpstatus 4")
	fmt.Println("  Look up all 41x codes:")
	fmt.Println("      httpstatus 41")
	fmt.Println("  Search for 'not found':")
	fmt.Println("      httpstatus --search \"not found\"")
	fmt.Println("  Get status 200 in JSON format:")
	fmt.Println("      httpstatus 200 --json")
	fmt.Println("  Get all details for status 500:")
	fmt.Println("      httpstatus 500 --all")

	fmt.Println("\nPARTIAL CODE LOOKUP:")
	fmt.Println("  You can enter just the first digit (e.g., '4') or first two digits (e.g., '41')")
	fmt.Println("  to list all HTTP status codes in that set. This is separate from --search.")

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
func printText(codes []StatusCode) {
	for i, sc := range codes {
		if i > 0 {
			fmt.Println()
			fmt.Println("---")
		}
		fmt.Printf("Code: %d\nType: %s\n", sc.Code, sc.Type)
		if sc.Short != nil && sc.Long != nil {
			fmt.Printf("Short: %s\nLong: %s\n", *sc.Short, *sc.Long)
		} else if sc.Long != nil {
			fmt.Printf("Long: %s\n", *sc.Long)
		} else if sc.Short != nil {
			fmt.Printf("Short: %s\n", *sc.Short)
		}
	}
}

// printJSON outputs JSON format
func printJSON(codes []StatusCode, pretty bool) {
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
	fmt.Println(string(data))
}

// printXML outputs XML format
func printXML(codes []StatusCode, pretty bool) {
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
	fmt.Println(xml.Header + string(data))
}

// printYAML outputs YAML format
func printYAML(codes []StatusCode, pretty bool) {
	if pretty {
		// Print each code as a separate document with proper indentation
		for i, sc := range codes {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(sc)
			if err != nil {
				log.Fatalf("YAML error: %v", err)
			}
			fmt.Println(string(data))
		}
	} else {
		data, err := yaml.Marshal(codes)
		if err != nil {
			log.Fatalf("YAML error: %v", err)
		}
		fmt.Println(string(data))
	}
}
