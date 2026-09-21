# GitHub Pages live demo

Static site for **Stateful Deployments Engine v1.1.0**.

## Enable

1. Push to `main` / `master`
2. Repo **Settings → Pages → Build and deployment → GitHub Actions**
3. Workflow: `.github/workflows/pages.yml`

## Local preview

Open `index.html` in a browser (assets under `brand/`, data under `data/`).

## Contents

| Page | Purpose |
|------|---------|
| `index.html` | Product home |
| `demo.html` | Interactive cutover / restore simulation |
| `architecture.html` | Component explorer |
| `benchmarks.html` | Evidence JSON |
| `failure-atlas.html` | Failure map summary |
| `railway-fit.html` | Correct Railway framing |
| `security.html` / `due-diligence.html` | Diligence pointers |

Simulations are clearly labeled. Real evidence: run `evaluate.ps1` in the repo root.
