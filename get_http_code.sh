#!/bin/bash

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
DB_FILE="${SCRIPT_DIR}/http_status_codes.json"

# Initialize variables
code=""
long=0
all=0
output_formats=()

# Parse options using getopts
while getopts ":c:laj-:x-:y-:" opt; do
  case "${opt}" in
    c)
      code="${OPTARG}"
      ;;
    l)
      long=1
      ;;
    a)
      all=1
      ;;
    j)
      output_formats+=("json")
      ;;
    -)
      case "${OPTARG}" in
        code)
          val="${!OPTIND}"; OPTIND=$((OPTIND + 1))
          code="${val}"
          ;;
        long)
          long=1
          ;;
        all)
          all=1
          ;;
        json)
          output_formats+=("json")
          ;;
        json-pretty)
          output_formats+=("json-pretty")
          ;;
        xml)
          output_formats+=("xml")
          ;;
        xml-pretty)
          output_formats+=("xml-pretty")
          ;;
        yaml)
          output_formats+=("yaml")
          ;;
        yaml-pretty)
          output_formats+=("yaml-pretty")
          ;;
        *)
          echo "Invalid option: --${OPTARG}" >&2
          exit 1
          ;;
      esac
      ;;
    x)
      output_formats+=("xml")
      ;;
    y)
      output_formats+=("yaml")
      ;;
    \?)
      echo "Invalid option: -${OPTARG}" >&2
      exit 1
      ;;
    :)
      echo "Option -${OPTARG} requires an argument." >&2
      exit 1
      ;;
  esac
done
shift $((OPTIND -1))

# Validate required parameters
if [[ -z "$code" ]]; then
  echo "Error: HTTP code must be specified with -c or --code" >&2
  exit 1
fi

# Check for conflicting options
if [[ $long -eq 1 && $all -eq 1 ]]; then
  echo "Note: --long ignored because --all was specified" >&2
  long=0
fi

# Find matching status code
result=$(jq -r ".[] | select(.code == $code)" "$DB_FILE")

if [[ -z "$result" ]]; then
  echo "Error: HTTP status code $code not found" >&2
  exit 1
fi

# Determine code type
first_digit=${code:0:1}
case $first_digit in
  1) type="Informational" ;;
  2) type="Success" ;;
  3) type="Redirection" ;;
  4) type="Client Error" ;;
  5) type="Server Error" ;;
  *) type="Unknown" ;;
esac

# Create temp file for XML processing
TMP_XML=$(mktemp)

# Prepare output based on flags
if [[ $all -eq 1 ]]; then
  json_output=$(jq -c --arg type "$type" '{code, type: $type, short, long}' <<< "$result")
  text_output="Code: $code\nType: $type\nShort: $(jq -r '.short' <<< "$result")\nLong: $(jq -r '.long' <<< "$result")"
  
  # Create proper YAML as an object
  yaml_output=$(jq -r --arg type "$type" '
    "code: \(.code)\n" +
    "type: \($type)\n" +
    "short: \(.short)\n" +
    "long: \(.long)"
  ' <<< "$result")
  
  # Create XML with proper structure
  jq -r --arg type "$type" '. | "<http_status><code>\(.code)</code><type>\($type)</type><short>\(.short)</short><long>\(.long)</long></http_status>"' <<< "$result" > "$TMP_XML"
elif [[ $long -eq 1 ]]; then
  json_output=$(jq -c --arg type "$type" '{code, type: $type, long}' <<< "$result")
  text_output="Code: $code\nType: $type\nLong: $(jq -r '.long' <<< "$result")"
  
  # Create proper YAML as an object
  yaml_output=$(jq -r --arg type "$type" '
    "code: \(.code)\n" +
    "type: \($type)\n" +
    "long: \(.long)"
  ' <<< "$result")
  
  # Create XML with proper structure
  jq -r --arg type "$type" '. | "<http_status><code>\(.code)</code><type>\($type)</type><long>\(.long)</long></http_status>"' <<< "$result" > "$TMP_XML"
else
  json_output=$(jq -c --arg type "$type" '{code, type: $type, short}' <<< "$result")
  text_output="Code: $code\nType: $type\nShort: $(jq -r '.short' <<< "$result")"
  
  # Create proper YAML as an object
  yaml_output=$(jq -r --arg type "$type" '
    "code: \(.code)\n" +
    "type: \($type)\n" +
    "short: \(.short)"
  ' <<< "$result")
  
  # Create XML with proper structure
  jq -r --arg type "$type" '. | "<http_status><code>\(.code)</code><type>\($type)</type><short>\(.short)</short></http_status>"' <<< "$result" > "$TMP_XML"
fi

# Function to output proper YAML objects
output_yaml() {
  local format="$1"
  case "$format" in
    yaml)
      echo -e "$yaml_output"
      ;;
    yaml-pretty)
      # Format as a YAML object with proper indentation
      echo -e "$yaml_output" | 
      awk '{print "  " $0}'
      ;;
  esac
}

# Output handling
if [[ ${#output_formats[@]} -eq 0 ]]; then
  echo -e "$text_output"
else
  for format in "${output_formats[@]}"; do
    case "$format" in
      json)
        echo "$json_output"
        ;;
      json-pretty)
        echo "$json_output" | jq .
        ;;
      xml)
        cat "$TMP_XML"
        ;;
      xml-pretty)
        xmllint --format "$TMP_XML"
        ;;
      yaml | yaml-pretty)
        output_yaml "$format"
        ;;
    esac
  done
fi

# Cleanup
rm -f "$TMP_XML"
