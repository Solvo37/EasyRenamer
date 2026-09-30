//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

type comGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type fileOpenDialogVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	Show                uintptr
	SetFileTypes        uintptr
	SetFileTypeIndex    uintptr
	GetFileTypeIndex    uintptr
	Advise              uintptr
	Unadvise            uintptr
	SetOptions          uintptr
	GetOptions          uintptr
	SetDefaultFolder    uintptr
	SetFolder           uintptr
	GetFolder           uintptr
	GetCurrentSelection uintptr
	SetFileName         uintptr
	GetFileName         uintptr
	SetTitle            uintptr
	SetOkButtonLabel    uintptr
	SetFileNameLabel    uintptr
	GetResult           uintptr
	AddPlace            uintptr
	SetDefaultExtension uintptr
	Close               uintptr
	SetClientGuid       uintptr
	ClearClientData     uintptr
	SetFilter           uintptr
	GetResults          uintptr
	GetSelectedItems    uintptr
}

type shellItemVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	BindToHandler  uintptr
	GetParent      uintptr
	GetDisplayName uintptr
}

type shellItemArrayVtbl struct {
	QueryInterface             uintptr
	AddRef                     uintptr
	Release                    uintptr
	BindToHandler              uintptr
	GetPropertyStore           uintptr
	GetPropertyDescriptionList uintptr
	GetAttributes              uintptr
	GetCount                   uintptr
	GetItemAt                  uintptr
}

var (
	ole32FolderPicker         = syscall.NewLazyDLL("ole32.dll")
	procCoInitializeExPicker  = ole32FolderPicker.NewProc("CoInitializeEx")
	procCoUninitializePicker  = ole32FolderPicker.NewProc("CoUninitialize")
	procCoCreateInstance      = ole32FolderPicker.NewProc("CoCreateInstance")
	procCoTaskMemFreePicker   = ole32FolderPicker.NewProc("CoTaskMemFree")
)

const (
	coinitApartmentThreaded = 0x2
	clsctxInprocServer      = 0x1

	fosPickFolders      = 0x20
	fosForceFileSystem  = 0x40
	fosAllowMultiSelect = 0x200

	sigdnFileSysPath = 0x80058000

	rpcEChangedMode = 0x80010106
	hresultCanceled = 0x800704C7
)

var (
	clsidFileOpenDialog = comGUID{
		Data1: 0xDC1C5A9C,
		Data2: 0xE88A,
		Data3: 0x4DDE,
		Data4: [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7},
	}
	iidIFileOpenDialog = comGUID{
		Data1: 0xD57C7288,
		Data2: 0xD4AD,
		Data3: 0x4768,
		Data4: [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60},
	}
)

func callCOMMethod(this unsafe.Pointer, method uintptr, args ...uintptr) uintptr {
	full := make([]uintptr, 0, len(args)+1)
	full = append(full, uintptr(this))
	full = append(full, args...)
	r1, _, _ := syscall.SyscallN(method, full...)
	return r1
}

func releaseCOM(this unsafe.Pointer) {
	if this == nil {
		return
	}
	vtbl := *(**[3]uintptr)(this)
	if vtbl == nil {
		return
	}
	callCOMMethod(this, vtbl[2])
}

func pickFoldersMulti(owner uintptr, title string) ([]string, error) {
	hr, _, _ := procCoInitializeExPicker.Call(0, coinitApartmentThreaded)
	hr32 := uint32(hr)
	shouldUninit := hr32 == 0 || hr32 == 1
	if hr32 != 0 && hr32 != 1 && hr32 != rpcEChangedMode {
		return nil, fmt.Errorf("CoInitializeEx failed: 0x%08X", hr32)
	}
	if shouldUninit {
		defer procCoUninitializePicker.Call()
	}

	var dialogPtr unsafe.Pointer
	hr, _, _ = procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidFileOpenDialog)),
		0,
		clsctxInprocServer,
		uintptr(unsafe.Pointer(&iidIFileOpenDialog)),
		uintptr(unsafe.Pointer(&dialogPtr)),
	)
	if uint32(hr) != 0 || dialogPtr == nil {
		return nil, fmt.Errorf("CoCreateInstance(IFileOpenDialog) failed: 0x%08X", uint32(hr))
	}
	defer releaseCOM(dialogPtr)

	vtbl := *(**fileOpenDialogVtbl)(dialogPtr)
	if vtbl == nil {
		return nil, fmt.Errorf("IFileOpenDialog vtable is nil")
	}

	var options uintptr
	if hr := callCOMMethod(dialogPtr, vtbl.GetOptions, uintptr(unsafe.Pointer(&options))); uint32(hr) == 0 {
		options |= fosPickFolders | fosForceFileSystem | fosAllowMultiSelect
	} else {
		options = fosPickFolders | fosForceFileSystem | fosAllowMultiSelect
	}
	if hr := callCOMMethod(dialogPtr, vtbl.SetOptions, options); uint32(hr) != 0 {
		return nil, fmt.Errorf("IFileOpenDialog.SetOptions failed: 0x%08X", uint32(hr))
	}

	if title != "" {
		titlePtr, err := syscall.UTF16PtrFromString(title)
		if err == nil {
			callCOMMethod(dialogPtr, vtbl.SetTitle, uintptr(unsafe.Pointer(titlePtr)))
		}
	}

	hr = callCOMMethod(dialogPtr, vtbl.Show, owner)
	if uint32(hr) == hresultCanceled {
		return nil, nil
	}
	if uint32(hr) != 0 {
		return nil, fmt.Errorf("IFileOpenDialog.Show failed: 0x%08X", uint32(hr))
	}

	var arrayPtr unsafe.Pointer
	if hr := callCOMMethod(dialogPtr, vtbl.GetResults, uintptr(unsafe.Pointer(&arrayPtr))); uint32(hr) != 0 || arrayPtr == nil {
		return nil, fmt.Errorf("IFileOpenDialog.GetResults failed: 0x%08X", uint32(hr))
	}
	defer releaseCOM(arrayPtr)

	arrayVtbl := *(**shellItemArrayVtbl)(arrayPtr)
	if arrayVtbl == nil {
		return nil, fmt.Errorf("IShellItemArray vtable is nil")
	}

	var count uint32
	if hr := callCOMMethod(arrayPtr, arrayVtbl.GetCount, uintptr(unsafe.Pointer(&count))); uint32(hr) != 0 {
		return nil, fmt.Errorf("IShellItemArray.GetCount failed: 0x%08X", uint32(hr))
	}

	paths := make([]string, 0, count)
	seen := make(map[string]struct{}, count)
	for i := uint32(0); i < count; i++ {
		var itemPtr unsafe.Pointer
		if hr := callCOMMethod(arrayPtr, arrayVtbl.GetItemAt, uintptr(i), uintptr(unsafe.Pointer(&itemPtr))); uint32(hr) != 0 || itemPtr == nil {
			continue
		}

		func() {
			defer releaseCOM(itemPtr)
			itemVtbl := *(**shellItemVtbl)(itemPtr)
			if itemVtbl == nil {
				return
			}

			var namePtr *uint16
			if hr := callCOMMethod(itemPtr, itemVtbl.GetDisplayName, sigdnFileSysPath, uintptr(unsafe.Pointer(&namePtr))); uint32(hr) != 0 || namePtr == nil {
				return
			}
			defer procCoTaskMemFreePicker.Call(uintptr(unsafe.Pointer(namePtr)))

			path := utf16PtrToString(namePtr)
			if path == "" {
				return
			}
			key := path
			if _, ok := seen[key]; ok {
				return
			}
			seen[key] = struct{}{}
			paths = append(paths, path)
		}()
	}

	return paths, nil
}

func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	n := 0
	for ptr := unsafe.Pointer(p); *(*uint16)(ptr) != 0; n++ {
		ptr = unsafe.Add(ptr, 2)
	}
	return syscall.UTF16ToString(unsafe.Slice(p, n))
}
