package ui

import _ "embed"

// keyboardSVG is the keyboard design embedded inline in the page (S23, S26).
//
// It is a generated copy of docs/keyboard.svg: `go generate ./internal/domain`
// rewrites it, and CI fails when the committed copy is stale. Inlining it —
// instead of an <img> or a separate document — is what lets the typing engine
// repaint individual keys: they have to be in this document's DOM.
//
//go:embed keyboard.svg
var keyboardSVG string
