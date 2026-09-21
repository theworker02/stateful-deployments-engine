# Glossary (v1.0.0)

| Term | Meaning |
|------|---------|
| **ACTIVE** | Currently serving deployment slot |
| **SHADOW / candidate** | Replacement deployment receiving synced state |
| **STANDBY** | Previous active retained for rollback |
| **Epoch** | Monotonic state generation; advances on commit |
| **Fence** | Token preventing stale coordinators from mutating newer state |
| **Journal** | Append-only log of filesystem mutations after checkpoint |
| **PSA** | Portable State Archive (`ENGINE_PROVIDED`) |
| **ENGINE_PROVIDED** | SDE-owned portability path (not platform-native backup) |
| **RAILWAY_NATIVE** | Platform volume backup behavior as Railway documents it |
| **Cutover write pause** | Measured barrier hold (`CUTOVER_WRITE_PAUSE_MS`) |
| **Fire drill** | Isolated restore test that never touches production |
| **UNKNOWN** | Verification/DR outcome that must never be treated as success |
| **Forward recovery** | Recover without destroying acknowledged post-cutover writes |
| **Receipt** | Immutable evidence artifact for an operation |

Independent project — not affiliated with Railway.
