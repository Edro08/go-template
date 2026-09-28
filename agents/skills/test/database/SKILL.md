# Go Database Test Skill

Use this skill when testing Go code that interacts with SQL databases through
`database/sql`.

## General Rules

- Use `github.com/DATA-DOG/go-sqlmock` for SQL database tests.
- Do not connect to a real database in unit tests.
- Create the mock database at the test-function level.
- Close the database with `defer db.Close()`.
- Use `t.Context()` for test contexts.
- Configure SQL expectations inside each table-driven test case.
- Call `mock.ExpectationsWereMet()` after the operation under test.
- Do not use `t.Parallel()` when subtests share one mock database.
- Verify SQL statements, arguments, results, affected rows, and errors.
- Test both successful and failing database operations.
- Keep database resources local to the test that owns them.

## SELECT

Cover the relevant cases:

- One matching row.
- Multiple rows.
- No rows returned.
- Query execution error.
- Row scanning or mapping error.
- Context cancellation or timeout.

Use `ExpectQuery` with `WithArgs` and `WillReturnRows`:

```go
mock.ExpectQuery(regexp.QuoteMeta(
    "SELECT id, name FROM users WHERE id = ?",
)).
    WithArgs(1).
    WillReturnRows(
        sqlmock.NewRows([]string{"id", "name"}).
            AddRow(1, "Alice"),
    )
```

## INSERT

Cover the relevant cases:

- Successful insert.
- Correct argument order and values.
- Returned ID when applicable.
- Affected rows.
- Constraint or validation error.
- Execution error.

```go
mock.ExpectExec(regexp.QuoteMeta(
    "INSERT INTO users (name) VALUES (?)",
)).
    WithArgs("Alice").
    WillReturnResult(sqlmock.NewResult(1, 1))
```

## UPDATE

Cover the relevant cases:

- Successful update.
- Correct argument order and values.
- Expected affected rows.
- No rows affected.
- Execution error.

```go
mock.ExpectExec(regexp.QuoteMeta(
    "UPDATE users SET name = ? WHERE id = ?",
)).
    WithArgs("Alice", 1).
    WillReturnResult(sqlmock.NewResult(0, 1))
```

## DELETE

Cover the relevant cases:

- Successful deletion.
- Correct parameters.
- Expected affected rows.
- No rows affected.
- Execution error.

```go
mock.ExpectExec(regexp.QuoteMeta(
    "DELETE FROM users WHERE id = ?",
)).
    WithArgs(1).
    WillReturnResult(sqlmock.NewResult(0, 1))
```

## Transactions

Verify the complete transaction lifecycle:

- `ExpectBegin` is configured.
- Expected SQL operations are executed.
- `ExpectCommit` is configured on success.
- `ExpectRollback` is configured when an operation fails.
- Commit and rollback errors are covered.

```go
mock.ExpectBegin()
mock.ExpectExec(regexp.QuoteMeta(query)).
    WithArgs(args...).
    WillReturnResult(sqlmock.NewResult(1, 1))
mock.ExpectCommit()
```

For failure paths:

```go
mock.ExpectBegin()
mock.ExpectExec(regexp.QuoteMeta(query)).
    WillReturnError(errDatabase)
mock.ExpectRollback()
```

## Table-Driven Tests

Prefer a setup function for related SQL cases:

```go
tests := []struct {
    name  string
    setup func(sqlmock.Sqlmock)
    want  Result
    err   error
}{}
```

Create the repository with the mock database, execute the operation with
`t.Context()`, assert the result and error, and verify SQL expectations.

## Avoid

- Real database connections in unit tests.
- Mocking repository methods instead of verifying SQL behavior.
- Ignoring `ExpectationsWereMet`.
- Omitting argument assertions.
- Testing only the happy path.
- Sharing one mock database between parallel subtests.
- Using `reflect.DeepEqual` to compare functions or callbacks.

## Verification

Run:

```bash
go test ./...
go test ./... -coverprofile=coverage.out
```

Remove `coverage.out` after reporting coverage. Do not overwrite an existing
user-owned coverage file.
