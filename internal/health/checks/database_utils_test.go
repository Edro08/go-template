package checks

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSQLVersionReader(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	tests := []struct {
		name    string
		setup   func(sqlmock.Sqlmock)
		want    string
		wantErr string
	}{
		{
			name: "returns database version",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT VERSION()")).
					WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.36"))
			},
			want: "8.0.36",
		},
		{
			name: "returns query error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT VERSION()")).
					WillReturnError(sqlmock.ErrCancelled)
			},
			wantErr: sqlmock.ErrCancelled.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(mock)
			version, err := NewSQLVersionReader(db)(t.Context())
			if version != tt.want {
				t.Errorf("version = %q, want %q", version, tt.want)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("reader() error = %v, want nil", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Errorf("reader() error = %v, want %q", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock expectations error = %v", err)
			}
		})
	}

	t.Run("nil database", func(t *testing.T) {
		_, err := NewSQLVersionReader(nil)(t.Context())
		if err == nil || err.Error() != "sql querier is nil" {
			t.Errorf("reader() error = %v, want sql querier is nil", err)
		}
	})
}
