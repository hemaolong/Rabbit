package main

import (
  "github.com/lxn/go-winapi"
  "syscall"
  "strconv"
)

func _TEXT(svt string) *uint16{
	return syscall.StringToUTF16Ptr(svt)
}

func toString(n int32) string{
	return strconv.Itoa(int(n));
}

func main(){
	var hwnd winapi.HWND
	// cxScreen := winapi.GetSystemMetrics(winapi.SM_CXSCREEN)
	// cyScreen := winapi.GetSystemMetrics(winapi.SM_CYSCREEN)
	winapi.MessageBox(hwnd, _TEXT("狗语言消息"), _TEXT("ff"), winapi.MB_OK)
}