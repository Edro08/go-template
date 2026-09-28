package checks

import "context"

const (
	// StatusOK indicates that a dependency is available.
	StatusOK = "ok"
	// StatusFail indicates that a dependency is unavailable.
	StatusFail = "failed"
)

// Check is a dependency check used by the readiness endpoint.
type Check interface {
	Name() string
	Type() string
	Technology() string
	Check(context.Context) Result
}

// Result is the public result of a readiness check.
type Result struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	Technology string `json:"technology"`
	Version    string `json:"version,omitempty"`
	Error      error  `json:"-"`
}
