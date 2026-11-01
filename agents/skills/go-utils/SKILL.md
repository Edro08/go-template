---
name: go-utils
description: Guía de uso y patrones de integración de la librería compartida go-utils (config, logger, cors, jwt y cipher) en microservicios.
---

# go-utils Integration Skill

Usa esta skill cuando necesites integrar, configurar o utilizar utilidades de la librería `github.com/edro08/go-utils` en el microservicio (`go-template`).

## Principios Generales

- **Dependencia compartida**: `go-utils` es la librería base transversal para todos los microservicios. No dupliques su funcionalidad dentro de `kit/` ni en paquetes internos.
- **Inyección por interfaces**: Siempre que un componente requiera configuración o registro de logs, inyecta las interfaces `config.IConfig` y `logger.ILogger` en lugar de referencias a structs concretos.
- **Inicialización en el Composition Root**: La creación e instanciación de configuraciones, loggers y middlewares CORS debe realizarse en `cmd/main.go` o `cmd/app/app.go`.

---

## 1. Configuración (`config`)

Import: `"github.com/edro08/go-utils/config"`

### Inicialización y Carga

```go
cfg := config.New(config.Options{
    Separator: ".", // opcional, por defecto es "."
})

// Cargar desde archivo YAML
if err := cfg.LoadFile("./config.yaml"); err != nil {
    // manejar error
}
```

### Acceso a Valores

El acceso es seguro contra concurrencia y soporta navegación por claves anidadas separadas por punto:

```go
// Valores primitivos
host := cfg.GetString("server.host")
port := cfg.GetInt("server.port")
enabled := cfg.GetBool("features.metrics.enabled")
timeout := cfg.GetFloat("server.timeout")

// Mapas y slices
endpoints := cfg.GetMapString("server.health.endpoints")
allowedHosts := cfg.GetSliceString("cors.allowedOrigins")
```

### Contrato de Interfaz (`IConfig`)

Usa `config.IConfig` como tipo en structs y firmas de funciones que consuman configuración:

```go
type Options struct {
    Cfg config.IConfig
}
```

---

## 2. Structured Logging (`logger`)

Import: `"github.com/edro08/go-utils/logger"`

### Inicialización

```go
log, err := logger.New(logger.Options{
    Output:   logger.Stdout,   // Stdout, File, o Both
    Format:   logger.FormatJSON, // FormatJSON o FormatText
    MinLevel: logger.INFO,       // DEBUG, INFO, WARN, ERROR, FATAL
    FilePath: "./logs/app.log", // Requerido si Output es File o Both
})
if err != nil {
    // manejar error
}
defer func() {
    _ = log.Close()
}()
```

### Escritura de Logs

Los métodos aceptan un `title` (categoría o componente) y argumentos clave-valor en pares:

```go
log.Info("SERVER", "status", "listening", "port", 8080)
log.Warn("STORAGE", "retry_attempt", 2, "error", err.Error())
log.Error("DATABASE", "error", err.Error(), "traceID", traceID)
log.Debug("PAYLOAD", "body", string(rawPayload))

// Fatal escribe el log con severidad FATAL y ejecuta os.Exit(1)
log.Fatal("MAIN", "error", "critical startup failure")
```

### Contextualización con `With` y `WithGroup`

```go
// Crear logger derivado con atributos persistentes
requestLogger := log.With("traceID", traceID, "userID", userID)
requestLogger.Info("REQUEST", "endpoint", "/api/v1/users")

// Crear logger derivado con grupo de atributos
dbLogger := log.WithGroup("database")
```

### Contrato de Interfaz (`ILogger`)

Inyecta siempre `logger.ILogger` en tus handlers, casos de uso y adapters:

```go
type Handler struct {
    logger logger.ILogger
}

func NewHandler(log logger.ILogger) *Handler {
    return &Handler{logger: log}
}
```

---

## 3. Middleware CORS (`cors`)

Import: `"github.com/edro08/go-utils/cors"`

### Configuración y Middleware

```go
c, err := cors.NewCors(cors.Options{
    AllowedOrigins:   []string{"https://mi-dominio.com", "http://localhost:3000"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Trace-ID"},
    ExposedHeaders:   []string{"X-Total-Count"},
    AllowCredentials: true,
    MaxAge:           3600,
})
if err != nil {
    // manejar error de configuración
}

// Envolver el ServeMux o router HTTP:
httpServer := &http.Server{
    Addr:    ":8080",
    Handler: c.Middleware(mux),
}
```

---

## 4. JSON Web Tokens (`jwt`)

Import: `"github.com/edro08/go-utils/jwt"`

### Generación y Firma

```go
token, err := jwt.New(jwt.NewTokenOpts{
    ID:             "token-uuid",
    Algorithm:      jwt.HS256,
    Issuer:         "auth-service",
    Subject:        "user-12345",
    ExpiredEnabled: true,
    ExpiredTime:    2 * time.Hour,
    Claims: map[string]any{
        "role":  "admin",
        "email": "user@example.com",
    },
})
if err != nil {
    // manejar error
}

// Firmar el token con clave simétrica
signedString, err := token.Sign([]byte("mi-clave-secreta"))
```

### Verificación y Parseo

```go
// Verificar validez y firma
if err := jwt.Verify(signedString, jwt.HS256, []byte("mi-clave-secreta")); err != nil {
    // token inválido o expirado
}

// Extraer claims
claims, err := jwt.Parse(signedString)
if err != nil {
    // error parseando claims
}
role := claims["role"].(string)
```

---

## 5. Criptografía (`cipher`)

### AES-GCM (Cifrado autenticado simétrico recomendado)

Import: `"github.com/edro08/go-utils/cipher/aesgcm"`

```go
key := []byte("clave-de-32-bytes-para-aes-256!!") // 16, 24 o 32 bytes

// Cifrar
ciphertext, err := aesgcm.Encrypt(key, []byte("texto plano"))

// Descifrar
plaintext, err := aesgcm.Decrypt(key, ciphertext)
```

### AES-CBC

Import: `"github.com/edro08/go-utils/cipher/aescbc"`

```go
key := []byte("clave-de-32-bytes-para-aes-256!!")

ciphertext, err := aescbc.Encrypt(key, []byte("texto plano"))
plaintext, err := aescbc.Decrypt(key, ciphertext)
```

### RSA (Asimétrico)

Import: `"github.com/edro08/go-utils/cipher/rsa"`

```go
// Cifrar con clave pública PEM
encrypted, err := rsa.EncryptWithPublicKey(publicKeyPEM, []byte("mensaje secreto"))

// Descifrar con clave privada PEM
decrypted, err := rsa.DecryptWithPrivateKey(privateKeyPEM, encrypted)
```
