package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPreviewResetsCounterPerDirectory(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"1", "2"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"10.jpg", "2.jpg", "1.jpg"} {
			if err := os.WriteFile(filepath.Join(root, dir, name), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	items, err := Preview(Config{
		Root: root, Recursive: true, Category: CategoryImages,
		Method:    MethodTemplate,
		Template:  "<Inc NrDir:01>-<Name>",
		BatchTime: time.Unix(1786000000, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 6 {
		t.Fatalf("got %d", len(items))
	}

	want := []string{"01-1.jpg", "02-2.jpg", "03-10.jpg", "01-1.jpg", "02-2.jpg", "03-10.jpg"}
	for i, w := range want {
		if items[i].NewName != w {
			t.Fatalf("item %d: got %q want %q", i, items[i].NewName, w)
		}
	}
}

func TestPreviewAcceptsMixedSourcesAndMethodStack(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "folder")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fileA := filepath.Join(root, "photo 2.jpg")
	fileB := filepath.Join(dir, "photo 10.jpg")
	for _, name := range []string{fileA, fileB} {
		if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	items, err := Preview(Config{
		Sources:   []string{fileA, dir, fileA}, // duplicate source must be ignored
		Recursive: false,
		Category:  CategoryImages,
		BatchTime: time.Unix(1786000000, 0),
		Methods: []RenameMethod{
			{Type: MethodTemplate, Template: "<Inc:001>-<Name>"},
			{Type: MethodReplace, Find: "photo", ReplaceWith: "image"},
			{Type: MethodCase, CaseMode: CaseUpper},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].NewName != "001-IMAGE 2.jpg" {
		t.Fatalf("unexpected first name: %q", items[0].NewName)
	}
	if items[1].NewName != "002-IMAGE 10.jpg" {
		t.Fatalf("unexpected second name: %q", items[1].NewName)
	}
}
