//go:build darwin

package ui

import (
	"fmt"
	"os/exec"
)

func posWinDarwin(title string, x, y int) error {
	script := fmt.Sprintf(`
        tell application "System Events"
            set winList to (every window of every process whose name contains "%s")
            if (count of winList) > 0 then
                repeat with w in item 1 of winList
                    try
                        set position of w to {%d, %d}
                    end try
                end repeat
            end if
        end tell
    `, title, x, y)

	cmd := exec.Command("osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript error: %v | %s", err, string(out))
	}

	return nil
}

// Stubs so macOS build doesn't fail when other OS functions aren't compiled.

func posWinLinux(title string, x, y int) error {
	return fmt.Errorf("posWinLinux called on macOS — shouldn't happen")
}

func posWinWindows(title string, x, y int) error {
	return fmt.Errorf("posWinWindows called on Linux — shouldn't happen")
}
