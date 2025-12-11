package validate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tsukumogami/tsuku/internal/recipe"
)

// DefaultValidationImage is the container image used for validation.
// Using Debian because the tsuku binary is dynamically linked against glibc.
const DefaultValidationImage = "debian:bookworm-slim"

// ValidationResult contains the result of a recipe validation.
type ValidationResult struct {
	Passed   bool   // Whether validation succeeded
	Skipped  bool   // Whether validation was skipped (no runtime)
	ExitCode int    // Container exit code
	Stdout   string // Container stdout
	Stderr   string // Container stderr
	Error    error  // Error if validation failed to run
}

// ExecutorLogger defines the logging interface for executor warnings.
type ExecutorLogger interface {
	Warn(msg string, args ...any)
	Debug(msg string, args ...any)
}

// noopExecutorLogger is a logger that discards all messages.
type noopExecutorLogger struct{}

func (noopExecutorLogger) Warn(string, ...any)  {}
func (noopExecutorLogger) Debug(string, ...any) {}

// Executor orchestrates container-based recipe validation.
// It combines runtime detection, asset pre-download, and isolated container execution.
type Executor struct {
	detector      *RuntimeDetector
	predownloader *PreDownloader
	logger        ExecutorLogger
	image         string
	limits        ResourceLimits
	tsukuBinary   string // Path to tsuku binary for container execution
}

// ExecutorOption configures an Executor.
type ExecutorOption func(*Executor)

// WithExecutorLogger sets a logger for executor warnings.
func WithExecutorLogger(logger ExecutorLogger) ExecutorOption {
	return func(e *Executor) {
		e.logger = logger
	}
}

// WithValidationImage sets the container image for validation.
func WithValidationImage(image string) ExecutorOption {
	return func(e *Executor) {
		e.image = image
	}
}

// WithResourceLimits sets resource limits for validation containers.
func WithResourceLimits(limits ResourceLimits) ExecutorOption {
	return func(e *Executor) {
		e.limits = limits
	}
}

// WithTsukuBinary sets the path to the tsuku binary for container execution.
func WithTsukuBinary(path string) ExecutorOption {
	return func(e *Executor) {
		e.tsukuBinary = path
	}
}

// NewExecutor creates a new Executor with the given dependencies.
func NewExecutor(detector *RuntimeDetector, predownloader *PreDownloader, opts ...ExecutorOption) *Executor {
	// Auto-detect tsuku binary path
	tsukuPath, _ := os.Executable()

	e := &Executor{
		detector:      detector,
		predownloader: predownloader,
		logger:        noopExecutorLogger{},
		image:         DefaultValidationImage,
		tsukuBinary:   tsukuPath,
		limits: ResourceLimits{
			Memory:   "2g",
			CPUs:     "2",
			PidsMax:  100,
			ReadOnly: true,
		},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Validate runs a recipe in an isolated container and checks the verification command.
// It returns a ValidationResult indicating whether the recipe installed correctly.
//
// The validation process:
// 1. Detect available container runtime
// 2. Serialize recipe to TOML file
// 3. Mount tsuku binary and recipe into container
// 4. Run tsuku install in isolated container
// 5. Run verification command
// 6. Check output against expected pattern
func (e *Executor) Validate(ctx context.Context, r *recipe.Recipe, assetURL string) (*ValidationResult, error) {
	// Detect container runtime
	runtime, err := e.detector.Detect(ctx)
	if err != nil {
		if err == ErrNoRuntime {
			e.logger.Warn("Container runtime not available. Skipping recipe validation.",
				"hint", "To enable validation, install Podman or Docker.")
			return &ValidationResult{
				Skipped: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to detect container runtime: %w", err)
	}

	// Emit security warning for Docker with group membership (non-rootless)
	if runtime.Name() == "docker" && !runtime.IsRootless() {
		e.logger.Warn("Using Docker with docker group membership.",
			"security", "This grants root-equivalent access on this machine.",
			"recommendation", "Consider configuring Docker rootless mode for better security.",
			"docs", "https://docs.docker.com/engine/security/rootless/")
	}

	e.logger.Debug("Using container runtime", "runtime", runtime.Name(), "rootless", runtime.IsRootless())

	// Create workspace directory
	workspaceDir, err := os.MkdirTemp("", TempDirPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}
	defer os.RemoveAll(workspaceDir)

	// Serialize recipe to TOML using custom method
	recipeData, err := r.ToTOML()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize recipe: %w", err)
	}

	// Write recipe to workspace
	recipePath := filepath.Join(workspaceDir, "recipe.toml")
	if err := os.WriteFile(recipePath, recipeData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write recipe file: %w", err)
	}

	// Build the validation script that runs tsuku install
	script := e.buildTsukuInstallScript(r)

	// Create the install script in workspace
	scriptPath := filepath.Join(workspaceDir, "validate.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return nil, fmt.Errorf("failed to write validation script: %w", err)
	}

	// Build run options
	// Override ReadOnly to false since we need to install packages
	limits := e.limits
	limits.ReadOnly = false

	opts := RunOptions{
		Image:   e.image,
		Command: []string{"/bin/sh", "/workspace/validate.sh"},
		Network: "host", // Need network for downloads
		WorkDir: "/workspace",
		Env: []string{
			"TSUKU_VALIDATION=1",
			"TSUKU_HOME=/workspace/tsuku",
			"HOME=/workspace",
		},
		Limits: limits,
		Labels: map[string]string{
			ContainerLabelPrefix: "true",
		},
		Mounts: []Mount{
			{
				Source:   workspaceDir,
				Target:   "/workspace",
				ReadOnly: false,
			},
		},
	}

	// Mount tsuku binary if available
	if e.tsukuBinary != "" {
		opts.Mounts = append(opts.Mounts, Mount{
			Source:   e.tsukuBinary,
			Target:   "/usr/local/bin/tsuku",
			ReadOnly: true,
		})
	}

	// Run the container
	result, err := runtime.Run(ctx, opts)
	if err != nil {
		return &ValidationResult{
			Passed:   false,
			ExitCode: -1,
			Stdout:   result.Stdout,
			Stderr:   result.Stderr,
			Error:    err,
		}, nil
	}

	// Check if verification passed
	passed := e.checkVerification(r, result)

	return &ValidationResult{
		Passed:   passed,
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
	}, nil
}

// buildTsukuInstallScript creates a shell script that runs tsuku install with the recipe.
func (e *Executor) buildTsukuInstallScript(r *recipe.Recipe) string {
	var sb strings.Builder

	sb.WriteString("#!/bin/sh\n")
	sb.WriteString("set -e\n\n")

	// Install ca-certificates for HTTPS downloads
	sb.WriteString("# Install required packages\n")
	sb.WriteString("apt-get update -qq && apt-get install -qq -y ca-certificates >/dev/null 2>&1 || true\n\n")

	// Setup tsuku home directory
	sb.WriteString("# Setup TSUKU_HOME\n")
	sb.WriteString("mkdir -p /workspace/tsuku/recipes\n")
	sb.WriteString("mkdir -p /workspace/tsuku/bin\n")
	sb.WriteString("mkdir -p /workspace/tsuku/tools\n\n")

	// Copy recipe to tsuku recipes directory
	sb.WriteString("# Copy recipe to tsuku recipes\n")
	sb.WriteString(fmt.Sprintf("cp /workspace/recipe.toml /workspace/tsuku/recipes/%s.toml\n\n", r.Metadata.Name))

	// Run tsuku install (which includes verification)
	sb.WriteString("# Run tsuku install\n")
	sb.WriteString(fmt.Sprintf("tsuku install %s --force\n", r.Metadata.Name))

	return sb.String()
}

// buildValidationScript creates a shell script that runs recipe steps and verification.
// DEPRECATED: Use buildTsukuInstallScript instead for proper recipe execution.
func (e *Executor) buildValidationScript(r *recipe.Recipe) string {
	var sb strings.Builder

	sb.WriteString("#!/bin/sh\n")
	sb.WriteString("set -e\n\n")

	// Add environment setup
	sb.WriteString("# Setup environment\n")
	sb.WriteString("export PATH=\"/workspace:/workspace/bin:$PATH\"\n")
	sb.WriteString("mkdir -p /workspace/bin\n\n")

	// Handle assets - copy, extract archives, and find binaries
	sb.WriteString("# Process assets\n")
	sb.WriteString("if [ -d /assets ]; then\n")
	sb.WriteString("  for asset in /assets/*; do\n")
	sb.WriteString("    case \"$asset\" in\n")
	// Handle tar.gz archives
	sb.WriteString("      *.tar.gz|*.tgz)\n")
	sb.WriteString("        tar -xzf \"$asset\" -C /workspace 2>/dev/null || true\n")
	sb.WriteString("        ;;\n")
	// Handle tar.xz archives
	sb.WriteString("      *.tar.xz)\n")
	sb.WriteString("        tar -xJf \"$asset\" -C /workspace 2>/dev/null || true\n")
	sb.WriteString("        ;;\n")
	// Handle tar.bz2 archives
	sb.WriteString("      *.tar.bz2)\n")
	sb.WriteString("        tar -xjf \"$asset\" -C /workspace 2>/dev/null || true\n")
	sb.WriteString("        ;;\n")
	// Handle plain tar archives
	sb.WriteString("      *.tar)\n")
	sb.WriteString("        tar -xf \"$asset\" -C /workspace 2>/dev/null || true\n")
	sb.WriteString("        ;;\n")
	// Handle zip archives
	sb.WriteString("      *.zip)\n")
	sb.WriteString("        unzip -q -o \"$asset\" -d /workspace 2>/dev/null || true\n")
	sb.WriteString("        ;;\n")
	// Handle bare binaries (copy and make executable)
	sb.WriteString("      *)\n")
	sb.WriteString("        cp \"$asset\" /workspace/ 2>/dev/null || true\n")
	sb.WriteString("        ;;\n")
	sb.WriteString("    esac\n")
	sb.WriteString("  done\n")
	sb.WriteString("fi\n\n")

	// Make all files in workspace and subdirectories executable
	sb.WriteString("# Make binaries executable\n")
	sb.WriteString("find /workspace -type f -exec chmod +x {} \\; 2>/dev/null || true\n\n")

	// Add workspace subdirectories to PATH (archives often have nested structure)
	sb.WriteString("# Add extracted directories to PATH\n")
	sb.WriteString("for dir in /workspace/*/bin /workspace/*/; do\n")
	sb.WriteString("  [ -d \"$dir\" ] && export PATH=\"$dir:$PATH\"\n")
	sb.WriteString("done\n\n")

	// Create symlinks for binaries that might have platform suffixes
	// Common patterns: tool_linux_amd64, tool-linux-amd64, etc.
	sb.WriteString("# Create symlinks for platform-suffixed binaries\n")
	sb.WriteString("cd /workspace\n")
	sb.WriteString("for f in *_linux_amd64 *_linux_arm64 *-linux-amd64 *-linux-arm64; do\n")
	sb.WriteString("  [ -f \"$f\" ] || continue\n")
	sb.WriteString("  # Extract base name (remove platform suffix)\n")
	sb.WriteString("  base=$(echo \"$f\" | sed -E 's/[_-]linux[_-](amd64|arm64|x86_64|aarch64)$//')\n")
	sb.WriteString("  [ -f \"$base\" ] || ln -sf \"$f\" \"$base\" 2>/dev/null || true\n")
	sb.WriteString("done\n")
	sb.WriteString("cd /\n\n")

	// Run verification command
	sb.WriteString("# Run verification\n")
	if r.Verify.Command != "" {
		// Handle the verification command
		// The command may reference the binary directly
		sb.WriteString(fmt.Sprintf("%s\n", r.Verify.Command))
	} else {
		sb.WriteString("echo 'No verification command specified'\n")
	}

	return sb.String()
}

// checkVerification checks if the verification output matches expectations.
func (e *Executor) checkVerification(r *recipe.Recipe, result *RunResult) bool {
	// If exit code is non-zero, verification failed
	expectedExitCode := 0
	if r.Verify.ExitCode != nil {
		expectedExitCode = *r.Verify.ExitCode
	}
	if result.ExitCode != expectedExitCode {
		return false
	}

	// If no pattern specified, just check exit code
	if r.Verify.Pattern == "" {
		return true
	}

	// Check if pattern appears in stdout or stderr
	output := result.Stdout + result.Stderr
	return strings.Contains(output, r.Verify.Pattern)
}

// GetAssetChecksum returns the SHA256 checksum of a downloaded asset.
// This is useful for embedding checksums in generated recipes.
func (e *Executor) GetAssetChecksum(ctx context.Context, url string) (string, error) {
	result, err := e.predownloader.Download(ctx, url)
	if err != nil {
		return "", err
	}
	defer func() { _ = result.Cleanup() }()
	return result.Checksum, nil
}
