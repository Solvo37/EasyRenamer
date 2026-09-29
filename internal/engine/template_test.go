package engine

import (
	"math/rand"
	"testing"
	"time"
)

func TestMarketplaceTemplate(t *testing.T) {
	tm := time.Unix(1786000000, 0)
	got, err := RenderTemplate("<Inc NrDir:01><Rand Alpha:9><UnixTimestamp>-<DirName:1>", TemplateContext{
		BaseName: "1", Extension: ".jpg", ParentDir: "24", DirIndex: 1, GlobalIndex: 7,
		BatchTime: tm, Rand: rand.New(rand.NewSource(1)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len("01")+9+10+1+2+4 {
		t.Fatalf("unexpected length %d: %q", len(got), got)
	}
	if got[:2] != "01" {
		t.Fatalf("expected 01 prefix, got %q", got)
	}
	if got[len(got)-7:] != "-24.jpg" {
		t.Fatalf("unexpected suffix: %q", got)
	}
}

func TestLegacyTemplateTokens(t *testing.T) {
	tm := time.Unix(1786000000, 0)
	got, err := RenderTemplate("<Inc NrDir:01><Rand><Rand Str:8><UnixTimestamp>-<DirName:1>", TemplateContext{
		Extension: ".png", ParentDir: "8", DirIndex: 2, GlobalIndex: 9,
		BatchTime: tm, Rand: rand.New(rand.NewSource(2)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got[:2] != "02" || got[len(got)-6:] != "-8.png" {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestUnknownTokenErrors(t *testing.T) {
	_, err := RenderTemplate("<Nope>", TemplateContext{Extension: ".jpg"})
	if err == nil {
		t.Fatal("expected error")
	}
}
