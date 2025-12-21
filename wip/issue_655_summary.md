# Issue 655 Summary

## What Was Implemented

Updated the curl recipe to use dynamic versioning with template variables instead of hardcoded version numbers. The recipe now uses the `download` action with `{version}` placeholders, allowing it to work with any version resolved by the Homebrew formula without requiring recipe file edits.

## Changes Made

- `internal/recipe/recipes/c/curl.toml`:
  - Changed `download_file` action to `download` action (line 14)
  - Updated download URL from `curl-8.11.1.tar.gz` to `curl-{version}.tar.gz` (line 15)
  - Removed static checksum parameter (was line 15)
  - Added explanatory comment about checksum computation (line 12)
  - Updated extract archive parameter from `curl-8.11.1.tar.gz` to `curl-{version}.tar.gz` (line 19)
  - Updated configure_make source_dir from `curl-8.11.1` to `curl-{version}` (line 27)

## Key Decisions

- **Use download action without checksum_url**: curl.se provides PGP signatures but not SHA256SUMS files. Following the pattern used by ruby and rust recipes, checksums are computed during plan generation by downloading the file.
- **Add explanatory comment**: Added a comment explaining why there's no checksum_url parameter, helping future maintainers understand the intentional difference from recipes like terraform/boundary/vault.
- **Keep all other steps unchanged**: Only the download, extract, and configure_make steps needed updates. All dependency handling, RPATH configuration, and verification remained untouched.

## Trade-offs Accepted

- **Checksum computation requires download during plan generation**: Unlike recipes with checksum_url (terraform, boundary), curl requires downloading the tarball during plan generation to compute the checksum. This is acceptable because the download is cached and reused for installation, following the established pattern for sources without checksum files.

## Test Coverage

- No new tests needed - existing recipe validation tests cover template expansion
- All 23 test packages pass successfully
- Recipe validates with `--strict` mode
- Existing CI will test curl recipe on Linux x86_64, macOS Intel, and macOS Apple Silicon

## Known Limitations

- RPATH dependency versioning still hardcoded (openssl-3.6.0, zlib-1.3.1) - tracked in issue #653, intentionally out of scope for this fix

## Future Improvements

None - this change brings curl into full alignment with modern recipe patterns.
