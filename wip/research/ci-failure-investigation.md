# Build Essentials CI Workflow Failure Investigation

**Date:** 2025-12-23
**Investigator:** CI Analysis Agent
**Scope:** Investigate build-essentials.yml workflow failures on main branch

## Executive Summary

The Build Essentials CI workflow is failing on main branch, but the failure is unusual: GitHub Actions reports the workflow run as "failed" with **zero check runs executed**. This suggests a workflow configuration issue rather than test failures.

## Affected Runs

| Run ID | Commit | Date | Status | Conclusion |
|--------|--------|------|--------|------------|
| 20471202216 | dbd3eb3 | 2025-12-23 20:45:29Z | completed | failure |
| 20470884897 | 45d8d4c | 2025-12-23 20:27:46Z | completed | failure |

**Last successful run:** 20449928292 (b93a74c, 2025-12-23 02:51:30Z)

## Analysis

### 1. Workflow Run Characteristics

Both failing runs share identical symptoms:
- Status: `completed`
- Conclusion: `failure`
- Jobs array: **empty** (no jobs were created or executed)
- Check runs: **empty** (no check runs exist)
- Error message from `gh run view`: "This run likely failed because of a workflow file issue"

### 2. Changes Between Success and Failure

**Commit dbd3eb3** (failing):
```
feat(recipes): add git recipe with curl dependency validation (#662)
```

**Commit 45d8d4c** (failing):
```
feat(recipes): add readline and sqlite recipes with dependency chain validation (#661)
```

**Commit b93a74c** (last success):
```
feat(recipes): add cmake and ninja recipes with cmake_build validation
```

### 3. Workflow File Changes

The failing commits added two new jobs to `.github/workflows/build-essentials.yml`:

1. **test-sqlite-source** (lines 187-222) - Added in commit 45d8d4c
2. **test-git-source** (lines 224-259) - Added in commit dbd3eb3

Both jobs follow the same pattern as existing jobs:
- Use matrix strategy with platform variations
- Have standard workflow steps (checkout, setup-go, build, install, verify)
- Reference test recipes that exist in `testdata/recipes/`

### 4. Workflow YAML Validation

Manual inspection of the workflow file shows:
- ✅ YAML structure appears valid
- ✅ Job names are unique
- ✅ Matrix syntax is consistent with other jobs
- ✅ Step names and commands are properly formatted
- ✅ All referenced files exist:
  - `testdata/recipes/sqlite-source.toml` (exists, 1212 bytes)
  - `testdata/recipes/git-source.toml` (exists, 1245 bytes)

### 5. GitHub Actions Check Suite

Check suite `52932853835` for commit dbd3eb3:
- Status: `completed`
- Conclusion: `failure`
- Check runs: **empty array** `[]`

This is the smoking gun: GitHub Actions created a check suite but **did not create any check runs**, which means the workflow file was rejected before job execution.

### 6. Other Check Suites on Same Commit

Interestingly, the same commit (dbd3eb3) has **other successful check suites**:
- 52932854488: success
- 52932854493: success
- 52932854505: success
- 52932854529: success
- 52944670575: success
- 52944948643: success

This indicates that:
1. Other workflows on the same commit are passing
2. Only the build-essentials workflow is affected
3. The issue is specific to this workflow file, not a repo-wide problem

## Root Cause Hypothesis

GitHub Actions is rejecting the workflow file **before creating any jobs**, but the reason is not visible through the API. Possible causes:

1. **Workflow file size limit** - Adding two large jobs may have exceeded a limit
2. **Matrix expansion limit** - Too many matrix combinations across all jobs
3. **Syntax issue not caught by YAML parser** - GitHub's workflow validator may have stricter rules
4. **Resource limit** - Total job count or step count may exceed GitHub Actions limits
5. **Runner availability** - Referenced runners may not be available (unlikely given `macos-15-intel` usage)

## Recommended Next Steps

### Immediate Actions

1. **Check GitHub Actions UI** - The web interface may show error details not available via API
2. **Validate workflow locally** - Use `actionlint` or GitHub's workflow validator
3. **Count total jobs** - Verify total job count doesn't exceed GitHub limits
4. **Test workflow incrementally** - Try adding one new job at a time to identify the threshold

### Potential Fixes

1. **Split workflow** - Move some jobs to a separate workflow file
2. **Reduce matrix combinations** - Combine or remove some platform/tool combinations
3. **Check runner labels** - Verify `macos-15-intel` is a valid runner label
4. **Simplify job structure** - Break complex jobs into smaller pieces

## Impact Assessment

### Does This Block M19/M20 Completion?

**YES - This is a blocker** for the following reasons:

1. **M19 Deliverable:** "All source build recipes validated in CI"
   - The test-sqlite-source and test-git-source jobs are critical validation
   - These tests prove the dependency provisioning system works end-to-end
   - Cannot claim M19 complete if CI isn't validating these recipes

2. **M20 Deliverable:** "Complete Homebrew Builder implementation"
   - Git recipe validates the most complex dependency chain
   - This is the capstone validation of the entire system
   - Failing CI means the validation isn't proven

### Severity

**HIGH** - This is not a code bug but a CI infrastructure issue that:
- Prevents validation of new features
- Blocks merging of completed work
- Affects main branch stability
- May indicate a systemic problem with workflow complexity

## Files Examined

- `/home/dangazineu/dev/workspace/tsuku/tsuku-1/public/tsuku/.github/workflows/build-essentials.yml`
- `testdata/recipes/sqlite-source.toml`
- `testdata/recipes/git-source.toml`

## API Calls Made

- `gh run list --workflow=build-essentials.yml --branch=main`
- `gh run view 20471202216`
- `gh api repos/tsukumogami/tsuku/actions/runs/20471202216`
- `gh api repos/tsukumogami/tsuku/actions/runs/20471202216/jobs`
- `gh api repos/tsukumogami/tsuku/commits/dbd3eb3/check-suites`
- `gh api repos/tsukumogami/tsuku/check-suites/52932853835/check-runs`
- `git diff b93a74c..dbd3eb3 -- .github/workflows/build-essentials.yml`
- `git show --stat dbd3eb3`
- `git show --stat 45d8d4c`

## Conclusion

The Build Essentials workflow is failing due to GitHub Actions rejecting the workflow file **before executing any jobs**. The exact error is not accessible via API, and manual YAML validation shows no obvious syntax errors.

The issue appeared when adding test-sqlite-source and test-git-source jobs, suggesting a resource limit or complexity threshold has been exceeded. This is a **blocker for M19/M20 completion** as it prevents validation of the dependency provisioning system.

**Recommended immediate action:** Access the GitHub Actions web UI directly to view the detailed error message that's not available through the API.
