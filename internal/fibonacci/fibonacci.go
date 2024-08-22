package fibonacci

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/josestg/yt-go-plugin/cache"
)

// Fibonacci calculates the nth Fibonacci number.
// This algorithm is not optimized and is used for demonstration purposes.
func Fibonacci(n int64) int64 {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

// NewHandler returns an HTTP handler that calculates the nth Fibonacci number.
func NewHandler(l *slog.Logger, c cache.Cache, exp time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		defer func() {
			l.Info("request completed", "duration", time.Since(started).String())
		}()

		param := r.PathValue("n")
		n, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			l.Error("cannot parse path value", "param", param, "error", err)
			sendJSON(l, w, map[string]any{"error": "invalid value"}, http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		result := make(chan int64)
		go func() {
			cached, err := c.Get(ctx, param)
			if err != nil {
				l.Debug("cache miss; calculating the fib(n)", "n", n, "cache_error", err)
				v := Fibonacci(n)
				l.Debug("fib(n) calculated", "n", n, "result", v)
				if err := c.Set(ctx, param, strconv.FormatInt(v, 10), exp); err != nil {
					l.Error("cannot set cache", "error", err)
				}
				result <- v
				return
			}

			l.Debug("cache hit; returning the cached value", "n", n, "value", cached)
			v, _ := strconv.ParseInt(cached, 10, 64)
			result <- v
		}()

		select {
		case v := <-result:
			sendJSON(l, w, map[string]any{"result": v}, http.StatusOK)
		case <-ctx.Done():
			l.Info("request cancelled")
		}
	}
}

func sendJSON(l *slog.Logger, w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		l.Error("cannot encode response", "error", err)
	}
}
