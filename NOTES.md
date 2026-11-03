# Notes (assumptions, one line each)
- Spark cache is keyed by call name (directions-m1, directions-m2, adversary-<id>), not by prompt hash: files are committed and hand-editable.
- Shock codes outside the catalog are dropped by engine.CleanCombos; spark.ParseShocks only extracts the lists.
- spark/cache/*.json were written by hand in the prompt output shape (no Muse API key at build time); cmd/sparkcache regenerates them when a key is set.
- Server handlers share one mutex via locked(); /api/setup does not recalibrate rates, it only changes budget/goal/target/market.
