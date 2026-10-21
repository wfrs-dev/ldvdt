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
  };

  // Build a char -> key lookup from the on-screen keyboard DOM.
  function buildKeyMap() {
    const map = new Map();
    document.querySelectorAll("#keyboard .key[data-bottom], #keyboard .key[data-top]").forEach((k) => {
      if (k.dataset.bottom) map.set(k.dataset.bottom, k);
      if (k.dataset.top) map.set(k.dataset.top, k);
    });
    return map;
  }

  let keyMap = buildKeyMap();

  function clearActiveKeys() {
    document.querySelectorAll("#keyboard .key.active").forEach((k) => k.classList.remove("active"));
  }

  function highlightPracticeKeys(keys) {
    clearActiveKeys();
    if (!keys) return;
    for (const ch of keys) {
      const k = keyMap.get(ch);
      if (k) k.classList.add("active");
    }
  }

  function highlightCurrentChar() {
    state.charEls.forEach((el) => el.classList.remove("current"));
    if (state.step < state.target.length && state.target[state.step] !== " ") {
      const el = state.charEls[state.charPos];
      if (el) el.classList.add("current");
    }
  }

  function highlightKey(ch) {
    document.querySelectorAll("#keyboard .key.current").forEach((k) => k.classList.remove("current"));
    const k = keyMap.get(ch);
    if (k) k.classList.add("current");
  }

  function reset() {
    state.active = false;
    state.step = 0;
    state.charPos = 0;
    state.correct = 0;
    state.total = 0;
    state.startTs = null;
    state.lastTs = null;
    state.intervals = [];
    document.querySelectorAll("#keyboard .key.current").forEach((k) => k.classList.remove("current"));
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

    highlightPracticeKeys(root.dataset.keys);
    highlightCurrentChar();
    highlightKey(state.target[0]);

    state.active = true;
  }

  function isIgnoredKey(e) {
    return e.key === "Shift" || e.key === "Control" || e.key === "Alt" ||
      e.key === "Meta" || e.key === "AltGraph" || e.key === "CapsLock" ||
      e.key === "Backspace" || e.key === "Tab" || e.key === "Escape" ||
      e.key === "Dead";
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
        // Show a brief summary in the words area.
        const root = document.getElementById("words");
        if (root) {
          const summary = document.createElement("p");
          summary.className = "hint";
          summary.textContent =
            "Listo — " + Math.round(wpm) + " WPM · " +
            (accuracy * 100).toFixed(1) + "% precisión · " +
            consistency.toFixed(0) + " ms ritmo";
          root.prepend(summary);
        }
      })
      .catch((err) => console.error(err));
  }

  function onKeydown(e) {
    if (!state.active) return;
    if (isIgnoredKey(e)) return;

    // Prevent the page from scrolling on space.
    if (e.key === " ") e.preventDefault();

    const expected = state.target[state.step];
    if (expected === undefined) return;

    const typed = e.key;
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

  document.addEventListener("keydown", onKeydown);
  document.body.addEventListener("htmx:afterSwap", (e) => {
    if (e.detail && e.detail.target && e.detail.target.id === "words") {
      startPractice();
    }
  });

  window.ldvdt = { startPractice };
})();
