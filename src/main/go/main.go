package main

import (
	. "github.com/lxn/go-winapi"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"reflect"
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
	libshell    uintptr
	libuser32   uintptr
	folderRoot  *uint16
	imageLoaded map[string]image.Image
)

func _TEXT(svt string) *uint16 {
	return syscall.StringToUTF16Ptr(svt)
}

/////////////////////Image operation
func readImge(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return img, err
}

func ReadImageList(path, ext string) (map[string]image.Image, error) {
	folder, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Read all the files in the folder
	fileList, err := folder.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// Read all png images
	fileMap := make(map[string]image.Image)
	for _, v := range fileList {
		if !v.IsDir() {
			fname := v.Name()
			ext := filepath.Ext(fname)
			if ext == ".png" {
				img, err := readImge(path + "/" + fname)
				if err == nil {
					fileMap[fname] = img
				}

			}
		}
	}
	return fileMap, nil
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
	bi.Root = folderRoot
	bi.Title = _TEXT("Select")
	bi.Flags = 0x10 | 0x40

	coInitialize := MustGetProcAddress(libuser32, "CoInitialize")
	syscall.Syscall(coInitialize, 1, 0, 0, 0)
	//w32.CoInitialize()
	sHBrowseForFolder := MustGetProcAddress(libshell, "SHBrowseForFolderW")
	ret, _, _ := syscall.Syscall(sHBrowseForFolder,
		1, uintptr(unsafe.Pointer(&bi)), 0, 0)
	if ret == 0 {
		return
	}

	coUninitialize := MustGetProcAddress(libuser32, "CoUninitialize")
	syscall.Syscall(coUninitialize, 0, 0, 0, 0)

	path := syscall.UTF16ToString(getPath(ret))
	folderRoot = (*uint16)(unsafe.Pointer(ret))

	imgList, err := ReadImageList(path, "png")
	if err != nil {
		panic(err)
		return
	}
	imageLoaded = imgList
	println(len(imageLoaded))
	drawFrame(parent)
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
	return DefWindowProc(hwnd, msg, wparam, lparam)
}

func drawFrame(hwnd HWND) {
	if imageLoaded == nil {
		return
	}
	// 	var ps GetWindowDCSTRUCT
	hdc := GetDC(hwnd)

	cdc := CreateCompatibleDC(0)
	defer DeleteDC(cdc)

	for _, img := range imageLoaded {
		var w int32 = int32(img.Bounds().Max.X - img.Bounds().Min.X)
		var h int32 = int32(img.Bounds().Max.Y - img.Bounds().Min.Y)

		println("type")
		println(reflect.TypeOf(img).Name())
		rgba:= img.(*image.NRGBA)
		print(rgba)


		// println(len(rgba.Pix))
		println("->")
		println(rgba.Pix)

		planes := GetDeviceCaps(hdc, PLANES)
		bmp := CreateBitmap(w, planes, 1, 24, unsafe.Pointer(&rgba.Pix[0]))
		hbmpOld := SelectObject(cdc, HGDIOBJ(bmp))
		defer SelectObject(cdc, HGDIOBJ(hbmpOld))

		BitBlt(hdc, 0, 0, w, h, cdc, 0, 0, SRCCOPY)
		break
	}
	// EndPaint(hwnd, &ps)
}

func WndProc(hwnd HWND, msg uint32, wparam uintptr, lparam uintptr) uintptr {
	openCb := syscall.NewCallback(LoadPath)

	switch msg {
	case WM_CREATE:
		createButton(10, 10, 100, 40, hwnd, "Open", OPEN_BTN_ID)
		return 0
	case WM_PAINT:
		drawFrame(hwnd)
		// return 0

	case WM_COMMAND:
		wid := LOWORD(uint32(wparam))
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

	return DefWindowProc(hwnd, msg, wparam, lparam)
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
