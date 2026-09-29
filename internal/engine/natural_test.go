package engine

import "testing"

func TestNaturalLess(t *testing.T) {
	cases := [][2]string{
		{"1.jpg", "2.jpg"},
		{"2.jpg", "10.jpg"},
		{"09.jpg", "10.jpg"},
		{"file2.jpg", "file10.jpg"},
	}
	for _, c := range cases {
		if !NaturalLess(c[0], c[1]) {
			t.Fatalf("expected %q < %q", c[0], c[1])
		}
	}
}
