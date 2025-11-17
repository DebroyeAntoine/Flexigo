//go:build linux

package ui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func posWinLinux(title string, x, y int) error {
	if os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		return errors.New("cannot move windows under Wayland")
	}

	if _, err := exec.LookPath("wmctrl"); err != nil {
		return fmt.Errorf("wmctrl is not installed")
	}

	time.Sleep(150 * time.Millisecond)

	out, err := exec.Command("wmctrl", "-l").Output()
	if err != nil {
		return err
	}

	var winID string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, title) {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				winID = fields[0]
			}
			break
		}
	}

	if winID == "" {
		return fmt.Errorf("window with title %q not found", title)
	}

	cmd := exec.Command("wmctrl", "-i", "-r", winID,
		"-e", fmt.Sprintf("0,%d,%d,-1,-1", x, y))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wmctrl error: %v | %s", err, stderr.String())
	}

	return nil
}

