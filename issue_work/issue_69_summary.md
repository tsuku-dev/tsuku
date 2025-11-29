# Issue 69 Summary

## What Was Implemented
Updated README installation instructions to use the new get.tsuku.dev URL and removed the obsolete install.sh file that is now managed in the tsuku.dev repository.

## Changes Made
- `README.md`: Replaced manual go build instructions with `curl -fsSL https://get.tsuku.dev/now | bash`
- `install.sh`: Deleted (now managed in tsuku-dev/tsuku.dev)

## Key Decisions
- Used the exact URL format specified in the issue (`https://get.tsuku.dev/now`)

## Test Coverage
- No code changes, documentation only
- All existing tests continue to pass

## Known Limitations
- None
