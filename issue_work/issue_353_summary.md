# Issue 353 Summary

## What Was Implemented

Consolidated the website and recipe registry deployments into a single Cloudflare Pages deployment. The website now generates and serves `recipes.json` from the same origin, eliminating cross-origin requests and enabling full preview deployment testing.

## Changes Made

- `.github/workflows/deploy-website.yml`: Added Python setup and recipe generation steps, expanded path triggers to include recipe TOML files
- `website/recipes/index.html`: Changed `API_URL` from `https://registry.tsuku.dev/recipes.json` to `/recipes.json`
- `.github/workflows/deploy-recipes.yml`: Added deprecation notice explaining the transition
- `website/.gitignore`: Added to ignore generated `recipes.json`

## Key Decisions

- **Keep deprecated workflow**: The GitHub Pages deployment is preserved during the transition period for backwards compatibility
- **Same-origin fetch**: Using relative path `/recipes.json` eliminates CORS requirements
- **Generated at deploy time**: `recipes.json` is not committed to the repo but generated during each deployment

## Trade-offs Accepted

- **Dual deployments temporarily**: Both workflows will run until registry.tsuku.dev is fully deprecated
- **Build time increase**: Website deployment now includes Python setup and registry generation (~10-15 seconds)

## Test Coverage

- N/A - CI/CD configuration changes, tested via deployment pipeline

## Known Limitations

- Design doc (`docs/DESIGN-recipe-browser.md`) still references `registry.tsuku.dev` - can be updated in a separate PR

## Future Improvements

- Remove `deploy-recipes.yml` entirely after confirming no external consumers
- Update DNS to redirect `registry.tsuku.dev` to `tsuku.dev/recipes.json`
