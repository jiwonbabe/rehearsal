# Demo script (under 3:00)

No live API calls; everything is served from `spark/cache/`. Restart the server to reset to month 1.

## Deploy

Fly CLI was not available when this was written. From the repo root:

```
fly launch --copy-config --no-deploy
fly deploy --ha=false
```

State (month, rates) lives in server memory, so run exactly one machine; restart it to reset to month 1.

Then put `https://rehearsal.fly.dev` in README's Live line. Local container: `docker build -t rehearsal . && docker run --rm -p 8080:8080 rehearsal`.

## Before any screen (0:00–0:20)

Say: "My sister runs a small jewelry brand. Every month she spends her budget on a guess. Rehearsal is for owners with no marketer: it stress-tests the plan before the money is spent."

## Screens

1. **Setup (0:20–0:40)**: point at the "Demo brand · synthetic data" badge and the backtest card (median error 9.5%, 8 of 13 weeks inside range). Say the engine checks itself on the brand's own history first. Click Use demo screenshot → Read my feed. Say: Spark read the real @bnd.jewelry profile: Korea, women 25–34, $200+, flat-lay and close-up photos lead. The owner confirms or fixes it, then Confirm. Click "See this month's directions".
2. **Directions (0:40–1:05)**: Spark ranks send_to_creators first. Say Spark proposes text and ranges only; the engine does all numbers. Select send_to_creators and pay_for_reach, click "Test these".
3. **Simulate (1:05–1:35)**: show the matrix, the planner JSON and the adversary's shock list. Say 20 shock combinations are run against each plan, plus a do-nothing baseline.
4. **Verdict (1:35–2:05)**: pay_for_reach survived 15/20, send_to_creators 14/20, baseline 2/20. Flip assumption: seed_post_rate, threshold 0.66 (currently 0.60); read the check line. Click "Commit to this plan" on the **send_to_creators** card (the owner follows Spark's #1 pick), because the Review screen reviews the seeding plan the brand ran in October. Then "Review October".
5. **Review (2:05–2:30)**: October: 1 of 4 creators posted, seed_post_rate 0.6 → 0.49. Click Plan November.
6. **Directions, month 2 (2:30–2:50)**: pay_for_reach now ranks first, send_to_creators last with a caution (survived 15 vs 12). Say: models at both ends, engine in the middle, a person at the end.
