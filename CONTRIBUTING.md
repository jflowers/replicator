# Contributing to Replicator

## Prerequisites

- Go 1.25+
- Git
- Make
- Node.js for workflow policy tests; local runs skip these tests when Node.js
  is unavailable, while CI requires them

## Development Setup

```bash
git clone git@github.com:unbound-force/replicator.git
cd replicator
make check   # builds, vets, and runs all tests
```

## Building and Testing

```bash
make build          # Build binary to bin/replicator
make test           # Run all tests
make vet            # Run go vet
make check          # Vet + test (use this before submitting PRs)
make coverage       # Run tests with race detection and print coverage
make check-coverage # Enforce per-package coverage ratchets locally
```

## Testing Conventions

- **Standard library only**: Use `testing` package. No testify, gomega, or
  external assertion libraries.
- **Assertions**: Use `t.Errorf` / `t.Fatalf` directly.
- **Naming**: `TestXxx_Description` (e.g., `TestCreateCell_Defaults`).
- **Database tests**: Use `db.OpenMemory()` for in-memory SQLite.
- **Filesystem tests**: Use `t.TempDir()` for temporary directories.
- **HTTP tests**: Use `httptest.NewServer` for mock servers.
- **No shared state**: Each test creates its own fixtures.
- **Git tests**: Guard with `if testing.Short() { t.Skip("requires git") }`.

Always run tests with `-count=1` to disable caching.

## Pull Request Workflow

1. **Create a branch**: Speckit features use `NNN-feature-name`, OpenSpec
   changes use `opsx/change-name`.
2. **Spec first**: Non-trivial changes require a spec (either Speckit under
   `specs/` or OpenSpec under `openspec/changes/`). When in doubt, use a spec.
3. **Conventional commits**: Use `type: description` format
   (feat, fix, docs, chore, refactor, test).
4. **CI must pass**: Run `make check` locally before pushing.
5. **One concern per PR**: Keep changes focused and minimal.

## Coding Conventions

- `gofmt` and `goimports` for formatting
- GoDoc comments on all exported functions and types
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Use `errors.Is` for sentinel errors (not string comparison)
- Import grouping: stdlib, then third-party, then internal
- JSON tags required on serialized struct fields
- No global mutable state

## Project Structure

See [AGENTS.md](AGENTS.md) for the full project structure, constitution,
behavioral constraints, and specification framework.

## License

By contributing, you agree that your contributions will be licensed under
the [MIT License](LICENSE).
