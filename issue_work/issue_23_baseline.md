# Issue 23 Baseline

## Issue Summary

**Title**: feat: add timeout handling for API requests
**Number**: #23
**Milestone**: v0.2.0

## Problem

GitHub API calls and version resolution requests can hang indefinitely if:
- Network issues occur
- API is slow to respond
- DNS resolution fails

The CLI appears frozen with no feedback.

## Expected Behavior

- Default timeout of 30 seconds for API requests
- Configurable via environment variable (e.g., TSUKU_API_TIMEOUT)
- Clear error message when timeout occurs
- Suggestion to check network or try again

## Baseline State

- **Branch**: feature/23-timeout-handling
- **Based on**: main (335a03f)
- **Tests**: All pass
- **Lint**: Clean

## Notes

This issue complements #22 (graceful cancellation) which added context propagation. The timeout handling can leverage the existing context infrastructure to implement request timeouts.
