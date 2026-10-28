// ldvdt trainer — client typing engine.
// Captures keystrokes, highlights the current key on the on-screen keyboard,
// computes WPM / accuracy / consistency, and submits the result.
(function () {
  "use strict";

  const state = {
    active: false,
    target: "", // full sequence to type (words joined by spaces)
    step: 0, // index into target
    charPos: 0, // index into charEls (spaces are not rendered as spans)
    charEls: [],
    correct: 0,
    total: 0,
    startTs: null,
    lastTs: null,
    intervals: [],
    level: 1,
    lang: "es",
    count: 20,
    accents: {}, // accented char -> { dead, base }
  };

  // A dead key was pressed and its composition is not resolved yet.
  let pendingDead = false;

  // ---- keyboard highlighting (S23) ----

  // The fill the key to press is painted with. It lives in app.css as
  // --ldvdt-key-current so the palette stays in one place; the fallback is the
  // colour the user asked for.
  const CURRENT_KEY_FILL_FALLBACK = "#bc8f8f";

  // char -> SVG element id ("key26", "space"), rendered by the server on
  // #keyboard from the ids in docs/layout.yaml.
  let keyMap = new Map();

  // element id -> the fill the SVG design declares, captured before the first
  // repaint so leaving a key restores its own colour exactly.
  const originalFill = new Map();

  // Elements currently painted as "the key to press".
  const paintedKeys = new Set();

  function buildKeyMap() {
    const wrapper = document.getElementById("keyboard");
    if (!wrapper || !wrapper.dataset || !wrapper.dataset.keymap) return new Map();
    try {
      return new Map(Object.entries(JSON.parse(wrapper.dataset.keymap)));
    } catch (err) {
      console.error("invalid keyboard keymap", err);
      return new Map();
    }
  }

  keyMap = buildKeyMap();

  // keyElement resolves a character to the SVG key that produces it. A missing
  // entry or a missing element is not an error: it only means that character
  // gets no highlight (the design may not draw every key yet).
  function keyElement(ch) {
    const id = keyMap.get(ch);
    if (!id) return null;
    return document.getElementById(id) || null;
  }

  function currentFill() {
    return cssVar("--ldvdt-key-current", CURRENT_KEY_FILL_FALLBACK);
  }

  function paintKey(el) {
    if (!originalFill.has(el.id)) originalFill.set(el.id, el.style.fill || "");
    el.style.fill = currentFill();
    paintedKeys.add(el);
  }

  function restoreKey(el) {
    if (originalFill.has(el.id)) el.style.fill = originalFill.get(el.id);
  }

  function clearCurrentKeys() {
    paintedKeys.forEach(restoreKey);
    paintedKeys.clear();
  }

  // markKeys paints every key that produces ch: a direct key, or the pair
  // (dead key + base letter) for an accented character (S14).
  function markKeys(ch) {
    const hint = state.accents[ch];
    const chars = hint ? [hint.dead, hint.base] : [ch];
    for (const c of chars) {
      const el = keyElement(c);
      if (el) paintKey(el);
    }
  }

  function highlightKey(ch) {
    clearCurrentKeys();
    markKeys(ch);
  }

  function highlightCurrentChar() {
    state.charEls.forEach((el) => el.classList.remove("current"));
    if (state.step < state.target.length && state.target[state.step] !== " ") {
      const el = state.charEls[state.charPos];
      if (el) el.classList.add("current");
    }
  }

  function reset() {
    state.active = false;
    pendingDead = false;
    state.step = 0;
    state.charPos = 0;
    state.correct = 0;
    state.total = 0;
    state.startTs = null;
    state.lastTs = null;
    state.intervals = [];
    clearCurrentKeys();
  }

  function startPractice() {
    const root = document.getElementById("words");
    if (!root) return;

    reset();

    const words = Array.from(root.querySelectorAll(".word")).map((w) =>
      Array.from(w.querySelectorAll(".char")).map((c) => c.textContent).join("")
    );
    if (words.length === 0) return;

    state.target = words.join(" ");
    state.charEls = Array.from(root.querySelectorAll(".char"));
    state.level = parseInt(root.dataset.level || "1", 10);
    state.lang = root.dataset.lang || "es";
    state.count = parseInt(root.dataset.count || "20", 10);
    state.accents = root.dataset.accents ? JSON.parse(root.dataset.accents) : {};

    highlightCurrentChar();
    highlightKey(state.target[0]);

    state.active = true;
  }

  // composeTyped resolves the character the user actually meant.
  //
  // Browsers disagree on dead keys: some deliver the composed character
  // ("á"), others deliver "Dead" followed by the base letter ("a"). When a dead
  // key is pending and the delivered letter is the expected character's base,
  // the expected accented character is what the user typed. Otherwise the
  // delivered key stands, so a wrong vowel still counts as a mistake.
  function composeTyped(deadPending, expected, key, accents) {
    if (!deadPending) return key;
    const hint = accents[expected];
    if (hint && key === hint.base) return expected;
    return key;
  }

  // isDeadKeyChar reports whether the delivered key is one of the layout's dead
  // keys (some platforms report the accent character instead of "Dead").
  function isDeadKeyChar(key, accents) {
    for (const ch in accents) {
      if (accents[ch].dead === key) return true;
    }
    return false;
  }

  function isIgnoredKey(e) {
    return e.key === "Shift" || e.key === "Control" || e.key === "Alt" ||
      e.key === "Meta" || e.key === "AltGraph" || e.key === "CapsLock" ||
      e.key === "Backspace" || e.key === "Tab" || e.key === "Escape" ||
      e.key === "Unidentified" || e.key === "Dead";
  }

  function mean(values) {
    if (values.length === 0) return 0;
    return values.reduce((a, b) => a + b, 0) / values.length;
  }

  function stddev(values) {
    if (values.length < 2) return 0;
    const m = mean(values);
    const variance = values.reduce((a, b) => a + (b - m) * (b - m), 0) / values.length;
    return Math.sqrt(variance);
  }

  function finish() {
    state.active = false;
    pendingDead = false;
    clearCurrentKeys();
    const now = performance.now();
    const durationMS = state.startTs === null ? 0 : Math.round(now - state.startTs);

    // Standard WPM: 5 characters = 1 word.
    const minutes = durationMS / 60000;
    const wpm = minutes > 0 ? (state.total / 5) / minutes : 0;
    const accuracy = state.total > 0 ? state.correct / state.total : 0;
    const consistency = stddev(state.intervals);

    const payload = {
      level: state.level,
      lang: state.lang,
      wordCount: state.count,
      wpm: Math.round(wpm * 10) / 10,
      accuracy: Math.round(accuracy * 1000) / 1000,
      consistency: Math.round(consistency * 10) / 10,
      durationMS: durationMS,
    };

    fetch("/practice", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then((resp) => {
        if (!resp.ok) throw new Error("POST /practice failed: " + resp.status);
        return resp.json();
      })
      .then(() => {
        // Refresh the progress panel.
        if (window.htmx) {
          window.htmx.ajax("GET", "/progress", { target: "#progress-panel", swap: "innerHTML" });
        }
        // Show a brief summary in the words area, with a button that starts
        // another practice using the same settings.
        showSummary(
          `Listo — ${Math.round(wpm)} WPM · ` +
          `${(accuracy * 100).toFixed(1)}% precisión · ` +
          `${consistency.toFixed(0)} ms ritmo`
        );
      })
      .catch((err) => console.error(err));
  }

  // startNewPractice loads a fresh practice with the settings of the one that
  // just finished. The swap on #words is what the htmx:afterSwap handler above
  // uses to re-initialise the typing engine, so no extra wiring is needed.
  function startNewPractice() {
    const root = document.getElementById("words");
    if (!root || !window.htmx) return;
    const data = root.dataset || {};
    const params = new URLSearchParams({
      level: data.level || "1",
      lang: data.lang || "es",
      count: data.count || "20",
    });
    window.htmx.ajax("GET", "/practice?" + params.toString(), {
      target: "#words",
      swap: "outerHTML",
    });
  }

  // buildPracticeSummary returns the completion summary: the result line plus
  // a button wired to startNewPractice. Kept out of the fetch callback so the
  // button wiring can be tested without driving a whole practice to the end.
  function buildPracticeSummary(text) {
    const summary = document.createElement("div");
    summary.className = "practice-summary";

    const line = document.createElement("p");
    line.textContent = text;
    summary.appendChild(line);

    const again = document.createElement("button");
    again.type = "button";
    again.className = "new-practice";
    again.textContent = "Nueva práctica";
    again.addEventListener("click", startNewPractice);
    summary.appendChild(again);

    return summary;
  }

  // showSummary prepends the completion summary to the practice area.
  function showSummary(text) {
    const root = document.getElementById("words");
    if (!root) return;
    root.prepend(buildPracticeSummary(text));
  }

  function onKeydown(e) {
    if (!state.active) return;

    // Dead key: remember it and wait for the letter that completes it.
    if (e.key === "Dead" || isDeadKeyChar(e.key, state.accents)) {
      pendingDead = true;
      return;
    }
    // Input method composition in progress; the result arrives later.
    if (e.key === "Process") return;

    if (isIgnoredKey(e)) return;

    // Prevent the page from scrolling on space.
    if (e.key === " ") e.preventDefault();

    const expected = state.target[state.step];
    if (expected === undefined) return;

    const typed = composeTyped(pendingDead, expected, e.key, state.accents);
    pendingDead = false;
    const now = performance.now();
    if (state.startTs === null) state.startTs = now;
    if (state.lastTs !== null) state.intervals.push(now - state.lastTs);
    state.lastTs = now;

    state.total++;
    const isCorrect = typed === expected;
    if (isCorrect) state.correct++;

    // Mark the current char (non-space steps only).
    if (expected !== " ") {
      const el = state.charEls[state.charPos];
      if (el) {
        el.classList.remove("current", "correct", "incorrect");
        el.classList.add(isCorrect ? "correct" : "incorrect");
      }
      state.charPos++;
    }

    highlightKey(expected);
    state.step++;

    if (state.step >= state.target.length) {
      finish();
      return;
    }

    highlightCurrentChar();
    highlightKey(state.target[state.step]);
  }

  // ---- progress chart (Chart.js) ----

  let progressChart = null;

  function parseSeries(canvas) {
    try {
      return JSON.parse(canvas.dataset.series || "[]");
    } catch (err) {
      console.error("invalid chart series", err);
      return [];
    }
  }

  // cssVar reads a theme colour from the document's computed style so the
  // chart follows the active theme (S20). The names are Bootstrap 5.3's own
  // tokens, so a palette change in docs/bootstrap-theme.scss reaches the chart.
  // Fallbacks keep the chart readable when the stylesheet or
  // getComputedStyle is unavailable (e.g. in tests).
  function cssVar(name, fallback) {
    try {
      if (typeof getComputedStyle === "undefined") return fallback;
      const style = getComputedStyle(document.documentElement);
      if (!style) return fallback;
      const value = style.getPropertyValue(name);
      return (value && value.trim()) || fallback;
    } catch (err) {
      return fallback;
    }
  }

  // hexToRgba converts a #rrggbb colour to an rgba() with the given alpha so
  // the chart fills can reuse the theme accent/correct colours at low opacity.
  function hexToRgba(hex, alpha) {
    const m = /^#?([0-9a-f]{6})$/i.exec(hex.trim());
    if (!m) return hex;
    const n = parseInt(m[1], 16);
    return "rgba(" + ((n >> 16) & 255) + "," + ((n >> 8) & 255) + "," + (n & 255) + "," + alpha + ")";
  }

  // renderProgressChart (re)builds the chart from the series embedded by the
  // server. htmx replaces the canvas on every /progress swap, so the previous
  // Chart.js instance is destroyed first. Colours are read from the CSS
  // variables at render time, so a theme change re-renders with the new palette.
  function renderProgressChart() {
    const canvas = document.getElementById("progress-chart");
    if (!canvas || typeof Chart === "undefined") return;

    if (progressChart) {
      progressChart.destroy();
      progressChart = null;
    }

    const series = parseSeries(canvas);
    if (series.length === 0) return;

    const accent = cssVar("--bs-primary", "#26a269");
    const correct = cssVar("--bs-success", "#198754");
    const text = cssVar("--bs-body-color", "#212529");
    const muted = cssVar("--bs-secondary", "#6c757d");

    progressChart = new Chart(canvas, {
      type: "line",
      data: {
        labels: series.map((p) => p.label),
        datasets: [
          {
            label: "WPM",
            data: series.map((p) => p.wpm),
            borderColor: accent,
            backgroundColor: hexToRgba(accent, 0.15),
            tension: 0.3,
            fill: true,
            yAxisID: "y",
          },
          {
            label: "Precisión %",
            data: series.map((p) => p.acc),
            borderColor: correct,
            backgroundColor: hexToRgba(correct, 0.12),
            tension: 0.3,
            fill: false,
            yAxisID: "y1",
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: "index", intersect: false },
        plugins: {
          legend: {
            labels: { color: text, boxWidth: 12, font: { size: 10 } },
          },
        },
        scales: {
          x: {
            ticks: { color: muted, maxRotation: 0, autoSkip: true, font: { size: 9 } },
            grid: { color: hexToRgba(muted, 0.15) },
          },
          y: {
            position: "left",
            beginAtZero: true,
            ticks: { color: accent, font: { size: 9 } },
            grid: { color: hexToRgba(muted, 0.15) },
          },
          y1: {
            position: "right",
            min: 0,
            max: 100,
            ticks: { color: correct, font: { size: 9 } },
            grid: { drawOnChartArea: false },
          },
        },
      },
    });
  }

  // ---- theme toggle (S20, S24) ----

  const THEME_KEY = "ldvdt-theme";

  // The theme lives in Bootstrap's own data-bs-theme attribute, so toggling it
  // also switches every Bootstrap component, not only this app's CSS.
  function currentTheme() {
    const el = document.documentElement;
    if (!el) return "dark";
    return el.getAttribute("data-bs-theme") === "light" ? "light" : "dark";
  }

  function savedTheme() {
    try {
      if (typeof localStorage === "undefined") return null;
      const theme = localStorage.getItem(THEME_KEY);
      return theme === "light" || theme === "dark" ? theme : null;
    } catch (err) {
      return null;
    }
  }

  function persistTheme(theme) {
    try {
      if (typeof localStorage !== "undefined") localStorage.setItem(THEME_KEY, theme);
    } catch (err) {
      // Storage can be unavailable (private mode); the toggle still works.
    }
  }

  // setToggleLabel keeps the button's accessible label in sync with the theme:
  // it announces the action clicking will perform, not the current state.
  function setToggleLabel(theme) {
    const toggle = document.getElementById("theme-toggle");
    if (!toggle) return;
    const toLight = theme !== "light";
    const label = toLight ? "Cambiar a tema claro" : "Cambiar a tema oscuro";
    const text = toLight ? "Tema claro" : "Tema oscuro";
    toggle.setAttribute("aria-label", label);
    toggle.textContent = text;
  }

  // applyTheme switches the document theme, keeps the toggle label in sync,
  // persists the choice and re-renders the chart with the new palette. Owning
  // persistence here means every entry point (the toggle and initTheme) leaves
  // the same state instead of only the toggle remembering it.
  function applyTheme(theme) {
    const el = document.documentElement;
    const previous = currentTheme();
    if (el) el.setAttribute("data-bs-theme", theme);
    persistTheme(theme);
    setToggleLabel(theme);
    if (previous !== theme) renderProgressChart();
  }

  function toggleTheme() {
    applyTheme(currentTheme() === "light" ? "dark" : "light");
  }

  // initTheme reuses the persisted choice (normally already applied by the
  // inline head script before first paint) and syncs the toggle label.
  function initTheme() {
    const saved = savedTheme();
    if (saved) applyTheme(saved);
    else setToggleLabel(currentTheme());
  }

  document.addEventListener("keydown", onKeydown);
  document.body.addEventListener("htmx:afterSwap", (e) => {
    const target = e.detail && e.detail.target;
    if (!target) return;
    if (target.id === "words") startPractice();
    if (target.id === "progress-panel") renderProgressChart();
  });

  const themeToggle = document.getElementById("theme-toggle");
  if (themeToggle) themeToggle.addEventListener("click", toggleTheme);

  initTheme();
  renderProgressChart();

  window.ldvdt = { startPractice, startNewPractice, buildPracticeSummary, showSummary, composeTyped, renderProgressChart, toggleTheme, applyTheme, initTheme, highlightKey, clearCurrentKeys };
})();
