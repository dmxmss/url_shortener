package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"dmxmss-project/internal/cache"
	"dmxmss-project/internal/config"
	"dmxmss-project/internal/events"
	"dmxmss-project/internal/httpx"
	"dmxmss-project/internal/metrics"
	"dmxmss-project/internal/storage"
)

type server struct {
	cfg       config.Config
	db        *storage.Store
	cache     *cache.Redis
	publisher events.Publisher
	log       *slog.Logger
	m         *metrics.Metrics
}

func main() {
	cfg := config.Load()
	log := httpx.NewLogger(cfg.LogLevel)
	ctx := context.Background()

	store, err := storage.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	redis, err := cache.New(ctx, cfg)
	if err != nil {
		log.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer redis.Close()

	publisher := events.NewNoop()
	if cfg.NATSURL != "" {
		publisher, err = events.NewNATS(cfg.NATSURL)
		if err != nil {
			log.Warn("nats disabled", "error", err)
			publisher = events.NewNoop()
		}
	}
	defer publisher.Close()

	s := &server{cfg: cfg, db: store, cache: redis, publisher: publisher, log: log, m: metrics.New("redirect_service")}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", httpx.Healthz)
	mux.HandleFunc("GET /readyz", httpx.Readyz(store, redis))
	mux.Handle("GET /metrics", s.m.Handler())
	mux.HandleFunc("GET /", s.redirect)

	handler := httpx.Logging(log, s.m)(mux)
	if err := httpx.ListenAndServe(cfg.RedirectAddr, handler, cfg.ShutdownTimeout, log); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func (s *server) redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.Trim(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if code == "" || strings.Contains(code, "/") {
		http.NotFound(w, r)
		return
	}

	longURL, err := s.cache.GetURL(r.Context(), code)
	if err == nil {
		s.m.CacheHits.Inc()
	} else if errors.Is(err, cache.ErrMiss) {
		s.m.CacheMisses.Inc()
		longURL, err = s.db.GetURL(r.Context(), code)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			s.log.Error("lookup url", "short_code", code, "error", err)
			httpx.JSONError(w, http.StatusInternalServerError, "lookup failed")
			return
		}
		_ = s.cache.SetURL(r.Context(), code, longURL, s.cfg.CacheTTL)
	} else {
		s.log.Warn("redis read failed", "error", err)
		longURL, err = s.db.GetURL(r.Context(), code)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	now := time.Now().UTC()
	userAgent := r.UserAgent()
	remoteAddr := r.RemoteAddr
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := s.db.RecordRedirect(ctx, code, now); err != nil {
			s.log.Warn("record redirect", "short_code", code, "error", err)
		}
		event := events.RedirectEvent{ShortCode: code, AccessedAt: now, UserAgent: userAgent, RemoteAddr: remoteAddr}
		if err := s.publisher.PublishRedirect(ctx, event); err != nil {
			s.log.Warn("publish redirect event", "error", err)
		}
	}()

	s.m.Redirects.Inc()
	http.Redirect(w, r, longURL, http.StatusFound)
}
