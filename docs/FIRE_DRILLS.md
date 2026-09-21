# Fire drills (v1.0.0)

A **fire drill** proves that an ENGINE_PROVIDED Portable State Archive actually restores.

## Command

```bash
sde fire-drill --archive <archive-dir> [--root .sde]
```

## Behavior

1. Select / load archive manifest  
2. Create an **isolated temporary** restore target (never production paths)  
3. Restore objects + optional journal position  
4. Verify root digest (UNKNOWN ≠ VERIFIED)  
5. Measure restore duration (RTO sample)  
6. Destroy temporary resources  
7. Emit DR / recovery receipt  
8. Update catalog `last_fire_drill` metadata when catalog is present  

## Guarantees

| Promise | Detail |
|---------|--------|
| Non-destructive to production | Temp target only |
| Integrity evidence | Digest comparison in receipt |
| Measured timing | Duration recorded; never fabricated |
| Catalog integration | Last drill status surfaces in `sde readiness` |

## When to run

- After every significant archive policy change  
- On a schedule for DR readiness scorecards  
- Before claiming RECOVERABLE in `sde dr-plan`  

## Related

- [`RECOVERY.md`](RECOVERY.md) · [`DR.md`](DR.md) · [`SAFETY_INVARIANTS.md`](SAFETY_INVARIANTS.md)
- Implementation: `internal/dr`, `internal/recovery`, CLI `fireDrillCmd`
