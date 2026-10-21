// Package browser opens the user's default web browser (S7), supporting Linux
// and Windows (and macOS for convenience).
package browser

import (
	"os/exec"
	"runtime"
)

// commandFor returns the program and arguments that open url in the default
// browser for the given GOOS. It is separated from Open so the platform mapping
// can be tested without launching anything.
func commandFor(goos, url string) (string, []string, error) {
	switch goos {
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", url}, nil
	case "darwin":
		return "open", []string{url}, nil
	default: // linux and other unixes, as before the refactor
		return "xdg-open", []string{url}, nil
	}
}

// Open launches the default browser pointing at url. It does not wait for the
// browser process to exit.
func Open(url string) error {
	name, args, err := commandFor(runtime.GOOS, url)
	if err != nil {
		return err
	}
	return exec.Command(name, args...).Start()
}
