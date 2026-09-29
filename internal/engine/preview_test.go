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
