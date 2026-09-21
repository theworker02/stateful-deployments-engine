/* Shared chrome for SDE Pages site */
(function () {
  const pages = [
    { href: "index.html", label: "Home" },
    { href: "demo.html", label: "Live Demo" },
    { href: "proof.html", label: "Proof" },
    { href: "architecture.html", label: "Architecture" },
    { href: "benchmarks.html", label: "Benchmarks" },
    { href: "failure-atlas.html", label: "Failure Atlas" },
    { href: "railway-fit.html", label: "Railway Fit" },
    { href: "security.html", label: "Security" },
    { href: "due-diligence.html", label: "Due Diligence" },
  ];

  const here = (location.pathname.split("/").pop() || "index.html").toLowerCase();

  function mountHeader() {
    const host = document.getElementById("site-header");
    if (!host) return;
    host.innerHTML = `
      <div class="brand-row">
        <a class="brand" href="index.html">
          <img src="brand/logo.svg" alt="SDE logo"/>
          <div class="brand-text">
            <strong>Stateful Deployments Engine</strong>
            <span>v1.1.1 · independent · not affiliated with Railway</span>
          </div>
        </a>
        <nav class="primary" aria-label="Primary">
          ${pages
            .map(
              (p) =>
                `<a href="${p.href}"${p.href === here ? ' aria-current="page"' : ""}>${p.label}</a>`
            )
            .join("")}
        </nav>
      </div>`;
  }

  function mountFooter() {
    const host = document.getElementById("site-footer");
    if (!host) return;
    host.innerHTML = `
      <p><strong>SDE v1.1.1 FROZEN</strong> — browser pages are labeled simulations.
      Local evaluation: <code>evaluate.ps1</code> / <code>evaluate.sh</code>.</p>
      <p>Independent project. Proprietary Pre-Acquisition License. No Railway trademarks used as branding.</p>`;
  }

  mountHeader();
  mountFooter();
})();
