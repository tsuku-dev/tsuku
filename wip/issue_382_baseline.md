# Issue 382 Baseline

## Environment
- Date: 2025-12-10T14:30:00Z
- Branch: feature/382-benchmark-harness
- Base commit: d6e00ef1626db3a4b08516ef0213827b36b34776

## Test Results
- Passed: All unit tests pass
- Failed: TestLLMGroundTruth (pre-existing) - requires GITHUB_TOKEN and real API calls

## Build Status
- go build: pass
- go vet: pass

## Dependencies Resolved
- #378 (rate limiting): merged
- #379 (budget enforcement): merged
- #381 (cost display): merged

## Task Overview
Create a benchmark harness (cmd/benchmark) to measure LLM recipe generation success rate:
- Separate from main CLI
- Reads corpus file of GitHub repos
- Runs generation + validation per repo
- Reports success rate with details
- Supports --limit flag for budget control
