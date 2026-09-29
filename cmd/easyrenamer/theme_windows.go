//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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
	procEnumThreadWindows     = user32Theme.NewProc("EnumThreadWindows")
	procGetCurrentThreadId    = syscall.NewLazyDLL("kernel32.dll").NewProc("GetCurrentThreadId")
	procSendMessageW          = user32Theme.NewProc("SendMessageW")
	procInvalidateRect        = user32Theme.NewProc("InvalidateRect")
	procGetClassNameW         = user32Theme.NewProc("GetClassNameW")
	procSystemParametersInfoW = user32Theme.NewProc("SystemParametersInfoW")
)

const (
	spiGetWorkArea     = 0x0030
	wmThemeChanged     = 0x031A
	lvmFirst           = 0x1000
	lvmSetBkColor      = lvmFirst + 1
	lvmSetTextColor    = lvmFirst + 36
	lvmSetTextBkColor  = lvmFirst + 38
)

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
		return declarative.SolidColorBrush{Color: walk.RGB(32, 33, 35)}
	}
	return declarative.SolidColorBrush{Color: walk.RGB(247, 248, 250)}
}

func uiPanelBrush(dark bool) declarative.Brush {
	if dark {
		return declarative.SolidColorBrush{Color: walk.RGB(42, 43, 46)}
	}
	return declarative.SolidColorBrush{Color: walk.RGB(255, 255, 255)}
}

func uiFieldBrush(dark bool) declarative.Brush {
	if dark {
		return declarative.SolidColorBrush{Color: walk.RGB(48, 49, 52)}
	}
	return declarative.SolidColorBrush{Color: walk.RGB(255, 255, 255)}
}

func uiTextColor(dark bool) walk.Color {
	if dark {
		return walk.RGB(242, 242, 242)
	}
	return walk.RGB(28, 30, 34)
}

func uiMutedTextColor(dark bool) walk.Color {
	if dark {
		return walk.RGB(166, 166, 166)
	}
	return walk.RGB(91, 96, 105)
}

func uiTableAltColor(dark bool, odd bool) walk.Color {
	if dark {
		if odd {
			return walk.RGB(45, 46, 49)
		}
		return walk.RGB(50, 51, 54)
	}
	if odd {
		return walk.RGB(247, 248, 250)
	}
	return walk.RGB(255, 255, 255)
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
	className := strings.ToLower(windowClassName(hwnd))
	name := "Explorer"

	if dark {
		switch className {
		case "combobox", "combolbox", "listbox", "edit", "richedit20w", "richedit50w":
			name = "DarkMode_CFD"
		default:
			name = "DarkMode_Explorer"
		}
	}

	ptr, err := syscall.UTF16PtrFromString(name)
	if err == nil {
		procSetWindowTheme.Call(hwnd, uintptr(unsafe.Pointer(ptr)), 0)
	}

	if className == "syslistview32" {
		bg := colorRef(255, 255, 255)
		text := colorRef(28, 30, 34)
		if dark {
			bg = colorRef(58, 60, 65)
			text = colorRef(232, 233, 236)
		}
		procSendMessageW.Call(hwnd, lvmSetBkColor, 0, bg)
		procSendMessageW.Call(hwnd, lvmSetTextBkColor, 0, bg)
		procSendMessageW.Call(hwnd, lvmSetTextColor, 0, text)
	}

	procSendMessageW.Call(hwnd, wmThemeChanged, 0, 0)
	procInvalidateRect.Call(hwnd, 0, 1)
}

func scheduleFloatingTheme(owner walk.Form, dark bool) {
	if owner == nil {
		return
	}
	time.AfterFunc(45*time.Millisecond, func() {
		owner.Synchronize(func() {
			// Theme only the transient popup. Re-theming the owner while a ComboBox
			// is opening can cause the native dropdown to close immediately.
			applyFloatingTheme(dark)
		})
	})
}

func applyFloatingTheme(dark bool) {
	threadID, _, _ := procGetCurrentThreadId.Call()
	callback := syscall.NewCallback(func(hwnd, lparam uintptr) uintptr {
		className := strings.ToLower(windowClassName(hwnd))
		if className == "combolbox" || className == "#32768" {
			applyThemeToHWND(hwnd, dark)
		}
		return 1
	})
	procEnumThreadWindows.Call(threadID, callback, 0)
}

func colorRef(r, g, b byte) uintptr {
	return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16)
}

func windowClassName(hwnd uintptr) string {
	var buf [128]uint16
	n, _, _ := procGetClassNameW.Call(
		hwnd,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:n])
}

type winRect struct {
	Left, Top, Right, Bottom int32
}

func initialWindowDimensions() (int, int) {
	// Keep a visible margin around the window. This prevents the first launch
	// from extending beyond the taskbar or putting the close button off-screen.
	var rc winRect
	ok, _, _ := procSystemParametersInfoW.Call(
		spiGetWorkArea,
		0,
		uintptr(unsafe.Pointer(&rc)),
		0,
	)
	if ok == 0 || rc.Right <= rc.Left || rc.Bottom <= rc.Top {
		return 1180, 720
	}

	workW := int(rc.Right - rc.Left)
	workH := int(rc.Bottom - rc.Top)

	width := workW - 72
	height := workH - 72

	if width > 1380 {
		width = 1380
	}
	if height > 840 {
		height = 840
	}
	if width < 940 {
		width = 940
	}
	if height < 600 {
		height = 600
	}
	return width, height
}
