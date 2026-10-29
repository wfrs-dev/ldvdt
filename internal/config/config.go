// Package config holds the application configuration with safe defaults.
package config

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
)

// Config is the runtime configuration of the ldvdt trainer.
type Config struct {
	Host        string
	Port        int
	LogLevel    string
	DBPath      string
	OpenBrowser bool
}

// Default returns the configuration with production defaults: the server
// listens on 127.0.0.1:10987 and the SQLite database lives under the user's
// config directory (`.config` on Linux, `%AppData%` on Windows) (S3, S7).
func Default() Config {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return Config{
		Host:        "127.0.0.1",
		Port:        10987,
		LogLevel:    "info",
		DBPath:      filepath.Join(dir, "ldvdt", "ldvdt.db"),
		OpenBrowser: true,
	}
}

// ListenAddr returns the host:port the HTTP server must bind to.
func (c Config) ListenAddr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// DBPath returns the database path, creating its parent directory.
func (c Config) EnsureDBDir() error {
	return os.MkdirAll(filepath.Dir(c.DBPath), 0o755)
}
