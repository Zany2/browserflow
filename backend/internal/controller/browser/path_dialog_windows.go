//go:build windows

package browser

import (
	"runtime"
	"syscall"
	"unsafe"
)

type fileFilter struct {
	name    string
	pattern string
}

type openfilename struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       uintptr
	lpstrCustomFilter uintptr
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         uintptr
	nMaxFile          uint32
	lpstrFileTitle    uintptr
	nMaxFileTitle     uint32
	lpstrInitialDir   uintptr
	lpstrTitle        uintptr
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       uintptr
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    uintptr
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

type browseinfo struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName uintptr
	lpszTitle      uintptr
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

var (
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")
)

const (
	ofnFileMustExist = 0x00001000
	ofnPathMustExist = 0x00000800

	bifReturnOnlyFSDirs = 0x00000001
	bifNewDialogStyle   = 0x00000040
)

func selectBrowserBinPath() (string, bool) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if hr, _, _ := ole32.NewProc("OleInitialize").Call(0); int32(hr) >= 0 {
		defer ole32.NewProc("OleUninitialize").Call()
	}

	return openFileDialog("选择浏览器程序", []fileFilter{
		{name: "浏览器程序 (*.exe)", pattern: "*.exe"},
		{name: "所有文件 (*.*)", pattern: "*.*"},
	}, "exe")
}

func selectUserDataDir() (string, bool) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if hr, _, _ := ole32.NewProc("OleInitialize").Call(0); int32(hr) >= 0 {
		defer ole32.NewProc("OleUninitialize").Call()
	}

	return openFolderDialog("选择用户数据目录")
}

func openFileDialog(title string, filters []fileFilter, defaultExt string) (string, bool) {
	buffer := make([]uint16, 1024)
	filter := buildFileFilter(filters)
	ofn := openfilename{
		lStructSize: uint32(unsafe.Sizeof(openfilename{})),
		lpstrFilter: uintptr(unsafe.Pointer(&filter[0])),
		lpstrFile:   uintptr(unsafe.Pointer(&buffer[0])),
		nMaxFile:    uint32(len(buffer)),
		lpstrTitle:  uintptr(unsafe.Pointer(utf16Ptr(title))),
		flags:       ofnFileMustExist | ofnPathMustExist,
		lpstrDefExt: uintptr(unsafe.Pointer(utf16Ptr(defaultExt))),
	}
	ret, _, _ := comdlg32.NewProc("GetOpenFileNameW").Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", false
	}
	return syscall.UTF16ToString(buffer), true
}

func openFolderDialog(title string) (string, bool) {
	displayName := make([]uint16, 260)
	dialog := browseinfo{
		pszDisplayName: uintptr(unsafe.Pointer(&displayName[0])),
		lpszTitle:      uintptr(unsafe.Pointer(utf16Ptr(title))),
		ulFlags:        bifReturnOnlyFSDirs | bifNewDialogStyle,
	}
	pidl, _, _ := shell32.NewProc("SHBrowseForFolderW").Call(uintptr(unsafe.Pointer(&dialog)))
	if pidl == 0 {
		return "", false
	}
	defer ole32.NewProc("CoTaskMemFree").Call(pidl)

	buffer := make([]uint16, 1024)
	ret, _, _ := shell32.NewProc("SHGetPathFromIDListW").Call(pidl, uintptr(unsafe.Pointer(&buffer[0])))
	if ret == 0 {
		return "", false
	}
	return syscall.UTF16ToString(buffer), true
}

func buildFileFilter(filters []fileFilter) []uint16 {
	var parts []string
	for _, filter := range filters {
		parts = append(parts, filter.name, filter.pattern)
	}
	return utf16Filter(parts...)
}

func utf16Filter(parts ...string) []uint16 {
	var filter []uint16
	for _, part := range parts {
		filter = append(filter, syscall.StringToUTF16(part)...)
	}
	return append(filter, 0)
}

func utf16Ptr(value string) *uint16 {
	return syscall.StringToUTF16Ptr(value)
}
