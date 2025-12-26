package actions

import (
	"testing"
)

func TestValidateAction_UnknownAction(t *testing.T) {
	result := ValidateAction("nonexistent_action", nil)
	if !result.HasErrors() {
		t.Error("expected error for unknown action")
	}
	if len(result.Errors) != 1 || result.Errors[0] != "unknown action 'nonexistent_action'" {
		t.Errorf("unexpected error message: %v", result.Errors)
	}
}

func TestValidateAction_ActionWithPreflight(t *testing.T) {
	// Actions implementing Preflight validate their parameters
	// download requires 'url' parameter
	result := ValidateAction("download", nil)
	if !result.HasErrors() {
		t.Error("expected error for download without url parameter")
	}

	// With valid params, should pass
	result = ValidateAction("download", map[string]interface{}{"url": "https://example.com"})
	if result.HasErrors() {
		t.Errorf("expected no errors for download with url, got: %v", result.Errors)
	}
}

func TestValidateAction_ActionWithoutPreflight(t *testing.T) {
	// Actions that don't implement Preflight pass validation
	// chmod is an example that doesn't require specific params in Preflight
	result := ValidateAction("chmod", nil)
	if result.HasErrors() {
		t.Errorf("expected no errors for action that passes Preflight validation, got: %v", result.Errors)
	}
}

func TestValidateAction_Warnings(t *testing.T) {
	// Test that warnings are returned separately from errors
	// For now, we just verify the structure works
	result := ValidateAction("download", map[string]interface{}{"url": "https://example.com"})
	if result.HasErrors() {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
	// Warnings may or may not be present depending on the action
	// The important thing is the structure supports them
}

func TestPreflightResult_ToError(t *testing.T) {
	// Test ToError with no errors
	result := &PreflightResult{}
	if result.ToError() != nil {
		t.Error("expected nil error for empty result")
	}

	// Test ToError with one error
	result.AddError("single error")
	err := result.ToError()
	if err == nil {
		t.Error("expected error for result with errors")
	}
	if err.Error() != "single error" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Test ToError with multiple errors
	result.AddError("second error")
	err = result.ToError()
	if err == nil {
		t.Error("expected error for result with multiple errors")
	}
	if err.Error() != "single error (and 1 more errors)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPreflightResult_AddMethods(t *testing.T) {
	result := &PreflightResult{}

	result.AddError("error1")
	result.AddErrorf("error %d", 2)
	result.AddWarning("warning1")
	result.AddWarningf("warning %d", 2)

	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
	if len(result.Warnings) != 2 {
		t.Errorf("expected 2 warnings, got %d", len(result.Warnings))
	}
	if !result.HasErrors() {
		t.Error("expected HasErrors to return true")
	}
	if !result.HasWarnings() {
		t.Error("expected HasWarnings to return true")
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
