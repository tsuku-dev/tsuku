# Issue 340 Baseline

## Environment
- Date: 2025-12-09
- Branch: feature/340-recipe-router
- Base commit: 93f8dbf4c2e0b8876462fb10b93e6bfef6916821

## Project Type
Static website (vanilla JavaScript) - no test suite

## Current State
- `website/recipes/index.html`: Grid view with search, fetches from `registry.tsuku.dev/recipes.json`
- `website/_redirects`: Cloudflare Pages routing rules
- `website/assets/style.css`: Dark theme styles

## Key Existing Functions
- `loadRecipes()`: Fetches JSON, validates, stores in `allRecipes`
- `renderRecipes()`: Clears grid, renders recipe cards
- `createRecipeCard()`: Safe DOM creation for cards
- `filterRecipes()`: Search filtering
- `handleSearch()`: Debounced search handler

## Pre-existing Issues
None - clean starting point
