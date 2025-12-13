package main

import (
	"log/slog"
	"os"
	"testing"
)

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1", true},
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"yes", true},
		{"YES", true},
		{"Yes", true},
		{"on", true},
		{"ON", true},
		{"On", true},
		{"0", false},
		{"false", false},
		{"FALSE", false},
		{"no", false},
		{"", false},
		{"off", false},
		{"random", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isTruthy(tt.input)
			if got != tt.want {
				t.Errorf("isTruthy(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetermineLogLevel(t *testing.T) {
	// Save original values
	origQuiet := quietFlag
	origVerbose := verboseFlag
	origDebug := debugFlag

	// Reset flags after each test
	defer func() {
		quietFlag = origQuiet
		verboseFlag = origVerbose
		debugFlag = origDebug
	}()

	tests := []struct {
		name       string
		quietF     bool
		verboseF   bool
		debugF     bool
		envQuiet   string
		envVerbose string
		envDebug   string
		want       slog.Level
	}{
		{
			name: "default is WARN",
			want: slog.LevelWarn,
		},
		{
			name:   "debug flag",
			debugF: true,
			want:   slog.LevelDebug,
		},
		{
			name:     "verbose flag",
			verboseF: true,
			want:     slog.LevelInfo,
		},
		{
			name:   "quiet flag",
			quietF: true,
			want:   slog.LevelError,
		},
		{
			name:     "debug env var",
			envDebug: "1",
			want:     slog.LevelDebug,
		},
		{
			name:       "verbose env var",
			envVerbose: "true",
			want:       slog.LevelInfo,
		},
		{
			name:     "quiet env var",
			envQuiet: "yes",
			want:     slog.LevelError,
		},
		{
			name:     "flag takes precedence over env var",
			quietF:   true,
			envDebug: "1",
			want:     slog.LevelError,
		},
		{
			name:     "debug flag overrides verbose flag",
			debugF:   true,
			verboseF: true,
			want:     slog.LevelDebug,
		},
		{
			name:     "verbose flag overrides quiet flag",
			verboseF: true,
			quietF:   true,
			want:     slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flags
			quietFlag = tt.quietF
			verboseFlag = tt.verboseF
			debugFlag = tt.debugF

			// Set env vars
			os.Unsetenv("TSUKU_QUIET")
			os.Unsetenv("TSUKU_VERBOSE")
			os.Unsetenv("TSUKU_DEBUG")
			if tt.envQuiet != "" {
				os.Setenv("TSUKU_QUIET", tt.envQuiet)
			}
			if tt.envVerbose != "" {
				os.Setenv("TSUKU_VERBOSE", tt.envVerbose)
			}
			if tt.envDebug != "" {
				os.Setenv("TSUKU_DEBUG", tt.envDebug)
			}

			got := determineLogLevel()
			if got != tt.want {
				t.Errorf("determineLogLevel() = %v, want %v", got, tt.want)
			}

			// Clean up env vars
			os.Unsetenv("TSUKU_QUIET")
			os.Unsetenv("TSUKU_VERBOSE")
			os.Unsetenv("TSUKU_DEBUG")
		})
	}
}
