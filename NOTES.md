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
