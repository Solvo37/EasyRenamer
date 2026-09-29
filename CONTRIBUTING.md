# Contributing

Thanks for helping improve EasyRenamer.

## Development

1. Fork or create a branch.
2. Keep rename logic in `internal/engine` independent from the GUI where possible.
3. Add or update tests for engine behavior.
4. Run:

```bash
go test ./internal/...
```

5. On Windows, build the GUI:

```powershell
go build ./cmd/easyrenamer
```

## Design principles

- Preview first. Never surprise the user.
- Preserve file extensions unless the rule explicitly changes them.
- Detect conflicts before execution.
- Use transactional/two-phase rename where collisions are possible.
- Keep the core testable without the GUI.
- Prefer a useful free feature over artificial edition limits.

## Pull requests

Explain what problem the change solves and include a before/after example when changing rename behavior.
