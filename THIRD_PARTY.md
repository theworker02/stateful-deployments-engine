# Third-Party Notices

Stateful Deployments Engine (SDE) is **proprietary** software. The
Proprietary Pre-Acquisition License in `LICENSE` applies to SDE itself.

This file lists **third-party Go dependencies** declared in `go.mod` and their
licenses, based on the modules’ upstream license metadata. Dependency licenses
apply **only** to those components. They do **not** relicense SDE as
open source.

Always verify current license texts in module source or
https://pkg.go.dev/ when performing formal diligence.

---

## Direct dependencies

| Module | Version (go.mod) | License (typical upstream) | Use in SDE |
|--------|------------------|----------------------------|------------|
| [github.com/spf13/cobra](https://github.com/spf13/cobra) | v1.8.1 | Apache-2.0 | CLI (`cmd/sde`) |
| [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) | v3.0.1 | MIT **and** Apache-2.0 (dual; see upstream NOTICE) | Config / YAML parsing |

## Indirect dependencies

| Module | Version (go.mod) | License (typical upstream) | Brought in by |
|--------|------------------|----------------------------|---------------|
| [github.com/inconshreveable/mousetrap](https://github.com/inconshreveable/mousetrap) | v1.1.0 | Apache-2.0 | cobra |
| [github.com/spf13/pflag](https://github.com/spf13/pflag) | v1.0.5 | BSD-3-Clause | cobra |

---

## Go standard library

The Go toolchain and standard library are used under the
[Go license](https://golang.org/LICENSE) (BSD-style). They are not vendored as
separate SDE source.

---

## Notices

- **Cobra / pflag / mousetrap:** retain Apache-2.0 / BSD notices as required when
  distributing binaries that link these libraries.
- **yaml.v3:** upstream includes dual licensing; consult the module’s LICENSE
  and NOTICE files in the module cache for redistribution obligations.
- **No copyleft dependencies** are intentionally included in `go.mod` at
  v0.1.0-dev. Proposed new dependencies must be disclosed in PRs and reviewed
  under `CONTRIBUTING.md` for proprietary-sale compatibility.

---

## SDE license reminder

```
SDE (this repository’s original code and docs)
  → Proprietary Pre-Acquisition License (LICENSE)

Third-party modules above
  → Their respective open-source licenses
```

Acquisition or commercial license of SDE does not by itself remove your
obligation to comply with third-party licenses for redistributed dependencies.

Contact: matthewlooney5@gmail.com / [@theworker02](https://github.com/theworker02)
