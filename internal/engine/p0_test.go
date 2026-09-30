package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewNaturalSortControlsNumbering(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"10.txt", "2.txt", "1.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	items, err := Preview(Config{
		Sources:       []string{dir},
		Category:      CategoryAll,
		SortBy:        SortName,
		SortPerFolder: true,
		Methods: []RenameMethod{{
			Type:     MethodTemplate,
			Template: "<Inc NrDir:01>_<Name>",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	wantOld := []string{"1.txt", "2.txt", "10.txt"}
	wantNew := []string{"01_1.txt", "02_2.txt", "03_10.txt"}
	if len(items) != len(wantOld) {
		t.Fatalf("got %d items, want %d", len(items), len(wantOld))
	}
	for i := range items {
		if items[i].OldName != wantOld[i] {
			t.Fatalf("item %d old name = %q, want %q", i, items[i].OldName, wantOld[i])
		}
		if items[i].NewName != wantNew[i] {
			t.Fatalf("item %d new name = %q, want %q", i, items[i].NewName, wantNew[i])
		}
	}
}

func TestExecuteContextCancellationRollsBackTemps(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	items, err := Preview(Config{
		Sources:  []string{dir},
		Category: CategoryAll,
		Methods: []RenameMethod{{
			Type:     MethodTemplate,
			Template: "renamed_<Name>",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	_, err = ExecuteContext(ctx, items, func(progress ExecuteProgress) {
		if progress.Phase == "staging" && progress.Completed == 1 {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecuteContext error = %v, want context.Canceled", err)
	}

	for _, name := range []string{"a.txt", "b.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("original file %s was not restored: %v", name, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".easyrenamer_tmp_") {
			t.Fatalf("temporary file left behind: %s", entry.Name())
		}
	}
}
