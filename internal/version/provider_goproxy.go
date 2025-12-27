package version

import (
	"context"
	"fmt"
	"strings"
)

// GoProxyProvider resolves versions from proxy.golang.org for Go modules.
// Implements both VersionResolver and VersionLister interfaces.
//
// Go module versions follow the go.mod convention with "v" prefix:
// - Module versions: "v1.2.3" (with "v" prefix)
// - This is distinct from Go toolchain versions which have no prefix
type GoProxyProvider struct {
	resolver   *Resolver
	modulePath string
}

// NewGoProxyProvider creates a provider for Go module versions
func NewGoProxyProvider(resolver *Resolver, modulePath string) *GoProxyProvider {
	return &GoProxyProvider{
		resolver:   resolver,
		modulePath: modulePath,
	}
}

// ListVersions returns all available versions for the module (newest first)
func (p *GoProxyProvider) ListVersions(ctx context.Context) ([]string, error) {
	return p.resolver.ListGoProxyVersions(ctx, p.modulePath)
}

// ResolveLatest returns the latest version for the module
func (p *GoProxyProvider) ResolveLatest(ctx context.Context) (*VersionInfo, error) {
	return p.resolver.ResolveGoProxy(ctx, p.modulePath)
}

// ResolveVersion resolves a specific version for the module.
// Validates that the requested version exists.
// Accepts versions with or without "v" prefix.
func (p *GoProxyProvider) ResolveVersion(ctx context.Context, version string) (*VersionInfo, error) {
	// Normalize: ensure version has "v" prefix for comparison
	normalizedVersion := version
	if !strings.HasPrefix(version, "v") {
		normalizedVersion = "v" + version
	}

	// Validate that the requested version exists
	versions, err := p.ListVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions for %s: %w", p.modulePath, err)
	}

	// Check if exact version exists
	for _, v := range versions {
		if v == normalizedVersion {
			return &VersionInfo{
				Tag:     normalizedVersion,
				Version: strings.TrimPrefix(normalizedVersion, "v"),
			}, nil
		}
	}

	return nil, fmt.Errorf("version %s not found for module %s", version, p.modulePath)
}

// SourceDescription returns a human-readable source description
func (p *GoProxyProvider) SourceDescription() string {
	return "proxy.golang.org"
}

// InferredGoProxyProvider resolves versions for Go modules with automatic
// path fallback. When the install path (e.g., github.com/go-delve/delve/cmd/dlv)
// differs from the module path (e.g., github.com/go-delve/delve), it tries
// the full path first, then falls back to stripping /cmd/... suffixes.
type InferredGoProxyProvider struct {
	resolver     *Resolver
	installPath  string
	resolvedPath string // cached after first successful resolution
}

// NewInferredGoProxyProvider creates a provider that infers the correct module path
func NewInferredGoProxyProvider(resolver *Resolver, installPath string) *InferredGoProxyProvider {
	return &InferredGoProxyProvider{
		resolver:    resolver,
		installPath: installPath,
	}
}

// extractModulePaths returns candidate module paths to try, in order.
// It handles common Go patterns where the install path differs from the module path:
//   - /cmd/xxx pattern: "github.com/owner/repo/cmd/tool" → "github.com/owner/repo"
//   - submodule pattern: "go.uber.org/mock/mockgen" → "go.uber.org/mock"
func extractModulePaths(installPath string) []string {
	paths := []string{installPath}

	// If path contains /cmd/, try stripping it and everything after
	if idx := strings.Index(installPath, "/cmd/"); idx != -1 {
		paths = append(paths, installPath[:idx])
	}

	// Also try stripping the last path component (for submodules like mockgen)
	if lastSlash := strings.LastIndex(installPath, "/"); lastSlash > 0 {
		parent := installPath[:lastSlash]
		// Avoid adding duplicates or paths that are too short (domain only)
		// A valid Go module path has at least domain/module, so >= 1 slash
		if parent != paths[len(paths)-1] && strings.Contains(parent, "/") {
			paths = append(paths, parent)
		}
	}

	return paths
}

// resolveModulePath determines which module path works for version resolution
func (p *InferredGoProxyProvider) resolveModulePath(ctx context.Context) (string, error) {
	if p.resolvedPath != "" {
		return p.resolvedPath, nil
	}

	candidates := extractModulePaths(p.installPath)
	var lastErr error

	for _, path := range candidates {
		versions, err := p.resolver.ListGoProxyVersions(ctx, path)
		if err == nil && len(versions) > 0 {
			p.resolvedPath = path
			return path, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed to resolve module path for %s: %w", p.installPath, lastErr)
	}
	return "", fmt.Errorf("no versions found for module %s", p.installPath)
}

// ListVersions returns all available versions for the module (newest first)
func (p *InferredGoProxyProvider) ListVersions(ctx context.Context) ([]string, error) {
	modulePath, err := p.resolveModulePath(ctx)
	if err != nil {
		return nil, err
	}
	return p.resolver.ListGoProxyVersions(ctx, modulePath)
}

// ResolveLatest returns the latest version for the module
func (p *InferredGoProxyProvider) ResolveLatest(ctx context.Context) (*VersionInfo, error) {
	modulePath, err := p.resolveModulePath(ctx)
	if err != nil {
		return nil, err
	}
	return p.resolver.ResolveGoProxy(ctx, modulePath)
}

// ResolveVersion resolves a specific version for the module
func (p *InferredGoProxyProvider) ResolveVersion(ctx context.Context, version string) (*VersionInfo, error) {
	modulePath, err := p.resolveModulePath(ctx)
	if err != nil {
		return nil, err
	}

	// Normalize: ensure version has "v" prefix for comparison
	normalizedVersion := version
	if !strings.HasPrefix(version, "v") {
		normalizedVersion = "v" + version
	}

	versions, err := p.resolver.ListGoProxyVersions(ctx, modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions for %s: %w", modulePath, err)
	}

	for _, v := range versions {
		if v == normalizedVersion {
			return &VersionInfo{
				Tag:     normalizedVersion,
				Version: strings.TrimPrefix(normalizedVersion, "v"),
			}, nil
		}
	}

	return nil, fmt.Errorf("version %s not found for module %s", version, modulePath)
}

// SourceDescription returns a human-readable source description
func (p *InferredGoProxyProvider) SourceDescription() string {
	return "proxy.golang.org (inferred)"
}
