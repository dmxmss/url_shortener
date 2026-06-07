package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"dmxmss-project/internal/cache"
	"dmxmss-project/internal/codegen"
	"dmxmss-project/internal/config"
	"dmxmss-project/internal/httpx"
	"dmxmss-project/internal/metrics"
	"dmxmss-project/internal/storage"
)

type server struct {
	cfg   config.Config
	db    *storage.Store
	cache *cache.Redis
	log   *slog.Logger
	m     *metrics.Metrics
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
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

	s := &server{cfg: cfg, db: store, cache: redis, log: log, m: metrics.New("shortener_api")}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", s.shorten)
	mux.HandleFunc("GET /healthz", httpx.Healthz)
	mux.HandleFunc("GET /readyz", httpx.Readyz(store, redis))
	mux.Handle("GET /metrics", s.m.Handler())

	handler := httpx.Logging(log, s.m)(mux)
	if err := httpx.ListenAndServe(cfg.APIAddr, handler, cfg.ShutdownTimeout, log); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func (s *server) shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpx.JSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	longURL := strings.TrimSpace(req.URL)
	if !validHTTPURL(longURL) {
		httpx.JSONError(w, http.StatusBadRequest, "url must be an absolute http or https URL")
		return
	}

	var code string
	var err error
	for i := 0; i < 5; i++ {
		code, err = codegen.RandomBase62(7)
		if err != nil {
			httpx.JSONError(w, http.StatusInternalServerError, "could not generate short code")
			return
		}
		err = s.db.CreateLink(r.Context(), code, longURL)
		if err == nil {
			break
		}
		if !errors.Is(err, storage.ErrConflict) {
			s.log.Error("create link", "error", err)
			httpx.JSONError(w, http.StatusInternalServerError, "could not create short link")
			return
		}
	}
	if err != nil {
		httpx.JSONError(w, http.StatusConflict, "could not allocate short code")
		return
	}

	_ = s.cache.SetURL(r.Context(), code, longURL, s.cfg.CacheTTL)
	resp := shortenResponse{
		ShortCode: code,
		ShortURL:  strings.TrimRight(s.cfg.PublicBaseURL, "/") + "/" + code,
	}
	httpx.JSON(w, http.StatusCreated, resp)
	s.m.LinksCreated.Inc()
}

func validHTTPURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
