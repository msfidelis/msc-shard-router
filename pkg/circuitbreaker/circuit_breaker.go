package circuitbreaker

import (
	"os"
	"strconv"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half_open"
	default:
		return "closed"
	}
}

type Config struct {
	FailureThreshold         int
	OpenTimeout              time.Duration
	HalfOpenSuccessThreshold int
}

func DefaultConfig() Config {
	return Config{
		FailureThreshold:         5,
		OpenTimeout:              30 * time.Second,
		HalfOpenSuccessThreshold: 2,
	}
}

func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	if value := os.Getenv("CB_FAILURE_THRESHOLD"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			cfg.FailureThreshold = parsed
		}
	}

	if value := os.Getenv("CB_OPEN_TIMEOUT_SEC"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			cfg.OpenTimeout = time.Duration(parsed) * time.Second
		}
	}

	if value := os.Getenv("CB_HALF_OPEN_SUCCESS_THRESHOLD"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			cfg.HalfOpenSuccessThreshold = parsed
		}
	}

	return cfg
}

type CircuitBreaker struct {
	mu                sync.Mutex
	cfg               Config
	state             State
	failures          int
	openedAt          time.Time
	halfOpenInFlight  bool
	halfOpenSuccesses int
}

func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	if cfg.FailureThreshold < 1 {
		cfg.FailureThreshold = 1
	}
	if cfg.OpenTimeout <= 0 {
		cfg.OpenTimeout = 30 * time.Second
	}
	if cfg.HalfOpenSuccessThreshold < 1 {
		cfg.HalfOpenSuccessThreshold = 1
	}

	return &CircuitBreaker{
		cfg:   cfg,
		state: StateClosed,
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.openedAt) < cb.cfg.OpenTimeout {
			return false
		}
		cb.state = StateHalfOpen
		cb.halfOpenInFlight = false
		cb.halfOpenSuccesses = 0
	case StateHalfOpen:
		if cb.halfOpenInFlight {
			return false
		}
		cb.halfOpenInFlight = true
	}

	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.halfOpenInFlight = false
		cb.halfOpenSuccesses++
		if cb.halfOpenSuccesses >= cb.cfg.HalfOpenSuccessThreshold {
			cb.state = StateClosed
			cb.failures = 0
		}
	case StateClosed:
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.halfOpenInFlight = false
		cb.open()
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.cfg.FailureThreshold {
			cb.open()
		}
	}
}

func (cb *CircuitBreaker) open() {
	cb.state = StateOpen
	cb.openedAt = time.Now()
	cb.failures = 0
	cb.halfOpenInFlight = false
	cb.halfOpenSuccesses = 0
}
