# Acquisition demo

One-command local destroy-source / restore / verify / receipt flow using
**ENGINE_PROVIDED** Portable State Archives (not Railway-native backup).

```bash
go run ./examples/acquisition-demo
```

Flow: create state → export PSA → verify → escrow catalog → recovery point →
destroy source → restore to new target → recovery receipt → fire drill → DR plan.
