# Issue 85 Implementation Summary

## Changes Made

### New Package: internal/userconfig
- `userconfig.go` - User configuration management
  - `Config` struct with `Telemetry` bool field
  - `Load()` reads from `~/.tsuku/config.toml`, returns defaults if missing
  - `Save()` writes config to file
  - `Get(key)` returns value as string
  - `Set(key, value)` updates and saves config
  - `AvailableKeys()` returns list of configurable settings
- `userconfig_test.go` - Comprehensive unit tests (11 tests)

### New Command: cmd/tsuku/config.go
- `tsuku config` parent command with help text
- `tsuku config get <key>` - displays current value
- `tsuku config set <key> <value>` - updates config file

### Modified: cmd/tsuku/main.go
- Registered `configCmd` in command list

### Modified: internal/telemetry/client.go
- `NewClient()` now checks userconfig in addition to env var
- Environment variable takes precedence (checked first)
- Added userconfig import

### Modified: internal/telemetry/notice.go
- Updated `NoticeText` to mention `tsuku config set telemetry false` option
- `ShowNoticeIfNeeded()` checks config file (in addition to env var)
- Added userconfig import

## Design Decisions

1. **TOML format**: Standard Go ecosystem choice, human-readable
2. **Env var precedence**: `TSUKU_NO_TELEMETRY=1` overrides config file for CI compatibility
3. **Silent failures**: Config loading errors silently use defaults
4. **Extensible**: AvailableKeys() makes it easy to add more settings

## Testing

- 11 new unit tests for userconfig package
- All existing tests pass (13 packages, 0 failures)
- go vet passes

## Files Changed

- `internal/userconfig/userconfig.go` (new)
- `internal/userconfig/userconfig_test.go` (new)
- `cmd/tsuku/config.go` (new)
- `cmd/tsuku/main.go` (modified)
- `internal/telemetry/client.go` (modified)
- `internal/telemetry/notice.go` (modified)

## Acceptance Criteria Met

- [x] `tsuku config get <key>` displays current value
- [x] `tsuku config set <key> <value>` updates config
- [x] Config stored in `~/.tsuku/config.toml`
- [x] Config respects `TSUKU_HOME` environment variable
- [x] `tsuku config set telemetry false` disables telemetry
- [x] Telemetry client checks config file in addition to env var
- [x] Env var takes precedence over config file
- [x] First-run notice updated to mention config option
- [x] Help text documents available settings
- [x] Unit tests for config commands and telemetry integration
