package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go-template/internal/health/checks"
	"go-template/kit/logger"
)

const readinessTimeout = 2 * time.Second

type response struct {
	Status string          `json:"status"`
	Checks []checks.Result `json:"checks"`
}

// Handler exposes the liveness, readiness, and startup health endpoints.
type Handler struct {
	checks  []checks.Check
	started atomic.Bool
	logger  logger.ILogger
}

// NewHandler creates a health handler with the supplied readiness checks.
func NewHandler(log logger.ILogger, dependencies ...checks.Check) *Handler {
	return &Handler{checks: dependencies, logger: log}
}

// Live handles the liveness endpoint.
func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Ready handles the readiness endpoint.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
	defer cancel()

	results := make([]checks.Result, len(h.checks))
	var waitGroup sync.WaitGroup
	var failed atomic.Bool
	for i, check := range h.checks {
		waitGroup.Add(1)
		go func(index int, dependency checks.Check) {
			defer waitGroup.Done()
			if dependency == nil {
				failed.Store(true)
				return
			}

			result := dependency.Check(ctx)
			results[index] = result
			if result.Status != checks.StatusOK {
				failed.Store(true)
				if h.logger != nil && result.Error != nil {
					h.logger.Error("readiness check failed", "check", result.Name, "error", result.Error)
				}
			}
		}(i, check)
	}

	waitGroup.Wait()
	status := http.StatusOK
	if failed.Load() {
		status = http.StatusServiceUnavailable
	}

	writeResponse(w, status, response{
		Status: statusText(status),
		Checks: results,
	})
}

// Startup handles the startup endpoint.
func (h *Handler) Startup(w http.ResponseWriter, _ *http.Request) {
	if !h.started.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// MarkStartupReady marks application startup as complete.
func (h *Handler) MarkStartupReady() {
	h.started.Store(true)
}

func statusText(status int) string {
	if status == http.StatusOK {
		return checks.StatusOK
	}

	return "unavailable"
}

func writeResponse(w http.ResponseWriter, status int, payload response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
