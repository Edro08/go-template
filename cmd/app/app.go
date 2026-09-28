package app

import (
	"context"
	"go-template/kit/config"
	"go-template/kit/logger"
	"net/http"
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

func Initializer(ao Options) error {
	return nil
}
