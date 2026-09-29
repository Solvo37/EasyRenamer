package engine

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultAndAdvancedTags(t *testing.T) {
	ctx := TemplateContext{
		BaseName: "The walking man",
		Extension: ".txt",
		ParentDir: "docs",
		GlobalIndex: 3,
		DirIndex: 2,
		TotalItems: 12,
		BatchTime: time.Date(2026, 9, 29, 15, 4, 5, 0, time.Local),
	}
	got, err := RenderTemplate("<Inc Nr:001:2>-<Word:2>-<Substr:5:2>-<Switch:A:B>", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "005-walking-wa-A.txt" {
		t.Fatalf("unexpected tag result: %q", got)
	}
}

func TestFallbackAndModifiers(t *testing.T) {
	ctx := TemplateContext{
		BaseName: "photo",
		Extension: ".jpg",
		Metadata: map[string]string{
			"performer": "David Bowie",
			"title":     "  Heroes  ",
		},
	}
	got, err := RenderTemplate("<Artist||Performer:upper>-<Title:trim:lower>", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "DAVID BOWIE-heroes.jpg" {
		t.Fatalf("unexpected fallback/modifier result: %q", got)
	}

	got, err = RenderTemplate("<Artist|Unknown artist:upper>", TemplateContext{BaseName: "x", Extension: ".txt"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "UNKNOWN ARTIST.txt" {
		t.Fatalf("unexpected default fallback result: %q", got)
	}
}

func TestDateAndFileTags(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.bin")
	data := []byte("hello world")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	modified := time.Date(2025, 3, 4, 5, 6, 7, 0, time.Local)
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}

	ctx := TemplateContext{
		Path: path,
		BaseName: "sample",
		Extension: ".bin",
		GlobalIndex: 1,
		DirIndex: 1,
		TotalItems: 1,
		BatchTime: time.Date(2026, 9, 29, 15, 4, 5, 0, time.Local),
	}
	got, err := RenderTemplate("<Date:yyyy-mm-dd>_<Date Modified:yyyy-mm-dd>_<Filesize B>_<MD5>", ctx)
	if err != nil {
		t.Fatal(err)
	}
	sum := md5.Sum(data)
	wantHash := hex.EncodeToString(sum[:])
	if !strings.HasPrefix(got, "2026-09-29_2025-03-04_11_"+wantHash) {
		t.Fatalf("unexpected date/file tag result: %q", got)
	}
}

func TestImageDimensionPadding(t *testing.T) {
	ctx := TemplateContext{
		BaseName: "img",
		Extension: ".png",
		Width: 10,
		Height: 7,
	}
	got, err := RenderTemplate("<Width:000>x<Height:3>", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "010x007.png" {
		t.Fatalf("unexpected dimension tag result: %q", got)
	}
}

func TestTagFallbackAlternative(t *testing.T) {
	ctx := TemplateContext{
		BaseName: "x",
		Extension: ".txt",
		Metadata: map[string]string{"singer": "Kate Bush"},
	}
	got, err := RenderTemplate("<Artist||Performer||Singer|Unknown>", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Kate Bush.txt" {
		t.Fatalf("unexpected fallback chain: %q", got)
	}
}
