// Package browser opens the user's default web browser (S7), supporting
// Windows, Linux and the other Unixes (and macOS for convenience).
package browser

import (
	"os/exec"
	"runtime"
)

// commandFor returns the program and arguments that open url in the default
// browser for the given GOOS. Windows uses rundll32, macOS uses open, and Linux
// shares xdg-open with every other Unix. It is separated from Open so the
// platform mapping can be tested without launching anything.
func commandFor(goos, url string) (string, []string, error) {
	switch goos {
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", url}, nil
	case "darwin":
		return "open", []string{url}, nil
	default: // Linux and every other Unix use the XDG opener.
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
