package main

import (
	. "github.com/lxn/go-winapi"
	"os"
	"syscall"
	"unsafe"
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
)

func _TEXT(svt string) *uint16 {
	return syscall.StringToUTF16Ptr(svt)
}

func createFolderBrower(parent HWND) {
    createWindowEx := MustGetProcAddress(libuser32, "CreateWindowExW")
    var BROWSEINFO
        bi.Owner = parent
        bi.Title = _TEXT("Select folder")
        bi.Flags = BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE

        w32.CoInitialize()
        ret := w32.SHBrowseForFolder(&bi)
        w32.CoUninitialize()

        folder = w32.SHGetPathFromIDList(ret)
        accepted = folder != ""

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

			GetOpenFileName(&dumb)
		}
		return 0
	case WM_DESTROY:
		os.Exit(0)
	}

	if winProc == 0 {
		println(=========================");
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
