# Issue 15 Baseline

## Issue Summary
- **Number**: 15
- **Title**: feat(cli): add --dry-run flag for install and update commands
- **Milestone**: v0.2.0

## Problem Statement
Users cannot preview what actions will be taken before running install or update. This makes it difficult to verify the correct version, review dependencies, or understand system changes.

## Expected Behavior
A `--dry-run` flag that shows planned actions without executing them:
- Show what would be installed (tool and version)
- List actions that would be taken (download, extract, install_binaries)
- No filesystem changes when flag is set

## Branch
- Feature branch: `feature/15-dry-run-flag`
- Base: `main` at commit f7f6ab3

## Acceptance Criteria
1. --dry-run flag available on install command
2. --dry-run flag available on update command
3. Shows planned version and actions
4. No filesystem changes when --dry-run is set
