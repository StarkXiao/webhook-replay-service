package main

import (
	"context"
	"fmt"
	"github.com/StarkXiao/webhook-replay-service/internal/config"
	"github.com/StarkXiao/webhook-replay-service/internal/handler"
	"github.com/StarkXiao/webhook-replay-service/internal/middleware"
	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/service"
	"github.com/StarkXiao/webhook-replay-service/pkg/clock"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	c := config.Load()
	if err := c.Validate(); err != nil {
		log.Fatal("invalid configuration: " + err.Error())
	}
	repo := repository.NewMemory()
	s := &service.Service{Repo: repo, Clock: clock.Real{}, MaxRetries: c.MaxRetries}
	h := handler.New(s)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/api/v1/webhooks", h.Webhooks)
	mux.HandleFunc("/api/v1/events", h.Events)
	mux.HandleFunc("/api/v1/events/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.Event(w, r)
			return
		}
		if r.Method == http.MethodPost && stringsHas(r.URL.Path, "/schedule") {
			h.Schedule(w, r)
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
	slog.Info("server starting", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	_ = context.Background()
	_ = fmt.Sprintf
	_ = os.Getenv
}
func stringsHas(s, x string) bool         { return len(s) >= len(x) && s[len(s)-len(x):] == x }
func writeNotFound(w http.ResponseWriter) { http.Error(w, "not found", 404) }
