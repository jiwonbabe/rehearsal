# Rehearsal — project rules

Rehearsal tests a small brand's monthly marketing plan before the money is spent.
Deadline: Nov 4, 2026, 6pm PT. Two days. Scope is in SPEC.md; do not widen it.

## Non-negotiable architecture

1. **Models propose, code computes.** Muse Spark produces direction text and parameter
   ranges (JSON). It never produces an order count, a revenue figure, or a rate.
   Every number on screen comes from `engine/`.
2. **Deterministic engine.** Same inputs → same outputs. All randomness goes through
   `rand.New(rand.NewSource(seed))` with a fixed seed. No `time.Now()` inside the engine.
3. **Oracle isolation.** `spark/` must not import `rehearsal/engine` or read `reference/`.
   `engine/` must not import `spark/`. Only `cmd/server` wires the two together.
4. **Backtest before any plan is shown.** `engine.RunBacktest()` runs on the brand data and
   its error is displayed in the UI. Never hide it.
5. **Honest labels.** The demo brand data is synthetic. Every screen shows
   "Demo brand · synthetic data". The words "forecast" and "prediction" do not appear
   anywhere in the product. Use: stress test, survived, dies if, flip assumption.

## Code style

- Go 1.22, **stdlib only** (net/http, html/template, encoding/json, math/rand with a fixed
  seed). No third-party modules. `gofmt` and `go vet` clean before every commit.
- Plain structs with plain field names (`Orders`, `SpendUsd`, `PostMix`); JSON tags match the
  Python keys in SPEC.md. Functions over frameworks. No generics unless unavoidable.
- Same variable names and line order across similar functions. Shared logic goes in a
  helper, not copy-paste with small differences.
- `reference/` holds the Python reference engine and golden fixtures. `go test ./...` compares
  the Go engine to those fixtures. If you change a formula, change it in BOTH, re-run
  `python reference/dump_fixtures.py`, then `go test ./...`. Never edit a fixture by hand.
- UI: `cmd/server` serves JSON from the engine and static pages from `web/`. Vanilla
  HTML/JS, no framework, no build step. Copy layouts from `mockup/`.
- Commit after each working step. Commit messages: what works now, one line.

## Working order (SPEC.md § Build order)

Finish each step fully, run tests, commit, then move on. Do not start the UI before
`go test ./...` passes. Do not call Spark live during the demo recording; cache responses in
`spark/cache/`.

## When unsure

Pick the simpler option, write one line in `NOTES.md` saying what you assumed, and keep going.
Do not ask the user about things SPEC.md already decides.
