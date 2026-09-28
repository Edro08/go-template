# Code Rules

Before writing or modifying Go code, read and follow `.golangci.yaml`.
Treat its enabled linters and settings as the source of truth; do not assume rules that are not configured.  
  
- Follow the project's formatting and linting requirements.  
- Keep functions within the configured limits for length, complexity, arguments, and return values.  
- Do not disable, bypass, or weaken a linter rule merely to make code pass. If a rule appears incompatible with the requested change, explain the issue and propose the smallest appropriate resolution.  
- Run the linter after code changes, when the tool is available:  
  
  ```bash  
    golangci-lint run 
  ```
  
- If the linter reports issues, fix those introduced by the change. Report any remaining relevant issues and explain if the linter could not be run.
