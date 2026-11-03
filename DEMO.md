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

## Language

Record in English (the default). The header toggle shows the same screens in Korean; mention it in one sentence if time allows.

## Before any screen (0:00–0:20)

Say: "My sister runs a small jewelry brand. Every month she spends her budget on a guess. Rehearsal is for owners with no marketer: it stress-tests the plan before the money is spent."

## Screens

1. **Setup (0:20–0:40)**: point at the "Demo brand · synthetic data" badge and the backtest card (median error 9.5%, 8 of 13 weeks inside range). Say the engine checks itself on the brand's own history first. Click Use demo screenshot → Read my feed. Say: Spark read the real @bnd.jewelry profile: Korea, women 25–34, $200+, flat-lay and close-up photos lead. The owner confirms or fixes it. Spark suggests Stretch on revenue (it saw the diamond wishlist promo in the grid); the three numbers come from the engine's last 13 weeks; say the owner can pull it back to Hold. Click "Connect Instagram (demo data)": followers 5,561, 46% are 25–34, 78% women, 41% Seoul; Reels reach 3.08× feed, on-hand photos get the most saves. Say: mock data shaped like the Instagram Graph API; the real account plugs in after App Review. Confirm, then "See this month's directions".
2. **Directions (0:40–1:05)**: top card "Where to post, and why": Reels and Naver blog lead, feed and Threads support, TikTok skip. Point at one "Your data" line (quotes 3.08× and 262 saves, numbers the engine computed) and one "Spark reasoning" line (industry judgment, labelled). Say: reasoned from the brand's own data and sourced statistics, not stress-tested. Then Spark ranks send_to_creators first. Point at the playbook lines under each card ("Week 1: look for small creators…", "Creators: small creators young women trust, soft clean look"): written for this brand, not generic advice. Say Spark proposes text and ranges only; the engine does all numbers. Select send_to_creators and pay_for_reach, click "Test these".
3. **Simulate (1:05–1:25)**: show the matrix and the adversary's shock list. Say 20 shock combinations are run against each plan, plus a do-nothing baseline.
4. **Verdict (1:25–1:50)**: pay_for_reach survived 15/20, send_to_creators 14/20, baseline 2/20. Flip assumption: seed_post_rate, threshold 0.66 (currently 0.60); read the check line. Click "Commit to this plan" on the **send_to_creators** card (the owner follows Spark's #1 pick), because the Review screen reviews the seeding plan the brand ran in October. Click "Make it happen →".
5. **Launch (1:50–2:15)**: the plan became actions. Creators card: search terms, the piece to send, the DM; click Copy DM; below it the six accounts found for those search terms with engine stats and Spark's fit read. Calendar: 4 weeks with post ideas; click Download calendar. Say: for a plan with ads this screen also shows the exact Meta Marketing API request and creates the campaign paused in the owner's ad account; nothing is spent until they switch it on. Then "4 Review" in the nav.
6. **Review (2:15–2:35)**: October: 1 of 4 creators posted, seed_post_rate 0.6 → 0.49. Click Plan November.
7. **Directions, month 2 (2:35–2:55)**: pay_for_reach now ranks first, send_to_creators last with a caution (survived 15 vs 12). Say: models at both ends, engine in the middle, a person before anything is sent.
