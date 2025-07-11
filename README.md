# httpstatus - HTTP Status Code Lookup Tool

[![Release Version](https://img.shields.io/github/v/release/yodanator/httpstatus)](https://github.com/yodanator/httpstatus/releases)
[![Build Status - main](https://github.com/yodanator/httpstatus/actions/workflows/release.yml/badge.svg)](https://github.com/yodanator/httpstatus/actions/workflows/release.yml)
[![Dev CI Build - dev](https://github.com/yodanator/httpstatus/actions/workflows/dev-ci.yml/badge.svg?branch=dev)](https://github.com/yodanator/httpstatus/actions/workflows/dev-ci.yml)
[![Forks](https://img.shields.io/github/forks/yodanator/httpstatus?style=social)](https://github.com/yodanator/httpstatus/network/members)
[![Stars](https://img.shields.io/github/stars/yodanator/httpstatus?style=social)](https://github.com/yodanator/httpstatus/stargazers)
[![Pull Requests](https://img.shields.io/github/issues-pr/yodanator/httpstatus)](https://github.com/yodanator/httpstatus/pulls)
[![Issues](https://img.shields.io/github/issues/yodanator/httpstatus)](https://github.com/yodanator/httpstatus/issues)
[![License](https://img.shields.io/github/license/yodanator/httpstatus)](https://github.com/yodanator/httpstatus/blob/main/LICENSE)
[![Codecov](https://codecov.io/gh/yodanator/httpstatus/branch/main/graph/badge.svg)](https://codecov.io/gh/yodanator/httpstatus)
[![Coveralls](https://coveralls.io/repos/github/yodanator/httpstatus/badge.svg?branch=main)](https://coveralls.io/github/yodanator/httpstatus?branch=main)
![GitHub Downloads (all assets, latest release)](https://img.shields.io/github/downloads/yodanator/httpstatus/latest/total)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/yodanator/httpstatus/total)




A CLI tool for looking up HTTP status codes with multiple output
formats. Instantly get detailed information about any HTTP status code
directly from your terminal.

------------------------------------------------------------------------

## Features

- Look up status codes by number (e.g., `404`) or category (e.g., `4`)
- Search descriptions with keywords (e.g., `not found`)
- Multiple output formats: JSON, XML, YAML, TOML, CSV, Markdown
- Save output to files
- Cross-platform support (Windows, Linux, macOS, FreeBSD)
- Multiple installation options

------------------------------------------------------------------------

## Installation

### Using Install Scripts (Recommended)

1.  Download the appropriate install script from the [latest
    release](https://github.com/yodanator/httpstatus/releases/latest):
    - `install.sh` for Linux/macOS
    - `install.bat` for Windows Command Prompt
    - `install.ps1` for Windows PowerShell
2.  Run the script with your preferred installation method:

#### Linux/macOS

    # User installation (default: ~/bin)
    chmod +x install.sh
    ./install.sh --user

    # System-wide installation (requires sudo)
    sudo ./install.sh --system-wide

#### Windows (Command Prompt)

    :: User installation (default: %USERPROFILE%\bin)
    install.bat --user

    :: System-wide installation (requires Admin)
    install.bat --system

#### Windows (PowerShell)

    # User installation
    .\install.ps1 -User

    # System-wide installation (requires Admin)
    .\install.ps1 -System

------------------------------------------------------------------------

### Using Go Install

#### User-level installation

    go install github.com/yodanator/httpstatus@latest

#### System-wide installation

    # Unix
    sudo GOBIN=/usr/local/bin go install github.com/yodanator/httpstatus@latest

    # Windows (Admin PowerShell)
    $env:GOBIN = "C:\Program Files\httpstatus"
    go install github.com/yodanator/httpstatus@latest

------------------------------------------------------------------------

### Manual Installation

Download the appropriate binary from the latest release and add it to
your `PATH`.

------------------------------------------------------------------------

### Supported Platforms

| OS      | Architectures                       | Filename Pattern                       |
|---------|-------------------------------------|----------------------------------------|
| Linux   | x86-32, x86-64, ARMv6, ARMv7, ARM64 | `httpstatus-linux-<arch>-v<version>`   |
| macOS   | Intel, Apple Silicon                | `httpstatus-darwin-<arch>-v<version>`  |
| Windows | x86-32, x86-64, ARM64               | `httpstatus-windows-<arch>-v<version>` |
| FreeBSD | x86-64                              | `httpstatus-freebsd-amd64-v<version>`  |

------------------------------------------------------------------------

## Usage

    httpstatus [flags] [status_code|partial_code]
    httpstatus --search "search term"
    httpstatus --code "200,404"
    httpstatus "4,5" --json-pretty
    httpstatus --to-file output --json --csv
    httpstatus --table

------------------------------------------------------------------------

## Examples

**Look up multiple status codes:**

    httpstatus -c "200,404"

**Look up all 4xx and 5xx codes:**

    httpstatus "4,5"

**Search for 'not found' and show 404:**

    httpstatus --search "not found" --code 404

**Get status 200 and 201 in JSON format:**

    httpstatus 200,201 --json

**Export all 2xx codes to CSV:**

    httpstatus 2 --csv --to-file success_codes

------------------------------------------------------------------------

## Flags

    -c, --code <codes>     HTTP status code(s) to look up (comma-separated)
    -s, --search <term>    Search status codes by keyword
    -l, --long             Show long description only
    -a, --all              Show both short and long descriptions
        --json             Output as JSON
        --json-pretty      Output as formatted JSON
        --xml              Output as XML
        --xml-pretty       Output as formatted XML
        --yaml             Output as YAML
        --yaml-pretty      Output as formatted YAML
        --toml             Output as TOML
        --table            Output as text table
        --markdown         Output as Markdown table
        --csv              Output as CSV
        --to-file <base>   Save output to files (automatic extensions)
        --help             Show help message
        --version          Show version information

------------------------------------------------------------------------

## Contributing

1.  Fork the repository
2.  Clone your fork: `git clone https://github.com/your-account-name/httpstatus.git`
3.  Create a new branch: `git checkout -b feature/your-feature`
4.  Commit your changes: `git commit -am 'Add some feature'`
5.  Push to the branch: `git push origin feature/your-feature`
6.  Open a pull request. The PR will run automated tests and require approval before merging.

------------------------------------------------------------------------

## Reporting Issues

Open an issue on the [GitHub Issues
page](https://github.com/yodanator/httpstatus/issues) with:

- A detailed description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, version, etc.)

------------------------------------------------------------------------

## License

This project is licensed under the [GNU General Public License
v3.0](https://www.gnu.org/licenses/gpl-3.0).

------------------------------------------------------------------------

## Support

For questions or support, visit the [GitHub
Discussions](https://github.com/yodanator/httpstatus/discussions) page.
