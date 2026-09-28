package app

import (
	"context"
	"fmt"
	"go-template/internal/health"
	"go-template/kit/config"
	"go-template/kit/logger"
	"net/http"
	"strings"
	"sync"
)

type Options struct {
	Ctx     context.Context
	TraceID string
	Server  *http.ServeMux
	Wg      *sync.WaitGroup
	Cfg     config.IConfig
	Logger  logger.ILogger
}

const (
	ckHealthLive    = "server.health.endpoints.live"
	ckHealthReady   = "server.health.endpoints.ready"
	ckHealthStartup = "server.health.endpoints.startup"
)

func Initializer(ao Options) error {
	if err := Health(ao); err != nil {
		return err
	}

	return nil
}

func Health(ao Options) error {
	handler := health.NewHandler(ao.Logger)

	routes := []struct {
		key     string
		handler http.HandlerFunc
	}{
		{key: ckHealthLive, handler: handler.Live},
		{key: ckHealthReady, handler: handler.Ready},
		{key: ckHealthStartup, handler: handler.Startup},
	}

	for _, route := range routes {
		path := ao.Cfg.GetString(route.key)
		if path == "" || !strings.HasPrefix(path, "/") {
			return fmt.Errorf("invalid health endpoint path for %s: %q", route.key, path)
		}

		ao.Server.HandleFunc("GET "+path, route.handler)
	}

	handler.MarkStartupReady()
	return nil
}
