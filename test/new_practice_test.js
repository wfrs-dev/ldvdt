// Regression tests for the "start another practice" button.
//
// Two things are pinned here. First, the request the button builds: it reuses
// the settings of the practice that just finished, read from the #words data
// attributes, so a wrong level/lang/count would silently start the next
// practice with the wrong configuration. Second, the wiring itself: the button
// that the completion summary renders must be connected to that request. The
// tests load the real app.js with a minimal DOM stub and a spy on htmx.ajax.
"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");

// Minimal element stub: just the surface app.js touches.
function element(tag) {
  return {
    tagName: tag.toUpperCase(),
    className: "",
    textContent: "",
    type: "",
    dataset: {},
    children: [],
    listeners: {},
    appendChild(child) {
      this.children.push(child);
      return child;
    },
    prepend(child) {
      this.children.unshift(child);
      return child;
    },
    addEventListener(type, handler) {
      this.listeners[type] = handler;
    },
  };
}

let wordsEl = element("div");
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
  createElement(tag) {
    return element(tag);
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
assert.equal(
  typeof api.buildPracticeSummary,
  "function",
  "app.js must expose buildPracticeSummary so its button wiring is testable"
);

// ---- the request the button builds -----------------------------------------

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

// ---- the wiring the summary renders ----------------------------------------

// Reporter: the completion summary carries the result line and a restart
// button, and clicking that button issues the new-practice request.
wordsEl.dataset = { level: "2", lang: "es", count: "20" };
const summary = api.buildPracticeSummary("Listo — 40 WPM");

assert.equal(summary.className, "practice-summary");
assert.equal(summary.children.length, 2, "summary holds a text line and a button");

const line = summary.children[0];
assert.equal(line.tagName, "P");
assert.equal(line.textContent, "Listo — 40 WPM");

const button = summary.children[1];
assert.equal(button.tagName, "BUTTON");
assert.equal(button.type, "button", "the button must not submit a form");
assert.equal(button.className, "new-practice");
assert.equal(button.textContent, "Nueva práctica");
assert.equal(
  typeof button.listeners.click,
  "function",
  "the restart button must be wired to a click handler"
);

calls.length = 0;
button.listeners.click();

assert.equal(calls.length, 1, "clicking the restart button starts a new practice");
assert.equal(calls[0].url, "/practice?level=2&lang=es&count=20");

// Reporter: showSummary puts that summary at the top of the practice area.
const fresh = element("div");
fresh.dataset = { level: "1", lang: "es", count: "20" };
wordsEl = fresh;
api.showSummary("Listo — 30 WPM");

assert.equal(fresh.children.length, 1, "the summary is prepended to #words");
assert.equal(fresh.children[0].className, "practice-summary");

// ---- nothing to render ------------------------------------------------------

// Without a #words element there is nothing to swap, so no request goes out.
wordsEl = null;
calls.length = 0;
api.startNewPractice();

assert.equal(calls.length, 0, "no request without a practice area");

// showSummary must not throw when the practice area is gone.
api.showSummary("Listo — 30 WPM");
