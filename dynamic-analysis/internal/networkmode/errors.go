package networkmode

import "errors"

var (
	// ErrInvalidMode indicates an invalid network mode
	ErrInvalidMode = errors.New("invalid network mode")

	// ErrInvalidAction indicates an invalid action
	ErrInvalidAction = errors.New("invalid action")

	// ErrMissingConfig indicates missing configuration
	ErrMissingConfig = errors.New("missing configuration")

	// ErrMissingServiceConfig indicates missing service configuration
	ErrMissingServiceConfig = errors.New("missing service configuration")

	// ErrMissingProxyConfig indicates missing proxy configuration
	ErrMissingProxyConfig = errors.New("missing proxy configuration")

	// ErrHalfModeNotEnabled indicates Half Mode is not enabled
	ErrHalfModeNotEnabled = errors.New("half mode is not enabled")

	// ErrTransparentModeNotEnabled indicates Transparent Mode is not enabled
	ErrTransparentModeNotEnabled = errors.New("transparent mode is not enabled")

	// ErrRequestBlocked indicates the request was blocked
	ErrRequestBlocked = errors.New("request blocked by policy")

	// ErrNoDecision indicates no decision could be made
	ErrNoDecision = errors.New("no decision could be made")

	// ErrModificationFailed indicates traffic modification failed
	ErrModificationFailed = errors.New("traffic modification failed")

	// ErrInterceptionFailed indicates traffic interception failed
	ErrInterceptionFailed = errors.New("traffic interception failed")
)
