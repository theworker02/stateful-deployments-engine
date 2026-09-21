# Security Model

**CONFIDENTIAL / Proprietary** — Stateful Deployments Engine (SDE)  
Companion to `ARCHITECTURE.md` and `SECURITY.md`. Describes the **threat model**
for journal, barrier, cutover, and rollback. This is an engineering model for
v0.1.0-dev — not a formal certification or guarantee.

---

## 1. Assets

| Asset | Why it matters |
|-------|----------------|
| Persistent application state (volume trees) | User data; integrity is the product promise |
| Mutation journal | Source of truth for catch-up; tampering → silent divergence |
| Checkpoints / manifests (`path → sha256`) | Gate for promote and rollback integrity |
| Active / shadow / standby slots | Wrong traffic target → outage or split-brain UX |
| Write barrier | Defines the operator-visible pause; bypass → races |
| Platform credentials (Railway token, etc.) | Adapter control plane; out of core crypto scope |

---

## 2. Trust boundaries

```
[App / workload writers]
         │  (must cooperate with agent/SDK or FUSE-later)
         ▼
[Agent / Writer] ──append──► [Journal store]
         │                         │
         │                         ▼
         │                  [Replay → Shadow state]
         ▼
[Coordinator] ◄──verify──► [Verifier / manifests]
         │
         ▼
[Platform Adapter] ──API──► [Hosting control plane + DNS/domain]
```

Assumptions (v0):

1. The **host OS and disk** under local demo are trusted for that evaluation.
2. The **platform API** (Railway, etc.) is trusted to perform the operations the
   adapter requests; SDE does not currently attest the platform’s honesty.
3. **Writers** that bypass the agent (raw FS writes while claiming journaled
   mode) are **out of model** until FUSE or mandatory interception exists —
   document this to operators.
4. Evaluation viewers are not trusted with production credentials.

---

## 3. Adversaries / failure classes

| Class | Examples |
|-------|----------|
| **Benign crash / partial failure** | Process kill mid-barrier; network blip during sync |
| **Corrupt storage** | Bit flips, truncated files, incomplete copies |
| **Buggy candidate image** | Shadow unhealthy; bad migrations |
| **Malicious or compromised shadow** | Candidate tries to serve traffic early |
| **Journal tampering** | Deleted/reordered entries; forged seq/epoch |
| **Confused deputy on adapter** | Wrong service ID / domain re-point |
| **Insider with repo access** | Pre-sale: follows `LICENSE`; post-sale: buyer’s IAM |

---

## 4. Journal threats

| Threat | Mitigation (current / planned) |
|--------|--------------------------------|
| Gap in sequence numbers | Replay / sync logic must not silently skip; treat gaps as hard errors (harden in near-term) |
| Epoch rollback confusion | Epoch advances only on successful cutover; rollback pins to standby epoch |
| Oversized / hostile payloads | Demo agent is trusted; production agent needs size limits and path sandboxing |
| Journal store deletion | Deploy abort; active remains source of truth if barrier not held |
| Replay onto wrong root | Slots carry distinct `StatePath`; adapter must never share one Railway volume |

**Residual risk:** Without cryptographic authentication of journal entries
(HMAC/signatures), a host attacker who can write the journal can forge history.
v0 treats journal authenticity as **host-trust**. Post-acquisition hardening may
add signed entries.

---

## 5. Barrier threats

| Threat | Mitigation |
|--------|------------|
| Writers continue during barrier | Agent must fence; apps bypassing agent break the model |
| Coordinator crash while fenced | **Planned:** agent TTL releases fence; operator runbook |
| Premature barrier release | Only release after promote **or** abort path that leaves active consistent |
| Long barrier (DoS UX) | Large final delta / slow disk; mitigate with better catch-up before barrier |

Barrier duration is the **only** intentional write pause. Read availability during
barrier depends on app design (reads may continue if not fenced).

---

## 6. Cutover / traffic threats

| Threat | Mitigation |
|--------|------------|
| Promote with divergent manifests | Hard verify under barrier; refuse cutover on mismatch |
| Health check false positive | Adapter health probes; still require verify |
| DNS / domain flip without standby | Protocol retains previous active as **standby** |
| Split traffic | Adapter must make cutover atomic relative to platform primitives (Railway domain binding); document eventual DNS TTL effects |
| Shadow exposed before ready | Do not TransferTraffic until verify + health succeed |

---

## 7. Rollback threats

| Threat | Mitigation |
|--------|------------|
| Rollback to corrupt standby | Integrity verify against pinned checkpoint / root hash |
| Journal epoch drift after rollback | Pin epoch to standby; rebuild checkpoint |
| Operator rolls back wrong deploy | CLI/report surfaces from/to image + epoch explicitly |

Rollback restores **application revision and state epoch**, not merely the
container image.

---

## 8. Adapter / Railway-specific

- Never mount the same Railway volume on two live deployments (platform forbids
  it; doing so would also void the safety model).
- Dual-service strategy isolates corruption of shadow volume from active.
- Tokens in env (`RAILWAY_TOKEN`) must not be committed; use secrets managers in
  any licensed production setting.
- Stub adapter returns `ErrNotWired` until live — reduces accidental “prod”
  calls against incomplete code.

See `docs/RAILWAY.md`.

---

## 9. Cryptography

| Use | Algorithm / approach |
|-----|----------------------|
| File / manifest hashing | SHA-256 content addressing |
| Transport to platform APIs | HTTPS as provided by platform SDKs (when wired) |
| Journal authentication | **Not in v0** (host-trust) |

---

## 10. Security invariants (acceptance)

A build is not “cutover-safe” unless:

1. Post-barrier active and shadow manifests match (`ConsistencyOK`).
2. Failed verify aborts without traffic transfer and releases barrier with
   active still serving.
3. Rollback reports `IntegrityOK` only after verification.
4. Railway adapter never attempts concurrent mount of one volume.

Eval procedures: `EVALUATION.md`. Reporting: `SECURITY.md`.
