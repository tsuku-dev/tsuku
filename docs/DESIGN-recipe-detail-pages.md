# Design: Recipe Detail Pages with Dependency Visualization

**Status**: Proposed

## Context and Problem Statement

The tsuku.dev website has a `/recipes/` page that shows a searchable grid of all available recipes with basic information (name, description, homepage link). However, users cannot view detailed information about individual recipes before installing them.

This creates several friction points:

1. **Hidden dependencies**: Users cannot see what dependencies a tool requires before installation. A user wanting to install Jekyll doesn't know it requires Ruby and Zig until after running `tsuku install jekyll`.

2. **No dedicated URLs**: There's no way to link to a specific tool's information. A blog post recommending k9s cannot link directly to tsuku's k9s page.

3. **Limited discoverability**: Users browsing the recipe list see only a brief description. They must visit external homepages to learn more about tools.

### Current State

Users can discover dependency information through the CLI:
- `tsuku info <tool>` shows recipe metadata but not dependencies
- `tsuku install <tool>` shows dependencies being installed as they happen
- No pre-installation preview of what will be installed

The website shows a searchable grid of tools with name, description, and homepage link only.

### Scope

**In scope:**
- Individual detail pages for each recipe at `/recipes/<tool>/`
- Display of install dependencies (runtime and build-time)
- Visual representation of dependency relationships
- Navigation between recipe grid and detail pages
- Progressive enhancement (core content accessible without JavaScript)

**Out of scope:**
- Version history or changelog
- Installation statistics per recipe (covered by `/stats/`)
- User reviews or ratings
- "Related tools" recommendations
- Recipe editing or submission through the website

### Assumptions

1. **Dependency data exists**: Recipe TOML files already include `dependencies` and `runtime_dependencies` fields where applicable.
2. **No circular dependencies**: The dependency graph is acyclic. (Verified: tsuku's dependency resolution validates this at build time.)
3. **Same deployment pipeline**: Website and registry can share a build step or trigger rebuilds together.
4. **Sparse dependencies**: Most recipes (pre-built binaries) have zero dependencies. This feature primarily benefits language-ecosystem tools (Ruby gems, Python packages, Rust crates).

## Decision Drivers

- **Static site architecture**: No build step beyond JSON generation; pages served as static HTML
- **Progressive enhancement**: Core content must be accessible without JavaScript
- **Dependency visibility**: Users should understand prerequisites before installing
- **URL stability**: Detail page URLs should be predictable and permanent
- **Minimal complexity**: Solution should integrate with existing vanilla JS patterns
- **Schema evolution**: Solution must handle recipes with or without dependency data

## Implementation Context

### Existing Patterns

**Recipe browser page** (`website/recipes/index.html`):
- Client-side fetches `https://registry.tsuku.dev/recipes.json`
- Vanilla JavaScript renders recipe cards using safe DOM APIs
- Debounced search filters in-memory array
- No build step - static HTML with inline script

**Current JSON schema** (v1.0.0):
```json
{
  "schema_version": "1.0.0",
  "recipes": [{
    "name": "k9s",
    "description": "Kubernetes CLI and TUI",
    "homepage": "https://k9scli.io/"
  }]
}
```

**Recipe TOML dependency fields**:
- `dependencies`: Tools required to build/install (e.g., `["ruby", "zig"]`)
- `runtime_dependencies`: Tools required to use after installation (e.g., `["golang"]`)
- Most recipes have neither field (pre-built binaries with no dependencies)

**JSON generation** (`scripts/generate-registry.py`):
- Reads TOML files, extracts metadata, writes JSON
- Currently only extracts name, description, homepage
- Would need modification to include dependency data

### Conventions to Follow

- Use `textContent` not `innerHTML` for user data
- All external links use `target="_blank" rel="noopener noreferrer"`
- Validate URLs are HTTPS before rendering
- Match existing dark theme CSS variables
- No build step - static files only

## Considered Options

This design involves three independent decisions:

### Decision 1: Page Generation Strategy

How should individual recipe detail pages be created and served?

#### Option 1A: Static HTML Pages (Build-Time)

Generate individual `recipes/<tool>/index.html` files during the registry build step.

**Pros:**
- Pages work without JavaScript (true progressive enhancement)
- SEO-friendly: search engines can index individual tool pages
- Fast initial render - no fetch required
- URL structure is guaranteed stable (real files)

**Cons:**
- Requires build step changes (Python script modifications)
- Site deployment depends on recipe registry build
- 267+ HTML files to generate and deploy
- Updating page template requires full rebuild

#### Option 1B: Single-Page with History API Routing

One HTML page at `/recipes/index.html` handles both grid and detail views using the History API for clean URLs.

**Pros:**
- No build step changes - pure JavaScript solution
- Recipe data already fetched; navigation between grid and detail is instant
- Simple deployment - one HTML file plus Cloudflare Pages catch-all redirect

**Cons:**
- Requires JavaScript for detail views (fails progressive enhancement)
- Requires catch-all redirect: `/recipes/*` → `/recipes/index.html`
- More JavaScript complexity for routing logic
- Direct links work but show loading state before content

#### Option 1C: Dynamic HTML Template with URL Rewriting

Single template file serves all `/recipes/<tool>/` URLs via Cloudflare Pages Functions or redirects.

**Pros:**
- Clean URL structure
- Single template to maintain
- No build of 267+ files

**Cons:**
- Cloudflare Pages Functions add complexity (separate deployment)
- Redirect-based approaches break direct linking
- Template still needs JavaScript to fetch tool-specific data
- Adds runtime dependency on Cloudflare infrastructure

### Decision 2: Dependency Data Location

Where should dependency information be stored and how should it be accessed?

#### Option 2A: Extend recipes.json Schema

Add `dependencies` and `runtime_dependencies` arrays to each recipe object in recipes.json.

**Pros:**
- Single fetch gets all data needed for detail pages
- Consistent with existing data flow
- Dependencies visible in same JSON used by grid
- Simple schema change (additive, backwards compatible)

**Cons:**
- Larger JSON payload (~20% increase estimated)
- All dependency data loaded even for grid view
- Requires `generate-registry.py` changes

#### Option 2B: Separate Dependencies JSON

Create a new `dependencies.json` file with dependency data only.

**Pros:**
- Grid page continues to use lean recipes.json
- Dependency data fetched only when needed
- Clear separation of concerns

**Cons:**
- Two fetches for detail pages (recipes + dependencies)
- More complex data joining in JavaScript
- Two files to keep in sync
- More build script complexity

*Note: A third option (inline dependencies in HTML) was considered but is really an implementation detail of Option 1A rather than a separate architectural choice. If 1A is chosen, dependency data would naturally be embedded in the generated HTML.*

### Decision 3: Dependency Visualization

How should dependencies be visually represented on detail pages?

#### Option 3A: Simple List

Display dependencies as a bulleted list, grouped by type (install vs runtime).

**Pros:**
- Simple to implement
- Works well without JavaScript
- Clear and scannable
- Accessible by default

**Cons:**
- Doesn't show transitive dependencies
- No visual hierarchy
- Less engaging than graphical display

#### Option 3B: Interactive Tree Diagram

Use JavaScript to render an expandable tree showing dependency chains.

**Pros:**
- Shows full dependency graph including transitives
- Visually engaging
- Collapsible for complex chains

**Cons:**
- Requires JavaScript
- Complex to implement well
- May be overkill for typical 1-2 dependency chains
- Accessibility concerns with interactive trees

#### Option 3C: Static Nested List

Show dependencies with their own dependencies as nested sublists.

**Pros:**
- Shows transitive dependencies without JavaScript
- Works with screen readers
- Moderate implementation complexity

**Cons:**
- Requires pre-computing transitive dependencies
- Deep nesting may be visually confusing
- Need to handle circular dependencies

### Option Evaluation Matrix

| Decision | Driver: No Build Step | Driver: Works without JS | Driver: URL Stability | Driver: Minimal Complexity |
|----------|----------------------|--------------------------|----------------------|---------------------------|
| 1A: Static HTML | Poor | Good | Good | Fair |
| 1B: SPA + History API | Good | Poor | Good | Fair |
| 1C: Dynamic Template | Poor | Poor | Good | Poor |
| 2A: Extend JSON | Good | N/A | N/A | Good |
| 2B: Separate JSON | Good | N/A | N/A | Poor |
| 3A: Simple List | Good | Good | N/A | Good |
| 3B: Tree Diagram | Good | Poor | N/A | Poor |
| 3C: Nested List | Fair | Good | N/A | Fair |

### Uncertainties

- Performance impact of larger recipes.json on slow connections not measured (current JSON is ~50KB; adding dependencies would add ~5KB estimated)
- Whether users value transitive dependency visibility or just immediate dependencies

## Decision Outcome

**Chosen: 1A + 2A + 3A**

### Summary

Generate static HTML detail pages at build time with dependency data embedded in an extended recipes.json schema. Display dependencies as simple grouped lists (install vs runtime).

### Rationale

This combination prioritizes **progressive enhancement** and **URL stability** - the two most important drivers for a documentation-style website intended to help users make informed decisions before installation.

**Decision 1: Static HTML Pages (1A)** was chosen because:
- Detail pages work without JavaScript, critical for users researching tools in constrained environments
- Clean, permanent URLs (`/recipes/k9s/`) enable direct linking from documentation, blog posts, and other resources
- SEO benefits allow search engines to index individual tool pages
- The "build step" concern is minimal since `generate-registry.py` already exists and generates JSON; extending it to generate HTML is straightforward

**Decision 2: Extend recipes.json (2A)** was chosen because:
- Single data source maintains consistency - the grid and detail pages always show the same data
- Minimal payload increase (~5KB for 267 recipes with sparse dependency data)
- Backwards compatible schema change - clients ignoring dependencies still work
- Simpler than maintaining separate JSON files

**Decision 3: Simple List (3A)** was chosen because:
- Most recipes have zero or 1-2 direct dependencies - elaborate visualization would be over-engineered
- Works without JavaScript, maintaining progressive enhancement
- Accessible by default (screen readers handle lists well)
- Can be enhanced later if transitive dependency visualization proves valuable

### Alternatives Rejected

- **Option 1B (SPA routing)**: Fails progressive enhancement - users without JavaScript see nothing
- **Option 1C (Dynamic template)**: Adds infrastructure complexity (Cloudflare Functions) for no additional benefit
- **Option 2B (Separate JSON)**: Extra complexity for marginal payload savings; complicates caching strategy
- **Option 3B (Tree diagram)**: Over-engineered for typical 0-2 dependency scenarios; fails progressive enhancement
- **Option 3C (Nested list)**: Adds complexity for transitive deps that most users don't need

### Trade-offs Accepted

By choosing static generation with simple lists, we accept:

1. **Build step dependency**: Detail pages update only when `generate-registry.py` runs. This is acceptable because:
   - Recipes change infrequently (weekly at most)
   - The build already runs on recipe changes
   - Stale data by hours/days is acceptable for documentation content

2. **No transitive dependency display**: Users see only direct dependencies. This is acceptable because:
   - Direct dependencies are the actionable information ("install these first")
   - Transitive deps can be explored by clicking through to dependency pages
   - Most users don't care about full dependency graphs

3. **Larger repository size**: 267+ HTML files add to repository. This is acceptable because:
   - Each file is ~5KB = ~1.3MB total
   - Files are generated, not hand-maintained
   - Modern hosting (Cloudflare Pages) handles this trivially

## Solution Architecture

### Overview

The solution extends the existing registry generation script to produce both JSON data and static HTML detail pages. Detail pages are self-contained HTML files with embedded recipe data.

```
Recipe TOML Files
       │
       ▼
┌─────────────────────────────┐
│  generate-registry.py       │
│  (extended script)          │
└─────────────────────────────┘
       │
       ├──────────────────────┐
       ▼                      ▼
┌─────────────────┐    ┌─────────────────────────┐
│ recipes.json    │    │ recipes/<tool>/         │
│ (extended)      │    │ index.html (×267+)      │
└─────────────────┘    └─────────────────────────┘
       │                      │
       ▼                      ▼
┌─────────────────┐    ┌─────────────────────────┐
│ /recipes/       │    │ /recipes/k9s/           │
│ Grid view       │    │ Detail page             │
└─────────────────┘    └─────────────────────────┘
```

### Components

#### 1. Extended JSON Schema (v1.1.0)

```json
{
  "schema_version": "1.1.0",
  "generated_at": "2025-12-07T12:00:00Z",
  "recipes": [{
    "name": "jekyll",
    "description": "Static site generator for personal, project, or organization sites",
    "homepage": "https://jekyllrb.com/",
    "dependencies": ["ruby", "zig"],
    "runtime_dependencies": []
  }, {
    "name": "k9s",
    "description": "Kubernetes CLI and TUI",
    "homepage": "https://k9scli.io/",
    "dependencies": [],
    "runtime_dependencies": []
  }]
}
```

**Schema changes:**
- `schema_version` bumped to "1.1.0" (minor version, backwards compatible)
- New optional fields: `dependencies` (array of strings), `runtime_dependencies` (array of strings)
- Both default to empty array if not present in TOML

#### 2. Static HTML Detail Page Template

Each detail page at `recipes/<tool>/index.html` contains:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="{description} - Install with tsuku">
    <title>{name} - tsuku</title>
    <link rel="stylesheet" href="/assets/style.css">
</head>
<body>
    <header><!-- standard nav --></header>
    <main>
        <section class="recipe-detail">
            <h1>{name}</h1>
            <p class="description">{description}</p>
            <p class="homepage">
                <a href="{homepage}" target="_blank" rel="noopener noreferrer">
                    Official Homepage
                </a>
            </p>

            <h2>Install</h2>
            <div class="code-block">
                <pre>tsuku install {name}</pre>
            </div>

            <!-- Only shown if dependencies exist -->
            <section class="dependencies">
                <h2>Dependencies</h2>

                <!-- Install dependencies -->
                <h3>Required to Install</h3>
                <ul>
                    <li><a href="/recipes/ruby/">ruby</a></li>
                    <li><a href="/recipes/zig/">zig</a></li>
                </ul>

                <!-- Runtime dependencies (if any) -->
                <h3>Required at Runtime</h3>
                <p>None</p>
            </section>
        </section>

        <section class="links">
            <a href="/recipes/" class="link-btn">Back to Recipes</a>
        </section>
    </main>
    <footer><!-- standard footer --></footer>
</body>
</html>
```

#### 3. Recipe Grid Navigation

Update `recipes/index.html` recipe cards to link to detail pages:

```html
<article class="recipe-card">
    <h3 class="recipe-name">
        <a href="/recipes/k9s/">k9s</a>
    </h3>
    <p class="recipe-description">Kubernetes CLI and TUI</p>
    <a class="recipe-homepage" href="https://k9scli.io/"
       target="_blank" rel="noopener noreferrer">Homepage</a>
</article>
```

### Data Flow

1. **Build time** (`generate-registry.py`):
   - Read all `recipes/*/*.toml` files
   - Extract metadata including `dependencies` and `runtime_dependencies`
   - Write `_site/recipes.json` with extended schema
   - For each recipe, render `_site/recipes/<name>/index.html` from template

2. **Grid page load** (`/recipes/`):
   - Fetch recipes.json (unchanged behavior)
   - Render cards with links to `/recipes/<name>/`
   - Dependency data in JSON ignored by grid (backwards compatible)

3. **Detail page load** (`/recipes/<name>/`):
   - Static HTML served directly by Cloudflare Pages
   - No JavaScript required for core content
   - Links to dependency pages are just `<a href="/recipes/dep/">`

### Key Interfaces

**Template variables for HTML generation:**

| Variable | Source | Example |
|----------|--------|---------|
| `{name}` | `metadata.name` | `jekyll` |
| `{description}` | `metadata.description` | `Static site generator...` |
| `{homepage}` | `metadata.homepage` | `https://jekyllrb.com/` |
| `{dependencies}` | `metadata.dependencies` | `["ruby", "zig"]` |
| `{runtime_dependencies}` | `metadata.runtime_dependencies` | `[]` |

**File output structure:**

```
_site/
├── recipes.json           # Extended JSON with dependencies
└── recipes/
    ├── index.html         # Grid page (existing, modified)
    ├── actionlint/
    │   └── index.html     # Detail page
    ├── k9s/
    │   └── index.html
    └── jekyll/
        └── index.html
```

## Implementation Approach

### Phase 0: Deployment Strategy

Clarify where generated files live and how they deploy:

1. HTML detail pages generate to `_site/recipes/<name>/index.html` (same location as recipes.json)
2. Update `deploy-website.yml` to copy `_site/recipes/` to website deployment
3. Both grid and detail pages will be on `tsuku.dev` (same domain)
4. `recipes.json` continues to be served from `registry.tsuku.dev` for backwards compatibility

**Deliverable:** Deployment workflow updated

### Phase 1: Extend JSON Schema

1. Modify `generate-registry.py` to extract `dependencies` and `runtime_dependencies` from TOML
2. Add validation: dependency arrays must be lists, each name matches `NAME_PATTERN`
3. Add validation: each dependency references an existing recipe (prevents broken links)
4. Update JSON output to include these fields (empty arrays if not present)
5. Bump `schema_version` to "1.1.0"
6. Verify existing recipe browser still works (backwards compatibility)

**Deliverable:** Extended recipes.json with dependency data

### Phase 2: Create HTML Generation

1. Add HTML template string to `generate-registry.py`
2. Implement template rendering with safe escaping
3. Generate `_site/recipes/<name>/index.html` for each recipe
4. Handle dependency section conditionally (hide if no deps)

**Deliverable:** Static HTML detail pages generated

### Phase 3: Style Detail Pages

1. Add CSS for `.recipe-detail` component to `style.css`
2. Style dependency lists (grouped by type)
3. Ensure responsive layout at mobile breakpoints
4. Match existing dark theme aesthetic

**Deliverable:** Styled detail pages

### Phase 4: Update Grid Navigation

1. Modify `recipes/index.html` JavaScript to render card names as links
2. Update card hover states for clickable cards
3. Maintain existing search/filter functionality

**Deliverable:** Grid cards link to detail pages

### Phase 5: Deployment Integration

1. Update GitHub Actions workflow to copy generated HTML to deployment
2. Verify Cloudflare Pages serves `/recipes/<name>/` correctly
3. Test direct URL access and navigation

**Deliverable:** End-to-end flow working in production

## Consequences

### Positive

- **Dependency visibility**: Users can see all prerequisites before installing
- **Direct linking**: Tools can be referenced with permanent URLs
- **SEO**: Search engines can index individual tool pages
- **Progressive enhancement**: Core content works without JavaScript
- **Consistent data**: Grid and detail pages share the same data source

### Negative

- **Build complexity**: Script now generates both JSON and HTML
- **Repository size**: 267+ generated HTML files (~1.3MB)
- **Template maintenance**: HTML template changes require regeneration

### Mitigations

- **Build complexity**: Keep template simple; use Python's built-in string formatting
- **Repository size**: Generated files go in `_site/` (excluded from main repo); only deployed to Cloudflare
- **Template maintenance**: Template is inline in Python script; changes automatically propagate on next build

## Security Considerations

### Download Verification

**Not applicable** - This feature does not download or execute binaries. It generates static HTML pages from trusted source data (recipe TOML files in the same repository) and displays metadata to users.

### Execution Isolation

**Not applicable** - No code execution occurs beyond:
1. Build-time Python script (already trusted, runs in GitHub Actions)
2. Browser rendering of static HTML (standard web security model)

The detail pages contain no JavaScript by design. The grid page's JavaScript is unchanged.

### Supply Chain Risks

**Data source trust model:**

Recipe metadata originates from TOML files in the tsuku repository, controlled by:
1. Branch protection requiring PR review
2. Maintainer approval for recipe changes
3. Automated validation in CI

**Risk: Malicious recipe metadata could be embedded in HTML**

If an attacker compromised the recipe repository and added malicious content to a recipe's `name`, `description`, or `homepage` field, this could be rendered in the detail pages.

| Attack Vector | Likelihood | Impact | Mitigation |
|--------------|------------|--------|------------|
| XSS via recipe name/description | Very Low | High | HTML escape all template variables |
| Phishing via homepage URL | Very Low | Medium | Validate HTTPS-only; use `rel="noopener noreferrer"` |
| Dependency link manipulation | Very Low | Low | Validate deps match NAME_PATTERN; link only to existing recipes |
| Path traversal in dependency names | Very Low | Low | Validate each dependency name matches `^[a-z0-9-]+$` |

**Mitigations implemented:**

1. **HTML escaping**: All template variables are escaped using Python's `html.escape()` before insertion
2. **URL validation**: Homepage URLs validated as HTTPS during generation (existing check)
3. **Dependency links**: Only link to dependencies that exist as recipes (prevent link injection)
4. **No user input**: Detail pages have no forms, inputs, or user-controllable content

### User Data Exposure

**Data accessed**: None. Detail pages are static HTML with no JavaScript accessing user data.

**Data transmitted**: Standard HTTP requests for page resources. No cookies, localStorage, or analytics on detail pages.

**Privacy implications**: None. The feature displays public recipe metadata only.

### Additional Security Measures

1. **Content Security Policy**: Add CSP header in `_headers` for detail pages:
   ```
   /recipes/*
     Content-Security-Policy: default-src 'none'; style-src 'self'; img-src 'self'; script-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'none'
   ```

   Key directives:
   - `default-src 'none'`: Deny all by default, whitelist what's needed
   - `script-src 'none'`: Enforce "no JavaScript" design principle
   - `frame-ancestors 'none'`: Prevent clickjacking via iframes
   - `img-src 'self'`: Only allow images from tsuku.dev (detail pages have no external images)

2. **Safe template rendering**: Use Python's `html.escape()` for all string interpolation
   ```python
   from html import escape
   html = f"<h1>{escape(recipe['name'])}</h1>"
   ```

3. **Link validation**: Validate and generate dependency links safely:
   ```python
   # During validation phase
   for dep in metadata.get("dependencies", []):
       if not NAME_PATTERN.match(dep):
           errors.append(f"invalid dependency name: {dep}")
       if dep not in all_recipe_names:
           errors.append(f"dependency does not exist: {dep}")

   # During HTML generation (validation already passed)
   from urllib.parse import quote
   link = f'<a href="/recipes/{quote(escape(dep), safe="")}/">{escape(dep)}</a>'
   ```

4. **Dependency array validation**: Ensure dependencies are arrays, not objects:
   ```python
   if "dependencies" in metadata:
       if not isinstance(metadata["dependencies"], list):
           errors.append("dependencies must be an array")
       if len(metadata["dependencies"]) > 50:  # Reasonable limit
           errors.append("too many dependencies")
   ```

