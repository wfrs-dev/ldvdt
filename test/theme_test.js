// Regression tests for the manual light/dark theme toggle (S20).
//
// The persisted theme is applied on startup (normally by the inline head script;
// app.js re-confirms it and syncs the button label). Clicking the toggle then
// alternates data-bs-theme and stores the choice in localStorage. These tests load
// the real app.js with a minimal DOM/localStorage stub.
"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");

const store = { "ldvdt-theme": "light" };

const htmlEl = {
  attrs: { "data-bs-theme": "dark" },
  getAttribute(name) {
    return this.attrs[name];
  },
  setAttribute(name, value) {
    this.attrs[name] = value;
  },
};

const toggle = {
  attrs: {},
  textContent: "",
  handlers: {},
  addEventListener(type, fn) {
    this.handlers[type] = fn;
  },
  setAttribute(name, value) {
    this.attrs[name] = value;
  },
};

global.localStorage = {
  getItem(key) {
    return key in store ? store[key] : null;
  },
  setItem(key, value) {
    store[key] = value;
  },
};

global.document = {
  documentElement: htmlEl,
  addEventListener() {},
  querySelectorAll() {
    return [];
  },
  getElementById(id) {
    return id === "theme-toggle" ? toggle : null;
  },
  body: { addEventListener() {} },
};

global.window = {};

require(path.join(__dirname, "..", "internal", "web", "static", "app.js"));

const api = global.window.ldvdt;
assert.ok(api, "app.js must expose window.ldvdt");
assert.equal(typeof api.toggleTheme, "function");

// The saved value is reused on startup.
assert.equal(htmlEl.getAttribute("data-bs-theme"), "light", "saved theme must be applied on load");
assert.equal(store["ldvdt-theme"], "light", "saved value must be left untouched on load");

// With light active, the label announces the action: switch to dark.
assert.equal(toggle.attrs["aria-label"], "Cambiar a tema oscuro");
assert.equal(toggle.textContent, "Tema oscuro");

// Clicking the toggle alternates the theme and persists the new value.
toggle.handlers.click();
assert.equal(htmlEl.getAttribute("data-bs-theme"), "dark", "toggle must switch to dark");
assert.equal(store["ldvdt-theme"], "dark", "toggle must persist the new theme");
assert.equal(toggle.attrs["aria-label"], "Cambiar a tema claro");
assert.equal(toggle.textContent, "Tema claro");

// And back again.
toggle.handlers.click();
assert.equal(htmlEl.getAttribute("data-bs-theme"), "light", "toggle must switch back to light");
assert.equal(store["ldvdt-theme"], "light", "toggle must persist the light theme");
assert.equal(toggle.attrs["aria-label"], "Cambiar a tema oscuro");
assert.equal(toggle.textContent, "Tema oscuro");

console.log("app.js theme toggle: all assertions passed");
