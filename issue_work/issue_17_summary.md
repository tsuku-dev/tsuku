# Issue #17 Implementation Summary

## Overview
Added `--json` flag to 5 commands for machine-readable output suitable for parsing with tools like jq.

## Commands Updated

### list
```bash
$ tsuku list --json
{
  "tools": [
    {"name": "nodejs", "version": "22.11.0", "path": "/home/user/.tsuku/tools/nodejs-22.11.0"}
  ]
}
```

### info
```bash
$ tsuku info nodejs --json
{
  "name": "nodejs",
  "description": "Node.js JavaScript runtime",
  "homepage": "https://nodejs.org",
  "version_format": "semver",
  "status": "installed",
  "installed_version": "22.11.0",
  "location": "/home/user/.tsuku/tools/nodejs-22.11.0",
  "verify_command": "node --version"
}
```

### versions
```bash
$ tsuku versions nodejs --json
{
  "versions": ["22.11.0", "22.10.0", "21.7.3"]
}
```

### outdated
```bash
$ tsuku outdated --json
{
  "updates": [
    {"name": "nodejs", "current": "22.10.0", "latest": "22.11.0"}
  ]
}
```

### search
```bash
$ tsuku search node --json
{
  "results": [
    {"name": "nodejs", "description": "Node.js JavaScript runtime", "installed": "22.11.0"}
  ]
}
```

## Files Changed
- `cmd/tsuku/helpers.go` - Added `printJSON` helper function
- `cmd/tsuku/list.go` - Added --json flag and JSON output
- `cmd/tsuku/info.go` - Added --json flag and JSON output
- `cmd/tsuku/versions.go` - Added --json flag and JSON output
- `cmd/tsuku/outdated.go` - Added --json flag and JSON output
- `cmd/tsuku/search.go` - Added --json flag and JSON output

## Design Decisions
1. Each command has its own local `--json` flag (not global)
2. JSON output uses pretty-printing with 2-space indentation
3. Empty arrays are initialized to `[]` instead of `null`
4. `omitempty` used for optional fields like `installed_version`
5. JSON output suppresses informational messages (similar to --quiet)
