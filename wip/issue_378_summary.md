# Issue 378 Summary

## What Was Implemented

Enforced hourly rate limit (default 10/hour) for LLM recipe generation in `tsuku create` when using GitHub sources. The implementation integrates with existing state tracking and user config infrastructure.

## Changes Made

- `cmd/tsuku/create.go`: Added rate limit check before LLM generation, recording after success, and helper function for formatting wait time

## Key Decisions

- Check rate limit only for GitHub builder: Ecosystem builders (crates.io, npm, etc.) don't use LLM
- Create StateManager regardless of limit: Needed for recording even when limit is disabled (0)
- Pass cost=0 for now: Actual cost tracking to be added in a future issue
- Error message includes actionable suggestions: Shows how to increase limit via config

## Trade-offs Accepted

- Cost is recorded as 0: BuildResult doesn't have cost field yet; can be enhanced later
- Rate limit check happens before builder creation: Could theoretically check earlier, but this is the logical point before any LLM work

## Test Coverage

- New tests added: 0 (rate limiting logic already tested in `internal/install/state_test.go`)
- Coverage change: N/A (integration code using existing tested functions)

## Known Limitations

- Cost tracking is a placeholder (0) until BuildResult includes cost information
- Wait time calculation assumes the oldest timestamp expiring first (correct for FIFO rate limiting)
