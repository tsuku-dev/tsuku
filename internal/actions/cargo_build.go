package actions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CargoBuildAction builds Rust crates with deterministic configuration.
// This is an ecosystem primitive that cannot be decomposed further.
// It achieves determinism through cargo's --locked flag and environment variables.
type CargoBuildAction struct{}

// Name returns the action name
func (a *CargoBuildAction) Name() string {
	return "cargo_build"
}

// Execute builds a Rust crate with deterministic configuration
//
// Parameters:
//   - source_dir (required): Directory containing Cargo.toml
//   - executables (required): List of executable names to verify and install
//   - target (optional): Build target triple (defaults to host)
//   - features (optional): Cargo features to enable
//   - locked (optional): Use Cargo.lock for reproducibility (default: true)
//   - output_binary (optional): Expected output binary path (relative to target dir)
//
// Deterministic Configuration:
//   - SOURCE_DATE_EPOCH: Set to recipe timestamp for reproducible embedded timestamps
//   - CARGO_INCREMENTAL=0: Disable incremental compilation for deterministic builds
//   - RUSTFLAGS="-C embed-bitcode=no": Smaller, more reproducible builds
//   - --locked: Require Cargo.lock to exist and be up-to-date
//
// Directory Structure Created:
//
//	<install_dir>/
//	  bin/<executable>     - Compiled binary
func (a *CargoBuildAction) Execute(ctx *ExecutionContext, params map[string]interface{}) error {
	// Get source directory (required)
	sourceDir, ok := GetString(params, "source_dir")
	if !ok {
		return fmt.Errorf("cargo_build action requires 'source_dir' parameter")
	}

	// Resolve source directory relative to work directory if not absolute
	if !filepath.IsAbs(sourceDir) {
		sourceDir = filepath.Join(ctx.WorkDir, sourceDir)
	}

	// Verify Cargo.toml exists
	cargoToml := filepath.Join(sourceDir, "Cargo.toml")
	if _, err := os.Stat(cargoToml); err != nil {
		return fmt.Errorf("Cargo.toml not found at %s: %w", cargoToml, err)
	}

	// Get executables list (required)
	executables, ok := GetStringSlice(params, "executables")
	if !ok || len(executables) == 0 {
		return fmt.Errorf("cargo_build action requires 'executables' parameter with at least one executable")
	}

	// SECURITY: Validate executable names to prevent path traversal
	for _, exe := range executables {
		if strings.Contains(exe, "/") || strings.Contains(exe, "\\") ||
			strings.Contains(exe, "..") || exe == "." || exe == "" {
			return fmt.Errorf("invalid executable name '%s': must not contain path separators", exe)
		}
	}

	// Get optional parameters
	target, _ := GetString(params, "target")
	features, _ := GetStringSlice(params, "features")
	locked := true // Default to locked builds
	if lockedVal, ok := params["locked"]; ok {
		if b, ok := lockedVal.(bool); ok {
			locked = b
		}
	}

	// Get cargo path
	cargoPath, _ := GetString(params, "cargo_path")
	if cargoPath == "" {
		cargoPath = ResolveCargo()
		if cargoPath == "" {
			cargoPath = "cargo"
		}
	}

	fmt.Printf("   Source: %s\n", sourceDir)
	fmt.Printf("   Executables: %v\n", executables)
	if target != "" {
		fmt.Printf("   Target: %s\n", target)
	}
	if len(features) > 0 {
		fmt.Printf("   Features: %v\n", features)
	}
	fmt.Printf("   Locked: %v\n", locked)
	fmt.Printf("   Using cargo: %s\n", cargoPath)

	// Build arguments
	args := []string{"build", "--release"}

	if locked {
		// Verify Cargo.lock exists when locked build is requested
		cargoLock := filepath.Join(sourceDir, "Cargo.lock")
		if _, err := os.Stat(cargoLock); err != nil {
			return fmt.Errorf("locked build requested but Cargo.lock not found in %s", sourceDir)
		}
		args = append(args, "--locked")
	}

	if target != "" {
		// SECURITY: Validate target triple
		if !isValidTargetTriple(target) {
			return fmt.Errorf("invalid target triple '%s'", target)
		}
		args = append(args, "--target", target)
	}

	for _, feature := range features {
		// SECURITY: Validate feature name
		if !isValidFeatureName(feature) {
			return fmt.Errorf("invalid feature name '%s'", feature)
		}
		args = append(args, "--features", feature)
	}

	fmt.Printf("   Building: cargo %s\n", strings.Join(args, " "))

	// Set up deterministic environment with isolated CARGO_HOME
	env := buildDeterministicCargoEnv(cargoPath, ctx.WorkDir)

	// Create bin directory in install dir
	binDir := filepath.Join(ctx.InstallDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Execute cargo build
	cmd := exec.CommandContext(ctx.Context, cargoPath, args...)
	cmd.Dir = sourceDir
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cargo build failed: %w\nOutput: %s", err, string(output))
	}

	// Show output if debugging
	outputStr := strings.TrimSpace(string(output))
	if outputStr != "" && os.Getenv("TSUKU_DEBUG") != "" {
		fmt.Printf("   cargo output:\n%s\n", outputStr)
	}

	// Determine target directory for built binaries
	targetDir := filepath.Join(sourceDir, "target")
	if target != "" {
		targetDir = filepath.Join(targetDir, target)
	}
	releaseDir := filepath.Join(targetDir, "release")

	// Copy executables to install directory
	for _, exe := range executables {
		srcPath := filepath.Join(releaseDir, exe)
		dstPath := filepath.Join(binDir, exe)

		// Check if executable exists
		if _, err := os.Stat(srcPath); err != nil {
			return fmt.Errorf("expected executable %s not found at %s", exe, srcPath)
		}

		// Copy the executable
		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy executable %s: %w", exe, err)
		}

		// Make executable
		if err := os.Chmod(dstPath, 0755); err != nil {
			return fmt.Errorf("failed to chmod executable %s: %w", exe, err)
		}
	}

	fmt.Printf("   Crate built successfully\n")
	fmt.Printf("   Installed %d executable(s)\n", len(executables))

	return nil
}

// buildDeterministicCargoEnv creates an environment with deterministic build settings.
// workDir is used to create an isolated CARGO_HOME for reproducible builds.
func buildDeterministicCargoEnv(cargoPath, workDir string) []string {
	cargoDir := filepath.Dir(cargoPath)
	env := os.Environ()

	// Filter existing variables that might affect determinism
	filteredEnv := make([]string, 0, len(env))
	for _, e := range env {
		// Keep most variables but filter some that could cause non-determinism
		if !strings.HasPrefix(e, "CARGO_INCREMENTAL=") &&
			!strings.HasPrefix(e, "SOURCE_DATE_EPOCH=") &&
			!strings.HasPrefix(e, "CARGO_HOME=") {
			filteredEnv = append(filteredEnv, e)
		}
	}

	// Add cargo's bin directory to PATH
	pathUpdated := false
	for i, e := range filteredEnv {
		if strings.HasPrefix(e, "PATH=") {
			filteredEnv[i] = fmt.Sprintf("PATH=%s:%s", cargoDir, e[5:])
			pathUpdated = true
			break
		}
	}
	if !pathUpdated {
		filteredEnv = append(filteredEnv, fmt.Sprintf("PATH=%s:%s", cargoDir, os.Getenv("PATH")))
	}

	// Set isolated CARGO_HOME to prevent cross-contamination between builds
	cargoHome := filepath.Join(workDir, ".cargo-home")
	filteredEnv = append(filteredEnv, "CARGO_HOME="+cargoHome)

	// Set deterministic build environment variables
	filteredEnv = append(filteredEnv,
		// Disable incremental compilation for deterministic builds
		"CARGO_INCREMENTAL=0",
		// Set SOURCE_DATE_EPOCH to Unix epoch (0) for reproducible timestamps
		"SOURCE_DATE_EPOCH=0",
	)

	// Set RUSTFLAGS for more reproducible builds
	// -C embed-bitcode=no: Don't embed LLVM bitcode (smaller, more reproducible)
	existingRustflags := ""
	for _, e := range filteredEnv {
		if strings.HasPrefix(e, "RUSTFLAGS=") {
			existingRustflags = e[10:]
			break
		}
	}

	rustflags := "-C embed-bitcode=no"
	if existingRustflags != "" {
		rustflags = existingRustflags + " " + rustflags
	}

	// Update or add RUSTFLAGS
	rustflagsSet := false
	for i, e := range filteredEnv {
		if strings.HasPrefix(e, "RUSTFLAGS=") {
			filteredEnv[i] = "RUSTFLAGS=" + rustflags
			rustflagsSet = true
			break
		}
	}
	if !rustflagsSet {
		filteredEnv = append(filteredEnv, "RUSTFLAGS="+rustflags)
	}

	// Set up C compiler for crates with native dependencies
	if !hasSystemCompiler() {
		if newEnv, found := SetupCCompilerEnv(filteredEnv); found {
			filteredEnv = newEnv
		}
	}

	return filteredEnv
}

// isValidTargetTriple validates Rust target triples
// Format: <arch><sub>-<vendor>-<sys>-<abi>
// Examples: x86_64-unknown-linux-gnu, aarch64-apple-darwin
func isValidTargetTriple(target string) bool {
	if target == "" || len(target) > 100 {
		return false
	}

	// Must contain at least two hyphens (arch-vendor-sys or arch-vendor-sys-abi)
	parts := strings.Split(target, "-")
	if len(parts) < 3 {
		return false
	}

	// Check allowed characters: alphanumeric and hyphens
	for _, c := range target {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}

	return true
}

// isValidFeatureName validates Cargo feature names
// Features can be alphanumeric with hyphens, underscores, and slashes (for namespaced features)
func isValidFeatureName(feature string) bool {
	if feature == "" || len(feature) > 100 {
		return false
	}

	for _, c := range feature {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '/') {
			return false
		}
	}

	return true
}
