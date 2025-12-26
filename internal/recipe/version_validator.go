package recipe

import "sync"

// VersionValidator validates version configuration for recipes.
// This interface is implemented by the version package and registered at init time,
// enabling the recipe package to validate version configuration without importing
// the version package (which would cause a circular import).
type VersionValidator interface {
	// CanResolveVersion returns true if a version provider can be created for this recipe.
	CanResolveVersion(r *Recipe) bool

	// KnownSources returns the list of known version source values.
	KnownSources() []string

	// ValidateVersionConfig performs detailed validation of version configuration.
	// Returns nil if valid, error describing the problem if invalid.
	ValidateVersionConfig(r *Recipe) error
}

var (
	versionValidator   VersionValidator
	versionValidatorMu sync.RWMutex
)

// SetVersionValidator registers the version validator.
// This is called from the version package init() to register the factory-based validator.
func SetVersionValidator(v VersionValidator) {
	versionValidatorMu.Lock()
	defer versionValidatorMu.Unlock()
	versionValidator = v
}

// GetVersionValidator returns the registered validator or nil if none is registered.
func GetVersionValidator() VersionValidator {
	versionValidatorMu.RLock()
	defer versionValidatorMu.RUnlock()
	return versionValidator
}

// ActionValidator validates action names and parameters for recipes.
// This interface is implemented by the actions package and registered at init time,
// enabling the recipe package to validate actions without importing the actions
// package (which would cause a circular import).
type ActionValidator interface {
	// RegisteredNames returns all registered action names.
	RegisteredNames() []string

	// ValidateAction checks if an action exists and validates its parameters.
	// Returns nil if valid, error describing the problem if invalid.
	ValidateAction(name string, params map[string]interface{}) error
}

var (
	actionValidator   ActionValidator
	actionValidatorMu sync.RWMutex
)

// SetActionValidator registers the action validator.
// This is called from the actions package init() to register the action validator.
func SetActionValidator(v ActionValidator) {
	actionValidatorMu.Lock()
	defer actionValidatorMu.Unlock()
	actionValidator = v
}

// GetActionValidator returns the registered validator or nil if none is registered.
func GetActionValidator() ActionValidator {
	actionValidatorMu.RLock()
	defer actionValidatorMu.RUnlock()
	return actionValidator
}
