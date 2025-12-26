package actions

import (
	"testing"
)

func TestValidateAction_UnknownAction(t *testing.T) {
	err := ValidateAction("nonexistent_action", nil)
	if err == nil {
		t.Error("expected error for unknown action")
	}
	if err.Error() != "unknown action 'nonexistent_action'" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateAction_ActionWithPreflight(t *testing.T) {
	// Actions implementing Preflight validate their parameters
	// download requires 'url' parameter
	err := ValidateAction("download", nil)
	if err == nil {
		t.Error("expected error for download without url parameter")
	}

	// With valid params, should pass
	err = ValidateAction("download", map[string]interface{}{"url": "https://example.com"})
	if err != nil {
		t.Errorf("expected nil for download with url, got: %v", err)
	}
}

func TestValidateAction_ActionWithoutPreflight(t *testing.T) {
	// Actions that don't implement Preflight pass validation
	// chmod is an example that doesn't require specific params in Preflight
	err := ValidateAction("chmod", nil)
	if err != nil {
		t.Errorf("expected nil for action that passes Preflight validation, got: %v", err)
	}
}

func TestRegisteredNames(t *testing.T) {
	names := RegisteredNames()

	// Should have actions registered
	if len(names) == 0 {
		t.Error("expected registered actions")
	}

	// Should include known actions
	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	expected := []string{"download", "extract", "chmod", "install_binaries"}
	for _, exp := range expected {
		if !found[exp] {
			t.Errorf("expected action '%s' to be registered", exp)
		}
	}

	// Should be sorted
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("names not sorted: %s comes after %s", names[i], names[i-1])
		}
	}
}
