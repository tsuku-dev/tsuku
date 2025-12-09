# Issue 308 Implementation Plan

## Summary
Implement the container executor that orchestrates container-based recipe validation by combining runtime detection, asset pre-download, and isolated container execution.

## Approach
Create an `Executor` struct in `internal/validate/executor.go` that:
1. Uses `RuntimeDetector` to find available container runtime
2. Uses `PreDownloader` to download assets with checksums
3. Runs recipe steps in an isolated container
4. Runs verification command and checks output
5. Returns `ValidationResult` with pass/fail status and output
6. Emits docker group security warning when applicable

### Alternatives Considered
- **Single monolithic validation function**: Rejected because it doesn't reuse existing components (RuntimeDetector, PreDownloader, Runtime.Run)
- **Integrate directly into LLM generation flow**: Rejected because executor should be testable in isolation

## Files to Modify
None - all new code

## Files to Create
- `internal/validate/executor.go` - Executor implementation
- `internal/validate/executor_test.go` - Unit tests

## Implementation Steps
- [ ] Create `Executor` struct with dependencies (RuntimeDetector, PreDownloader)
- [ ] Create `ValidationResult` type (pass/fail, stdout, stderr, exit code)
- [ ] Create `ExecutorLogger` interface for warnings
- [ ] Implement `Validate()` method that orchestrates: detect runtime, download assets, run container, check verification
- [ ] Add docker group security warning when `runtime.Name() == "docker" && !runtime.IsRootless()`
- [ ] Handle case when no runtime is available (skip with warning)
- [ ] Write unit tests with mocked components
- [ ] Run linting and tests

## Testing Strategy
- Unit tests: Mock RuntimeDetector, PreDownloader, and Runtime
- Test cases:
  - No runtime available (skip with warning)
  - Docker group runtime (emit warning)
  - Successful validation (verification passes)
  - Failed validation (verification fails)
  - Container execution error
  - Download error

## Risks and Mitigations
- **Risk**: Recipe execution is complex
  - **Mitigation**: Start with simple command execution, iterate for full recipe support
- **Risk**: Container image may not have required tools
  - **Mitigation**: Use alpine:latest as base, document requirements

## Success Criteria
- [ ] `Executor.Validate()` runs recipe steps in container
- [ ] Pre-downloaded assets mounted read-only at `/assets`
- [ ] Workspace mounted for installation
- [ ] Runs verification command and checks output
- [ ] Returns `ValidationResult` (pass/fail + output)
- [ ] Skips with warning when no runtime available
- [ ] Warns when using Docker with group membership (non-rootless)

## Open Questions
None - design is clear from the design document
