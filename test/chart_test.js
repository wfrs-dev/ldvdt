// Regression tests for the progress chart rendered with Chart.js.
//
// The /progress fragment is swapped by htmx, so the canvas is recreated and the
// chart must be rebuilt from the series embedded by the server. These tests load
// the real app.js with a DOM/Chart stub.
"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");

const created = [];
let destroyed = 0;

const fakeCanvas = {
  dataset: {
    series: JSON.stringify([
      { label: "10:00", wpm: 22.5, acc: 91.0 },
      { label: "10:05", wpm: 28.0, acc: 95.5 },
      { label: "10:12", wpm: 31.2, acc: 97.0 },
    ]),
  },
  getContext() {
    return {};
  },
};

global.document = {
  addEventListener() {},
  querySelectorAll() {
    return [];
  },
  getElementById(id) {
    return id === "progress-chart" ? fakeCanvas : null;
  },
  body: { addEventListener() {} },
};

global.Chart = class Chart {
  constructor(canvas, config) {
    this.canvas = canvas;
    this.config = config;
    created.push(this);
  }
  destroy() {
    destroyed += 1;
  }
};

global.window = {};

// The chart must read its colours from the CSS theme variables (S20, S21), so
// stub getComputedStyle with a known palette and pin the values below. The name
// is Bootstrap's own token, so the Sass theme reaches the chart.
global.getComputedStyle = function () {
  return {
    getPropertyValue(name) {
      return {
        "--bs-primary": "#26a269",
        "--bs-success": "#198754",
        "--bs-body-color": "#212529",
        "--bs-secondary": "#6c757d",
      }[name] || "";
    },
  };
};

require(path.join(__dirname, "..", "internal", "web", "static", "app.js"));

const api = global.window.ldvdt;
assert.ok(api, "app.js must expose window.ldvdt");
assert.equal(
  typeof api.renderProgressChart,
  "function",
  "app.js must expose renderProgressChart"
);

// The chart is built on load from the server-embedded series.
assert.equal(created.length, 1, "chart must be created on load");
const chart = created[0];
assert.equal(chart.canvas, fakeCanvas);
assert.deepEqual(chart.config.data.labels, ["10:00", "10:05", "10:12"]);

const datasets = chart.config.data.datasets;
assert.equal(datasets.length, 2, "expected WPM and accuracy datasets");
assert.deepEqual(datasets[0].data, [22.5, 28.0, 31.2]);
assert.deepEqual(datasets[1].data, [91.0, 95.5, 97.0]);

// Colours come from the CSS variables read at render time.
assert.equal(datasets[0].borderColor, "#26a269", "WPM line must use --bs-primary");
assert.equal(datasets[0].backgroundColor, "rgba(38,162,105,0.15)");
assert.equal(datasets[1].borderColor, "#198754", "accuracy line must use --bs-success");
assert.equal(datasets[1].backgroundColor, "rgba(25,135,84,0.12)");
assert.equal(chart.config.options.plugins.legend.labels.color, "#212529");
assert.equal(chart.config.options.scales.x.ticks.color, "#6c757d");

// The chart is responsive and fills its box.
assert.equal(chart.config.options.maintainAspectRatio, false);
assert.equal(chart.config.options.responsive, true);

// Re-rendering (after an htmx swap) must destroy the previous instance.
api.renderProgressChart();
assert.equal(destroyed, 1, "previous chart must be destroyed");
assert.equal(created.length, 2, "a new chart must be created");

// An empty series renders nothing instead of an empty chart.
fakeCanvas.dataset.series = "[]";
api.renderProgressChart();
assert.equal(created.length, 2, "no chart is created for an empty series");
assert.equal(destroyed, 2, "the previous chart is still destroyed");

console.log("app.js progress chart: all assertions passed");
