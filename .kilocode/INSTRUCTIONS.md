# Claude Code Guidelines — Go Edition
> Adapted from the original guidelines by Sabrina Ramonov

## Implementation Best Practices

### 0 — Purpose

These rules ensure maintainability, safety, and developer velocity.
**MUST** rules are enforced by CI; **SHOULD** rules are strongly recommended.

---

### 1 — Before Coding

- **BP-1 (MUST)** Ask the user clarifying questions when requirements are ambiguous or ≥ 2 fundamentally different approaches exist.
- **BP-2 (SHOULD)** Draft and confirm an approach for complex work.
- **BP-3 (SHOULD)** If ≥ 2 approaches exist, list clear pros and cons.

---

### 2 — While Coding

- **C-1 (MUST)** Follow TDD: scaffold stub → write failing test → implement.
- **C-2 (MUST)** Name functions and variables using existing domain vocabulary for consistency.
- **C-3 (SHOULD NOT)** Introduce structs with methods when small, testable functions suffice.
- **C-4 (SHOULD)** Prefer simple, composable, testable functions.
- **C-5 (MUST)** Use strongly typed ID types to prevent mix-ups:
  ```go
  // ✅ Good — distinct type prevents accidental substitution
  type UserID string
  type PostID string

  // ❌ Bad — plain string allows silent bugs
  type UserID = string
  ```
- **C-6 (MUST)** Return errors explicitly; never swallow them silently:
  ```go
  // ✅ Good
  result, err := doSomething()
  if err != nil {
      return fmt.Errorf("doSomething: %w", err)
  }

  // ❌ Bad
  result, _ := doSomething()
  ```
- **C-7 (SHOULD NOT)** Add comments except for exported symbols (godoc) and critical non-obvious caveats (e.g. algorithm quirks, library workarounds, regulatory constraints). Rely on self-explanatory code.
- **C-8 (SHOULD)** Prefer small interfaces defined at the point of use (consumer side), not in the package that implements them.
- **C-9 (SHOULD NOT)** Extract a new function unless it will be reused elsewhere, is the only way to unit-test otherwise untestable logic, or drastically improves readability of an opaque block.
- **C-10 (MUST)** Use `context.Context` as the first argument for any function that performs I/O or long-running work. When using Gin, always extract context from the request via `c.Request.Context()` — never pass `*gin.Context` directly into service or repository layers:
  ```go
  // ✅ Good — service layer stays framework-agnostic
  func (h *UserHandler) GetUser(c *gin.Context) {
      user, err := h.svc.GetUser(c.Request.Context(), UserID(c.Param("id")))
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, user)
  }

  // ❌ Bad — couples service to Gin
  func (s *UserService) GetUser(c *gin.Context) (*User, error) { ... }
  ```

---

### 3 — Error Handling

- **E-1 (MUST)** Wrap errors with context using `fmt.Errorf("operation: %w", err)` so callers can trace the failure chain.
- **E-2 (MUST)** Define sentinel errors with `errors.New` or typed errors for conditions callers need to handle explicitly:
  ```go
  var ErrNotFound = errors.New("not found")

  type ValidationError struct {
      Field   string
      Message string
  }
  func (e *ValidationError) Error() string {
      return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
  }
  ```
- **E-3 (SHOULD NOT)** Use `panic` for expected error conditions; reserve it for truly unrecoverable programmer errors.
- **E-4 (SHOULD)** Use `errors.Is` / `errors.As` for error inspection, never string comparison.
- **E-5 (MUST — Gin)** Never write HTTP error responses inline in each handler. Use a single centralized error middleware via `c.Error(err)` + `c.Next()`, and map error types to HTTP status codes in one place:
  ```go
  // internal/handler/middleware/error.go
  func ErrorMiddleware() gin.HandlerFunc {
      return func(c *gin.Context) {
          c.Next()
          if len(c.Errors) == 0 {
              return
          }
          err := c.Errors.Last().Err

          var valErr *ValidationError
          switch {
          case errors.As(err, &valErr):
              c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Error()})
          case errors.Is(err, ErrNotFound):
              c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
          default:
              c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
          }
      }
  }

  // Handler hanya perlu:
  func (h *UserHandler) GetUser(c *gin.Context) {
      user, err := h.svc.GetUser(c.Request.Context(), UserID(c.Param("id")))
      if err != nil {
          _ = c.Error(err) // delegasikan ke middleware
          return
      }
      c.JSON(http.StatusOK, user)
  }
  ```

---

### 4 — Testing

- **T-1 (MUST)** Colocate unit tests in `*_test.go` in the same package as the source file (use `package foo_test` for black-box tests).
- **T-2 (MUST)** For any API or DB change, add/extend integration tests in a dedicated `integration/` or `testdata/` directory, separated from unit tests.
- **T-3 (MUST)** ALWAYS separate pure-logic unit tests from DB/network-touching integration tests. Use build tags if needed:
  ```go
  //go:build integration
  ```
- **T-4 (SHOULD)** Prefer integration tests over heavy mocking; mock only at clear boundaries (e.g. interfaces for external services).
- **T-5 (SHOULD)** Unit-test complex algorithms thoroughly.
- **T-6 (SHOULD)** Use table-driven tests for multiple input/output cases:
  ```go
  // ✅ Good
  tests := []struct {
      name  string
      input int
      want  int
  }{
      {"positive", 2, 4},
      {"zero", 0, 0},
      {"negative", -1, 1},
  }
  for _, tc := range tests {
      t.Run(tc.name, func(t *testing.T) {
          got := square(tc.input)
          if got != tc.want {
              t.Errorf("square(%d) = %d, want %d", tc.input, got, tc.want)
          }
      })
  }
  ```
- **T-7 (SHOULD)** Use `testify/assert` or `testify/require` for cleaner assertions:
  ```go
  require.NoError(t, err)
  assert.Equal(t, want, got)
  ```
- **T-8 (SHOULD)** Express invariants with property-based testing using `pgregory.net/rapid` or `github.com/leanovate/gopter` when practical.

---

### 5 — Database

- **D-1 (MUST)** Accept `*sql.DB` or `*sql.Tx` via an interface so helpers work for both regular queries and transactions:
  ```go
  type Querier interface {
      QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
      ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
  }
  ```
- **D-2 (MUST)** Always pass `context.Context` to every DB call — never use context-less variants.
- **D-3 (SHOULD)** Keep SQL queries in `.sql` files or typed query builders (e.g. `sqlc`, `squirrel`); avoid raw string concatenation for dynamic queries.
- **D-4 (SHOULD)** Override incorrect auto-generated types in a dedicated `db_overrides.go` file, with a comment explaining why.

---

### 6 — Security

- **S-1 (MUST)** Never interpolate user input directly into SQL — always use parameterized queries.
- **S-2 (MUST)** Store secrets via environment variables or a secrets manager; never hardcode credentials.
- **S-3 (MUST)** Validate and sanitize all external inputs before processing or storing.
- **S-4 (SHOULD)** Use `crypto/rand` for all cryptographically sensitive random values; never use `math/rand`.
- **S-5 (SHOULD)** Set timeouts on all outbound HTTP clients:
  ```go
  client := &http.Client{Timeout: 10 * time.Second}
  ```

---

### 7 — Code Organization

- **O-1 (MUST)** Place code in a shared package (e.g. `internal/shared`) only if used by ≥ 2 other packages.
- **O-2 (MUST)** Use `internal/` to prevent unintended external imports of implementation details.
- **O-3 (SHOULD)** Follow standard Go project layout:
  ```
  cmd/          # main entrypoints
  internal/     # private application packages
  pkg/          # reusable public packages (if any)
  api/          # API schema definitions (proto, OpenAPI, etc.)
  ```
- **O-4 (SHOULD NOT)** Create packages named `util`, `common`, or `helpers` — name packages by what they provide, not what they are.

---

### 8 — Tooling Gates

- **G-1 (MUST)** `gofmt` or `goimports` passes — no unformatted code.
- **G-2 (MUST)** `go vet ./...` passes with zero warnings.
- **G-3 (MUST)** `golangci-lint run` passes (configure via `.golangci.yml`).
- **G-4 (MUST)** `go test ./...` passes, including the race detector: `go test -race ./...`.

---

### 9 — Git

- **GH-1 (MUST)** Use Conventional Commits format: https://www.conventionalcommits.org/en/v1.0.0
- **GH-2 (SHOULD NOT)** Refer to Claude or Anthropic in commit messages.

---

## Writing Functions Best Practices

When evaluating whether a function you implemented is good or not, use this checklist:

1. Can you read the function and HONESTLY easily follow what it's doing? If yes, stop here.
2. Does the function have very high cyclomatic complexity (deeply nested if/else, many switch cases)? If so, it's likely problematic.
3. Are there common data structures or algorithms (parsers, stacks, maps) that would make this function simpler and more robust?
4. Are there unused parameters in the function?
5. Are there unnecessary type assertions/conversions that could be moved to function arguments?
6. Is the function easily testable without mocking core features (DB, HTTP, Redis)? If not, can it be covered by an integration test?
7. Does it have hidden dependencies that could be factored into arguments instead?
8. Brainstorm 3 better function names; verify the current name is the most expressive and consistent with the rest of the codebase.

**IMPORTANT:** Do NOT extract a new function unless:
- It is used in more than one place, **or**
- It is easily unit testable while the original is not AND no other testing path exists, **or**
- The original function is extremely hard to follow and comments are needed just to explain it.

---

## Writing Tests Best Practices

When evaluating whether a test you implemented is good or not, use this checklist:

1. SHOULD use table-driven tests; never embed unexplained literals like `42` or `"foo"` directly.
2. SHOULD NOT add a test unless it can fail for a real defect. Trivial assertions are forbidden.
3. SHOULD ensure the test name (`t.Run("name", ...)`) states exactly what is being verified.
4. SHOULD compare results to independent, pre-computed expectations — never to the function's own output reused as the oracle.
5. SHOULD follow the same lint, formatting, and style rules as production code (`gofmt`, `golangci-lint`).
6. SHOULD express invariants (commutativity, idempotence, round-trip) with property-based tests when practical.
7. Unit tests for a function SHOULD be grouped under `TestFunctionName`.
8. Use `assert.IsType` or interface checks when testing for polymorphic returns.
9. ALWAYS use strong assertions: `assert.Equal(t, expected, got)` over `assert.NotNil(t, got)`.
10. SHOULD test edge cases: empty input, nil pointers, boundary values, concurrent access.
11. SHOULD NOT test conditions that are already caught by the compiler (e.g. wrong types).

---

## Code Organization Reference

```
cmd/
  api/main.go            # API server entrypoint, router setup

internal/
  handler/               # Gin handlers — HTTP concern only (parse request, call service, write response)
    middleware/          # Gin middlewares: error, auth, logging, recovery
    user.go
    post.go
  service/               # Business logic — tidak kenal Gin sama sekali
    user.go
    post.go
  repository/            # DB access via Querier interface
    user.go
    post.go
  domain/                # Types, sentinel errors, interfaces (no external dependencies)
    user.go
    errors.go
  shared/                # Shared utilities (used by ≥ 2 packages)
    social.go            # Character size and media validations

api/                     # API contract schemas (OpenAPI / protobuf)
```

> **Aturan ketergantungan antar layer (dependency rule):**
> `handler` → `service` → `repository` → `domain`
> Tidak boleh ada layer yang mengimport layer di atasnya.
> `*gin.Context` hanya boleh muncul di package `handler` dan `middleware`.

---

## Remember Shortcuts

### QNEW

When I type "qnew", this means:

```
Understand all BEST PRACTICES listed in CLAUDE_GO.md.
Your code SHOULD ALWAYS follow these best practices.
```

### QPLAN

When I type "qplan", this means:

```
Analyze similar parts of the codebase and determine whether your plan:
- is consistent with the rest of the codebase
- introduces minimal changes
- reuses existing code
```

### QCODE

When I type "qcode", this means:

```
Implement your plan and make sure your new tests pass.
Always run tests (including -race) to make sure nothing is broken:
  go test -race ./...
Always run goimports on the newly created files:
  goimports -w .
Always run linting:
  golangci-lint run
Always run vet:
  go vet ./...
```

### QCHECK

When I type "qcheck", this means:

```
You are a SKEPTICAL senior Go engineer.
Perform this analysis for every MAJOR code change you introduced (skip minor changes):

1. CLAUDE_GO.md checklist: Writing Functions Best Practices.
2. CLAUDE_GO.md checklist: Writing Tests Best Practices.
3. CLAUDE_GO.md checklist: Implementation Best Practices.
```

### QCHECKF

When I type "qcheckf", this means:

```
You are a SKEPTICAL senior Go engineer.
Perform this analysis for every MAJOR function you added or edited (skip minor changes):

1. CLAUDE_GO.md checklist: Writing Functions Best Practices.
```

### QCHECKT

When I type "qcheckt", this means:

```
You are a SKEPTICAL senior Go engineer.
Perform this analysis for every MAJOR test you added or edited (skip minor changes):

1. CLAUDE_GO.md checklist: Writing Tests Best Practices.
```

### QUX

When I type "qux", this means:

```
Imagine you are a human UX tester of the feature you implemented.
Output a comprehensive list of scenarios you would test, sorted by highest priority.
```

### QGIT

When I type "qgit", this means:

```
Create a new branch refer to skills create-branch, then add all changes to staging, create a commit, and push to remote.

Follow this checklist for writing your commit message:
- SHOULD use Conventional Commits format: https://www.conventionalcommits.org/en/v1.0.0
- SHOULD NOT refer to Claude or Anthropic in the commit message.
- SHOULD structure commit message as follows:

<type>[optional scope]: <description>

[optional body]

[optional footer(s)]

Allowed types: fix, feat, build, chore, ci, docs, style, refactor, perf, test.
BREAKING CHANGE footer for breaking API changes.
```