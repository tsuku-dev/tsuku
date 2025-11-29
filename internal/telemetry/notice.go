package telemetry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tsuku-dev/tsuku/internal/config"
)

const (
	// NoticeMarkerFile is the filename used to track if the notice has been shown.
	NoticeMarkerFile = "telemetry_notice_shown"

	// NoticeText is the message displayed to users on first run.
	NoticeText = `tsuku collects anonymous usage statistics to improve the tool.
No personal information is collected. See: https://tsuku.dev/telemetry

To opt out: export TSUKU_NO_TELEMETRY=1
`
)

// ShowNoticeIfNeeded displays the telemetry notice on first run.
// It writes to stderr and creates a marker file to prevent future displays.
// Returns silently on any error (file permissions, etc.).
func ShowNoticeIfNeeded() {
	// Don't show notice if telemetry is disabled
	if os.Getenv(EnvNoTelemetry) != "" {
		return
	}

	cfg, err := config.DefaultConfig()
	if err != nil {
		return // Silent failure
	}

	markerPath := filepath.Join(cfg.HomeDir, NoticeMarkerFile)

	// Check if marker file exists
	if _, err := os.Stat(markerPath); err == nil {
		return // Already shown
	}

	// Show notice to stderr
	fmt.Fprint(os.Stderr, NoticeText)

	// Create marker file (ensure directory exists)
	if err := os.MkdirAll(cfg.HomeDir, 0755); err != nil {
		return // Silent failure
	}

	// Create empty marker file
	f, err := os.Create(markerPath)
	if err != nil {
		return // Silent failure
	}
	f.Close()
}
