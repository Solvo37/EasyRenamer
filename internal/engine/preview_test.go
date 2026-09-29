package engine

import (
	"image"
	"image/color"
	"image/png"
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

func TestDisabledMethodIsSkipped(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Photo.jpg")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryImages,
		Methods: []RenameMethod{
			{Type: MethodPrefixSuffix, Prefix: "prefix-"},
			{Type: MethodCase, CaseMode: CaseUpper, Disabled: true},
		},
		BatchTime: time.Unix(1786000000, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0].NewName; got != "prefix-Photo.jpg" {
		t.Fatalf("disabled case method was applied: %q", got)
	}
}

func TestPreviewReadsPNGDimensions(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.png")

	img := image.NewRGBA(image.Rect(0, 0, 13, 7))
	img.Set(0, 0, color.White)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryImages,
		Methods:  []RenameMethod{{Type: MethodPrefixSuffix, Prefix: "x-"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].Width != 13 || items[0].Height != 7 {
		t.Fatalf("unexpected dimensions %dx%d", items[0].Width, items[0].Height)
	}
	if items[0].Size <= 0 {
		t.Fatal("expected file size metadata")
	}
}
