# Issue 353 Baseline

## Environment
- Date: 2025-12-09
- Branch: chore/353-consolidate-deployments
- Base commit: ec8aa2456528c7ca37bf4707afe9ab1c28b97459

## Build Status
Pass - `go build ./cmd/tsuku` succeeds without errors

## Pre-existing State
- Website deploys via Cloudflare Pages (`deploy-website.yml`)
- Registry deploys via GitHub Pages (`deploy-recipes.yml`)
- Website fetches `recipes.json` from `registry.tsuku.dev` (cross-origin)

## Notes
- This is a CI/deployment consolidation task, not Go code changes
- Go tests are not directly relevant to this work
