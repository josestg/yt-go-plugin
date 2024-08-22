package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"plugin"
	"time"

	"github.com/josestg/yt-go-plugin/cache"
	"github.com/josestg/yt-go-plugin/internal/fibonacci"
)

type conf struct {
	Port                   int
	LogLevel               slog.Level
	CacheExpiration        time.Duration
	CachePluginPath        string
	CachePluginFactoryName string
}

func main() {
	var cfg conf
	flag.IntVar(&cfg.Port, "port", 8080, "port to listen on")
	flag.TextVar(&cfg.LogLevel, "log-level", slog.LevelInfo, "log level")
	flag.StringVar(&cfg.CachePluginPath, "cache-plugin-path", "", "path to the cache plugin")
	flag.StringVar(&cfg.CachePluginFactoryName, "cache-plugin-factory-name", "Factory", "name of the factory function in the cache plugin")
	flag.DurationVar(&cfg.CacheExpiration, "cache-expiration", 15*time.Second, "duration that a cache entry will be valid for")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel}))
	if err := run(cfg, log); err != nil {
		log.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run(cfg conf, log *slog.Logger) error {
	log.Info("application started")
	log.Debug("using configuration", "config", cfg)
	defer log.Info("application stopped")

	plug, err := loadCachePlugin(log, cfg.CachePluginPath, cfg.CachePluginFactoryName)
	if err != nil {
		return fmt.Errorf("load plugin: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /fib/{n}", fibonacci.NewHandler(log, plug, cfg.CacheExpiration))

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Info("listening", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

// loadCachePlugin loads a cache implementation from a shared object (.so) file at the specified path.
// It calls the constructor function by name, passing the necessary dependencies, and returns the initialized cache.
// If path is empty, it returns the NopCache implementation.
func loadCachePlugin(log *slog.Logger, path, name string) (cache.Cache, error) {
	if path == "" {
		log.Info("no cache plugin configured; using nop cache")
		return cache.NopCache, nil
	}

	plug, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open plugin %q: %w", path, err)
	}

	sym, err := plug.Lookup(name)
	if err != nil {
		return nil, fmt.Errorf("lookup symbol New: %w", err)
	}

	factoryPtr, ok := sym.(*cache.Factory)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T; want %T", sym, factoryPtr)
	}

	factory := *factoryPtr
	return factory(log)
}
