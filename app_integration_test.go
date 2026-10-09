//go:build windows

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Solvo37/easyrenamer/internal/engine"
)

// No parallel tests: UserCacheDir reads process-wide environment variables.
func integrationApp(t *testing.T) (*App, string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", filepath.Join(root, "cache"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
	dir := filepath.Join(root, "files")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	return NewApp(), dir
}

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

// Exact tree comparison catches missing/extra files (including staging names),
// and compares bytes rather than just file existence or length.
func assertFiles(t *testing.T, dir string, want map[string]string) {
	t.Helper()
	got := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(d.Name(), ".easyrenamer_tmp_") {
			t.Errorf("staging name remains: %s", p)
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		got[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v; want %#v", got, want)
	}
}

func previewTemplate(t *testing.T, a *App, sources []string, template, policy string) PreviewResult {
	t.Helper()
	r, err := a.Preview(sources, nil, false, string(engine.CategoryAll), "", []engine.RenameMethod{{Type: engine.MethodTemplate, Template: template}}, string(engine.SortName), false, true, nil, policy)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func assertHistory(t *testing.T, a *App, count int) []HistoryEntry {
	t.Helper()
	h, err := a.History()
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != count {
		t.Fatalf("history = %#v; want %d entries", h, count)
	}
	return h
}

func TestAppIntegrationPreviewExecuteUndo(t *testing.T) {
	a, dir := integrationApp(t)
	p := writeFixture(t, dir, "a.txt", "first\x00payload")
	writeFixture(t, dir, "b.txt", "second payload")
	original := map[string]string{"a.txt": "first\x00payload", "b.txt": "second payload"}
	r := previewTemplate(t, a, []string{dir}, "renamed_<Name>", "skip")
	if len(r.Items) != 2 {
		t.Fatalf("preview = %#v", r)
	}
	for _, it := range r.Items {
		if it.Status != engine.StatusOK || it.NewName != "renamed_"+it.OldName {
			t.Fatalf("item = %#v", it)
		}
	}
	assertFiles(t, dir, original)
	assertHistory(t, a, 0)
	result, err := a.Execute([]string{p})
	if err != nil || result.Count != 1 || len(result.Pairs) != 1 || result.Cancelled {
		t.Fatalf("execute = %#v, %v", result, err)
	}
	if result.Pairs[0].From != p || result.Pairs[0].To != filepath.Join(dir, "renamed_a.txt") {
		t.Fatalf("pairs = %#v", result.Pairs)
	}
	assertFiles(t, dir, map[string]string{"renamed_a.txt": "first\x00payload", "b.txt": "second payload"})
	h := assertHistory(t, a, 1)
	if h[0].Undone || h[0].Count != 1 || h[0].Folder != dir {
		t.Fatalf("history = %#v", h)
	}
	result, err = a.Undo()
	if err != nil || result.Count != 1 {
		t.Fatalf("undo = %#v, %v", result, err)
	}
	assertFiles(t, dir, original)
	if !assertHistory(t, a, 1)[0].Undone {
		t.Fatal("history not marked undone")
	}
	if _, err = a.Undo(); err == nil {
		t.Fatal("repeated undo succeeded")
	}
	assertFiles(t, dir, original)
}

func TestAppIntegrationHistorySurvivesNewApp(t *testing.T) {
	a, dir := integrationApp(t)
	p := writeFixture(t, dir, "a.txt", "A")
	previewTemplate(t, a, []string{p}, "one_<Name>", "skip")
	if _, err := a.Execute([]string{p}); err != nil {
		t.Fatal(err)
	}
	id := assertHistory(t, a, 1)[0].ID
	q := writeFixture(t, dir, "b.txt", "B")
	previewTemplate(t, a, []string{q}, "two_<Name>", "skip")
	if _, err := a.Execute([]string{q}); err != nil {
		t.Fatal(err)
	}
	a = NewApp()
	assertHistory(t, a, 2)
	r, err := a.UndoHistory(id)
	if err != nil || r.Count != 1 {
		t.Fatalf("UndoHistory = %#v, %v", r, err)
	}
	assertFiles(t, dir, map[string]string{"a.txt": "A", "two_b.txt": "B"})
	if _, err := a.UndoHistory(id); err == nil {
		t.Fatal("repeated history undo succeeded")
	}
	if _, err := a.UndoHistory("missing"); err == nil {
		t.Fatal("unknown history undo succeeded")
	}
	if _, err := a.Undo(); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, dir, map[string]string{"a.txt": "A", "b.txt": "B"})
	for _, h := range assertHistory(t, a, 2) {
		if !h.Undone {
			t.Fatalf("active record: %#v", h)
		}
	}
}

func TestAppIntegrationBlockedNames(t *testing.T) {
	for _, name := range []string{"CON.txt", "NUL.txt", "COM1.txt", "LPT9.txt", "bad?.txt", "bad:.txt", "bad\x01.txt"} {
		t.Run(name, func(t *testing.T) {
			a, dir := integrationApp(t)
			p := writeFixture(t, dir, "source.txt", "unchanged")
			r := previewTemplate(t, a, []string{p}, name, "skip")
			if len(r.Items) != 1 || r.Items[0].Status != engine.StatusInvalid || r.Items[0].Error == "" {
				t.Fatalf("preview = %#v", r)
			}
			if _, err := a.Execute([]string{p}); err == nil {
				t.Fatal("invalid rename succeeded")
			}
			assertFiles(t, dir, map[string]string{"source.txt": "unchanged"})
			assertHistory(t, a, 0)
		})
	}
}

func TestAppIntegrationConflicts(t *testing.T) {
	for _, policy := range []string{"skip", "stop", "overwrite", "auto-number"} {
		t.Run(policy, func(t *testing.T) {
			a, dir := integrationApp(t)
			p := writeFixture(t, dir, "source.txt", "source")
			writeFixture(t, dir, "fixed.txt", "existing")
			r := previewTemplate(t, a, []string{p}, "fixed.txt", policy)
			if len(r.Items) != 1 {
				t.Fatalf("preview = %#v", r)
			}
			if policy == "auto-number" {
				if r.Items[0].Status != engine.StatusOK || r.Items[0].NewName != "fixed (1).txt" {
					t.Fatalf("preview = %#v", r)
				}
				if result, err := a.Execute([]string{p}); err != nil || result.Count != 1 {
					t.Fatalf("execute = %#v, %v", result, err)
				}
				assertFiles(t, dir, map[string]string{"fixed.txt": "existing", "fixed (1).txt": "source"})
				if _, err := a.Undo(); err != nil {
					t.Fatal(err)
				}
			} else {
				if r.Items[0].Status != engine.StatusConflict {
					t.Fatalf("preview = %#v", r)
				}
				if _, err := a.Execute([]string{p}); err == nil {
					t.Fatal("conflict rename succeeded")
				}
				assertHistory(t, a, 0)
			}
			assertFiles(t, dir, map[string]string{"source.txt": "source", "fixed.txt": "existing"})
		})
	}
}

func TestAppIntegrationLateConflictAndUndoConflict(t *testing.T) {
	a, dir := integrationApp(t)
	p := writeFixture(t, dir, "a.txt", "original")
	previewTemplate(t, a, []string{p}, "new_<Name>", "skip")
	target := writeFixture(t, dir, "new_a.txt", "intruder")
	if _, err := a.Execute([]string{p}); err == nil {
		t.Fatal("late conflict succeeded")
	}
	assertFiles(t, dir, map[string]string{"a.txt": "original", "new_a.txt": "intruder"})
	assertHistory(t, a, 0)
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Execute([]string{p}); err != nil {
		t.Fatal(err)
	}
	id := assertHistory(t, a, 1)[0].ID
	writeFixture(t, dir, "a.txt", "intruder")
	for _, undo := range []func() (OperationResult, error){a.Undo, func() (OperationResult, error) { return a.UndoHistory(id) }} {
		if _, err := undo(); err == nil {
			t.Fatal("undo overwrote original path")
		}
		assertFiles(t, dir, map[string]string{"a.txt": "intruder", "new_a.txt": "original"})
		if assertHistory(t, a, 1)[0].Undone {
			t.Fatal("failed undo marked history undone")
		}
	}
}

func TestAppIntegrationCancelExecute(t *testing.T) {
	for _, phase := range []string{"staging", "renaming"} {
		t.Run(phase, func(t *testing.T) {
			a, dir := integrationApp(t)
			p := writeFixture(t, dir, "a.txt", "A")
			q := writeFixture(t, dir, "b.txt", "B")
			previewTemplate(t, a, []string{dir}, "new_<Name>", "skip")
			reached := false
			result, err := a.execute([]string{p, q}, func(progress engine.ExecuteProgress) {
				if progress.Phase == phase && progress.Completed == 1 {
					reached = true
					a.CancelExecute()
				}
			})
			if !reached || err != nil || !result.Cancelled || result.Count != 0 || len(result.Pairs) != 0 {
				t.Fatalf("cancel = %#v, %v; reached=%v", result, err, reached)
			}
			assertFiles(t, dir, map[string]string{"a.txt": "A", "b.txt": "B"})
			assertHistory(t, a, 0)
			// Cancellation state must not poison a subsequent public operation.
			result, err = a.Execute([]string{p, q})
			if err != nil || result.Count != 2 {
				t.Fatalf("retry = %#v, %v", result, err)
			}
			if _, err := a.Undo(); err != nil {
				t.Fatal(err)
			}
			assertFiles(t, dir, map[string]string{"a.txt": "A", "b.txt": "B"})
		})
	}
}

func TestAppIntegrationCancelPreview(t *testing.T) {
	a, dir := integrationApp(t)
	p := writeFixture(t, dir, "a.txt", "A")
	// Install the same cancellation state Preview registers, then exercise the
	// public cancellation method without depending on filesystem scan timing.
	ctx, cancel := context.WithCancel(context.Background())
	a.previewCancel = cancel
	seq := a.previewSeq
	a.CancelPreview()
	if !errors.Is(ctx.Err(), context.Canceled) || a.previewCancel != nil || a.previewSeq != seq+1 {
		t.Fatal("preview cancellation state not cleared/invalidated")
	}
	a.CancelPreview() // safe when idle
	a.CancelExecute()
	a.ctx = ctx
	_, err := a.Preview([]string{p}, nil, false, string(engine.CategoryAll), "", []engine.RenameMethod{{Type: engine.MethodTemplate, Template: "new_<Name>"}}, "name", false, true, nil, "skip")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled preview error = %v", err)
	}
	if _, err := a.Execute([]string{p}); err == nil {
		t.Fatal("cancelled preview became executable")
	}
	assertFiles(t, dir, map[string]string{"a.txt": "A"})
	assertHistory(t, a, 0)
	a.ctx = nil
	previewTemplate(t, a, []string{p}, "new_<Name>", "skip")
	if _, err := a.Execute([]string{p}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Undo(); err != nil {
		t.Fatal(err)
	}
	assertFiles(t, dir, map[string]string{"a.txt": "A"})
}

func TestAppIntegrationDuplicateTargets(t *testing.T) {
	a, dir := integrationApp(t)
	p := writeFixture(t, dir, "a.txt", "A")
	q := writeFixture(t, dir, "b.txt", "B")
	r := previewTemplate(t, a, []string{dir}, "same.txt", "skip")
	if len(r.Items) != 2 {
		t.Fatalf("preview = %#v", r)
	}
	for _, it := range r.Items {
		if it.Status != engine.StatusConflict {
			t.Fatalf("item = %#v", it)
		}
	}
	if _, err := a.Execute([]string{p, q}); err == nil {
		t.Fatal("duplicate targets succeeded")
	}
	assertFiles(t, dir, map[string]string{"a.txt": "A", "b.txt": "B"})
	assertHistory(t, a, 0)
}
