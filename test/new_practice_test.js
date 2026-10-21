// Regression tests for the "start another practice" button.
//
// The button reuses the settings of the practice that just finished, read from
// the #words data attributes. A wrong level/lang/count would silently start the
// next practice with the wrong configuration, so those are pinned here. The
// tests load the real app.js with a minimal DOM stub and a spy on htmx.ajax.
"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");

let wordsEl = { dataset: {} };
const calls = [];

global.window = {
  htmx: {
    ajax(method, url, options) {
      calls.push({ method, url, options });
    },
  },
};

global.document = {
  addEventListener() {},
  querySelectorAll() {
    return [];
  },
  getElementById(id) {
    return id === "words" ? wordsEl : null;
  },
  body: { addEventListener() {} },
};

require(path.join(__dirname, "..", "internal", "web", "static", "app.js"));

const api = global.window.ldvdt;
assert.ok(api, "app.js must expose window.ldvdt");
assert.equal(
  typeof api.startNewPractice,
  "function",
  "app.js must expose startNewPractice for the restart button"
);

// Reporter: the next practice keeps the level, language and word count that
// the finished one used, and swaps #words so the engine re-initialises.
wordsEl.dataset = { level: "4", lang: "en", count: "35" };
calls.length = 0;
api.startNewPractice();

assert.equal(calls.length, 1, "one htmx request per click");
assert.equal(calls[0].method, "GET");
assert.equal(calls[0].url, "/practice?level=4&lang=en&count=35");
assert.equal(calls[0].options.target, "#words");
assert.equal(calls[0].options.swap, "outerHTML");

// Missing data attributes fall back to the same defaults as the controls form.
wordsEl.dataset = {};
calls.length = 0;
api.startNewPractice();

assert.equal(calls[0].url, "/practice?level=1&lang=es&count=20");

// Without a #words element there is nothing to swap, so no request goes out.
wordsEl = null;
calls.length = 0;
api.startNewPractice();

assert.equal(calls.length, 0, "no request without a practice area");
