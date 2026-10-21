// Regression tests for dead-key composition in the typing engine.
//
// The browser does not always deliver the composed character. With a dead-key
// layout it commonly emits keydown "Dead" and then keydown with the base letter
// ("a"), so the engine must compose the accent itself only when the base letter
// matches the expected character's base. These tests load the real app.js with a
// minimal DOM stub.
"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");

global.window = {};
global.document = {
  addEventListener() {},
  querySelectorAll() {
    return [];
  },
  getElementById() {
    return null;
  },
  body: { addEventListener() {} },
};

require(path.join(__dirname, "..", "internal", "web", "static", "app.js"));

const api = global.window.ldvdt;
assert.ok(api, "app.js must expose window.ldvdt");
assert.equal(
  typeof api.composeTyped,
  "function",
  "app.js must expose composeTyped for dead-key composition"
);

const accents = {
  "á": { dead: "´", base: "a" },
  "Á": { dead: "´", base: "A" },
};

const compose = api.composeTyped;

// Reporter: dead key + base letter produces the expected accented character.
assert.equal(compose(true, "á", "a", accents), "á");

// Reporter with uppercase base: SHIFT + vowel.
assert.equal(compose(true, "Á", "A", accents), "Á");

// Browser already composed the character: keep it as delivered.
assert.equal(compose(true, "á", "á", accents), "á");

// Wrong vowel after the dead key must stay wrong (counted as a mistake).
assert.equal(compose(true, "é", "i", accents), "i");

// Uppercase expected but lowercase typed is a mistake (missing SHIFT).
assert.equal(compose(true, "Á", "a", accents), "a");

// No dead key: the delivered character is used verbatim.
assert.equal(compose(false, "á", "a", accents), "a");

// Dead key pressed but the expected character is not accented.
assert.equal(compose(true, "q", "q", accents), "q");

console.log("app.js dead-key composition: all assertions passed");
