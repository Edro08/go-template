package checks

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewPostgreSQLDatabase(t *testing.T) {
	type args struct {
		name string
		db   *sql.DB
	}
	tests := []struct {
		name string
		args args
		want Check
	}{
		{
			name: "creates postgresql database check",
			args: args{name: "database-primary"},
			want: &postgresqlDatabase{name: "database-primary"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NewPostgreSQLDatabase(tt.args.name, tt.args.db).(*postgresqlDatabase)
			want := tt.want.(*postgresqlDatabase)
			if !ok || got.name != want.name || got.db != want.db || got.version == nil {
				t.Errorf("NewPostgreSQLDatabase() = %#v, want name %q and version reader", got, want.name)
			}
		})
	}
}

func Test_postgresqlDatabase_Check(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	databaseUnavailable := errors.New("database unavailable")
	type fields struct {
		name      string
		db        *sql.DB
		version   VersionReader
		versionMu sync.Mutex
		loaded    bool
		cached    string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		setup  func(sqlmock.Sqlmock)
		want   Result
	}{
		{
			name: "ping succeeds and version is loaded",
			fields: fields{
				name: "database-primary",
				version: func(context.Context) (string, error) {
					return "16.4", nil
				},
			},
			args: args{ctx: context.Background()},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			want: Result{
				Name: "database-primary", Status: StatusOK, Type: "database",
				Technology: "postgresql", Version: "16.4",
			},
		},
		{
			name:   "ping fails",
			fields: fields{name: "database-primary"},
			args:   args{ctx: context.Background()},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(databaseUnavailable)
			},
			want: Result{
				Name: "database-primary", Status: StatusFail, Type: "database",
				Technology: "postgresql", Error: databaseUnavailable,
			},
		},
		{
			name:   "database client is nil",
			fields: fields{name: "database-primary"},
			args:   args{ctx: context.Background()},
			want: Result{
				Name: "database-primary", Status: StatusFail, Type: "database",
				Technology: "postgresql", Error: errors.New("postgresql database client is nil"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(mock)
			}
			c := &postgresqlDatabase{
				name:      tt.fields.name,
				db:        tt.fields.db,
				version:   tt.fields.version,
				versionMu: tt.fields.versionMu,
				loaded:    tt.fields.loaded,
				cached:    tt.fields.cached,
			}
			if c.db == nil && tt.setup != nil {
				c.db = db
			}
			if got := c.Check(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
			if tt.setup != nil {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("sqlmock expectations error = %v", err)
				}
			}
		})
	}
}

func Test_postgresqlDatabase_Name(t *testing.T) {
	type fields struct {
		name      string
		db        *sql.DB
		version   VersionReader
		versionMu sync.Mutex
		loaded    bool
		cached    string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "returns configured name", fields: fields{name: "database-primary"}, want: "database-primary"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &postgresqlDatabase{
				name:      tt.fields.name,
				db:        tt.fields.db,
				version:   tt.fields.version,
				versionMu: tt.fields.versionMu,
				loaded:    tt.fields.loaded,
				cached:    tt.fields.cached,
			}
			if got := c.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_postgresqlDatabase_Technology(t *testing.T) {
	type fields struct {
		name      string
		db        *sql.DB
		version   VersionReader
		versionMu sync.Mutex
		loaded    bool
		cached    string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "returns postgresql technology", fields: fields{}, want: "postgresql"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &postgresqlDatabase{
				name:      tt.fields.name,
				db:        tt.fields.db,
				version:   tt.fields.version,
				versionMu: tt.fields.versionMu,
				loaded:    tt.fields.loaded,
				cached:    tt.fields.cached,
			}
			if got := c.Technology(); got != tt.want {
				t.Errorf("Technology() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_postgresqlDatabase_Type(t *testing.T) {
	type fields struct {
		name      string
		db        *sql.DB
		version   VersionReader
		versionMu sync.Mutex
		loaded    bool
		cached    string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "returns database type", fields: fields{}, want: "database"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &postgresqlDatabase{
				name:      tt.fields.name,
				db:        tt.fields.db,
				version:   tt.fields.version,
				versionMu: tt.fields.versionMu,
				loaded:    tt.fields.loaded,
				cached:    tt.fields.cached,
			}
			if got := c.Type(); got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
