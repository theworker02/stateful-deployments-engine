# Contributing

Thank you for your interest in Stateful Deployments Engine (SDE).

## Important: this is not open source

SDE is distributed under the **Proprietary Pre-Acquisition License** (`LICENSE`).
There is **no implied open-source grant**. Viewing the repository for evaluation
or diligence does **not** authorize you to use, modify, distribute, or
commercialize the Software.

Contributions are accepted **only** under terms that assign or license rights to
the Owner so the IP remains clean for sale or commercial licensing.

## How to discuss ideas (preferred first step)

Before sending code:

1. Open a **non-security** GitHub Discussion or Issue describing the idea, **or**
2. Email matthewlooney5@gmail.com with subject `[SDE IDEA] …`

Security issues: follow `SECURITY.md` (private coordinated disclosure).

Acquisition / commercial license: see `ACQUISITION.md` and `LICENSE` §11.

## Contributor License Agreement (CLA) / assignment

**No pull request will be merged** without a signed CLA or IP assignment
agreeable to the Owner. At minimum, contributors must confirm that:

1. You are legally entitled to submit the contribution.
2. You grant the Owner a perpetual, worldwide, royalty-free, irrevocable right
   to use, modify, sublicense, sell, and relicense the contribution as part of
   SDE (including under proprietary terms).
3. You waive (to the extent permitted by law) claims that would block the Owner
   from commercializing or assigning SDE.
4. Your contribution does not knowingly include third-party code under a
   copyleft or conflicting license without prior written disclosure.

A formal CLA document will be provided before first merge. Until then, do not
assume that submitting a PR creates any license *to you* or joint ownership.

## What kinds of contributions are useful

| Welcome (after CLA) | Generally not accepted |
|---------------------|------------------------|
| Bug fixes with tests | Large refactors without prior discussion |
| Bench / fault-injection harness improvements | Dependencies that complicate proprietary relicensing |
| Docs accuracy (citing platform docs) | Features that assume production use without Agreement |
| Adapter stubs that match `internal/adapter` | Code copied from GPL/AGPL or unknown-license sources |

## Development setup

```bash
go test ./...
go build -o sde ./cmd/sde
./sde demo
```

See `README.md`, `EVALUATION.md`, and `examples/local-demo.md`.

## Code style

- Match existing Go style in `internal/` and `cmd/`.
- Prefer small, reviewed PRs with tests for coordinator / journal / verifier paths.
- Do not add telemetry that phones home without explicit discussion.

## Conduct

Participation is governed by `CODE_OF_CONDUCT.md`.

## License of submissions

By submitting a contribution (including issues with substantial code snippets),
you agree that if accepted it will be owned/licensed per the CLA and may be
distributed solely under the Owner’s proprietary terms — not under Apache-2.0,
MIT, or any other open-source license unless the Owner later chooses otherwise
in writing.
