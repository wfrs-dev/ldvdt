// Regression tests for the SVG keyboard highlight (S23).
//
// The keyboard is the user's SVG design embedded inline, so the engine paints
// the key to press by setting its fill and restores the fill the design declares
// when the key is no longer current. The character -> element mapping comes from
// data-keymap, rendered by the server from the ids in docs/layout.yaml. These
// tests load the real app.js with a minimal DOM stub.
"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");

const CURRENT = "#bc8f8f";

// Fake SVG keys carrying their own inline fill, exactly like the real artwork.
function svgKey(id, fill) {
  return { id, style: { fill } };
}

const keys = {
  key24: svgKey("key24", "#ffabff"), // ´ dead key
  key26: svgKey("key26", "#dfffbd"), // a / A
  key27: svgKey("key27", "#ffffaf"), // o / O
  space: svgKey("space", "#dbdbdb"),
};

// "z" is mapped to a key the design does not draw (the SVG currently has no
// #key7 either), so a gap must degrade to no highlight instead of throwing.
const keymap = {
  a: "key26",
  A: "key26",
  o: "key27",
  O: "key27",
  "´": "key24",
  z: "key99",
  " ": "space",
};

const accents = { "á": { dead: "´", base: "a" } };

const charEl = { textContent: "á", classList: { add() {}, remove() {} } };
const wordEl = { querySelectorAll: (sel) => (sel === ".char" ? [charEl] : []) };

const wordsRoot = {
  dataset: { level: "3", lang: "es", count: "1", accents: JSON.stringify(accents) },
  querySelectorAll(sel) {
    if (sel === ".word") return [wordEl];
    if (sel === ".char") return [charEl];
    return [];
  },
};

const keyboardRoot = { dataset: { keymap: JSON.stringify(keymap) } };

global.document = {
  addEventListener() {},
  querySelectorAll() {
    return [];
  },
  getElementById(id) {
    if (id === "keyboard") return keyboardRoot;
    if (id === "words") return wordsRoot;
    return keys[id] || null;
  },
  body: { addEventListener() {} },
};

global.window = {};

require(path.join(__dirname, "..", "internal", "web", "static", "app.js"));

const api = global.window.ldvdt;
assert.ok(api, "app.js must expose window.ldvdt");
assert.equal(typeof api.highlightKey, "function", "app.js must expose highlightKey");
assert.equal(typeof api.clearCurrentKeys, "function", "app.js must expose clearCurrentKeys");

// The key to press gets the sage fill.
api.highlightKey("a");
assert.equal(keys.key26.style.fill, CURRENT, "the key to press must be painted");

// Moving to the next key restores the previous key's own colour.
api.highlightKey("o");
assert.equal(keys.key26.style.fill, "#dfffbd", "the previous key must get its colour back");
assert.equal(keys.key27.style.fill, CURRENT, "the next key must be painted");

// Clearing restores the last painted key too.
api.clearCurrentKeys();
assert.equal(keys.key27.style.fill, "#ffffaf", "clearing must restore the design colour");

// Repainting the same key twice must not lose the original colour.
api.highlightKey("a");
api.highlightKey("a");
api.clearCurrentKeys();
assert.equal(keys.key26.style.fill, "#dfffbd", "the original fill must survive repaints");

// The spacebar is not part of layout.yaml; it resolves through the design's id.
api.highlightKey(" ");
assert.equal(keys.space.style.fill, CURRENT, "the spacebar must be painted");
api.clearCurrentKeys();
assert.equal(keys.space.style.fill, "#dbdbdb");

// A character whose key the design does not draw is a no-op, not a crash.
api.highlightKey("z");
api.highlightKey("€");
api.clearCurrentKeys();

// An accented character paints the dead key and the base letter (S14).
api.startPractice();
assert.equal(keys.key24.style.fill, CURRENT, "the dead key must be painted");
assert.equal(keys.key26.style.fill, CURRENT, "the base letter must be painted");
api.clearCurrentKeys();
assert.equal(keys.key24.style.fill, "#ffabff", "the dead key must be restored");
assert.equal(keys.key26.style.fill, "#dfffbd", "the base letter must be restored");

console.log("app.js SVG keyboard highlight: all assertions passed");
