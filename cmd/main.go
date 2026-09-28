package main

import (
	"context"
	"errors"
	"fmt"
	"go-template/cmd/app"
	"go-template/kit/config"
	"go-template/kit/constants"
	"go-template/kit/logger"
	"go-template/kit/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	tMain = "MAIN"

	ckServerPort              = "server.port"
	ckServerName              = "server.name"
	ckServerReadHeaderTimeout = "server.readHeaderTimeout"
	ckServerReadTimeout       = "server.readTimeout"
	ckServerWriteTimeout      = "server.writeTimeout"
	ckServerIdleTimeout       = "server.idleTimeout"
	ckServerShutdownTimeout   = "server.shutdown.timeout"
)

type serverContext struct {
	ctx       context.Context
	cancel    context.CancelFunc
	traceID   string
	serveMux  *http.ServeMux
	server    *http.Server
	serverErr chan error
	wg        *sync.WaitGroup
	cfg       config.IConfig
	logger    logger.ILogger
}

func main() {
	// 1. Start
	// ____________________________________________________________________________________________________
	var sc serverContext
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	appCtx, cancel := context.WithCancel(signalCtx)
	defer cancel()

	sc.traceID = utils.NewUUID()
	sc.ctx = context.WithValue(appCtx, constants.LogTraceID, sc.traceID)
	sc.cancel = cancel

	sc.serveMux = http.NewServeMux()
	sc.wg = &sync.WaitGroup{}

	loadLogger(&sc)
	defer sc.logger.Close()
	loadConfigurations(&sc)

	// 2. Apps
	// ____________________________________________________________________________________________________
	ao := app.Options{
		Ctx:     sc.ctx,
		TraceID: sc.traceID,
		Server:  sc.serveMux,
		Wg:      sc.wg,
		Cfg:     sc.cfg,
		Logger:  sc.logger,
	}

	if err := app.Initializer(ao); err != nil {
		sc.logger.Fatal(tMain, constants.LogStatus, constants.LogFailure, constants.LogError, err.Error(), constants.LogTraceID, sc.traceID)
	}

	// 3. Server Run and Stop
	// ____________________________________________________________________________________________________
	serverStart(&sc)
	serverStop(&sc)
}

func loadLogger(sc *serverContext) {
	l, err := logger.New(logger.Options{})
	if err != nil {
		log.Fatal(tMain, constants.LogStatus, constants.LogFailure, constants.LogError, err.Error(), constants.LogTraceID, sc.traceID)
	}
	sc.logger = l
}

func loadConfigurations(sc *serverContext) {
	cfg := config.New(config.Options{})
	if err := cfg.LoadFile("./config.yaml"); err != nil {
		sc.logger.Fatal(tMain, constants.LogStatus, constants.LogFailure, constants.LogError, err.Error(), constants.LogTraceID, sc.traceID)
	}

	sc.cfg = cfg
}

func serverStart(sc *serverContext) {
	port := sc.cfg.GetString(ckServerPort)
	serverName := sc.cfg.GetString(ckServerName)
	if _, err := strconv.Atoi(port); err != nil {
		sc.logger.Fatal(tMain, constants.LogStatus, constants.LogFailure, constants.LogError, err.Error(), constants.LogTraceID, sc.traceID)
	}

	sc.server = &http.Server{
		Addr:              ":" + port,
		Handler:           sc.serveMux,
		ReadHeaderTimeout: sc.getDurationWithDefault(ckServerReadHeaderTimeout, 5*time.Second),
		ReadTimeout:       sc.getDurationWithDefault(ckServerReadTimeout, 15*time.Second),
		WriteTimeout:      sc.getDurationWithDefault(ckServerWriteTimeout, 30*time.Second),
		IdleTimeout:       sc.getDurationWithDefault(ckServerIdleTimeout, 60*time.Second),
	}

	sc.serverErr = make(chan error, 1)
	go func() {
		if err := sc.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sc.serverErr <- err
		}
	}()

	sc.logger.Info(tMain, constants.LogStatus, fmt.Sprintf("Start server %s in port %s", serverName, port), constants.LogTraceID, sc.traceID)
}

func serverStop(sc *serverContext) {
	select {
	case <-sc.ctx.Done():
		sc.logger.Info(tMain, constants.LogStatus, "termination signal received via context", constants.LogTraceID, sc.traceID)
		break
	case err := <-sc.serverErr:
		sc.logger.Error(tMain, constants.LogStatus, constants.LogFailure, constants.LogError, err.Error(), constants.LogTraceID, sc.traceID)
		break
	}

	// Notify workers and other components that shutdown has started.
	sc.cancel()

	timeout := sc.getDurationWithDefault(ckServerShutdownTimeout, 30*time.Second)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	allDone := make(chan struct{})
	go func() {
		var internalWg sync.WaitGroup

		// Sub-tarea A: Apagar Servidor HTTP
		internalWg.Add(1)
		go func() {
			defer internalWg.Done()
			sc.server.SetKeepAlivesEnabled(false)
			if err := sc.server.Shutdown(shutdownCtx); err != nil {
				sc.logger.Error(tMain, "http server shutdown error", constants.LogError, err.Error(), constants.LogTraceID, sc.traceID)
				_ = sc.server.Close()
			}
		}()

		// Sub-tarea B: Esperar WaitGroup genérico (Bots, workers, etc.)
		internalWg.Add(1)
		go func() {
			defer internalWg.Done()
			sc.wg.Wait()
		}()

		sc.logger.Info(tMain, constants.LogStatus, "waiting for all components to close...", constants.LogTraceID, sc.traceID)
		internalWg.Wait()
		close(allDone)
	}()

	select {
	case <-allDone:
		sc.logger.Info(tMain, constants.LogStatus, "all components shut down cleanly", constants.LogTraceID, sc.traceID)
	case <-shutdownCtx.Done():
		sc.logger.Warn(tMain, constants.LogStatus, "shutdown limit reached", "timeout", timeout.String(), constants.LogTraceID, sc.traceID)
	}

	sc.logger.Info(tMain, constants.LogStatus, "Application shutdown finished", constants.LogTraceID, sc.traceID)
}

func (sc *serverContext) getDurationWithDefault(key string, df time.Duration) time.Duration {
	if val := sc.cfg.GetString(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}

	return df
}
