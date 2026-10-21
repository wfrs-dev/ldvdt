// Package resources embeds the static assets shipped inside the binary
// (S5). The practice word lists live next to this file.
package resources

import "embed"

// FS holds the embedded word lists. All assets are loaded through embed so the
// binary is self-contained (no runtime file reads for these resources).
//
//go:embed eng.dat spa.dat
var FS embed.FS
