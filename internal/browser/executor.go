package browser

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// LaunchOptions définit les options de lancement du navigateur
type LaunchOptions struct {
	KioskMode bool // Si true, lance en mode kiosque plein écran
	NewWindow bool // Si true, force une nouvelle fenêtre
	Incognito bool // Si true, mode navigation privée
}

// BrowserExecutor gère le lancement et l'arrêt du navigateur
type BrowserExecutor struct {
	process *exec.Cmd
	path    string
}

// NewBrowserExecutor crée un nouveau gestionnaire de navigateur
func NewBrowserExecutor(path string) *BrowserExecutor {
	return &BrowserExecutor{
		path: path,
	}
}

// Launch lance le navigateur avec l'URL optionnelle et les options par défaut
func (b *BrowserExecutor) Launch(url string) error {
	return b.LaunchWithOptions(url, LaunchOptions{
		KioskMode: false, // Mode normal par défaut pour cohabiter avec Flexigo
		NewWindow: true,
		Incognito: false,
	})
}

// LaunchWithOptions lance le navigateur avec des options personnalisées
func (b *BrowserExecutor) LaunchWithOptions(url string, opts LaunchOptions) error {
	if b.process != nil && b.process.Process != nil {
		return fmt.Errorf("browser already running")
	}

	var args []string

	// Arguments selon les options
	if opts.KioskMode {
		args = append(args, "--kiosk")
	}

	if opts.NewWindow {
		args = append(args, "--new-window")
	}

	if opts.Incognito {
		switch runtime.GOOS {
		case "windows", "linux":
			args = append(args, "--incognito")
		case "darwin":
			args = append(args, "--incognito")
		}
	}

	// Ajoute l'URL si fournie
	if url != "" {
		args = append(args, url)
	}

	b.process = exec.Command(b.path, args...)

	// Démarre le processus sans bloquer
	err := b.process.Start()
	if err != nil {
		b.process = nil
		return fmt.Errorf("failed to start browser: %w", err)
	}

	log.Printf("Browser launched with PID: %d (kiosk=%v)", b.process.Process.Pid, opts.KioskMode)
	return nil
}

// IsRunning vérifie si le navigateur est toujours en cours d'exécution
func (b *BrowserExecutor) IsRunning() bool {
	if b.process == nil || b.process.Process == nil {
		return false
	}

	// Essaie de vérifier si le processus existe encore
	err := b.process.Process.Signal(nil)
	return err == nil
}

// Close ferme le navigateur proprement
func (b *BrowserExecutor) Close() error {
	if b.process == nil || b.process.Process == nil {
		return nil
	}

	log.Printf("Closing browser (PID: %d)", b.process.Process.Pid)

	// Essaie de tuer le processus proprement
	if err := b.process.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill browser process: %w", err)
	}

	// Attend que le processus se termine
	_ = b.process.Wait()

	b.process = nil
	log.Println("Browser closed successfully")
	return nil
}

// GetDefaultBrowserPath retourne le chemin par défaut du navigateur selon l'OS
func GetDefaultBrowserPath() string {
	switch runtime.GOOS {
	case "windows":
		// Chrome par défaut sur Windows
		return "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	case "linux":
		// Essaie plusieurs navigateurs courants
		browsers := []string{
			"/usr/bin/google-chrome",
			"/usr/bin/chromium-browser",
			"/usr/bin/firefox",
		}
		for _, browser := range browsers {
			if _, err := exec.LookPath(browser); err == nil {
				return browser
			}
		}
		return "google-chrome"
	case "darwin":
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	default:
		return "google-chrome"
	}
}

