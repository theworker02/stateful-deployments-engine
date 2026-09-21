# Transfer Checklist — v1.1.1 FROZEN

Complete at close. Independent project — not affiliated with Railway.

## Repository & build

- [ ] Source repository admin access transferred (or mirrored under buyer org)
- [ ] Freeze tag `v1.1.1` recorded as diligence baseline
- [ ] CI green on buyer-controlled runners (`go test ./...`)
- [ ] Evaluation artifacts regenerated (`evaluate.ps1` / `evaluate.sh`)
- [ ] Release assets downloaded: binaries + `SHA256SUMS.txt` + Cosign bundles
- [ ] Cosign verify performed on each artifact

## Legal & IP

- [ ] `LICENSE` / commercial terms attorney-reviewed
- [ ] Sole Author IP Declaration reviewed (`legal/SOLE_AUTHOR_IP_DECLARATION.md`)
- [ ] `EXECUTED_OWNERSHIP_RECORD.json` archived by buyer counsel
- [ ] APA / assignment executed from counsel-reviewed drafts (`legal-review/`)
- [ ] SBOM reviewed (`sbom/sde-1.1.1.spdx.json`)
- [ ] Third-party IP gate: no UNKNOWN (`THIRD_PARTY_IP_GATE.md`)

## Brand & docs

- [ ] Brand assets + provenance confirmed original
- [ ] Docs current: gap matrix, failure atlas, integration API, RAILWAY_LIVE
- [ ] Pages site URL / fork decision documented
- [ ] FUNDING.yml / public contact updated if ownership changes

## Secrets & operations

- [ ] Secrets rotated (none should be in git history)
- [ ] Railway credentials **not** transferred via this repo
- [ ] Buyer secret store holds any pilot tokens
- [ ] Risk register residual items accepted or scheduled (`RISK_REGISTER.md`)

## Inventory

See `TRANSFER_INVENTORY.json` for path-level asset list.

## Sign-off

| Role | Name | Date |
|------|------|------|
| Seller technical | | |
| Buyer engineering | | |
| Buyer counsel | | |
