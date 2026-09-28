/*
Package logger provee un logger estructurado basado en la biblioteca estándar log/slog para aplicaciones escritas en Go.

Permite registrar mensajes en diferentes niveles, utilizar formato JSON o texto
y, opcionalmente, escribir los registros tanto en consola como en un archivo.

Uso básico:

```

	logger, err := logger.New(logger.Options{
		MinLevel: logger.INFO,
		Format:   logger.FormatJSON,
	})

	if err != nil {
		log.Fatal(err)
	}

logger.Info("server started", "port", 8080)
logger.Warn("high memory usage", "usage", "85%")
logger.Error("failed to process request", "error", err)
```

Para asociar atributos a un logger:

```
requestLogger := logger.With(

	"request_id", requestID,
	"user_id", userID,

)

requestLogger.Info("request processed")
```

Comportamiento:
  - Los registros por debajo de MinLevel son ignorados.
  - Los registros incluyen la ubicación del código donde se realizó la llamada.
  - Permite utilizar los niveles DEBUG, INFO, WARN, ERROR y FATAL.
  - FATAL registra el mensaje y termina inmediatamente el proceso con código 1.
  - Permite registrar atributos estructurados mediante pares clave-valor.
  - Permite asociar atributos permanentes mediante With.
  - Permite agrupar atributos mediante WithGroup.
  - Permite escribir los registros en formato JSON o texto.
  - Puede escribir simultáneamente en stdout y en un archivo.
  - Cuando se habilita el registro en archivo y no se especifica una ruta,
    se genera un archivo con timestamp dentro del directorio ./log.
*/
package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options define la configuración para inicializar un Logger.
type Options struct {
	MinLevel   Level
	Format     Format
	EnableFile bool
	FilePath   string
}

// Format define el formato de salida de los logs.
type Format string

const (
	FormatJSON Format = "JSON"
	FormatText Format = "TEXT"
)

// Level representa los niveles de log soportados.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger es el wrapper estructurado sobre log/slog.
type Logger struct {
	opts   Options
	slog   *slog.Logger
	closer io.Closer
}

// New crea e inicializa una nueva instancia de Logger basado en log/slog.
func New(opts Options) (*Logger, error) {
	validateOptions(&opts)

	w, closer, err := setupWriter(opts)
	if err != nil {
		return nil, err
	}

	handler := buildHandler(w, opts)

	return &Logger{
		opts:   opts,
		slog:   slog.New(handler),
		closer: closer,
	}, nil
}

// validateOptions ajusta valores por defecto para opciones no inicializadas.
func validateOptions(opts *Options) {
	if opts.Format != FormatJSON && opts.Format != FormatText {
		opts.Format = FormatJSON
	}
	if opts.MinLevel < DEBUG || opts.MinLevel > FATAL {
		opts.MinLevel = INFO
	}
}

// setupWriter configura la salida del logger (stdout y archivo si EnableFile es true).
func setupWriter(opts Options) (io.Writer, io.Closer, error) {
	if !opts.EnableFile {
		return os.Stdout, nil, nil
	}

	dir := opts.FilePath
	if dir == "" {
		dir = "./log"
	}

	var path string

	if strings.HasSuffix(dir, ".log") {
		path = dir
		dir = filepath.Dir(path)
	} else {
		timestamp := time.Now().Format("20060102T150405.000")
		path = filepath.Join(dir, timestamp+".log")
	}

	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, err
		}
	}

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		return nil, nil, err
	}

	return io.MultiWriter(os.Stdout, file), file, nil
}

// buildHandler genera el slog.Handler correspondiente (JSON o Text).
func buildHandler(w io.Writer, opts Options) slog.Handler {
	handlerOpts := &slog.HandlerOptions{
		Level:     opts.MinLevel.toSlog(),
		AddSource: true,
	}

	if opts.Format == FormatJSON {
		return slog.NewJSONHandler(w, handlerOpts)
	}
	return slog.NewTextHandler(w, handlerOpts)
}

func (l *Logger) Close() error {
	if l.closer == nil {
		return nil
	}

	return l.closer.Close()
}
