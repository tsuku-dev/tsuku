package telemetry

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestShowNoticeIfNeeded_FirstRun(t *testing.T) {
	// Setup temp directory as TSUKU_HOME
	tmpDir := t.TempDir()
	t.Setenv("TSUKU_HOME", tmpDir)
	_ = os.Unsetenv(EnvNoTelemetry)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	ShowNoticeIfNeeded()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify notice was shown
	if output != NoticeText {
		t.Errorf("notice text mismatch:\ngot:  %q\nwant: %q", output, NoticeText)
	}

	// Verify marker file was created
	markerPath := filepath.Join(tmpDir, NoticeMarkerFile)
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("marker file was not created")
	}
}

func TestShowNoticeIfNeeded_AlreadyShown(t *testing.T) {
	// Setup temp directory with marker file
	tmpDir := t.TempDir()
	t.Setenv("TSUKU_HOME", tmpDir)
	_ = os.Unsetenv(EnvNoTelemetry)

	// Create marker file
	markerPath := filepath.Join(tmpDir, NoticeMarkerFile)
	f, err := os.Create(markerPath)
	if err != nil {
		t.Fatalf("failed to create marker file: %v", err)
	}
	f.Close()

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	ShowNoticeIfNeeded()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify notice was NOT shown
	if output != "" {
		t.Errorf("notice was shown when marker file exists: %q", output)
	}
}

func TestShowNoticeIfNeeded_TelemetryDisabled(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	t.Setenv("TSUKU_HOME", tmpDir)
	t.Setenv(EnvNoTelemetry, "1")

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	ShowNoticeIfNeeded()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify notice was NOT shown
	if output != "" {
		t.Errorf("notice was shown when telemetry disabled: %q", output)
	}

	// Verify marker file was NOT created
	markerPath := filepath.Join(tmpDir, NoticeMarkerFile)
	if _, err := os.Stat(markerPath); err == nil {
		t.Error("marker file was created when telemetry disabled")
	}
}

func TestShowNoticeIfNeeded_RespectsHome(t *testing.T) {
	// Setup custom TSUKU_HOME
	tmpDir := t.TempDir()
	customHome := filepath.Join(tmpDir, "custom", "tsuku")
	t.Setenv("TSUKU_HOME", customHome)
	_ = os.Unsetenv(EnvNoTelemetry)

	// Capture stderr (ignore output)
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	ShowNoticeIfNeeded()

	w.Close()
	os.Stderr = oldStderr

	// Verify marker file was created in custom location
	markerPath := filepath.Join(customHome, NoticeMarkerFile)
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Errorf("marker file not created at custom TSUKU_HOME: %s", markerPath)
	}
}

func TestNoticeText_Content(t *testing.T) {
	// Verify expected content per issue requirements
	expectedSubstrings := []string{
		"tsuku collects anonymous usage statistics",
		"No personal information is collected",
		"https://tsuku.dev/telemetry",
		"TSUKU_NO_TELEMETRY=1",
	}

	for _, expected := range expectedSubstrings {
		if !bytes.Contains([]byte(NoticeText), []byte(expected)) {
			t.Errorf("NoticeText missing expected content: %q", expected)
		}
	}
}
