package actions

import (
	"runtime"
	"strings"
	"testing"
)

func TestNixRealizeAction_Name(t *testing.T) {
	action := &NixRealizeAction{}
	if action.Name() != "nix_realize" {
		t.Errorf("Name() = %q, want %q", action.Name(), "nix_realize")
	}
}

func TestIsValidFlakeRef(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected bool
	}{
		// Valid flake references
		{"nixpkgs package", "nixpkgs#hello", true},
		{"github flake", "github:user/repo#package", true},
		{"github with rev", "github:user/repo/abc123#pkg", true},
		{"path flake", "path:/some/path#attr", true},
		{"nested attribute", "nixpkgs#python3Packages.pytorch", true},
		{"with hyphen", "nixpkgs#cargo-audit", true},
		{"with underscore", "nixpkgs#my_package", true},
		{"with at sign", "github:user/repo@v1.0.0#pkg", true},

		// Invalid - no hash separator
		{"no hash", "nixpkgs", false},
		{"no hash github", "github:user/repo", false},

		// Invalid - empty or too long
		{"empty", "", false},
		{"too long", string(make([]byte, 513)), false},

		// Invalid - shell metacharacters
		{"semicolon", "nixpkgs#hello;rm -rf /", false},
		{"pipe", "nixpkgs#hello|evil", false},
		{"ampersand", "nixpkgs#pkg&&evil", false},
		{"dollar", "nixpkgs#$(evil)", false},
		{"backtick", "nixpkgs#`evil`", false},
		{"redirect", "nixpkgs#pkg>file", false},
		{"redirect in", "nixpkgs#pkg<file", false},
		{"parentheses", "nixpkgs#pkg()", false},
		{"brackets", "nixpkgs#pkg[]", false},
		{"braces", "nixpkgs#pkg{}", false},
		{"space", "nixpkgs#hello world", false},

		// Invalid - path traversal
		{"path traversal", "nixpkgs#../etc/passwd", false},
		{"dot dot", "nixpkgs#foo..bar", false},

		// Invalid - special characters
		{"single quote", "nixpkgs#pkg'foo", false},
		{"double quote", "nixpkgs#pkg\"foo", false},
		{"newline", "nixpkgs#pkg\nfoo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFlakeRef(tt.ref)
			if result != tt.expected {
				t.Errorf("isValidFlakeRef(%q) = %v, expected %v", tt.ref, result, tt.expected)
			}
		})
	}
}

func TestIsValidNixStorePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Valid store paths
		{"derivation", "/nix/store/abc123-hello-1.0.0.drv", true},
		{"output", "/nix/store/xyz789-hello-1.0.0", true},
		{"with plus", "/nix/store/abc123-gcc++-12.0", true},
		{"long hash", "/nix/store/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-pkg", true},

		// Invalid - not starting with /nix/store/
		{"no prefix", "abc123-hello", false},
		{"wrong prefix", "/usr/store/abc123", false},
		{"partial prefix", "/nix/abc123", false},

		// Invalid - empty or too long
		{"empty", "", false},
		{"too long", "/nix/store/" + string(make([]byte, 300)), false},

		// Invalid - shell metacharacters
		{"semicolon", "/nix/store/abc;evil", false},
		{"pipe", "/nix/store/abc|evil", false},
		{"dollar", "/nix/store/$(evil)", false},
		{"backtick", "/nix/store/abc`evil`", false},
		{"space", "/nix/store/abc def", false},

		// Invalid - path traversal
		{"path traversal", "/nix/store/../etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidNixStorePath(tt.path)
			if result != tt.expected {
				t.Errorf("isValidNixStorePath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGetLocksMap(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]interface{}
		expectOK   bool
		expectKeys []string
	}{
		{
			name:     "no locks",
			params:   map[string]interface{}{},
			expectOK: false,
		},
		{
			name: "interface map",
			params: map[string]interface{}{
				"locks": map[string]interface{}{
					"locked_ref": "github:NixOS/nixpkgs/abc123",
					"system":     "x86_64-linux",
				},
			},
			expectOK:   true,
			expectKeys: []string{"locked_ref", "system"},
		},
		{
			name: "string map",
			params: map[string]interface{}{
				"locks": map[string]string{
					"locked_ref": "test",
				},
			},
			expectOK:   true,
			expectKeys: []string{"locked_ref"},
		},
		{
			name: "json string",
			params: map[string]interface{}{
				"locks": `{"locked_ref": "test", "system": "linux"}`,
			},
			expectOK:   true,
			expectKeys: []string{"locked_ref", "system"},
		},
		{
			name: "invalid type",
			params: map[string]interface{}{
				"locks": 123,
			},
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := getLocksMap(tt.params)
			if ok != tt.expectOK {
				t.Errorf("getLocksMap() ok = %v, expected %v", ok, tt.expectOK)
			}
			if tt.expectOK && result != nil {
				for _, key := range tt.expectKeys {
					if _, exists := result[key]; !exists {
						t.Errorf("getLocksMap() missing expected key %q", key)
					}
				}
			}
		})
	}
}

func TestNixRealizeAction_Execute_PlatformCheck(t *testing.T) {
	// Skip on Linux - the platform check passes there
	if runtime.GOOS == "linux" {
		t.Skip("Skipping platform check test on Linux")
	}

	action := &NixRealizeAction{}
	ctx := &ExecutionContext{}
	params := map[string]interface{}{
		"flake_ref":   "nixpkgs#hello",
		"executables": []string{"hello"},
	}

	err := action.Execute(ctx, params)
	if err == nil {
		t.Error("Expected platform error on non-Linux")
	}
	if err != nil && !strings.Contains(err.Error(), "only supports Linux") {
		t.Errorf("Expected 'only supports Linux' error, got: %v", err)
	}
}

func TestNixRealizeAction_Execute_MissingParams(t *testing.T) {
	// Skip on non-Linux - will fail at platform check
	if runtime.GOOS != "linux" {
		t.Skip("Skipping on non-Linux")
	}

	action := &NixRealizeAction{}
	ctx := &ExecutionContext{}

	tests := []struct {
		name           string
		params         map[string]interface{}
		expectedErrMsg string
	}{
		{
			name:           "missing flake_ref and package",
			params:         map[string]interface{}{},
			expectedErrMsg: "requires 'flake_ref' or 'package'",
		},
		{
			name: "missing executables",
			params: map[string]interface{}{
				"flake_ref": "nixpkgs#hello",
			},
			expectedErrMsg: "requires 'executables'",
		},
		{
			name: "empty executables",
			params: map[string]interface{}{
				"flake_ref":   "nixpkgs#hello",
				"executables": []string{},
			},
			expectedErrMsg: "requires 'executables'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.Execute(ctx, tt.params)
			if err == nil {
				t.Error("Expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectedErrMsg, err)
			}
		})
	}
}

func TestNixRealizeAction_Execute_InvalidInputs(t *testing.T) {
	// Skip on non-Linux - will fail at platform check
	if runtime.GOOS != "linux" {
		t.Skip("Skipping on non-Linux")
	}

	action := &NixRealizeAction{}
	ctx := &ExecutionContext{}

	tests := []struct {
		name           string
		params         map[string]interface{}
		expectedErrMsg string
	}{
		{
			name: "invalid flake ref",
			params: map[string]interface{}{
				"flake_ref":   "invalid;injection",
				"executables": []string{"hello"},
			},
			expectedErrMsg: "invalid flake reference",
		},
		{
			name: "invalid package name",
			params: map[string]interface{}{
				"package":     "pkg;rm -rf /",
				"executables": []string{"hello"},
			},
			expectedErrMsg: "invalid nixpkgs package name",
		},
		{
			name: "invalid executable name",
			params: map[string]interface{}{
				"flake_ref":   "nixpkgs#hello",
				"executables": []string{"../evil"},
			},
			expectedErrMsg: "invalid executable name",
		},
		{
			name: "invalid derivation path",
			params: map[string]interface{}{
				"flake_ref":       "nixpkgs#hello",
				"executables":     []string{"hello"},
				"derivation_path": "/tmp/evil.drv",
			},
			expectedErrMsg: "invalid derivation path",
		},
		{
			name: "invalid output path",
			params: map[string]interface{}{
				"flake_ref":   "nixpkgs#hello",
				"executables": []string{"hello"},
				"output_path": "/home/user/evil",
			},
			expectedErrMsg: "invalid output path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.Execute(ctx, tt.params)
			if err == nil {
				t.Error("Expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectedErrMsg, err)
			}
		})
	}
}

func TestNixRealizeIsPrimitive(t *testing.T) {
	if !IsPrimitive("nix_realize") {
		t.Error("nix_realize should be registered as a primitive")
	}
}
