//go:build windows

package engine

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	versionDLL                 = syscall.NewLazyDLL("version.dll")
	procGetFileVersionInfoSize = versionDLL.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfo     = versionDLL.NewProc("GetFileVersionInfoW")
	procVerQueryValue          = versionDLL.NewProc("VerQueryValueW")
)

type langCodePage struct {
	Language uint16
	CodePage uint16
}

func readExecutableMetadata(path string) map[string]string {
	out := map[string]string{}
	if !strings.EqualFold(filepath.Ext(path), ".exe") && !strings.EqualFold(filepath.Ext(path), ".dll") {
		return out
	}
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return out
	}
	size, _, _ := procGetFileVersionInfoSize.Call(uintptr(unsafe.Pointer(p)), 0)
	if size == 0 {
		return out
	}
	buf := make([]byte, int(size))
	ok, _, _ := procGetFileVersionInfo.Call(
		uintptr(unsafe.Pointer(p)),
		0,
		size,
		uintptr(unsafe.Pointer(&buf[0])),
	)
	if ok == 0 {
		return out
	}

	lang, codepage := uint16(0x0409), uint16(0x04B0)
	if ptr, length, ok := versionQueryRaw(buf, "\\VarFileInfo\\Translation"); ok && length >= 4 {
		cp := (*langCodePage)(ptr)
		lang, codepage = cp.Language, cp.CodePage
	}

	base := fmt.Sprintf("\\StringFileInfo\\%04x%04x\\", lang, codepage)
	product := versionQueryString(buf, base+"ProductName")
	version := versionQueryString(buf, base+"ProductVersion")
	fileVersion := versionQueryString(buf, base+"FileVersion")
	company := versionQueryString(buf, base+"CompanyName")
	description := versionQueryString(buf, base+"FileDescription")

	setMeta(out, "exe product", product)
	setMeta(out, "exe version", version)
	setMeta(out, "exe fileversion", fileVersion)
	setMeta(out, "exe company", company)
	setMeta(out, "description", description)

	major, minor := splitVersion(version)
	setMeta(out, "exe versionmajor", major)
	setMeta(out, "exe versionminor", minor)
	return out
}

func versionQueryRaw(buf []byte, subBlock string) (unsafe.Pointer, uint32, bool) {
	if len(buf) == 0 {
		return nil, 0, false
	}
	sub, err := syscall.UTF16PtrFromString(subBlock)
	if err != nil {
		return nil, 0, false
	}
	var ptr unsafe.Pointer
	var length uint32
	ok, _, _ := procVerQueryValue.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(sub)),
		uintptr(unsafe.Pointer(&ptr)),
		uintptr(unsafe.Pointer(&length)),
	)
	return ptr, length, ok != 0 && ptr != nil
}

func versionQueryString(buf []byte, subBlock string) string {
	ptr, length, ok := versionQueryRaw(buf, subBlock)
	if !ok || length == 0 {
		return ""
	}
	u := unsafe.Slice((*uint16)(ptr), int(length))
	if len(u) > 0 && u[len(u)-1] == 0 {
		u = u[:len(u)-1]
	}
	return strings.TrimSpace(syscall.UTF16ToString(u))
}

func splitVersion(value string) (string, string) {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r < '0' || r > '9'
	})
	if len(fields) == 0 {
		return "", ""
	}
	major := fields[0]
	minor := ""
	if len(fields) > 1 {
		minor = fields[1]
	}
	if _, err := strconv.Atoi(major); err != nil {
		major = ""
	}
	if _, err := strconv.Atoi(minor); err != nil {
		minor = ""
	}
	return major, minor
}
