package main

import (
	. "github.com/lxn/go-winapi"
	"image"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
	// "gameimg"
	//"fmt"
)

type BROWSEINFO struct {
	Owner        HWND
	Root         *uint16
	DisplayName  *uint16
	Title        *uint16
	Flags        uint32
	CallbackFunc uintptr
	LParam       uintptr
	Image        int32
}

const (
	winWidth  int32 = 800
	winHeight int32 = 600

	OPEN_BTN_ID int32 = 1001
)

var (
	winProc  HWND
	replaced bool = false

	libshell  uintptr
	libuser32 uintptr
)

func _TEXT(svt string) *uint16 {
	return syscall.StringToUTF16Ptr(svt)
}

/////////////////////Image operation
func ReadImage(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	return image.Decode(f)
}

func ReadImageList(path, ext string) []image.Image {
	// Iterate the path, find all the image files as the ext.
	imgFiles := make(map[string]int)
	filepath.Walk(path,
		func(p string, f os.FileInfo, err error) error {
			if f.IsDir() {
				return nil
			}
			imgFiles[f.Name()] = 1
			return nil
		})

	// Read all the images
	result := make([]image.Image, len(imgFiles))
	return result
}

/////////////////////End Image operation

func getPath(path uintptr) []uint16 {
	var nameBuf [100]uint16
	getPath := MustGetProcAddress(libshell, "SHGetPathFromIDListW")
	syscall.Syscall(getPath,
		2, uintptr(path), uintptr(unsafe.Pointer(&nameBuf[0])), 0)

	return nameBuf[0:100]
}
func createFolderBrower(parent HWND) {
	if libshell == 0 {
		libshell = MustLoadLibrary("Shell32.dll")
	}
	if libuser32 == 0 {
	}
	libuser32 = MustLoadLibrary("Ole32.dll")

	var bi BROWSEINFO
	bi.Owner = parent
	bi.Title = _TEXT("Select")
	bi.Flags = 1 | 2

	coInitialize := MustGetProcAddress(libuser32, "CoInitialize")
	syscall.Syscall(coInitialize, 1, 0, 0, 0)
	//w32.CoInitialize()
	sHBrowseForFolder := MustGetProcAddress(libshell, "SHBrowseForFolderW")
	ret, _, _ := syscall.Syscall(sHBrowseForFolder,
		1, uintptr(unsafe.Pointer(&bi)), 0, 0)

	coUninitialize := MustGetProcAddress(libuser32, "CoUninitialize")
	syscall.Syscall(coUninitialize, 0, 0, 0, 0)

	path := syscall.UTF16ToString(getPath(ret))
	println(path)
	// ReadImageList(path, "png")

	//println(len(imgList))
}

func createButton(x, y, w, h int32, parent HWND, text string, id int32) (result HWND) {
	result = CreateWindowEx(
		WS_EX_TRANSPARENT,
		_TEXT("Button"),
		_TEXT(text),
		WS_CHILD|WS_VISIBLE|BS_PUSHBUTTON,
		x, y, w, h,
		parent,
		HMENU(id),
		GetModuleHandle(nil),
		unsafe.Pointer(nil))

	return result
}

func LoadPath(hwnd HWND, msg uint32, wparam uintptr, lparam uintptr) uintptr {

	if hwnd != 0 && !replaced {
		println("replace the win proc")
		parentHwnd := GetParent(hwnd)
		if parentHwnd == hwnd {
			replaced = true
			winProc = HWND(SetWindowLong(parentHwnd, GWL_WNDPROC,
				int32(syscall.NewCallback(WndProc))))
		}
	}

	if msg == WM_NOTIFY {
		println("notifyffffffffffff")
	}

	// println("fffffffffff");
	// return DefWindowProc(winProc, msg, wparam, lparam)
	return 0
}

func WndProc(hwnd HWND, msg uint32, wparam uintptr, lparam uintptr) uintptr {
	openCb := syscall.NewCallback(LoadPath)

	switch msg {
	case WM_CREATE:
		createButton(10, 10, 100, 40, hwnd, "Open", OPEN_BTN_ID)
		return 0

	case WM_COMMAND:
		wid := LOWORD(uint32(wparam))
		if wparam == IDOK {
			println("OK       clicked")
		}
		if wid == uint16(OPEN_BTN_ID) {
			// MessageBox(hwnd, _TEXT("hi"), _TEXT("kill U"), MB_OK)
			var nameBuf [100]uint16
			var dumb OPENFILENAME
			dumb.LStructSize = uint32(unsafe.Sizeof(dumb))
			dumb.HwndOwner = hwnd
			dumb.LpstrFile = (*uint16)(unsafe.Pointer(&nameBuf))
			dumb.NMaxFile = 100
			dumb.NFilterIndex = 1
			dumb.LpstrFilter = (*uint16)(_TEXT("All Files (*.png)"))
			dumb.Flags = OFN_ENABLEHOOK | OFN_EXPLORER

			// Trigger the open
			dumb.LpfnHook = LPOFNHOOKPROC(openCb)

			// GetOpenFileName(&dumb)

			createFolderBrower(hwnd)
		}
		return 0
	case WM_DESTROY:
		os.Exit(0)
	}

	if winProc == 0 {
		println("=========================")
		winProc = hwnd
	}

	return DefWindowProc(winProc, msg, wparam, lparam)
}

//Register WndClass
func RegisterClass() {
	var wndProcPtr uintptr = syscall.NewCallback(WndProc)
	hInst := GetModuleHandle(nil)
	if hInst == 0 {
		panic("GetModuleHandle")
	}
	hIcon := LoadIcon(0, (*uint16)(unsafe.Pointer(uintptr(IDI_APPLICATION))))
	if hIcon == 0 {
		panic("LoadIcon")
	}
	hCursor := LoadCursor(0, (*uint16)(unsafe.Pointer(uintptr(IDC_ARROW))))
	if hCursor == 0 {
		panic("LoadCursor")
	}
	var wc WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = wndProcPtr
	wc.HInstance = hInst
	wc.HIcon = hIcon
	wc.HCursor = hCursor
	wc.HbrBackground = COLOR_BTNFACE + 1
	wc.LpszClassName = syscall.StringToUTF16Ptr("test")
	if atom := RegisterClassEx(&wc); atom == 0 {
		panic("RegisterClassEx")
	}
}

func main() {
	// Register the "test" window to the system
	RegisterClass()

	// Create a window of "test" type
	var hwnd HWND = CreateWindowEx(
		WS_EX_CLIENTEDGE,
		// hwnd,
		_TEXT("test"),
		_TEXT("test"),
		WS_OVERLAPPEDWINDOW|WS_CLIPSIBLINGS,
		(GetSystemMetrics(SM_CXSCREEN)-winWidth)>>1,
		// 0,
		(GetSystemMetrics(SM_CYSCREEN)-winHeight)>>1,
		// 0,
		winWidth,
		winHeight,
		0,
		0,
		GetModuleHandle(nil),
		unsafe.Pointer(nil))

	//ffffffff
	ShowWindow(hwnd, SW_SHOW)

	var message MSG
	for {
		if GetMessage(&message, 0, 0, 0) == 0 {
			break
		}
		TranslateMessage(&message)
		DispatchMessage(&message)
		// fmt.Print(message)
	}
	os.Exit(int(message.WParam))
}
