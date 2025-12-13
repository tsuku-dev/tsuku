# Issue 436 Baseline

## Environment
- Date: 2025-12-13
- Branch: feature/436-decomposable-interface
- Base commit: 1e3168d209beafead9e431be85e0c0a488d1ed55

## Test Results
- Total: 18 packages tested
- Passed: 17 packages
- Failed: 1 package (github.com/tsukumogami/tsuku/internal/builders)

### Pre-existing Failure
```
TestLLMGroundTruth/L16_minikube - Action mismatch: got github_archive, want github_file
```
This is an LLM integration test failure unrelated to issue 436.

## Build Status
- Build: PASS
- Vet: PASS

## Pre-existing Issues
- LLM ground truth test for minikube expects `github_file` action but model generates `github_archive`
- This is unrelated to the decomposable interface work
