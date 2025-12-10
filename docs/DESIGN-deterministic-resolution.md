# Design: Deterministic Recipe Resolution and Immutable Snapshots

- **Status**: Proposed
- **Issue**: #227
- **Author**: @dangazineu
- **Created**: 2025-12-09

## Context and Problem Statement

Tsuku recipes are dynamic by design - they contain templates and version provider configuration that determine how to resolve and install tools. This dynamism is a feature: recipes like `go.toml` work across all Go versions without hardcoding URLs.

However, this creates a **reproducibility problem**:

**Current flow:**
```
tsuku install ripgrep@14.1.0
    ↓
Recipe (templates) → Version Resolution (GitHub API) → Template Expansion → Download → Install
```

Even when a user specifies an exact version like `ripgrep@14.1.0`, several factors are resolved at runtime:
1. **Platform detection**: `{os}` and `{arch}` are expanded based on the current machine
2. **URL construction**: The download URL is constructed dynamically from templates
3. **Asset selection**: For GitHub releases, the specific asset to download is matched at runtime

This means:
- Running `tsuku install ripgrep@14.1.0` today vs. tomorrow could theoretically yield different binaries
- There's no record of exactly what was downloaded - only what version was requested
- Teams cannot guarantee identical tool installations across machines
- CI/CD builds may diverge from developer environments

**Concrete non-determinism scenarios:**

1. **Upstream re-tags a release**: A maintainer force-pushes a tag to fix a security issue. `tsuku install tool@1.0.0` now downloads a different binary than last week, but the version number is unchanged.

2. **Asset naming changes**: GitHub release assets can be renamed or replaced. A recipe pattern like `tool-{version}-{os}.tar.gz` might match different files over time.

3. **Recipe updates**: The recipe registry updates `ripgrep.toml` to use a different download mirror or change the asset selection pattern. The same `tsuku install ripgrep@14.1.0` command now resolves to a different URL.

4. **Platform drift**: A developer installs on Linux x64, but CI runs on arm64. Without explicit platform tracking, reinstallation on a different machine silently uses different binaries.

**Why this matters now:**

1. **Team workflows**: As tsuku gains adoption, teams need confidence that `tsuku install` produces identical results across all machines
2. **CI/CD reproducibility**: Build pipelines should install the exact same tool binaries every time
3. **Audit trail**: For security-sensitive environments, knowing exactly what was installed (URLs, checksums) matters
4. **Testing**: Recipe builders need to verify their changes don't alter the effective installation plan

### Scope

**In scope:**
- Two-phase installation model: evaluation (produces plan) and execution (runs plan)
- Immutable installation plans that capture all resolved state
- New `tsuku eval` command to produce plans without installing
- Re-install semantics that replay plans by default
- Simple lock files that reference version pins

**Out of scope:**
- Changes to recipe format (recipes remain dynamic)
- Cryptographic signature verification (separate concern, see #208)
- Remote plan sharing/caching (future enhancement)

## Decision Drivers

1. **Determinism by default**: Every installation should be reproducible without opt-in flags
2. **Transparency**: Users should be able to inspect exactly what will be downloaded before it happens
3. **Testability**: Recipe changes should be verifiable without actual installation
4. **Simplicity**: Lock files should be simple version pins, not duplicated resolution metadata
5. **Separation of concerns**: "What to install" (eval) should be separate from "how to install" (exec)

## Core Insight

**A recipe is a program that produces a deterministic installation plan.**

The key architectural shift is separating recipe evaluation from plan execution:

```
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 1: Evaluation (dynamic, may call external APIs)             │
│                                                                     │
│  Recipe + Version + Platform → Installation Plan                    │
│                                                                     │
│  - Query version providers (GitHub, npm, etc.)                      │
│  - Expand templates ({version}, {os}, {arch})                       │
│  - Select assets from releases                                      │
│  - Compute expected checksums                                       │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    Installation Plan (JSON)
                    - Fully resolved URLs
                    - Expected checksums
                    - Concrete installation steps
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 2: Execution (deterministic, no external API calls)         │
│                                                                     │
│  Installation Plan → Download → Verify → Install                    │
│                                                                     │
│  - Download from exact URLs in plan                                 │
│  - Verify checksums match plan                                      │
│  - Execute concrete installation steps                              │
│  - Fail if any checksum mismatches                                  │
└─────────────────────────────────────────────────────────────────────┘
```

## Considered Options

### Decision 1: Installation Plan Storage

**Question:** Where should installation plans be stored?

#### Option 1A: Inline in state.json

Store plans as part of the existing state file:

```json
{
  "installed": {
    "ripgrep": {
      "active_version": "14.1.0",
      "versions": {
        "14.1.0": {
          "plan": { /* full installation plan */ }
        }
      }
    }
  }
}
```

**Pros:**
- No new files
- Plans naturally associated with installed versions

**Cons:**
- state.json grows large
- Plans not easily shareable or inspectable

#### Option 1B: Separate Plan Files

Store plans in dedicated directory:

```
$TSUKU_HOME/
├── plans/
│   ├── ripgrep@14.1.0-linux-x64.json
│   └── go@1.22.0-linux-x64.json
└── tools/
    └── ...
```

**Pros:**
- Plans are inspectable files
- Can be copied/shared independently
- state.json stays lightweight

**Cons:**
- Additional files to manage
- Need to handle orphaned plans

#### Option 1C: Plans in state.json, exportable on demand

Store plans inline but provide export command:

```bash
tsuku plan show ripgrep@14.1.0    # Display stored plan
tsuku plan export ripgrep@14.1.0  # Export to file
```

**Pros:**
- Simple storage model
- Export when needed for sharing/testing

**Cons:**
- Two representations to keep in sync

---

### Decision 2: Re-install Behavior

**Question:** What happens when you run `tsuku install ripgrep@14.1.0` for a tool that's already installed?

#### Option 2A: Re-evaluate (current behavior)

Always re-run evaluation, potentially getting different results.

**Pros:**
- Picks up upstream changes automatically
- Simpler mental model

**Cons:**
- Not deterministic
- Can't guarantee same binary

#### Option 2B: Replay Plan by Default

If a plan exists for this tool+version+platform, replay it:

```bash
tsuku install ripgrep@14.1.0
# → Uses stored plan, verifies checksums
# → Fails if upstream content changed

tsuku install ripgrep@14.1.0 --refresh
# → Re-evaluates recipe, creates new plan
```

**Pros:**
- Deterministic by default
- Explicit opt-in to get new content
- Detects upstream changes via checksum failure

**Cons:**
- Stale plans if upstream fixes security issues
- Users must learn `--refresh` flag

#### Option 2C: Verify and Warn

Replay plan but warn if checksums differ without failing:

**Pros:**
- Non-breaking behavior change

**Cons:**
- Weaker guarantee
- Warning fatigue

---

### Decision 3: Lock File Format

**Question:** What should lock files contain?

#### Option 3A: Full Plans (current design)

Lock files contain complete resolution metadata:

```toml
[[tools.ripgrep]]
version = "14.1.0"

[tools.ripgrep.platforms.linux-x64]
url = "https://github.com/..."
checksum = "sha256:..."
```

**Pros:**
- Self-contained
- Works without plan storage

**Cons:**
- Duplicates plan information
- Complex structure
- Platform-specific sections

#### Option 3B: Version Pins Only

Lock files are simple version declarations:

```toml
[tools]
ripgrep = "14.1.0"
go = "1.22.0"
node = "20.11.0"
```

**Pros:**
- Simple, human-readable
- Cross-platform (same lock file for all platforms)
- Plans handle the complexity

**Cons:**
- Requires plan storage/regeneration
- Less self-contained

#### Option 3C: Version Pins with Checksum Validation

Lock files declare versions, execution validates checksums:

```toml
[tools]
ripgrep = "14.1.0"
go = "1.22.0"

[checksums.linux-x64]
"ripgrep@14.1.0" = "sha256:abc123..."
"go@1.22.0" = "sha256:def456..."
```

**Pros:**
- Simple version section
- Per-platform checksums for validation
- Can detect upstream changes

**Cons:**
- Two sections to manage
- Checksums must be updated per-platform

---

### Evaluation Against Decision Drivers

| Decision | Option | Determinism | Transparency | Testability | Simplicity |
|----------|--------|-------------|--------------|-------------|------------|
| Storage | 1A: Inline | Good | Poor | Fair | Good |
| Storage | 1B: Separate files | Good | Good | Good | Fair |
| Storage | 1C: Inline + export | Good | Good | Good | Good |
| Re-install | 2A: Re-evaluate | Poor | Fair | Poor | Good |
| Re-install | 2B: Replay plan | Good | Good | Good | Fair |
| Re-install | 2C: Warn only | Fair | Fair | Fair | Good |
| Lock format | 3A: Full plans | Good | Fair | Fair | Poor |
| Lock format | 3B: Version pins | Good | Good | Good | Good |
| Lock format | 3C: Pins + checksums | Good | Good | Good | Fair |

## Decision Outcome

**Chosen: 1C (Inline + export) + 2B (Replay plan) + 3B (Version pins)**

### Summary

Installation becomes a two-phase process: evaluation produces an immutable plan, execution replays that plan deterministically. Plans are stored in state.json and can be exported for inspection or testing. Lock files are simple version pins - the complexity lives in plans, not lock files.

### Rationale

**Storage (1C: Inline + export):**
- Plans stored in state.json keeps the model simple
- Export enables inspection and testing use cases
- No orphaned files to manage

**Re-install (2B: Replay plan by default):**
- Determinism is the default, not opt-in
- Users get explicit `--refresh` when they want new content
- Checksum mismatches are failures, not warnings - this is a security feature

**Lock format (3B: Version pins only):**
- Lock files become trivially simple
- Plans handle all the complexity
- Cross-platform teams use the same lock file
- `tsuku install --lock` = "install these versions, evaluate fresh plans for each"

### Trade-offs Accepted

1. **Learning curve**: Users must understand eval vs exec phases. Mitigated by good documentation and intuitive defaults.

2. **Stale plans**: Plans don't auto-update. Mitigated by `--refresh` flag and `tsuku outdated` warnings.

3. **Lock files don't contain checksums**: Validation happens at plan level. Acceptable because plans are the source of truth.

## Solution Architecture

### Overview

```
                            ┌──────────────────────────────────────────┐
                            │           tsuku install ripgrep          │
                            └──────────────────────────────────────────┘
                                               │
                        ┌──────────────────────┴──────────────────────┐
                        │                                              │
                        ▼                                              ▼
               Plan exists for                                 No plan exists
               ripgrep@<version>                                      │
               on this platform?                                      │
                        │                                              │
                        ▼                                              │
                  Replay plan                              ┌──────────┴──────────┐
                  (Phase 2 only)                           │   tsuku eval ripgrep │
                        │                                  │   (Phase 1)          │
                        │                                  └──────────┬──────────┘
                        │                                              │
                        │                                              ▼
                        │                                     Installation Plan
                        │                                              │
                        │                                              ▼
                        │                                  ┌──────────────────────┐
                        │                                  │   Execute Plan       │
                        │                                  │   (Phase 2)          │
                        │                                  └──────────┬──────────┘
                        │                                              │
                        └──────────────────────┬───────────────────────┘
                                               │
                                               ▼
                                    Download → Verify → Install
                                               │
                                               ▼
                                    Store plan in state.json
```

### New Commands

| Command | Description |
|---------|-------------|
| `tsuku eval <tool>[@version]` | Produce installation plan without installing |
| `tsuku eval <tool> --output plan.json` | Save plan to file |
| `tsuku install <tool> --refresh` | Re-evaluate plan even if one exists |
| `tsuku install --plan <file>` | Install from a pre-computed plan file |
| `tsuku plan show <tool>[@version]` | Display stored plan for installed tool |
| `tsuku plan export <tool>[@version]` | Export plan to file |

### Installation Plan Format

```json
{
  "schema_version": 1,
  "tool": "ripgrep",
  "version": "14.1.0",
  "platform": "linux-x64",
  "evaluated_at": "2025-12-09T10:30:00Z",
  "recipe_hash": "sha256:...",
  "downloads": [
    {
      "url": "https://github.com/BurntSushi/ripgrep/releases/download/14.1.0/ripgrep-14.1.0-x86_64-unknown-linux-musl.tar.gz",
      "checksum": "sha256:abc123...",
      "size": 1234567,
      "extract": {
        "format": "tar.gz",
        "strip_components": 1
      }
    }
  ],
  "binaries": ["rg"],
  "verify": {
    "command": "rg --version",
    "pattern": "ripgrep {version}"
  }
}
```

### Lock File Format

```toml
# tsuku.lock - simple version pins
# Plans are evaluated per-platform at install time

[tools]
ripgrep = "14.1.0"
go = "1.22.0"
node = "20.11.0"
```

When `tsuku install --lock` runs:
1. Read versions from lock file
2. For each tool, check if plan exists for this version+platform
3. If plan exists: replay it
4. If no plan: evaluate recipe for locked version, save plan, execute

### Key Data Structures

```go
// InstallationPlan represents a fully-resolved, deterministic installation.
type InstallationPlan struct {
    SchemaVersion int       `json:"schema_version"`
    Tool          string    `json:"tool"`
    Version       string    `json:"version"`
    Platform      string    `json:"platform"`
    EvaluatedAt   time.Time `json:"evaluated_at"`
    RecipeHash    string    `json:"recipe_hash,omitempty"`
    Downloads     []Download `json:"downloads"`
    Binaries      []string  `json:"binaries"`
    Verify        *VerifySpec `json:"verify,omitempty"`
}

type Download struct {
    URL      string       `json:"url"`
    Checksum string       `json:"checksum"`
    Size     int64        `json:"size,omitempty"`
    Extract  *ExtractSpec `json:"extract,omitempty"`
}

// LockFile represents a simple version-pinning lock file.
type LockFile struct {
    Tools map[string]string `toml:"tools"` // tool name → version
}
```

### Testing Use Case

The `tsuku eval` command enables golden file testing for recipes:

```bash
# Generate expected plans for a recipe
tsuku eval ripgrep@14.1.0 --output testdata/ripgrep-14.1.0-linux-x64.json

# In CI, verify recipe still produces same plan
tsuku eval ripgrep@14.1.0 --output actual.json
diff testdata/ripgrep-14.1.0-linux-x64.json actual.json
```

This allows:
- Testing builder changes without downloading binaries
- Verifying recipe updates don't change effective output
- Comparing plans across recipe versions

## Implementation Approach

### Phase 1: Installation Plans

**Goal:** Introduce the plan concept and `tsuku eval` command.

**Changes:**
- Add `InstallationPlan` type
- Create plan evaluation logic (refactor from executor)
- Add `tsuku eval` command
- Store plans in state.json alongside version info
- Add `tsuku plan show` and `tsuku plan export`

**Deliverables:**
- `tsuku eval ripgrep` outputs a plan
- Plans are stored after installation
- Plans can be inspected and exported

### Phase 2: Deterministic Execution

**Goal:** Make plan replay the default for re-installs.

**Changes:**
- Refactor executor to accept plans
- Check for existing plan before evaluation
- Add `--refresh` flag to force re-evaluation
- Implement checksum verification during execution
- Fail on checksum mismatch

**Deliverables:**
- Re-installing uses stored plan by default
- Checksum mismatches fail installation
- `--refresh` forces fresh evaluation

### Phase 3: Plan-Based Installation

**Goal:** Enable installation from plan files.

**Changes:**
- Add `tsuku install --plan <file>` flag
- Validate plan schema and checksums
- Support piping: `tsuku eval ripgrep | tsuku install --plan -`

**Deliverables:**
- Can install from exported plan files
- Enables offline installation (if artifacts are cached)

### Phase 4: Lock Files

**Goal:** Simple version-pinning lock files.

**Changes:**
- Add lock file parser (TOML, version pins only)
- Add `tsuku install --lock` flag
- Read versions from lock file, use plans for execution

**Deliverables:**
- `tsuku.lock` with simple `[tools]` section
- `tsuku install --lock` uses lock file versions

## Consequences

### Positive

1. **Determinism by default**: Re-installs are reproducible without opt-in
2. **Testability**: `tsuku eval` enables recipe testing without installation
3. **Transparency**: Users can inspect exactly what will be downloaded
4. **Simple lock files**: Version pins are easy to read, write, and merge
5. **Clear separation**: Evaluation vs execution have distinct responsibilities

### Negative

1. **Behavior change**: Re-install no longer picks up upstream changes automatically
2. **Plan storage**: state.json grows with plan data
3. **Learning curve**: Users must understand when to use `--refresh`

### Mitigations

| Consequence | Mitigation |
|-------------|------------|
| Behavior change | Clear documentation; `tsuku outdated` warns of stale plans |
| Plan storage | Plans are compact JSON; consider pruning old plans |
| Learning curve | Sensible defaults; `--refresh` is explicit opt-in |

## Security Considerations

### Download Verification

**How are downloaded artifacts validated?**

Checksums are mandatory in plans. During execution:
1. Download artifact to temporary location
2. Compute SHA256 of downloaded content
3. Compare with plan's expected checksum
4. **Fail immediately** if mismatch - this indicates upstream content changed

**TOCTOU mitigation:**
- Download to temp directory with restricted permissions (0700)
- Verify checksum before any extraction
- Use atomic rename to move verified content to final location

### Execution Isolation

**What permissions does this feature require?**

No change to execution model:
- File system: write to `$TSUKU_HOME/` directory
- Network: HTTPS to upstream sources
- No privilege escalation

**Plan files as trusted input:**
When using `--plan`, the file is trusted:
- URLs are used directly for downloads
- Checksums are used for verification
- A malicious plan could point to malicious URLs
- **Mitigation**: Plans should be generated via `tsuku eval`, not hand-crafted

### Supply Chain Risks

**Where do artifacts come from?**

Plans capture exact URLs and checksums:
- If upstream re-tags a release, checksum mismatch is detected
- Plans provide audit trail of what was intended
- `--refresh` explicitly opts into new upstream content

**Residual risk:** Initial plan creation inherits any existing compromise.

### User Data Exposure

**What user data does this feature access or transmit?**

Minimal exposure:
- Plans contain URLs and checksums (not sensitive)
- Lock files contain tool names and versions (not sensitive)
- No new network traffic beyond current behavior

## Future Enhancements

### Plan Caching/Sharing

Plans could be cached centrally:
- `tsuku eval` checks plan cache before evaluating
- Common tool+version+platform combinations are pre-computed
- Reduces API calls to version providers

### Offline Installation

With pre-computed plans and cached artifacts:
- `tsuku install --offline` uses only local plans and cache
- Useful for air-gapped environments

### Plan Diffing

Compare plans to understand changes:
```bash
tsuku plan diff ripgrep@14.0.0 ripgrep@14.1.0
```
