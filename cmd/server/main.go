package main

import (
	"context"
	"github.com/StarkXiao/webhook-replay-service/internal/config"
	"github.com/StarkXiao/webhook-replay-service/internal/handler"
	"github.com/StarkXiao/webhook-replay-service/internal/middleware"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/retry"
	"github.com/StarkXiao/webhook-replay-service/internal/scheduler"
	"github.com/StarkXiao/webhook-replay-service/internal/service"
	"github.com/StarkXiao/webhook-replay-service/internal/worker"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c := config.Load()
	if err := c.Validate(); err != nil {
		log.Fatal("invalid configuration: " + err.Error())
	}
	repo := repository.NewMemory()
	s := &service.Service{Repo: repo, Clock: clock.Real{}, MaxRetries: c.MaxRetries, RetryPolicy: retry.Policy{Initial: c.InitialBackoff, Max: c.MaxBackoff}}
	h := handler.New(s)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/api/v1/webhooks", method(http.MethodPost, h.Webhooks))
	mux.HandleFunc("/api/v1/events", method(http.MethodGet, h.Events))
	mux.HandleFunc("/api/v1/events/replay", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.ReplayBatch(w, r)
	})
	mux.HandleFunc("/api/v1/statistics", method(http.MethodGet, h.Statistics))
	mux.HandleFunc("/api/v1/dead-letters", method(http.MethodGet, h.DeadLetters))
	mux.HandleFunc("/api/v1/delivery-tasks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.TaskAction(w, r)
	})
	mux.HandleFunc("/api/v1/events/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.Event(w, r)
			return
		}
		if r.Method == http.MethodPost && stringsHas(r.URL.Path, "/schedule") {
			h.Schedule(w, r)
			return
		}
		if r.Method == http.MethodPost && stringsHas(r.URL.Path, "/replay") {
			h.Replay(w, r)
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeNotFound(w)
	})
	handler := middleware.Request(middleware.CORS(middleware.Limit(c.RequestBodyLimit, mux)))
	if c.RatePerSecond > 0 {
		handler = middleware.NewLimiter(c.RatePerSecond).Middleware(handler)
	}
	srv := &http.Server{Addr: ":" + c.Port, Handler: handler}
	pool := worker.New(s, c.WorkerCount)
	runner := &scheduler.Scheduler{Repo: repo, Pool: pool, Every: time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	schedulerDone := make(chan struct{})
	go func() { defer close(schedulerDone); runner.Run(ctx) }()
	recoveryDone := make(chan struct{})
	go func() {
		defer close(recoveryDone)
		(worker.Recovery{Repo: repo, Timeout: 30 * time.Second}).Run(ctx, time.Second)
	}()
	slog.Info("server starting", "addr", srv.Addr)
	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.ListenAndServe() }()
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}
	<-schedulerDone
	<-recoveryDone
	pool.Close()
}
func stringsHas(s, x string) bool         { return len(s) >= len(x) && s[len(s)-len(x):] == x }
func writeNotFound(w http.ResponseWriter) { http.Error(w, "not found", 404) }
func method(allowed string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowed {
			w.Header().Set("Allow", allowed)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}
