# Issue 524 Implementation Plan

## Summary

Add comprehensive test coverage for Homebrew source build functionality and related parsing functions.

## Analysis

After reviewing the codebase, the existing tests cover:
- Build system step generation (autotools, cmake, cargo, go, make)
- Platform conditional steps
- Source recipe data validation
- Basic parseFromFlag tests

Missing test coverage per the issue requirements:
1. parseFromFlag edge cases (trailing colons, empty values, multiple colons)
2. Builder robustness tests (LLM returning wrong types, malformed JSON)
3. Multiple executables handling in build steps
4. Configure/CMake argument edge cases

## Implementation Steps

### Priority 1 - parseFromFlag Edge Cases

- [x] 1. Add parseFromFlag edge case tests in cmd/tsuku/create_test.go
  - PFF-1: Trailing colon (`github:`)
  - PFF-2: Colon only (`:`)
  - PFF-3: Double colon (`github::cli`)
  - PFF-4: No colon - already covered
  - PFF-5: Multiple colons (`homebrew:pg@15:source`)
  - PFF-6: Empty string handling

### Priority 2 - Builder Robustness Tests

- [x] 2. Add tool call validation tests in internal/builders/homebrew_test.go
  - HBR-1: Tool call with wrong type for argument (number instead of string)
  - HBR-2: Empty executables array validation (already exists)
  - HBR-3: Executable with empty string (already exists)
  - HBR-8: Formula name validation (special chars) (already exists)

- [x] 3. Add source recipe output tests for multiple executables
  - EX-2: Multiple executables
  - EX-3: Versioned executable names

- [x] 4. Add configure/cmake argument edge cases
  - Test valid edge cases (paths with spaces, equals signs)
  - Test rejection of dangerous patterns (already exists)

### Priority 3 - Platform/Build System Edge Cases

- [x] 5. Add multiple platform conditional tests
  - PS-5: Combined os and arch conditionals
  - Test empty platform steps handling (already exists)

- [x] 6. Add version specifier handling tests
  - VS-1: Formula with @ version (`postgresql@15`)

## Files Modified

- `cmd/tsuku/create_test.go` - Added parseFromFlag edge cases (7 new tests)
- `internal/builders/homebrew_test.go` - Added builder robustness and coverage tests:
  - TestHomebrewBuilder_buildSourceSteps_MultipleExecutables
  - TestValidateSourceRecipeData_VersionedExecutables
  - TestValidateSourceRecipeData_ConfigureArgsValidCases
  - TestValidateSourceRecipeData_CMakeArgsValidCases
  - TestValidateSourceRecipeData_CombinedPlatformConditionals
  - TestHomebrewBuilder_CanBuild_VersionedFormula
  - TestHomebrewBuilder_executeToolCall_TypeValidation
  - TestHomebrewBuilder_executeToolCall_SourceTypeValidation

## Testing Strategy

Each test group was added and verified locally:
1. Run specific tests: `go test -v -run TestName ./path/...`
2. Run full package tests: `go test ./path/...`
3. Verify no regressions: `go test ./...`

## Success Criteria

- [x] All parseFromFlag edge cases tested
- [x] Builder robustness tests pass
- [x] Multiple executables handling tested
- [x] All tests pass locally
- [ ] CI passes
