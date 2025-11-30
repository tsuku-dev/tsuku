# Issue 16 Summary

## Overview

Added a global `--quiet` (`-q`) flag to suppress informational output across all commands while preserving error messages on stderr.

## Implementation

### Core Changes

1. **main.go**: Added `quietFlag` global variable and registered persistent flag on root command
2. **helpers.go**: Added `printInfo()` and `printInfof()` helper functions that respect quiet mode

### Commands Updated

All commands with informational output were updated to use quiet-aware print functions:

- **install.go**: Progress messages, warnings, success confirmations
- **update.go**: Progress messages
- **remove.go**: Success confirmations, warnings, orphan cleanup messages
- **list.go**: Header text, empty list message, system dependency footnote
- **recipes.go**: Header text, empty list message, help text
- **outdated.go**: Progress messages, summary messages
- **create.go**: Progress messages, success confirmations
- **search.go**: Empty results help text
- **update_registry.go**: Progress and success messages
- **verify.go**: Step progress, verification output

### Commands NOT Modified

Output from these commands is the command's primary purpose and should always be shown:

- **config.go**: `config get` output IS the result
- **info.go**: Tool info IS the result
- **versions.go**: Version list IS the result

## Testing

- All existing tests pass
- Manual verification:
  - `./tsuku --help` shows `-q, --quiet` flag
  - `./tsuku list --help` shows inherited `--quiet` flag
  - Flag suppresses informational output as expected

## Files Changed

- `cmd/tsuku/main.go`
- `cmd/tsuku/helpers.go`
- `cmd/tsuku/install.go`
- `cmd/tsuku/update.go`
- `cmd/tsuku/remove.go`
- `cmd/tsuku/list.go`
- `cmd/tsuku/recipes.go`
- `cmd/tsuku/outdated.go`
- `cmd/tsuku/create.go`
- `cmd/tsuku/search.go`
- `cmd/tsuku/update_registry.go`
- `cmd/tsuku/verify.go`
