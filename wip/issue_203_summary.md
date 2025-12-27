# Issue 203 Summary

## What Was Implemented

Post-install binary checksum pinning (Layer 3 of the defense-in-depth verification strategy). After successful installation, tsuku computes SHA256 checksums of installed binaries and stores them in state.json. The `tsuku verify` command now includes an integrity verification step that recomputes checksums and compares against stored values to detect tampering.

## Changes Made

- `internal/install/state.go`: Added `BinaryChecksums` field to `VersionState` struct
- `internal/install/checksum.go`: New file with `ComputeBinaryChecksums()` and `VerifyBinaryChecksums()` functions
- `internal/install/checksum_test.go`: Comprehensive unit tests for checksum functions
- `internal/install/manager.go`: Integrated checksum computation after installation completes
- `cmd/tsuku/verify.go`: Added integrity verification step (Step 4 for visible tools) and updated help text
- `docs/DESIGN-checksum-pinning.md`: Design document with 6 key decisions

## Key Decisions

1. **Checksum binaries only** (not all files): Binaries are the primary attack surface; already tracked in `VersionState.Binaries`
2. **Compute after all actions complete**: Captures final installed state reliably
3. **Store in `VersionState.BinaryChecksums`**: Simple flat map, colocated with related metadata, backward compatible
4. **Graceful verification**: Old installations without checksums show "SKIPPED" rather than failing
5. **SHA256 algorithm**: Already used throughout codebase for download verification
6. **Warning on mismatch**: Report modification but don't fail (user may have intentionally modified)

## Trade-offs Accepted

- **Partial coverage**: Only binaries are checksummed, not libraries or configs. Acceptable because binaries are the most security-sensitive files.
- **State file trust**: Attacker with write access to state.json could modify both binary and checksum. This is inherent to the threat model - if attacker can write to state, they can write anywhere.

## Test Coverage

- New tests added: 12 tests in `checksum_test.go`
- Tests cover: file checksums, multiple binaries, symlinks, missing files, mismatches, backward compatibility

## Known Limitations

- Checksums only computed for binaries listed in recipe; unlisted files in tool directory are not tracked
- No automatic periodic verification (planned for `tsuku doctor` in future)
- State file tampering not detected (would require external key storage)

## Future Improvements

- Integration with `tsuku doctor` for scheduled integrity checks
- Optional `--strict` flag for verify to fail on any integrity issue
- Extended checksums for library dependencies (if needed)
