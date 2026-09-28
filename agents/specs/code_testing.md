# Code Testing

- **Mocks tests**:
    - For function or interface dependencies invoked directly by the code under test, use the project's existing mocks or `testify/mock` when it fits the package's patterns.  
    - For REST or other HTTP integrations, use `net/http/httptest` to run a test server and verify requests, responses, status codes, and error handling. Avoid mocking the HTTP client when a test server can exercise the real request flow.  
- **Table-driven tests**: Use table-driven tests for related cases when they improve clarity and reduce duplication; use separate tests when that is clearer.  
- **Package placement**: Place tests in the same package as the code under test by default. Use an external `_test` package when testing the public API or when isolation requires it.  
- **Context**: Use `t.Context()` for test contexts. Add context values only when they are required by the behavior under test.  
- **Naming**: Test functions should follow the `Test_<Receiver>_<MethodName>` or `Test<FunctionName>` convention.  
- **Validation**: Run both commands after making changes:  
    - `go test ./...` to verify that tests pass across all packages.  
    - `go test ./... -coverprofile=coverage.out` to generate coverage data and verify the coverage target.  
- **Coverage artifacts**: If the coverage command creates `coverage.out`, remove that generated file before finishing. Do not delete or overwrite a pre-existing or user-owned file. Report the total coverage and any packages or code that prevent reaching 100%.  
- **Reporting**: Report the commands run, their results, the measured coverage, and any validation that could not be completed.