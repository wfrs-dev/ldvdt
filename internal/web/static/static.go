// Package static embeds the web assets (CSS, JS, and vendored libraries)
// served by the binary (S5).
package static

import "embed"

// FS holds the static web assets.
//
//go:embed vendor app.css app.js
var FS embed.FS
