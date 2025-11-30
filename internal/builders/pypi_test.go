package builders

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPyPIBuilder_Name(t *testing.T) {
	builder := NewPyPIBuilder(nil)
	if builder.Name() != "pypi" {
		t.Errorf("Name() = %q, want %q", builder.Name(), "pypi")
	}
}

func TestPyPIBuilder_CanBuild(t *testing.T) {
	// Test server that returns 200 for ruff, 404 for everything else
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pypi/ruff/json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"info":{"name":"ruff","summary":"Linter"}}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	builder := NewPyPIBuilderWithBaseURL(nil, server.URL)
	ctx := context.Background()

	tests := []struct {
		name    string
		pkg     string
		wantOK  bool
		wantErr bool
		useReal bool // use real builder (for invalid name check)
	}{
		{"valid package", "ruff", true, false, false},
		{"not found", "nonexistent", false, false, false},
		{"invalid name with spaces", "invalid name", false, false, true},
		{"invalid name path traversal", "../etc/passwd", false, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := builder
			if tc.useReal {
				b = NewPyPIBuilder(nil)
			}
			canBuild, err := b.CanBuild(ctx, tc.pkg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CanBuild() error = %v, wantErr %v", err, tc.wantErr)
			}
			if canBuild != tc.wantOK {
				t.Errorf("CanBuild() = %v, want %v", canBuild, tc.wantOK)
			}
		})
	}
}

func TestPyPIBuilder_Build(t *testing.T) {
	// Test server returning different responses per package
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pypi/ruff/json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"info": {
					"name": "ruff",
					"summary": "An extremely fast Python linter.",
					"home_page": "",
					"project_urls": {"Homepage": "https://docs.astral.sh/ruff", "Repository": "https://github.com/astral-sh/ruff"}
				}
			}`))
		case "/pypi/no-source/json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"info":{"name":"no-source","summary":"Tool","home_page":"","project_urls":null}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	builder := NewPyPIBuilderWithBaseURL(nil, server.URL)
	ctx := context.Background()

	t.Run("build ruff recipe", func(t *testing.T) {
		result, err := builder.Build(ctx, "ruff", "")
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if result.Recipe == nil {
			t.Fatal("result.Recipe is nil")
		}
		if result.Recipe.Metadata.Name != "ruff" {
			t.Errorf("Metadata.Name = %q, want %q", result.Recipe.Metadata.Name, "ruff")
		}
		if result.Recipe.Version.Source != "pypi" {
			t.Errorf("Version.Source = %q, want %q", result.Recipe.Version.Source, "pypi")
		}
		if result.Recipe.Steps[0].Action != "pipx_install" {
			t.Errorf("Steps[0].Action = %q, want %q", result.Recipe.Steps[0].Action, "pipx_install")
		}
		if result.Source != "pypi:ruff" {
			t.Errorf("Source = %q, want %q", result.Source, "pypi:ruff")
		}
		// Homepage from project_urls.Homepage
		if result.Recipe.Metadata.Homepage != "https://docs.astral.sh/ruff" {
			t.Errorf("Homepage = %q, want %q", result.Recipe.Metadata.Homepage, "https://docs.astral.sh/ruff")
		}
	})

	t.Run("fallback to package name when no source URL", func(t *testing.T) {
		result, err := builder.Build(ctx, "no-source", "")
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
		if len(result.Warnings) == 0 {
			t.Error("expected warning about no source URL")
		}
		executables := result.Recipe.Steps[0].Params["executables"].([]string)
		if len(executables) != 1 || executables[0] != "no-source" {
			t.Errorf("executables = %v, want [\"no-source\"]", executables)
		}
		if result.Recipe.Verify.Command != "no-source --version" {
			t.Errorf("Verify.Command = %q", result.Recipe.Verify.Command)
		}
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := builder.Build(ctx, "nonexistent", "")
		if err == nil {
			t.Error("Build() should fail for nonexistent package")
		}
	})
}

func TestIsValidPyPIPackageName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"ruff", true},
		{"black", true},
		{"some_package", true},
		{"package-name", true},
		{"a", true},
		{"A", true},
		{"package.name", true},
		{"a1", true},
		{"1a", true},
		{"", false},
		{"-invalid", false},
		{"../path/traversal", false},
		{"path/traversal", false},
		{"path\\traversal", false},
		{"has spaces", false},
		// 215 characters (too long)
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidPyPIPackageName(tc.name)
			if got != tc.valid {
				t.Errorf("isValidPyPIPackageName(%q) = %v, want %v", tc.name, got, tc.valid)
			}
		})
	}
}

func TestPyPIBuilder_buildPyprojectURL(t *testing.T) {
	builder := NewPyPIBuilder(nil)

	tests := []struct {
		sourceURL string
		want      string
	}{
		{"https://github.com/astral-sh/ruff", "https://raw.githubusercontent.com/astral-sh/ruff/HEAD/pyproject.toml"},
		{"https://github.com/psf/black.git", "https://raw.githubusercontent.com/psf/black/HEAD/pyproject.toml"},
		{"https://github.com/owner/repo/", "https://raw.githubusercontent.com/owner/repo/HEAD/pyproject.toml"},
		{"https://gitlab.com/owner/repo", ""}, // Not GitHub
		{"not-a-url", ""},                     // Invalid URL
	}

	for _, tc := range tests {
		t.Run(tc.sourceURL, func(t *testing.T) {
			got := builder.buildPyprojectURL(tc.sourceURL)
			if got != tc.want {
				t.Errorf("buildPyprojectURL(%q) = %q, want %q", tc.sourceURL, got, tc.want)
			}
		})
	}
}
