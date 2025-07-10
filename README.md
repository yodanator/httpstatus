# httpstatus

A CLI tool for looking up HTTP status codes in multiple formats.

---

## Overview

**httpstatus** is a command-line utility that allows you to quickly look up HTTP status codes, their types, and their descriptions. It supports searching by code, partial code, or keyword and can output results in plain text, JSON, XML, or YAML formats (including pretty-printed versions).

---

## Usage

```sh
httpstatus [flags] [status_code|partial_code]
```

### Examples

- **Look up status code 404:**
  ```
  httpstatus -c 404
  ```
- **Look up all 4xx codes (client errors):**
  ```
  httpstatus 4
  ```
- **Look up all 41x codes:**
  ```
  httpstatus 41
  ```
- **Search for codes with "not found":**
  ```
  httpstatus --search "not found"
  ```
- **Get status 200 in JSON format:**
  ```
  httpstatus 200 --json
  ```
- **See help for all flags and output options:**
  ```
  httpstatus --help
  ```
  
### Partial Code Lookup

You can enter just the first digit (e.g., `4`) or first two digits (e.g., `41`) to list all HTTP status codes in that set.  
This is separate from `--search` and is useful for quickly listing all codes in a category.

---

## Contributing & Feedback

- **Found a missing HTTP status code?**  
  Please [open an issue](https://github.com/yodanator/httpstatus/issues) or submit an enhancement request!
- **Have a suggestion or improvement?**  
  Pull requests and feature requests are welcome.
- **Bug reports** are appreciated—please use the [issues page](https://github.com/yodanator/httpstatus/issues).

---
