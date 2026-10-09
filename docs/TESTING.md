# Regression testing

## Go engine (no desktop window)

Run `go test -race -count=1 ./internal/...` on Linux or Windows.
`executor_integration_test.go` exercises real two-stage renames, byte preservation,
Undo, duplicate targets, cancellation before execution and during both phases,
and rollback after a filesystem error. Cancellation/error injection uses progress
callbacks, not sleeps. Exact directory comparisons reject leftover staging names.

## Application methods (Windows, no desktop window)

Build the frontend assets first (`npm install` and `npm run build` in `frontend`), or
run the normal Wails build. `main.go` embeds `frontend/dist`, so those assets must
exist even though these tests never start the desktop window.

Run `go test -race -count=1 -timeout 2m .` on Windows. The tests cover:

- Preview without filesystem changes; selected-file Execute and Undo.
- Persistent history loaded by a new App, undo by ID, repeated/unknown undo.
- Existing targets for skip/stop/overwrite and safe auto-numbering.
- Targets appearing after Preview and original paths blocking Undo.
- Reserved Windows names, forbidden characters and control characters.
- CancelExecute during staging and final renaming, rollback and a successful retry.
- CancelPreview after operation registration, cancelled-context rejection and recovery.

All fixtures use `t.TempDir`. `LOCALAPPDATA` points inside the fixture root so
history cannot read or overwrite the user's real EasyRenamer journal. Do not run
these environment-mutating tests in parallel. Every scenario checks file names
and bytes; exact tree comparisons also detect `.easyrenamer_tmp_*` leftovers.
The private `preview` helper observes the operation-registration boundary; the private `execute` helper runs the same code as the public Execute method and
adds a progress observer for deterministic cancellation without a Wails runtime.
CancelPreview is invoked after registration and before scanning starts, deterministically rejecting publication/execution of the cancelled preview. This is not a GUI scan test.

CI runs engine tests separately from application tests, with the race detector.
Release builds also run both suites before packaging/publishing. Changing tests
or pushing a PR does not create a release; release triggers remain unchanged.

## Desktop/UI checks (separate coverage)

The six-second EXE check in CI/release only proves process liveness. It does not
prove the window rendered or that a user can finish a rename. These Go suites
exercise engine/application behavior, not React components or Wails bindings.
`frontend/package.json` currently provides build/type checking, not a UI test suite.

For manual Windows verification, use a disposable directory and cache profile:
open the window, add/drop files, inspect preview rows, select files, execute,
observe progress/cancel, undo via toolbar and history, and check displayed errors.
Also verify dialogs, keyboard interactions and list updates after undo/cancel.
A future automated GUI suite must interact with an actual WebView/window; keep
it labelled separately from headless Go tests and process-liveness smoke checks.
