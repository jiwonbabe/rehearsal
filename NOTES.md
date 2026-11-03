# Notes (assumptions, one line each)
- Spark cache is keyed by call name (directions-m1, directions-m2, adversary-<id>), not by prompt hash: files are committed and hand-editable.
- Shock codes outside the catalog are dropped by engine.CleanCombos; spark.ParseShocks only extracts the lists.
- Server handlers share one mutex via locked(); /api/setup does not recalibrate rates, it only changes budget/goal/target/market.
- sparkcache builds the month-1 planner summary from the weeks before October (pre-review view); month 2 adds the October review block. Engine rates in the server are always calibrated on the full data.
- spark/cache/*.json were generated live on Nov 2, 2026 with muse-spark-1.3 via `go run ./cmd/sparkcache` (api.meta.ai/v1). Two prompt rules were added to prompts/directions.md so ranges reach what the budget allows; without them Spark anchored creator_count to the 2-3 the brand had tried and no flip existed.
- Live month-1 directions include a `mix` direction; the demo picks pay_for_reach + send_to_creators.
- Docker CLI is installed but its daemon was not running, so the Dockerfile was not smoke-tested; run the build command in DEMO.md before relying on it.
- Fly CLI (fly/flyctl) is not installed, so nothing was deployed; the two deploy commands are at the top of DEMO.md and README's Live line says "Deploy: see DEMO.md".
- Server is cache-only: main() unsets MUSE_API_KEY so the demo never calls Spark live; cmd/sparkcache is the only live caller.
- The profile read is cached by handle (spark/cache/profile-<handle>.json): the same handle returns the same read regardless of the image sent.
- /api/profile takes `source`; only "screenshot" exists. Instagram Login (Graph API) will add `source: "instagram"` with the same prompt and ValidateProfile after Meta App Review.

- Success metric revenue is display-only (goal ÷ avg price); the engine tests orders.
- spark/cache/playbook-<id>.json generated live on Nov 3, 2026 (Spark role 5, prompts/playbook.md) from the cached profile read + brand basics; one file per direction id, month 2 reuses it.
- Creator fit read (Spark role 6, prompts/creator.md) is cached by handle in spark/cache/creator-<handle>.json; the demo server has none until screenshots are dropped into web/demo-creators/<handle>.png and sparkcache is rerun.
- Launch numbers: creator_dms = creator_count × 2 (assume half reply); ads daily budget = ads_usd / 28 days; posts per type by largest remainder of posts_per_week × mix.
- Meta ads adapter creates campaign + ad set only (OUTCOME_TRAFFIC + LINK_CLICKS, no pixel; not yet tried against a real ad account); the owner picks the boosted posts in Ads Manager. META_* env vars are read per request; MUSE_API_KEY stays unset in the server.
