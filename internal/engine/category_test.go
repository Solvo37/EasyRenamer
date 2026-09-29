package engine

import "testing"

func TestCategories(t *testing.T) {
	if !MatchesCategory(CategoryImages, "", ".JPG") {
		t.Fatal("JPG should be image")
	}
	if MatchesCategory(CategoryAudio, "", ".jpg") {
		t.Fatal("JPG should not be audio")
	}
	if !MatchesCategory(CategoryCustom, "psd; ai, .svg", ".svg") {
		t.Fatal("custom extension failed")
	}
}
