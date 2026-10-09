package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func assertExecutorFiles(t *testing.T, dir string, want map[string]string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".easyrenamer_tmp_") {
			t.Fatalf("temporary name remains: %s", entry.Name())
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		got[entry.Name()] = string(data)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("files = %#v; want %#v", got, want)
	}
}

func TestExecutorIntegrationRollback(t *testing.T) {
	for _, mode := range []string{"cancel-before-start", "cancel-staging", "cancel-renaming", "rename-error"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			original := map[string]string{"a.txt": "A\x00content", "b.txt": "different B content"}
			for name, data := range original {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			items, err := Preview(Config{Sources: []string{dir}, Category: CategoryAll, SortBy: SortName, Methods: []RenameMethod{{Type: MethodTemplate, Template: "new_<Name>"}}})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "cancel-before-start" {
				cancel()
			}
			reached := mode == "cancel-before-start"
			pairs, err := ExecuteContext(ctx, items, func(p ExecuteProgress) {
				if mode == "cancel-staging" && p.Phase == "staging" && p.Completed == 1 {
					reached = true
					cancel()
				}
				if mode == "cancel-renaming" && p.Phase == "renaming" && p.Completed == 1 {
					reached = true
					cancel()
				}
				if mode == "rename-error" && p.Phase == "staging" && p.Completed == len(items) {
					// Obstruct the second destination after preflight. This forces a
					// filesystem error after the first final rename on Windows and Unix.
					reached = true
					if err := os.Mkdir(filepath.Join(dir, "new_b.txt"), 0700); err != nil {
						t.Fatal(err)
					}
				}
			})
			if !reached || err == nil || len(pairs) != 0 {
				t.Fatalf("result = %#v, %v; reached=%v", pairs, err, reached)
			}
			if mode != "rename-error" && !errors.Is(err, context.Canceled) {
				t.Fatalf("error = %v", err)
			}
			if mode == "rename-error" {
				if err := os.Remove(filepath.Join(dir, "new_b.txt")); err != nil {
					t.Fatal(err)
				}
			}
			assertExecutorFiles(t, dir, original)
			for _, item := range items {
				item.NewName = "new_" + item.OldName
			}
			pairs, err = Execute(items)
			if err != nil || len(pairs) != 2 {
				t.Fatalf("retry = %#v, %v", pairs, err)
			}
			assertExecutorFiles(t, dir, map[string]string{"new_a.txt": "A\x00content", "new_b.txt": "different B content"})
			if err := Undo(pairs); err != nil {
				t.Fatal(err)
			}
			assertExecutorFiles(t, dir, original)
		})
	}
}

func TestExecutorIntegrationDuplicateTargets(t *testing.T) {
	dir := t.TempDir()
	original := map[string]string{"a.txt": "A", "b.txt": "B"}
	for name, data := range original {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	items, err := Preview(Config{Sources: []string{dir}, Category: CategoryAll, Methods: []RenameMethod{{Type: MethodTemplate, Template: "same.txt"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %#v", items)
	}
	for _, item := range items {
		if item.Status != StatusConflict {
			t.Fatalf("item = %#v", item)
		}
	}
	if _, err := Execute(items); err == nil {
		t.Fatal("duplicate targets succeeded")
	}
	assertExecutorFiles(t, dir, original)
}
