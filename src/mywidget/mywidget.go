// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mywidget

import (
	"github.com/lxn/go-winapi"
	"syscall"
	"unsafe"
)

type (
	FolderOpen struct {
		Title          string
		FilePath       string
		InitialDirPath string
		Filter         string
		FilterIndex    int
	}
	BROWSEINFO struct {
		Owner        winapi.HWND
		Root         *uint16
		DisplayName  *uint16
		Title        *uint16
		Flags        uint32
		CallbackFunc uintptr
		LParam       uintptr
		Image        int32
	}
)

var (
	libshell uintptr
)

func init() {
	libshell = winapi.MustLoadLibrary("Shell32.dll")
}

func GetPath(parent winapi.HWND, path uintptr) string {
	var bi BROWSEINFO
	bi.Owner = parent
	// bi.Root = path
	bi.Title = syscall.StringToUTF16Ptr("Select")
	bi.Flags = 0x10 | 0x40

	// coInitialize := MustGetProcAddress(libuser32, "CoInitialize")
	// syscall.Syscall(coInitialize, 1, 0, 0, 0)
	//w32.CoInitialize()
	sHBrowseForFolder := winapi.MustGetProcAddress(libshell, "SHBrowseForFolderW")
	ret, _, _ := syscall.Syscall(sHBrowseForFolder,
		1, uintptr(unsafe.Pointer(&bi)), 0, 0)

	var nameBuf [100]uint16
	getPath := winapi.MustGetProcAddress(libshell, "SHGetPathFromIDListW")
	syscall.Syscall(getPath,
		2, uintptr(ret), uintptr(unsafe.Pointer(&nameBuf[0])), 0)

	return syscall.UTF16ToString(nameBuf[:])

}

func GetSavePath(parent winapi.HWND) string {
    ofn := &winapi.OPENFILENAME{}

	ofn.LStructSize = uint32(unsafe.Sizeof(*ofn))
	ofn.HwndOwner = parent

	filter := make([]uint16, 124)
	ofn.LpstrFilter = &filter[0]
	ofn.NFilterIndex = 1

	filePath := make([]uint16, 124)
	ofn.LpstrFile = &filePath[0]
	ofn.NMaxFile = uint32(len(filePath))

	// ofn.LpstrInitialDir = syscall.StringToUTF16Ptr(dlg.InitialDirPath)
	// ofn.LpstrTitle = syscall.StringToUTF16Ptr(dlg.Title)
	ofn.Flags = winapi.OFN_FILEMUSTEXIST

	if ret := winapi.GetSaveFileName(ofn); !ret {
		return ""
	}

	return syscall.UTF16ToString(filePath)
}
