package checks

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

type postgresqlDatabase struct {
	name      string
	db        *sql.DB
	version   VersionReader
	versionMu sync.Mutex
	loaded    bool
	cached    string
}

// NewPostgreSQLDatabase creates a readiness check for a PostgreSQL database.
func NewPostgreSQLDatabase(name string, db *sql.DB) Check {
	return &postgresqlDatabase{
		name:    name,
		db:      db,
		version: NewSQLVersionReader(db),
	}
}

func (c *postgresqlDatabase) Name() string {
	return c.name
}

func (c *postgresqlDatabase) Type() string {
	return "database"
}

func (c *postgresqlDatabase) Technology() string {
	return "postgresql"
}

func (c *postgresqlDatabase) Check(ctx context.Context) Result {
	result := Result{
		Name:       c.Name(),
		Status:     StatusFail,
		Type:       c.Type(),
		Technology: c.Technology(),
	}

	if c.db == nil {
		result.Error = errors.New("postgresql database client is nil")
		return result
	}
	if err := c.db.PingContext(ctx); err != nil {
		result.Error = err
		return result
	}

	result.Status = StatusOK
	if c.version != nil {
		c.versionMu.Lock()
		if !c.loaded {
			if version, err := c.version(ctx); err == nil {
				c.cached = version
				c.loaded = true
			}
		}
		c.versionMu.Unlock()
		result.Version = c.cached
	}

	return result
}
