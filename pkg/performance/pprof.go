package performance

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
)

// RegisterPprofHandlers registers pprof debugging endpoints on the given router.
// These endpoints are essential for production debugging and performance analysis.
//
// Available endpoints:
//   - /debug/pprof/          - Index page with available profiles
//   - /debug/pprof/cmdline   - Command line invocation
//   - /debug/pprof/profile   - CPU profile (use ?seconds=30 for 30-second profile)
//   - /debug/pprof/symbol    - Symbol lookup
//   - /debug/pprof/trace     - Execution trace (use ?seconds=5 for 5-second trace)
//   - /debug/pprof/heap      - Heap profile
//   - /debug/pprof/goroutine - Goroutine dump
//   - /debug/pprof/block     - Block profile
//   - /debug/pprof/mutex     - Mutex profile
//   - /debug/pprof/allocs    - All past memory allocations
//   - /debug/pprof/threadcreate - Thread creation profile
//
// Usage with curl:
//   curl http://localhost:8080/debug/pprof/heap > heap.out
//   curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.out
//   curl http://localhost:8080/debug/pprof/trace?seconds=5 > trace.out
//
// Usage with go tool:
//   go tool pprof http://localhost:8080/debug/pprof/heap
//   go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
//   go tool trace http://localhost:8080/debug/pprof/trace?seconds=5
func RegisterPprofHandlers(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Additional profiles
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
	router.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	router.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
}

// PprofConfig holds configuration for pprof profiling.
type PprofConfig struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	Path           string `json:"path" yaml:"path"`                       // Default: /debug/pprof
	BlockRate      int    `json:"block_rate" yaml:"block_rate"`           // Block profile rate (0=disabled)
	MutexFraction  int    `json:"mutex_fraction" yaml:"mutex_fraction"`   // Mutex profile fraction (0=disabled)
	MemProfileRate int    `json:"mem_profile_rate" yaml:"mem_profile_rate"` // Memory profile rate (0=default)
}

// DefaultPprofConfig returns default pprof configuration.
func DefaultPprofConfig() PprofConfig {
	return PprofConfig{
		Enabled:        true,
		Path:           "/debug/pprof",
		BlockRate:      1,      // Enable block profiling
		MutexFraction:  1,      // Enable mutex profiling
		MemProfileRate: 512*1024, // Sample 1 allocation per 512KB (default)
	}
}
