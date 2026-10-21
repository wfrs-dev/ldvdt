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

  function clearCurrentKeys() {
    document.querySelectorAll("#keyboard .key.current").forEach((k) => k.classList.remove("current"));
  }

  // markKeys adds a class to every key that produces ch: a direct key, or the
  // pair (dead key + base letter) for an accented character (S14).
  function markKeys(ch, cls) {
    const hint = state.accents[ch];
    if (hint) {
      const dead = keyMap.get(hint.dead);
      if (dead) dead.classList.add(cls);
      const base = keyMap.get(hint.base);
      if (base) base.classList.add(cls);
      return;
    }
    const k = keyMap.get(ch);
    if (k) k.classList.add(cls);
  }

  function highlightPracticeKeys(keys) {
    clearActiveKeys();
    if (!keys) return;
    for (const ch of keys) markKeys(ch, "active");
  }

  function highlightCurrentChar() {
    state.charEls.forEach((el) => el.classList.remove("current"));
    if (state.step < state.target.length && state.target[state.step] !== " ") {
      const el = state.charEls[state.charPos];
      if (el) el.classList.add("current");
    }
  }

  function highlightKey(ch) {
    clearCurrentKeys();
    markKeys(ch, "current");
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

    highlightPracticeKeys(root.dataset.keys);
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

  // renderProgressChart (re)builds the chart from the series embedded by the
  // server. htmx replaces the canvas on every /progress swap, so the previous
  // Chart.js instance is destroyed first.
  function renderProgressChart() {
    const canvas = document.getElementById("progress-chart");
    if (!canvas || typeof Chart === "undefined") return;

    if (progressChart) {
      progressChart.destroy();
      progressChart = null;
    }

    const series = parseSeries(canvas);
    if (series.length === 0) return;

    progressChart = new Chart(canvas, {
      type: "line",
      data: {
        labels: series.map((p) => p.label),
        datasets: [
          {
            label: "WPM",
            data: series.map((p) => p.wpm),
            borderColor: "#7aa2f7",
            backgroundColor: "rgba(122,162,247,0.15)",
            tension: 0.3,
            fill: true,
            yAxisID: "y",
          },
          {
            label: "Precisión %",
            data: series.map((p) => p.acc),
            borderColor: "#9ece6a",
            backgroundColor: "rgba(158,206,106,0.12)",
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
            labels: { color: "#c0caf5", boxWidth: 12, font: { size: 10 } },
          },
        },
        scales: {
          x: {
            ticks: { color: "#7c84a8", maxRotation: 0, autoSkip: true, font: { size: 9 } },
            grid: { color: "rgba(124,132,168,0.15)" },
          },
          y: {
            position: "left",
            beginAtZero: true,
            ticks: { color: "#7aa2f7", font: { size: 9 } },
            grid: { color: "rgba(124,132,168,0.15)" },
          },
          y1: {
            position: "right",
            min: 0,
            max: 100,
            ticks: { color: "#9ece6a", font: { size: 9 } },
            grid: { drawOnChartArea: false },
          },
        },
      },
    });
  }

  document.addEventListener("keydown", onKeydown);
  document.body.addEventListener("htmx:afterSwap", (e) => {
    const target = e.detail && e.detail.target;
    if (!target) return;
    if (target.id === "words") startPractice();
    if (target.id === "progress-panel") renderProgressChart();
  });

  renderProgressChart();

  window.ldvdt = { startPractice, composeTyped, renderProgressChart };
})();
