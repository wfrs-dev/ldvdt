// Package browser opens the user's default web browser (S7), supporting Linux
// and Windows.
package browser

import (
	"os/exec"
	"runtime"
)

// Open launches the default browser pointing at url. It does not wait for the
// browser process to exit.
func Open(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default: // linux and other unixes
		return exec.Command("xdg-open", url).Start()
	}
}
