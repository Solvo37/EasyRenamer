//go:build windows

package i18n

import (
	"syscall"
	"unsafe"
)

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultLocaleName = kernel32.NewProc("GetUserDefaultLocaleName")
)

func detectSystemLanguage() Language {
	var buf [85]uint16
	n, _, _ := procGetUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if n == 0 {
		return English
	}
	return languageFromLocale(syscall.UTF16ToString(buf[:]))
}
