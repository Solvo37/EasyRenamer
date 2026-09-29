//go:build !windows

package engine

import (
	"os"
	"time"
)

func createdTime(path string, st os.FileInfo) time.Time {
	return st.ModTime()
}
