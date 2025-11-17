package ui

import (
	"fmt"
	"runtime"
)

func (ui *UIManager) PositionWindow(title string, x, y int) error {
	switch runtime.GOOS {
	case "linux":
		return posWinLinux(title, x, y)
	case "windows":
		return posWinWindows(title, x, y)
	case "darwin":
		return posWinDarwin(title, x, y)
	default:
		return fmt.Errorf("window positioning not supported on %s", runtime.GOOS)
	}
}

