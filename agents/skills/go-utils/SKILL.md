---
name: go-utils
description: Usage guidelines and integration patterns for the shared go-utils library (config, logger, cors, jwt, and cipher) across microservices.
---

# go-utils Integration Skill

Use this skill whenever you need to integrate, configure, or use utilities from the shared library `github.com/edro08/go-utils` in this microservice (`go-template`).

## General Principles

- **Shared dependency**: `go-utils` is the foundational cross-cutting library for all microservices. Do not duplicate its functionality inside `kit/` or internal packages.
- **Interface injection**: Whenever a component requires configuration or logging, inject the `config.IConfig` and `logger.ILogger` interfaces instead of concrete struct references.
- **Composition Root initialization**: Instantiation and setup of configurations, loggers, and CORS middlewares must occur in the Composition Root (`cmd/main.go` or `cmd/app/app.go`).

---

## 1. Configuration (`config`)

Import: `"github.com/edro08/go-utils/config"`

### Initialization and Loading

```go
cfg := config.New(config.Options{
    Separator: ".", // optional, defaults to "."
})

// Load from YAML file
if err := cfg.LoadFile("./config.yaml"); err != nil {
    // handle error
}
```

### Accessing Values

Access is thread-safe and supports dot-separated nested key traversal:

```go
// Primitive values
host := cfg.GetString("server.host")
port := cfg.GetInt("server.port")
enabled := cfg.GetBool("features.metrics.enabled")
timeout := cfg.GetFloat("server.timeout")

// Maps and slices
endpoints := cfg.GetMapString("server.health.endpoints")
allowedHosts := cfg.GetSliceString("cors.allowedOrigins")
```

### Interface Contract (`IConfig`)

Use `config.IConfig` in structs and function signatures that consume configuration:

```go
type Options struct {
    Cfg config.IConfig
}
```

---

## 2. Structured Logging (`logger`)

Import: `"github.com/edro08/go-utils/logger"`

### Initialization

```go
log, err := logger.New(logger.Options{
    Output:   logger.Stdout,     // Stdout, File, or Both
    Format:   logger.FormatJSON, // FormatJSON or FormatText
    MinLevel: logger.INFO,       // DEBUG, INFO, WARN, ERROR, FATAL
    FilePath: "./logs/app.log",   // Required if Output is File or Both
})
if err != nil {
    // handle error
}
defer func() {
    _ = log.Close()
}()
```

### Logging Entries

Methods accept a `title` (category or component tag) and key-value arguments in pairs:

```go
log.Info("SERVER", "status", "listening", "port", 8080)
log.Warn("STORAGE", "retry_attempt", 2, "error", err.Error())
log.Error("DATABASE", "error", err.Error(), "traceID", traceID)
log.Debug("PAYLOAD", "body", string(rawPayload))

// Fatal logs with FATAL severity and executes os.Exit(1)
log.Fatal("MAIN", "error", "critical startup failure")
```

### Contextual Attributes with `With` and `WithGroup`

```go
// Create a derived logger with persistent contextual attributes
requestLogger := log.With("traceID", traceID, "userID", userID)
requestLogger.Info("REQUEST", "endpoint", "/api/v1/users")

// Create a derived logger with grouped attributes
dbLogger := log.WithGroup("database")
```

### Interface Contract (`ILogger`)

Always inject `logger.ILogger` into handlers, use cases, and adapters:

```go
type Handler struct {
    logger logger.ILogger
}

func NewHandler(log logger.ILogger) *Handler {
    return &Handler{logger: log}
}
```

---

## 3. CORS Middleware (`cors`)

Import: `"github.com/edro08/go-utils/cors"`

### Configuration and Middleware

```go
c, err := cors.NewCors(cors.Options{
    AllowedOrigins:   []string{"https://my-domain.com", "http://localhost:3000"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Trace-ID"},
    ExposedHeaders:   []string{"X-Total-Count"},
    AllowCredentials: true,
    MaxAge:           3600,
})
if err != nil {
    // handle configuration error
}

// Wrap ServeMux or HTTP router:
httpServer := &http.Server{
    Addr:    ":8080",
    Handler: c.Middleware(mux),
}
```

### Interface Contract (`CorsMiddleware`)

If injecting the CORS handler:

```go
type Server struct {
    cors cors.CorsMiddleware
}
```

---

## 4. JSON Web Tokens (`jwt`)

Import: `"github.com/edro08/go-utils/jwt"`

### Generation and Signing

```go
token, err := jwt.New(jwt.NewTokenOpts{
    ID:             "token-uuid",
    Algorithm:      jwt.HS256,
    Issuer:         "auth-service",
    Subject:        "user-12345",
    ExpiredEnabled: true,
    ExpiredTime:    2 * time.Hour,
    CustomClaims: map[string]any{
        "role":  "admin",
        "email": "user@example.com",
    },
})
if err != nil {
    // handle error
}

// Sign token using symmetric or asymmetric key
signedString, err := token.Sign([]byte("my-secret-key"))
```

### Verification and Parsing

```go
// Verify signature and algorithm
if err := jwt.Verify(signedString, jwt.HS256, []byte("my-secret-key")); err != nil {
    // invalid or expired token
}

// Extract claims without signature verification
claims, err := jwt.Parse(signedString)
if err != nil {
    // error parsing claims
}
role := claims["role"].(string)
```

---

## 5. Cryptography (`cipher`)

Direct utility invocation without internal state or anchored keys.

### AES-GCM (Recommended Authenticated Symmetric Encryption)

Import: `"github.com/edro08/go-utils/cipher/aesgcm"`

```go
key := []byte("32-byte-secret-key-for-aes256!!") // 16, 24, or 32 bytes

// Encrypt
ciphertext, err := aesgcm.Encrypt(key, []byte("plaintext payload"))

// Decrypt
plaintext, err := aesgcm.Decrypt(key, ciphertext)
```

### AES-CBC

Import: `"github.com/edro08/go-utils/cipher/aescbc"`

```go
key := []byte("32-byte-secret-key-for-aes256!!")
iv := []byte("16-byte-init-vec")

ciphertext, err := aescbc.Encrypt(key, iv, []byte("plaintext payload"))
plaintext, err := aescbc.Decrypt(key, iv, ciphertext)
```

### RSA (Asymmetric)

Import: `"github.com/edro08/go-utils/cipher/rsa"`

```go
// Parse public key from PEM bytes or file
pubKey, err := rsa.ParsePublicKeyFromPEM(publicKeyPEM)
// or: pubKey, err := rsa.ParsePublicKeyFromPEMFile("./keys/public.pem")

// Encrypt with RSA public key
ciphertext, err := rsa.Encrypt(pubKey, []byte("secret message"))

// Parse private key and decrypt
privKey, err := rsa.ParsePrivateKeyFromPEM(privateKeyPEM)
// or: privKey, err := rsa.ParsePrivateKeyFromPEMFile("./keys/private.pem")

plaintext, err := rsa.Decrypt(privKey, ciphertext)
```
