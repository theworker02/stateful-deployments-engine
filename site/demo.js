/**
 * Browser-safe interactive simulation of SDE cutover + recovery.
 * Clearly labeled: NOT a live control plane; no network calls.
 */
(function () {
  const CUTOVER_PHASES = [
    "ACTIVE",
    "CHECKPOINT",
    "SHADOW",
    "SYNCHRONIZING",
    "VERIFYING",
    "QUIESCING",
    "FINAL_DELTA",
    "CUTOVER",
    "OBSERVING",
    "COMMITTED",
  ];

  const RESTORE_PHASES = [
    "EXPORT_PSA",
    "VERIFY_ARCHIVE",
    "DESTROY_SOURCE",
    "PROVISION_TARGET",
    "RESTORE",
    "REPLAY_JOURNAL",
    "DIGEST_MATCH",
    "RECOVERY_RECEIPT",
  ];

  const state = {
    running: false,
    aborted: false,
    injectFailAt: null,
    pauseMs: null,
    mutations: 0,
  };

  const $ = (id) => document.getElementById(id);

  function sleep(ms) {
    return new Promise((r) => setTimeout(r, ms));
  }

  function log(line) {
    const el = $("log");
    if (!el) return;
    el.textContent += line + "\n";
    el.scrollTop = el.scrollHeight;
  }

  function clearLog() {
    const el = $("log");
    if (el) el.textContent = "";
  }

  function renderPipeline(phases, current, failed) {
    const el = $("pipeline");
    if (!el) return;
    el.innerHTML = phases
      .map((p, i) => {
        let cls = "phase";
        if (failed === p) cls += " fail";
        else if (i < current) cls += " done";
        else if (i === current) cls += " active";
        return `<span class="${cls}">${p}</span>`;
      })
      .join("");
  }

  function setMetric(id, value) {
    const el = $(id);
    if (el) el.textContent = value;
  }

  function setBusy(busy) {
    state.running = busy;
    ["btn-cutover", "btn-restore", "btn-fail", "btn-clone"].forEach((id) => {
      const b = $(id);
      if (b) b.disabled = busy && id !== "btn-fail";
    });
  }

  async function runCutover() {
    if (state.running) return;
    setBusy(true);
    state.aborted = false;
    clearLog();
    log("=== SIMULATED TRANSACTIONAL CUTOVER ===");
    log("Mode: browser simulation (ENGINE_PROVIDED semantics)");
    log("Independent project — not affiliated with Railway\n");

    state.mutations = 12000 + Math.floor(Math.random() * 8000);
    setMetric("m-mutations", state.mutations.toLocaleString());
    setMetric("m-pause", "—");
    setMetric("m-result", "RUNNING");
    setMetric("m-rpo", "—");

    for (let i = 0; i < CUTOVER_PHASES.length; i++) {
      if (state.aborted) {
        renderPipeline(CUTOVER_PHASES, i, CUTOVER_PHASES[i]);
        log("✗ ABORTED by operator / injected failure at " + CUTOVER_PHASES[i]);
        setMetric("m-result", "ABORTED → RECOVERING");
        setBusy(false);
        return;
      }
      if (state.injectFailAt === CUTOVER_PHASES[i]) {
        renderPipeline(CUTOVER_PHASES, i, CUTOVER_PHASES[i]);
        log("✗ FAULT INJECTED at " + CUTOVER_PHASES[i]);
        log("Invariant: no false VERIFIED; standby retained");
        log("Recovery path: ROLLBACK_SAFE (no post-cutover writes yet)");
        setMetric("m-result", "FAULT → ROLLED_BACK");
        state.injectFailAt = null;
        setBusy(false);
        return;
      }
      renderPipeline(CUTOVER_PHASES, i, null);
      const p = CUTOVER_PHASES[i];
      if (p === "SYNCHRONIZING") {
        log(`✓ ${p} — ${state.mutations.toLocaleString()} mutations synchronized`);
      } else if (p === "VERIFYING") {
        log(`✓ ${p} — CHECKSUM receipt VERIFIED (UNKNOWN≠VERIFIED)`);
      } else if (p === "QUIESCING") {
        log(`✓ ${p} — write barrier acquired`);
      } else if (p === "FINAL_DELTA") {
        const delta = 5 + Math.floor(Math.random() * 40);
        log(`✓ ${p} — final delta: ${delta} writes`);
      } else if (p === "CUTOVER") {
        state.pauseMs = +(8 + Math.random() * 45).toFixed(2);
        setMetric("m-pause", state.pauseMs + " ms");
        log(`✓ ${p} — CUTOVER_WRITE_PAUSE_MS=${state.pauseMs} (simulated)`);
      } else if (p === "OBSERVING") {
        log(`✓ ${p} — TARGET_ACTIVE_UNCOMMITTED window`);
      } else if (p === "COMMITTED") {
        log(`✓ ${p}`);
        log("\nSTATEFUL ZERO-DOWNTIME DEPLOYMENT (simulated)");
        log("Portability class: ENGINE_PROVIDED");
      } else {
        log(`✓ ${p}`);
      }
      await sleep(p === "SYNCHRONIZING" ? 650 : 380);
    }
    setMetric("m-result", "COMMITTED");
    setMetric("m-rpo", "0 mutations");
    renderPipeline(CUTOVER_PHASES, CUTOVER_PHASES.length, null);
    setBusy(false);
  }

  async function runRestore() {
    if (state.running) return;
    setBusy(true);
    clearLog();
    log("=== SIMULATED RECOVERY INDEPENDENCE ===");
    log("Source environment will be destroyed after archive seal\n");
    setMetric("m-result", "RESTORING");
    for (let i = 0; i < RESTORE_PHASES.length; i++) {
      renderPipeline(RESTORE_PHASES, i, null);
      const p = RESTORE_PHASES[i];
      if (p === "DESTROY_SOURCE") {
        log(`✓ ${p} — simulated project/env removed`);
      } else if (p === "DIGEST_MATCH") {
        log(`✓ ${p} — root digest equal (VERIFIED)`);
      } else if (p === "RECOVERY_RECEIPT") {
        log(`✓ ${p} — RECOVERY_RECEIPT.json`);
        log("\nHeadline: archive usable after source disappearance");
      } else {
        log(`✓ ${p}`);
      }
      await sleep(420);
    }
    setMetric("m-result", "RECOVERABLE");
    setMetric("m-rpo", "epoch pinned");
    setMetric("m-pause", "n/a");
    renderPipeline(RESTORE_PHASES, RESTORE_PHASES.length, null);
    setBusy(false);
  }

  function injectFail() {
    if (!state.running) {
      state.injectFailAt = "FINAL_DELTA";
      log("Armed: fail at FINAL_DELTA on next cutover run");
      return;
    }
    state.aborted = true;
    log("…operator abort requested");
  }

  async function runClone() {
    if (state.running) return;
    setBusy(true);
    clearLog();
    const phases = [
      "CAPTURE_SOURCE",
      "EXPORT_PSA",
      "VERIFY_ARCHIVE",
      "NEW_IDENTITY",
      "MATERIALIZE_CLONE",
      "SANITIZE_HOOK",
      "VERIFY_DIGEST",
      "CLONE_RECEIPT",
    ];
    log("=== SIMULATED ENVIRONMENT CLONE (ENGINE_PROVIDED) ===");
    log("Writable clone with new identity — original archive immutable\n");
    setMetric("m-result", "CLONING");
    for (let i = 0; i < phases.length; i++) {
      renderPipeline(phases, i, null);
      log("✓ " + phases[i]);
      await sleep(380);
    }
    setMetric("m-result", "CLONE_VERIFIED");
    setMetric("m-rpo", "new epoch id");
    setMetric("m-pause", "n/a");
    renderPipeline(phases, phases.length, null);
    log("\nCLONE_RECEIPT.json — production archive unchanged");
    setBusy(false);
  }

  window.SDEDemo = { runCutover, runRestore, injectFail, runClone };

  document.addEventListener("DOMContentLoaded", () => {
    renderPipeline(CUTOVER_PHASES, -1, null);
    const a = $("btn-cutover");
    const b = $("btn-restore");
    const c = $("btn-fail");
    const d = $("btn-clone");
    if (a) a.onclick = runCutover;
    if (b) b.onclick = runRestore;
    if (c) c.onclick = injectFail;
    if (d) d.onclick = runClone;
  });
})();
