# Transfer Checklist

- [ ] Source repository access transferred
- [ ] `LICENSE` / commercial terms attorney-reviewed
- [ ] SBOM reviewed (`acquisition/sbom/`)
- [ ] Third-party IP gate: no UNKNOWN
- [ ] Brand assets + provenance confirmed original
- [ ] CI green (`go test ./...`)
- [ ] Evaluation artifacts regenerated
- [ ] Secrets rotated (none should be in git history)
- [ ] Railway credentials **not** transferred via this repo
- [ ] Docs: gap matrix, failure atlas, integration API current
- [ ] Version tag `v1.1.0` + checksums recorded (`dist/SHA256SUMS.txt`)

See `TRANSFER_INVENTORY.json`.
