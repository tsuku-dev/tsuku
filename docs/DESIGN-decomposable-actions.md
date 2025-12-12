# Design: Decomposable Actions and Primitive Operations

- **Status**: Proposed
- **Issue**: #368
- **Author**: @dangazineu
- **Created**: 2025-12-12
- **Scope**: Tactical

## Upstream Design Reference

This design implements part of [DESIGN-deterministic-resolution.md](DESIGN-deterministic-resolution.md).

**Relevant sections:**
- Core Insight: Split installation into evaluation and execution phases
- Milestone 2: Deterministic Execution

## Context and Problem Statement

The deterministic recipe execution design establishes a two-phase model: evaluation produces an installation plan, execution replays that plan. However, the current implementation has a structural flaw: **composite actions are treated as atomic units in plans, but contain runtime logic that can produce different outcomes**.

For example, `github_archive` is marked as "evaluable" in `ActionEvaluability`, but during execution it:
1. Resolves wildcard asset patterns via GitHub API
2. Constructs download URLs with runtime logic
3. Internally calls download, extract, chmod, and install_binaries actions

This means the plan captures `github_archive` with parameters, but execution may behave differently than what was evaluated - violating the determinism guarantee.

**The core problem**: Plans should contain only primitive operations that execute deterministically. Composite actions are recipe-authoring conveniences that must decompose during evaluation.

## Decision Drivers

1. **Single execution path**: `tsuku install foo` and `tsuku eval foo | tsuku install --plan -` must be semantically equivalent
2. **Structural determinism**: Determinism should be guaranteed by architecture, not by careful implementation
3. **Honest barriers**: Where true determinism is impossible (ecosystem installers), capture maximum constraint and be explicit about limitations
4. **Incremental adoption**: Design must work with existing recipes without breaking changes

## Considered Options

### Option 1: Composite Decomposition at Eval Time

Composite actions implement a `Decompose()` method that returns primitive steps. The plan generator calls this instead of just recording the composite action.

**Pros:**
- Plans contain only primitives - execution is trivially deterministic
- Clear separation between recipe DSL and execution model
- Testable: can unit test decomposition logic

**Cons:**
- Requires refactoring all composite actions
- Decomposition logic may duplicate execution logic initially

### Option 2: Interpreter Pattern with Action Rewriting

The plan generator "interprets" composite actions, rewriting them to primitive sequences. Composite actions remain unchanged but the plan generator understands their structure.

**Pros:**
- No changes to existing action implementations
- Centralized decomposition logic

**Cons:**
- Plan generator becomes complex, tightly coupled to action internals
- Decomposition logic diverges from execution logic over time

### Option 3: Unified Execute/Decompose with Mode Flag

Actions receive an execution mode (Eval vs Execute). In Eval mode, they record their operations to a plan builder rather than executing them.

**Pros:**
- Single code path for both modes
- Decomposition is always in sync with execution

**Cons:**
- Invasive change to action interface
- Every action must handle both modes
- Testing becomes more complex

## Decision Outcome

**Chosen: Option 1 (Composite Decomposition at Eval Time)**

This option provides the cleanest architectural separation. The plan generator produces primitive-only plans, and the executor only understands primitives. The complexity is localized to composite action implementations.

### Rationale

- Maintains single responsibility: composites know how to decompose, primitives know how to execute
- Plans become self-describing - no need to understand composite action semantics to read a plan
- Execution path is simple: iterate over primitives and execute each one
- Future ecosystem primitives follow the same pattern

## Solution Architecture

### Primitive Action Classification

Actions are classified into two tiers based on their decomposition barrier:

#### Tier 1: File Operation Primitives

These are fully atomic operations with deterministic, reproducible behavior:

| Primitive | Purpose | Key Parameters |
|-----------|---------|----------------|
| `download` | Fetch URL to file | `url`, `dest`, `checksum` |
| `extract` | Decompress archive | `archive`, `format`, `strip_dirs` |
| `chmod` | Set file permissions | `files`, `mode` |
| `install_binaries` | Copy to install dir, create symlinks | `binaries`, `install_mode` |
| `set_env` | Set environment variables | `vars` |
| `set_rpath` | Modify binary rpath | `binary`, `rpath` |
| `link_dependencies` | Create dependency symlinks | `dependencies` |
| `install_libraries` | Install shared libraries | `libraries` |

#### Tier 2: Ecosystem Primitives

These represent the **decomposition barrier** for ecosystem-specific operations. They are atomic from tsuku's perspective but internally invoke external tooling. The plan captures maximum constraint to minimize non-determinism.

| Primitive | Ecosystem | Locked at Eval | Residual Non-determinism |
|-----------|-----------|----------------|--------------------------|
| `go_build` | Go | go.sum, module versions | Compiler version, CGO |
| `cargo_build` | Rust | Cargo.lock | Compiler version |
| `npm_exec` | Node.js | package-lock.json | Native addon builds |
| `pip_install` | Python | requirements.txt with hashes | Native extensions |
| `gem_exec` | Ruby | Gemfile.lock | Native extensions |
| `nix_realize` | Nix | Derivation hash | None (fully deterministic) |

Each ecosystem primitive requires dedicated investigation to determine:
1. What can be locked at eval time
2. What reproducibility guarantees the ecosystem provides
3. Minimal invocation that respects locks
4. Residual non-determinism that must be accepted

### Composite Action Decomposition

Composite actions implement the `Decomposable` interface:

```go
// Decomposable indicates an action can be broken into primitive steps
type Decomposable interface {
    // Decompose returns the primitive steps this action expands to.
    // Called during plan generation, not execution.
    Decompose(ctx *EvalContext, params map[string]interface{}) ([]PrimitiveStep, error)
}

// PrimitiveStep represents a single atomic operation in a plan
type PrimitiveStep struct {
    Action   string                 // Primitive action name
    Params   map[string]interface{} // Fully resolved parameters
    Checksum string                 // For download actions: expected SHA256
    Size     int64                  // For download actions: expected size
}

// EvalContext provides context during decomposition
type EvalContext struct {
    Version    string
    VersionTag string
    OS         string
    Arch       string
    Recipe     *recipe.Recipe
    Resolver   *version.Resolver    // For API calls (asset resolution, etc.)
    Downloader *validate.PreDownloader // For checksum computation
}
```

#### Example: github_archive Decomposition

```go
func (a *GitHubArchiveAction) Decompose(ctx *EvalContext, params map[string]interface{}) ([]PrimitiveStep, error) {
    // 1. Resolve asset pattern (may involve API call)
    assetName := resolveAssetPattern(ctx, params)

    // 2. Construct download URL
    url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s",
        params["repo"], ctx.VersionTag, assetName)

    // 3. Download to compute checksum
    result, err := ctx.Downloader.Download(ctx.Context, url)
    if err != nil {
        return nil, err
    }
    defer result.Cleanup()

    // 4. Return primitive steps
    return []PrimitiveStep{
        {
            Action:   "download",
            Params:   map[string]interface{}{"url": url, "dest": assetName},
            Checksum: result.Checksum,
            Size:     result.Size,
        },
        {
            Action: "extract",
            Params: map[string]interface{}{
                "archive":    assetName,
                "format":     params["archive_format"],
                "strip_dirs": params["strip_dirs"],
            },
        },
        {
            Action: "chmod",
            Params: map[string]interface{}{"files": extractBinaries(params)},
        },
        {
            Action: "install_binaries",
            Params: map[string]interface{}{
                "binaries":     params["binaries"],
                "install_mode": params["install_mode"],
            },
        },
    }, nil
}
```

### Plan Structure

Plans contain only primitive operations:

```json
{
  "format_version": 2,
  "tool": "ripgrep",
  "version": "14.1.0",
  "platform": {"os": "linux", "arch": "amd64"},
  "generated_at": "2025-12-12T10:00:00Z",
  "recipe_hash": "sha256:abc...",
  "steps": [
    {
      "action": "download",
      "params": {
        "url": "https://github.com/BurntSushi/ripgrep/releases/download/14.1.0/ripgrep-14.1.0-x86_64-unknown-linux-musl.tar.gz",
        "dest": "ripgrep-14.1.0-x86_64-unknown-linux-musl.tar.gz"
      },
      "checksum": "sha256:1234567890abcdef...",
      "size": 2048576
    },
    {
      "action": "extract",
      "params": {
        "archive": "ripgrep-14.1.0-x86_64-unknown-linux-musl.tar.gz",
        "format": "tar.gz",
        "strip_dirs": 1
      }
    },
    {
      "action": "chmod",
      "params": {"files": ["rg"]}
    },
    {
      "action": "install_binaries",
      "params": {"binaries": ["rg"]}
    }
  ]
}
```

### Ecosystem Primitive Example: go_build

For ecosystem primitives, the plan captures the lock information:

```json
{
  "action": "go_build",
  "params": {
    "module": "github.com/jesseduffield/lazygit",
    "version": "v0.40.2",
    "executables": ["lazygit"]
  },
  "locks": {
    "go_version": "1.21.0",
    "go_sum": "h1:abc...=\ngithub.com/foo/bar v1.0.0 h1:xyz...=\n..."
  },
  "deterministic": false
}
```

The `deterministic: false` flag explicitly marks this step as having residual non-determinism.

### Unified Execution Flow

```
tsuku install foo
       │
       ▼
┌─────────────────┐
│  Load Recipe    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Generate Plan  │◄── Decompose composites
│  (eval phase)   │    Resolve versions
└────────┬────────┘    Compute checksums
         │             Lock ecosystem deps
         ▼
┌─────────────────┐
│  Installation   │◄── Plan (primitives only)
│     Plan        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Execute Plan   │◄── Iterate primitives
│  (exec phase)   │    Verify checksums
└────────┬────────┘    Fail on mismatch
         │
         ▼
    Installed Tool
```

Both `tsuku install foo` and `tsuku eval foo | tsuku install --plan -` follow this exact flow. The only difference is whether the plan is transient (piped) or persisted (file).

## Implementation Approach

### Phase 1: Decomposable Interface and File Primitives

1. Define `Decomposable` interface in `internal/actions/`
2. Update `ResolvedStep` to remove composite actions, only allow primitives
3. Implement `Decompose()` for:
   - `DownloadArchiveAction`
   - `GitHubArchiveAction`
   - `GitHubFileAction`
   - `HashiCorpReleaseAction`
4. Update plan generator to call `Decompose()` for composite actions
5. Executor validates plan contains only primitives

### Phase 2: Ecosystem Primitive Investigation

Launch parallel investigation for each ecosystem:
- Go modules: lock format, reproducibility guarantees
- Cargo: Cargo.lock integration, build reproducibility
- npm: package-lock.json, native addon handling
- pip: requirements.txt with hashes, wheel reproducibility
- gem: Gemfile.lock format
- nix: derivation hashing, store path capture

### Phase 3: Ecosystem Primitive Implementation

Based on investigation results, implement each ecosystem primitive:
1. Define lock capture at eval time
2. Implement locked execution
3. Document residual non-determinism
4. Add `deterministic` flag to plan schema

## Security Considerations

### Download Verification

- All `download` primitives in plans include checksums
- Executor verifies checksum before proceeding
- Mismatch is a hard failure (security feature)

### Execution Isolation

- Ecosystem primitives run in controlled environments
- Go: `CGO_ENABLED=0`, isolated `GOMODCACHE`
- npm: `--ignore-scripts` option for untrusted packages
- Each ecosystem primitive documents its isolation model

### Supply Chain Risks

- Plans capture checksums at eval time
- Re-evaluation detects upstream changes (modified releases)
- Ecosystem locks (go.sum, Cargo.lock) provide additional verification
- Residual risk: initial eval inherits any existing compromise

### User Data Exposure

- Plans may contain URLs that reveal tool preferences
- No credentials in plans (URLs are public download links)
- Ecosystem locks don't contain sensitive data

## Consequences

### Positive

- **Guaranteed determinism for decomposable recipes**: Plans with only Tier 1 primitives are fully reproducible
- **Explicit non-determinism**: Ecosystem primitives clearly mark residual non-determinism
- **Simpler executor**: Only understands primitives, no composite action logic
- **Better testability**: Can unit test decomposition separate from execution
- **Auditable plans**: Plans are self-describing, no hidden logic

### Negative

- **Increased eval time**: Decomposition requires API calls and downloads upfront
- **Migration effort**: All composite actions need `Decompose()` implementation
- **Plan size**: Decomposed plans are larger than composite plans
- **Ecosystem complexity**: Each ecosystem primitive requires dedicated investigation

### Neutral

- **Recipe syntax unchanged**: Authors still write `github_archive`, it decomposes transparently
- **Backward compatibility**: Old plans with composite actions can be re-evaluated

## Open Questions

1. **Plan format version**: Should decomposed plans use format version 2 to distinguish from current plans?

2. **Partial determinism**: Should plans have an overall `deterministic` flag, or is per-step sufficient?

3. **Ecosystem primitive naming**: `go_build` vs `go_install_locked` vs keeping `go_install` with enhanced behavior?

4. **Lock storage**: Should ecosystem locks be inline in the plan or referenced files?

## Appendix: Ecosystem Investigation Template

For each ecosystem, the investigation agent should answer:

### [Ecosystem Name]

**Lock mechanism**: What file/format captures the dependency graph?

**Eval-time capture**: What commands extract lock information?

**Locked execution**: What flags/env ensure the lock is respected?

**Reproducibility guarantees**: What does the ecosystem guarantee about builds?

**Residual non-determinism**: What can still vary between runs?

**Recommended primitive interface**:
```go
type [Ecosystem]BuildParams struct {
    // Fields
}
```

**Security considerations**: Specific risks for this ecosystem.
