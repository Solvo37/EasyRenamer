//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows/registry"
)

type themeMode string

const (
	themeSystem themeMode = "system"
	themeLight  themeMode = "light"
	themeDark   themeMode = "dark"
)

var (
	dwmapi                    = syscall.NewLazyDLL("dwmapi.dll")
	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	uxtheme                   = syscall.NewLazyDLL("uxtheme.dll")
	procSetWindowTheme        = uxtheme.NewProc("SetWindowTheme")
	user32Theme               = syscall.NewLazyDLL("user32.dll")
	procEnumChildWindows      = user32Theme.NewProc("EnumChildWindows")
	procSendMessageW          = user32Theme.NewProc("SendMessageW")
	procInvalidateRect        = user32Theme.NewProc("InvalidateRect")
)

const wmThemeChanged = 0x031A

func loadThemeMode() themeMode {
	path := themeSettingsPath()
	if data, err := os.ReadFile(path); err == nil {
		switch themeMode(strings.TrimSpace(string(data))) {
		case themeLight:
			return themeLight
		case themeDark:
			return themeDark
		case themeSystem:
			return themeSystem
		}
	}
	return themeSystem
}

func saveThemeMode(mode themeMode) error {
	switch mode {
	case themeSystem, themeLight, themeDark:
	default:
		mode = themeSystem
	}
	path := themeSettingsPath()
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(mode), 0o644)
}

func themeSettingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "EasyRenamer", "theme.txt")
}

func effectiveDarkTheme(mode themeMode) bool {
	switch mode {
	case themeDark:
		return true
	case themeLight:
		return false
	default:
		return windowsAppsUseDarkTheme()
	}
}

func windowsAppsUseDarkTheme() bool {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false
	}
	defer key.Close()

	value, _, err := key.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false
	}
	return value == 0
}

func uiWindowBrush(dark bool) declarative.Brush {
	if dark {
		return declarative.SolidColorBrush{Color: walk.RGB(31, 31, 34)}
	}
	return declarative.SolidColorBrush{Color: walk.RGB(248, 249, 251)}
}

func uiPanelBrush(dark bool) declarative.Brush {
	if dark {
		return declarative.SolidColorBrush{Color: walk.RGB(38, 38, 42)}
	}
	return declarative.SolidColorBrush{Color: walk.RGB(255, 255, 255)}
}

func uiFieldBrush(dark bool) declarative.Brush {
	if dark {
		return declarative.SolidColorBrush{Color: walk.RGB(46, 46, 50)}
	}
	return declarative.SolidColorBrush{Color: walk.RGB(255, 255, 255)}
}

func uiTextColor(dark bool) walk.Color {
	if dark {
		return walk.RGB(238, 238, 242)
	}
	return walk.RGB(28, 30, 34)
}

func uiMutedTextColor(dark bool) walk.Color {
	if dark {
		return walk.RGB(172, 174, 181)
	}
	return walk.RGB(91, 96, 105)
}

func uiDangerTextColor(dark bool) walk.Color {
	if dark {
		return walk.RGB(255, 111, 111)
	}
	return walk.RGB(190, 30, 30)
}

func uiUnchangedTextColor(dark bool) walk.Color {
	if dark {
		return walk.RGB(145, 147, 154)
	}
	return walk.RGB(110, 110, 110)
}

func applyNativeTheme(hwnd uintptr, dark bool) {
	if hwnd == 0 {
		return
	}

	useDark := int32(0)
	if dark {
		useDark = 1
	}
	for _, attr := range []uintptr{20, 19} {
		r, _, _ := procDwmSetWindowAttribute.Call(
			hwnd,
			attr,
			uintptr(unsafe.Pointer(&useDark)),
			unsafe.Sizeof(useDark),
		)
		if int32(r) >= 0 {
			break
		}
	}

	applyThemeToHWND(hwnd, dark)

	callback := syscall.NewCallback(func(child, lparam uintptr) uintptr {
		applyThemeToHWND(child, dark)
		return 1
	})
	procEnumChildWindows.Call(hwnd, callback, 0)

	procInvalidateRect.Call(hwnd, 0, 1)
}

func applyThemeToHWND(hwnd uintptr, dark bool) {
	name := "Explorer"
	if dark {
		name = "DarkMode_Explorer"
	}
	ptr, err := syscall.UTF16PtrFromString(name)
	if err == nil {
		procSetWindowTheme.Call(hwnd, uintptr(unsafe.Pointer(ptr)), 0)
	}
	procSendMessageW.Call(hwnd, wmThemeChanged, 0, 0)
	procInvalidateRect.Call(hwnd, 0, 1)
}
