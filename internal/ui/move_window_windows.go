//go:build windows

package ui

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW  = user32.NewProc("FindWindowW")
	procSetWindowPos = user32.NewProc("SetWindowPos")

	SWP_NOSIZE uint32 = 0x0001
)

func posWinWindows(title string, x, y int) error {
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return err
	}

	hwnd, _, _ := procFindWindowW.Call(
		0,
		uintptr(unsafe.Pointer(titlePtr)),
	)
	if hwnd == 0 {
		return fmt.Errorf("window %q not found", title)
	}

	r, _, err2 := procSetWindowPos.Call(
		hwnd,
		0,
		uintptr(x),
		uintptr(y),
		0,
		0,
		uintptr(SWP_NOSIZE),
	)
	if r == 0 {
		return fmt.Errorf("SetWindowPos failed: %v", err2)
	}

	return nil
}

func posWinDarwin(title string, x, y int) error {
	return fmt.Errorf("posWinLinux called on macOS — shouldn't happen")
}

func posWinLinux(title string, x, y int) error {
	return fmt.Errorf("posWinWindows called on Linux — shouldn't happen")
}
