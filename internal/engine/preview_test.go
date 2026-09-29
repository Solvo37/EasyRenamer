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

func TestAdvancedRenameMethods(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "  Sample 123 File  .txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryAll,
		Methods: []RenameMethod{
			{Type: MethodTrim, TrimNormalizeSpaces: true},
			{Type: MethodRemovePattern, RemovePattern: "Sample "},
			{Type: MethodRemove, RemoveStart: 1, RemoveCount: 4},
			{Type: MethodRenumber, RenumberStart: 7, RenumberStep: 2, RenumberPadding: 3, RenumberPosition: PositionPrefix, RenumberSeparator: "-"},
		},
		BatchTime: time.Unix(1786000000, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if got := items[0].NewName; got != "007-File.txt" {
		t.Fatalf("unexpected advanced pipeline result: %q", got)
	}
}

func TestTimestampUsesModifiedTime(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "photo.jpg")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	modified := time.Date(2024, 3, 4, 5, 6, 7, 0, time.Local)
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryImages,
		Methods: []RenameMethod{{
			Type:               MethodTimestamp,
			TimestampSource:    TimestampModified,
			TimestampFormat:    "yyyyMMdd-HHmmss",
			TimestampPosition:  PositionSuffix,
			TimestampSeparator: "-",
		}},
		BatchTime: time.Date(2030, 1, 1, 0, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0].NewName; got != "photo-20240304-050607.jpg" {
		t.Fatalf("unexpected timestamp result: %q", got)
	}
}

func TestRenumberPerDirectory(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"a", "b"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"one.txt", "two.txt"} {
			if err := os.WriteFile(filepath.Join(root, dir, name), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	items, err := Preview(Config{
		Root:      root,
		Recursive: true,
		Category:  CategoryAll,
		Methods: []RenameMethod{{
			Type:              MethodRenumber,
			RenumberStart:     1,
			RenumberStep:      1,
			RenumberPadding:   2,
			RenumberPerDir:    true,
			RenumberPosition:  PositionPrefix,
			RenumberSeparator: "_",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"01_one.txt", "02_two.txt", "01_one.txt", "02_two.txt"}
	for i, expected := range want {
		if items[i].NewName != expected {
			t.Fatalf("item %d: got %q want %q", i, items[i].NewName, expected)
		}
	}
}

func TestMoveMethod(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "abcdef.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryAll,
		Methods: []RenameMethod{{
			Type: MethodMove, MoveStart: 2, MoveCount: 2, MoveTo: 4,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0].NewName; got != "adebcf.txt" {
		t.Fatalf("unexpected move result: %q", got)
	}
}

func TestListMethod(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"1.jpg", "2.jpg"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	items, err := Preview(Config{
		Root:      root,
		Recursive: false,
		Category:  CategoryImages,
		Methods: []RenameMethod{{
			Type: MethodList,
			ListText: "alpha\nbeta",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].NewName != "alpha.jpg" || items[1].NewName != "beta.jpg" {
		t.Fatalf("unexpected list results: %q, %q", items[0].NewName, items[1].NewName)
	}
}

func TestListReplaceMethod(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Hello_red-blue.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryAll,
		Methods: []RenameMethod{{
			Type: MethodListReplace,
			ListReplaceText: "hello => hi\nred => green\n- => _",
			ListReplaceCaseSensitive: false,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0].NewName; got != "hi_green_blue.txt" {
		t.Fatalf("unexpected list replace result: %q", got)
	}
}

func TestSwapMethod(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Michael Jackson - Thriller.mp3")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryAll,
		Methods: []RenameMethod{{
			Type: MethodSwap, SwapSeparator: " - ", SwapOccurrence: 1,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0].NewName; got != "Thriller - Michael Jackson.mp3" {
		t.Fatalf("unexpected swap result: %q", got)
	}
}

func TestScriptExpressionMethod(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Photo 01.JPG")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryImages,
		Methods: []RenameMethod{{
			Type: MethodScript,
			ScriptExpression: "concat(lower(Name), '-', Index, Ext)",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0].NewName; got != "photo 01-1.JPG" {
		t.Fatalf("unexpected script result: %q", got)
	}
}

func TestScriptPreservesExtensionWhenOmitted(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Sample.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := Preview(Config{
		Sources:  []string{path},
		Category: CategoryAll,
		Methods: []RenameMethod{{
			Type: MethodScript,
			ScriptExpression: "upper(Name)",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := items[0].NewName; got != "SAMPLE.txt" {
		t.Fatalf("unexpected script extension preservation: %q", got)
	}
}
