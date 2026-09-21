# Local demo (filesystem adapter)

```bash
go test ./...
go build -o sde ./cmd/sde
./sde init --root .sde
./sde demo --root .sde
./sde status --root .sde
```

See also:

- `examples/sqlite-like/` — journaled SQLite-shaped writes via SDK
- `examples/media-like/` — large cold objects + few hot metadata files
- `examples/log-append/` — append-only hot path
- `examples/dry-run-plan/` — sample MigrationPlan JSON
