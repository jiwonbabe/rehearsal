# Rehearsal

Built Nov 2–4, 2026 during the Meta Global AI Developer Hackathon. No external code; Go stdlib only.

Stress-test a small brand's monthly marketing plan before the money is spent. Go, stdlib only. See SPEC.md.

```
python data/synthetic.py            # synthetic demo brand
python reference/dump_fixtures.py   # golden fixtures from the Python reference engine
go test ./...                       # Go engine must match them
go run ./cmd/server                 # http://localhost:8080
```

`go run ./cmd/sparkcache` regenerates `spark/cache/` when `MUSE_API_KEY` is set.

Rehearsal stress-tests a small brand's monthly marketing plan against shocks and finds the one assumption that flips the verdict. It runs an **agent loop**: planner (Muse Spark) → engine → adversary (Muse Spark) → engine → verdict + check line (Muse Spark) → human commits. Models at both ends, engine in the middle, a person at the end. A fourth Spark role reads the owner's Instagram profile screenshot into labels only (market, customer guess, post types); the owner confirms or fixes them. Muse Spark only writes direction text and parameter ranges; every number on screen comes from the deterministic Go engine. Oracle isolation is enforced by imports: `spark/` cannot import `engine/`, and only `cmd/` wires them together. Two implementations produce the same numbers: the Python `reference/engine` writes golden fixtures and `go test ./...` compares the Go engine to them. Before any plan is shown the engine backtests itself on the brand data: median absolute error 9.5%, 8 of 13 weeks inside the stated range (`/api/backtest`). The demo brand is synthetic (`data/synthetic.py`, seed 7).

## Model access

Default is the Meta Model API: `MUSE_API_KEY`, `MUSE_BASE_URL` (default `https://api.meta.ai/v1`), `MUSE_MODEL` (default `muse-spark-1.3`).
Fallback is Baseten's OpenAI-compatible endpoint: set `MUSE_MODEL=meta-models/muse-spark-open-1026`, `MUSE_BASE_URL` to Baseten's base URL (confirm it on Baseten's docs) and `MUSE_API_KEY` to the Baseten key. Same prompts and code; only env vars change.
The demo runs from `spark/cache/` with no live calls.

## Live

Deploy: see DEMO.md

## Next

Instagram Login (Graph API, `instagram_business_basic`) replaces the screenshot: `/api/profile` gains `source: "instagram"` and the server fetches recent media itself; same prompt, same validation. Needs Meta App Review (2–4 weeks).

Followers and inquiries as success metrics (the engine models orders only today).

Shopify OAuth, Meta Marketing API, Instagram Insights, SAM 3.1 post classification (synthetic data carries type labels), paste-your-plan, web_search priors, mid-month check, discount code / UTM generation, CSV upload, multi-brand.
