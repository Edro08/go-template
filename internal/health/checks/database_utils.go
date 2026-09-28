package checks

import (
	"context"
	"database/sql"
	"errors"
)

// VersionReader obtains the server version without exposing connection details.
type VersionReader func(context.Context) (string, error)

// NewSQLVersionReader creates an optional version reader for database/sql.
func NewSQLVersionReader(db *sql.DB) VersionReader {
	return func(ctx context.Context) (string, error) {
		if db == nil {
			return "", errors.New("sql querier is nil")
		}

		var version string
		if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
			return "", err
		}

		return version, nil
	}
}
