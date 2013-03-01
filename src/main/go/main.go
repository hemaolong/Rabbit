package main

import (
	// "fmt"
	. "github.com/lxn/go-winapi"
	"os"
	"syscall"
	"unsafe"
	// "strconv"
)


const (
	winWidth  int32 = 800
	winHeight int32 = 600

	OPEN_BTN_ID int32 = 1
)

func _TEXT(svt string) *uint16 {
	return syscall.StringToUTF16Ptr(svt)
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

func WndProc(hwnd HWND, msg uint32, wparam uintptr, lparam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		createButton(10, 10, 100, 40, hwnd, "Open", OPEN_BTN_ID)
		return 0

	case WM_COMMAND:
	    if LOWORD(uint32(wparam)) == 1 {
	        // MessageBox(hwnd, _TEXT("hi"), _TEXT("kill U"), MB_OK)
	        nameBuf := new([]uint16)
	        var dumb OPENFILENAME
	        dumb.LStructSize = 200;
	        dumb.HwndOwner = hwnd
	        dumb.LpstrFile = (*uint16)(unsafe.Pointer(nameBuf))
	        dumb.NMaxFile  = 100
	        dumb.LpstrFilter = (*uint16)(_TEXT("All/0*.*/0Text/0*.TXT/0"))
	        dumb.Flags = OFN_PATHMUSTEXIST | OFN_FILEMUSTEXIST

////////////////
            libcomdlg32 := MustLoadLibrary("comdlg32.dll")

            getOpenFileName := MustGetProcAddress(libcomdlg32, "GetOpenFileNameW")
            ret, ret1, ret2 := syscall.Syscall(getOpenFileName, 1,
            		uintptr(unsafe.Pointer(&dumb)),
            		0,
            		0)
            println(ret);
            println(ret1);
            println(ret2)

            ///////////////////////
	        print(GetOpenFileName(&dumb))
	    }
	    return 0
	case WM_DESTROY:
		os.Exit(0)
	}
	return DefWindowProc(hwnd, msg, wparam, lparam)
}

  //Register WndClass
 func RegisterClass(){
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
