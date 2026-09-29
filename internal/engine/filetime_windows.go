//go:build windows

package engine

import (
	"os"
	"syscall"
	"time"
)

func createdTime(path string, st os.FileInfo) time.Time {
	if data, ok := st.Sys().(*syscall.Win32FileAttributeData); ok {
		return time.Unix(0, data.CreationTime.Nanoseconds())
	}
	return st.ModTime()
}
