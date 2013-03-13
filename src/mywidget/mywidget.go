// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mywidget

import (
	"github.com/lxn/go-winapi"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
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

	prevFilePath string

	preSavePath *uint16
)

func init() {
	libshell = winapi.MustLoadLibrary("Shell32.dll")
}

func GetOpenFileName(parent walk.RootWidget, title, filter string) string {
	dlg := &walk.FileDialog{}

	dlg.FilePath = prevFilePath
	dlg.Filter = filter
	dlg.Title = title

	if ok, _ := dlg.ShowOpen(parent); !ok {
		return ""
	}

	prevFilePath = dlg.FilePath
	return prevFilePath
}

func GetSaveFileName(parent walk.RootWidget, title, filter string) string {
	dlg := &walk.FileDialog{}

	dlg.InitialDirPath = prevFilePath
	dlg.Filter = filter
	dlg.Title = title

	if ok, _ := dlg.ShowSave(parent); !ok {
		return ""
	}

	prevFilePath = dlg.FilePath
	return prevFilePath
}

func GetPath(parent walk.RootWidget, title string) string {
	var bi BROWSEINFO
	bi.Owner = 0
	// bi.Root = path
	bi.Title = syscall.StringToUTF16Ptr(title)
	bi.Flags = 0x10 | 0x40

	// bi.Root = (*uint16)(unsafe.Pointer(preSavePath))

	sHBrowseForFolder := winapi.MustGetProcAddress(libshell, "SHBrowseForFolderW")
	ret, _, _ := syscall.Syscall(sHBrowseForFolder,
		1, uintptr(unsafe.Pointer(&bi)), 0, 0)

	var nameBuf [100]uint16
	getPath := winapi.MustGetProcAddress(libshell, "SHGetPathFromIDListW")
	syscall.Syscall(getPath,
		2, uintptr(ret), uintptr(unsafe.Pointer(&nameBuf[0])), 0)

	preSavePath = (*uint16)(unsafe.Pointer(ret))

	return syscall.UTF16ToString(nameBuf[:])

}
