package actions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PipExecAction is the primitive execution action for Python packages.
// It installs packages from a locked requirements.txt with hash verification
// into an isolated virtual environment.
//
// This is an ecosystem primitive that achieves determinism through
// lockfile enforcement and reproducible build configuration.
type PipExecAction struct{ BaseAction }

// Dependencies returns python-standalone as both install-time and runtime dependency.
func (PipExecAction) Dependencies() ActionDeps {
	return ActionDeps{InstallTime: []string{"python-standalone"}, Runtime: []string{"python-standalone"}}
}

// RequiresNetwork returns true because pip needs to download packages from PyPI.
func (PipExecAction) RequiresNetwork() bool {
	return true
}

// Name returns the action name
func (a *PipExecAction) Name() string {
	return "pip_exec"
}

// IsDeterministic returns false because pip installation has residual non-determinism.
// While hash verification ensures identical packages, bytecode generation and
// platform-specific wheel selection introduce variance.
func (a *PipExecAction) IsDeterministic() bool {
	return false
}

// Execute installs Python packages from locked requirements into an isolated venv.
//
// Parameters:
//   - package (required): Primary package name (informational)
//   - version (required): Primary package version (informational)
//   - executables (required): List of executable names to verify and symlink
//   - locked_requirements (required): Full requirements.txt content with hashes
//   - python_version (optional): Expected Python version for validation
//   - has_native_addons (optional): Whether package includes native extensions
//
// Installation process:
//  1. Verify Python interpreter version (if python_version specified)
//  2. Create isolated venv in installDir/venvs/<package>/
//  3. Write locked_requirements to venv/requirements.txt
//  4. Install with pip --require-hashes --no-deps --only-binary :all:
//  5. Verify executables exist
//  6. Create symlinks in installDir/bin/
func (a *PipExecAction) Execute(ctx *ExecutionContext, params map[string]interface{}) error {
	// Get package name (required)
	packageName, ok := GetString(params, "package")
	if !ok {
		return fmt.Errorf("pip_exec action requires 'package' parameter")
	}

	// Get version (required for info)
	version, _ := GetString(params, "version")
	if version == "" {
		version = ctx.Version
	}

	// Get executables list (required)
	executables, ok := GetStringSlice(params, "executables")
	if !ok || len(executables) == 0 {
		return fmt.Errorf("pip_exec action requires 'executables' parameter with at least one executable")
	}

	// Get locked requirements (required)
	lockedRequirements, ok := GetString(params, "locked_requirements")
	if !ok || lockedRequirements == "" {
		return fmt.Errorf("pip_exec action requires 'locked_requirements' parameter")
	}

	// Get optional parameters
	expectedPythonVersion, _ := GetString(params, "python_version")
	hasNativeAddons, _ := GetBool(params, "has_native_addons")

	// Find Python interpreter from python-standalone installation
	pythonPath := ResolvePythonStandalone()
	if pythonPath == "" {
		// Check ExecPaths from dependencies
		for _, p := range ctx.ExecPaths {
			candidatePath := filepath.Join(p, "python3")
			if _, err := os.Stat(candidatePath); err == nil {
				pythonPath = candidatePath
				break
			}
		}
	}
	if pythonPath == "" {
		return fmt.Errorf("python not found: install python-standalone first (tsuku install python-standalone)")
	}

	fmt.Printf("   Package: %s@%s\n", packageName, version)
	fmt.Printf("   Executables: %v\n", executables)
	fmt.Printf("   Using python: %s\n", pythonPath)
	if hasNativeAddons {
		fmt.Printf("   Warning: Package contains native addons (may have platform-specific behavior)\n")
	}

	// Step 1: Verify Python version if specified
	if expectedPythonVersion != "" {
		actualVersion, err := getPythonVersion(pythonPath)
		if err != nil {
			return fmt.Errorf("failed to get Python version: %w", err)
		}
		if !strings.HasPrefix(actualVersion, expectedPythonVersion) {
			fmt.Printf("   Warning: Python version mismatch - expected %s, got %s\n",
				expectedPythonVersion, actualVersion)
		}
	}

	// Step 2: Create isolated venv
	venvDir := filepath.Join(ctx.InstallDir, "venvs", packageName)
	fmt.Printf("   Creating venv: %s\n", venvDir)

	if err := os.MkdirAll(filepath.Dir(venvDir), 0755); err != nil {
		return fmt.Errorf("failed to create venvs directory: %w", err)
	}

	venvCmd := exec.CommandContext(ctx.Context, pythonPath, "-m", "venv", venvDir)
	if output, err := venvCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create venv: %w\nOutput: %s", err, string(output))
	}

	// Step 3: Write locked requirements
	reqFile := filepath.Join(venvDir, "requirements.txt")
	if err := os.WriteFile(reqFile, []byte(lockedRequirements), 0644); err != nil {
		return fmt.Errorf("failed to write requirements.txt: %w", err)
	}

	// Count packages for progress reporting
	packageCount := countRequirementsPackages(lockedRequirements)
	fmt.Printf("   Installing %d packages with hash verification\n", packageCount)

	// Step 4: Install with safety flags
	pipBin := filepath.Join(venvDir, "bin", "pip")

	// Build pip install command with security flags
	pipArgs := []string{
		"install",
		"--require-hashes",
		"--no-deps",
		"--only-binary", ":all:",
		"--disable-pip-version-check",
		"-r", reqFile,
	}

	pipCmd := exec.CommandContext(ctx.Context, pipBin, pipArgs...)
	pipCmd.Dir = venvDir

	// Set up environment - filter out PIP_USER which conflicts with venv installs
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "PIP_USER=") {
			env = append(env, e)
		}
	}
	// Set PYTHONHASHSEED for deterministic bytecode
	env = append(env, "PYTHONHASHSEED=0")
	pipCmd.Env = env

	output, err := pipCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip install failed: %w\nOutput: %s", err, string(output))
	}

	// Step 5: Verify executables exist in venv
	venvBinDir := filepath.Join(venvDir, "bin")
	for _, exe := range executables {
		exePath := filepath.Join(venvBinDir, exe)
		if _, err := os.Stat(exePath); err != nil {
			return fmt.Errorf("expected executable %s not found at %s", exe, exePath)
		}
	}

	// Step 6: Fix python3 symlink in venv to use relative path to python-standalone
	// This matches what pipx_install does and ensures the venv's python points to the right interpreter
	python3Link := filepath.Join(venvBinDir, "python3")
	if target, err := os.Readlink(python3Link); err == nil && filepath.IsAbs(target) {
		// Remove the absolute symlink
		os.Remove(python3Link)
		// Create relative symlink to tsuku's python-standalone
		// From: venvs/<package>/bin/python3
		// To: ../../../python-standalone-XXXXXXXX/bin/python3
		if pythonPath != "" {
			relPath, err := filepath.Rel(venvBinDir, pythonPath)
			if err == nil {
				_ = os.Symlink(relPath, python3Link) // Ignore error if symlink fails
			}
		}
	}

	// Step 6b: Fix shebangs in entry point scripts
	// Entry point scripts have shebangs with absolute paths to the venv's python,
	// which become invalid after the executor moves the directory.
	// Rewrite them to use ./python3 (relative to script location).
	for _, exe := range executables {
		exePath := filepath.Join(venvBinDir, exe)
		if err := fixPythonShebang(exePath); err != nil {
			// Log warning but don't fail
			fmt.Printf("   Warning: failed to fix shebang in %s: %v\n", exe, err)
		}
	}

	// Step 7: Create symlinks in bin/ directory (where executor expects them)
	binDir := filepath.Join(ctx.InstallDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	for _, exe := range executables {
		// Create relative symlink: bin/<exe> -> ../venvs/<package>/bin/<exe>
		srcPath := filepath.Join("..", "venvs", packageName, "bin", exe)
		dstPath := filepath.Join(binDir, exe)

		// Remove existing symlink if present
		os.Remove(dstPath)

		if err := os.Symlink(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", exe, err)
		}
	}

	fmt.Printf("   Package installed successfully\n")
	fmt.Printf("   Verified %d executable(s)\n", len(executables))

	return nil
}

// fixPythonShebang rewrites the shebang in a Python entry point script to use ./python3.
// This fixes shebangs that have absolute paths to the temporary staging directory.
func fixPythonShebang(scriptPath string) error {
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	// Check if it has a Python shebang
	if len(content) < 2 || content[0] != '#' || content[1] != '!' {
		return nil // Not a script
	}

	// Find first newline
	newlineIdx := strings.IndexByte(string(content), '\n')
	if newlineIdx == -1 {
		return nil // No newline found
	}

	shebang := string(content[:newlineIdx])
	rest := content[newlineIdx:]

	// Only fix if it's a Python shebang
	if !strings.Contains(shebang, "python") {
		return nil
	}

	// Replace with a shebang that uses ./python3 (same directory as script)
	// This works because both the script and python3 are in venvs/<package>/bin/
	newShebang := "#!/bin/sh\nexec \"$(dirname \"$0\")/python3\" \"$0\" \"$@\""

	// Write back
	newContent := []byte(newShebang + string(rest))
	return os.WriteFile(scriptPath, newContent, 0755)
}

// getPythonVersion returns the Python version string (e.g., "3.11.7")
func getPythonVersion(pythonPath string) (string, error) {
	cmd := exec.Command(pythonPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Output is "Python 3.11.7" - extract version
	versionStr := strings.TrimSpace(string(output))
	parts := strings.SplitN(versionStr, " ", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected Python version output: %s", versionStr)
	}
	return parts[1], nil
}

// countRequirementsPackages counts the number of packages in requirements.txt
func countRequirementsPackages(requirements string) int {
	count := 0
	for _, line := range strings.Split(requirements, "\n") {
		line = strings.TrimSpace(line)
		// Skip empty lines, comments, and continuation lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "\\") {
			continue
		}
		// Count lines that look like package specs (contain == or @ for URL)
		if strings.Contains(line, "==") || strings.Contains(line, " @ ") {
			count++
		}
	}
	return count
}
