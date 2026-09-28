# Architecture Specification

## Purpose

This project uses a domain-oriented architecture with explicit separation between
application use cases, domain rules, infrastructure adapters, and process
composition.

The architecture must keep business rules independent from transport protocols,
storage technologies, external services, and framework details.

## Project Structure

```text
cmd/
├── main.go
└── app/

internal/
└── example/
    ├── application/
    │   └── usecase/
    ├── domain/
    │   ├── model/
    │   └── ports/
    └── infrastructure/
        ├── handler/
        ├── messaging/
        ├── repository/
        └── storage/

kit/
├── config/
├── constants/
├── logger/
└── utils/

agents/
├── prompts/
└── specs/
```

`example` is a placeholder for a bounded context or business module. It should
be replaced by a meaningful domain name when a real module is created.

## Package Responsibilities

### `cmd`

Contains executable entry points and process lifecycle management.

#### `cmd/main.go`

Responsible for:

- Creating the root context.
- Handling operating system signals.
- Loading configuration and logging.
- Starting and stopping the HTTP server.
- Coordinating graceful shutdown.

It must not contain business rules or concrete domain use cases.

#### `cmd/app`

Acts as the composition root of the application. It is responsible for wiring
interfaces to their concrete implementations.

It may:

- Create storage, repository, and messaging adapters.
- Create application use cases.
- Create handlers.
- Register routes and consumers.
- Configure lifecycle dependencies.

It must not implement business rules. Its role is construction and assembly.

### `internal/<context>/application/usecase`

Contains application use cases and orchestration logic.

This package is responsible for:

- Coordinating domain operations.
- Defining application workflows.
- Calling domain services and ports.
- Managing transaction or workflow boundaries when required.
- Translating application input into domain operations.
- Returning application results and errors.

Use cases may depend on domain models and port interfaces, but must not depend
on concrete infrastructure implementations.

Use cases must not contain protocol-specific request or response types. HTTP,
SOAP, GraphQL, and messaging DTOs belong to their respective handlers or
adapters.

### `internal/<context>/domain/model`

Contains the business model and rules.

This package may contain:

- Entities.
- Value objects.
- Aggregates.
- Domain services.
- Domain events.
- Domain-specific errors.
- Invariants and validation rules.

The domain model must not import infrastructure, handlers, repositories,
storage implementations, or transport libraries.

### `internal/<context>/domain/ports`

Contains interfaces required by the domain or application layer.

Examples include ports for:

- Loading or saving domain data.
- Calling external capabilities.
- Publishing events.
- Reading files or objects.
- Sending notifications.

Ports define contracts only. They must not contain concrete clients, SQL,
HTTP code, SDK calls, or serialization details.

Interfaces should be declared close to the code that consumes them. If a port
is required by a use case rather than by the domain model, it may be placed in
the application layer when that makes ownership clearer.

### `internal/<context>/infrastructure/handler`

Contains inbound adapters that expose application capabilities.

Examples include:

```text
handler/
├── http/
├── soap/
└── graphql/
```

Handlers are responsible for:

- Reading protocol-specific input.
- Validating transport-level data.
- Mapping requests to use case input.
- Calling use cases.
- Mapping results and errors to protocol responses.
- Handling protocol concerns such as status codes, headers, and serialization.

Handlers must not access storage or repositories directly and must not contain
business rules.

### `internal/<context>/infrastructure/repository`

Contains outbound adapters for remote services and external APIs.

Examples include:

```text
repository/
├── rest/
├── soap/
└── graphql/
```

These adapters may contain:

- HTTP or SOAP clients.
- GraphQL queries.
- External authentication.
- Retry and timeout configuration.
- External request and response DTOs.
- Mapping between external contracts and internal models.

Repositories implement ports and hide external service details from the
application and domain layers.

In this project, `repository` refers to integrations with remote services. It
does not mean that every repository must persist data in a database.

### `internal/<context>/infrastructure/storage`

Contains adapters for storing or retrieving data and objects.

Examples include:

```text
storage/
├── database/
├── s3/
├── sftp/
└── filesystem/
```

Storage adapters may contain:

- SQL queries and database drivers.
- S3 clients and object operations.
- SFTP clients.
- Filesystem operations.
- Storage-specific serialization.
- Connection and transaction handling.

Storage implementations must satisfy ports and must not expose technology-
specific details to the domain or application layers.

### `internal/<context>/infrastructure/messaging`

Contains messaging adapters.

This package may contain:

- Message producers.
- Message consumers.
- Broker clients.
- Message serialization and deserialization.
- Consumer lifecycle management.
- Mapping between external messages and application commands or domain events.

Messaging adapters must not implement business rules. Consumers should invoke
application use cases instead.

## Shared Packages

### `kit/config`

Provides configuration loading and typed access to configuration values.

It must remain independent from business modules and must not import packages
under `internal`.

### `kit/logger`

Provides structured logging and logger configuration.

It must not contain business logic or depend on a specific bounded context.

### `kit/constants`

Contains stable constants shared across technical packages. It must not become
a general-purpose package for arbitrary domain values.

Domain-specific constants belong to the corresponding domain package.

### `kit/utils`

Contains small, reusable technical utilities such as UUID or time helpers.

Utilities must remain generic. Business-specific helpers belong to the owning
domain or application package.

## Dependency Rules

The intended dependency direction is:

```text
cmd/app
├── application/usecase
├── infrastructure/handler
├── infrastructure/repository
├── infrastructure/storage
└── infrastructure/messaging

application/usecase
└── domain

infrastructure
└── domain/ports

domain
└── no infrastructure dependency
```

The following rules are mandatory:

- The domain must not import infrastructure, `cmd`, or transport packages.
- Use cases must depend on interfaces, not concrete adapters.
- Infrastructure implementations must satisfy ports.
- Handlers must call use cases and must not access storage directly.
- Repositories and storage adapters must not call handlers.
- `cmd/app` is the only place where concrete implementations are assembled.
- `kit` packages must not import `internal` packages.
- External SDKs must be isolated inside infrastructure adapters whenever
  possible.
- Circular dependencies must be resolved by moving interfaces to the consuming
  layer or by introducing a domain/application port.

## Initialization Flow

The application should be assembled in this order:

```text
main
├── create root context
├── load configuration and logging
└── cmd/app initializer
    ├── create infrastructure adapters
    ├── create use cases
    ├── create handlers and consumers
    ├── register routes and message handlers
    └── return the running application components
```

The initializer should fail fast when a required dependency cannot be created.
It should not hide initialization errors or silently replace required
dependencies with nil values.

## Protocol Direction

The location of REST, SOAP, and GraphQL code depends on communication
direction:

- `infrastructure/handler`: the application exposes the protocol.
- `infrastructure/repository`: the application consumes an external service
  through the protocol.

For example:

```text
infrastructure/handler/http       # application exposes HTTP endpoints
infrastructure/handler/soap       # application exposes SOAP operations
infrastructure/handler/graphql    # application exposes GraphQL

infrastructure/repository/rest    # client for an external REST API
infrastructure/repository/soap    # client for an external SOAP service
infrastructure/repository/graphql # client for an external GraphQL API
```

## Naming and Implementation Conventions

- Prefer names based on business concepts over names based only on technology.
- Keep external DTOs inside their adapter package.
- Map external DTOs to internal models at the infrastructure boundary.
- Keep transport validation separate from domain validation.
- Return explicit errors from adapters and use cases.
- Do not place SQL, HTTP, SOAP, GraphQL, or broker code in domain packages.
- Add tests for domain rules and use cases independently from infrastructure.
- Add adapter tests for external contract mapping and failure handling.
