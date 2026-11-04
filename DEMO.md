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

## Meta ad account (optional, makes Launch live)

Before recording, in the shell that starts the server: `export META_AD_ACCOUNT_ID=<act id> META_ACCESS_TOKEN=<60-day token>`. Launch then shows Meta's live delivery estimate for the audience and "Create paused campaign" really creates a paused campaign + ad set in that account (nothing spends; delete it in Ads Manager afterwards). Without the two variables the button shows the request it would send. Never commit the token.

## Language

Record in English (the default). The header toggle shows the same screens in Korean; mention it in one sentence if time allows.

## Before any screen (0:00–0:20)

Say: "My sister runs a small jewelry brand. Every month she spends her budget on a guess. Rehearsal is for owners with no marketer: it stress-tests the plan before the money is spent."

## Screens

1. **Setup (0:20–0:40)**: point at the "Demo brand · synthetic data" badge and the trust card: holdout backtest (the engine learned from the first 39 weeks, the last 13 were hidden: typical miss 14.7%, 8 of 13 inside range) and the known-answer table (the generator's rates next to what the engine recovered, most within 4%). Say: the inputs are demo data, so we can show the engine recovering a known answer; a real brand gets the holdout check on its own history. Click Use demo screenshot → Read my feed. Say: Spark read the real @bnd.jewelry profile: Korea, women 25–34, $200+, flat-lay and close-up photos lead. The owner confirms or fixes it. Spark suggests Stretch on revenue (it saw the diamond wishlist promo in the grid); the three numbers come from the engine's last 13 weeks; click **Hold** (46 orders, the current goal) so the numbers below match. Click "Connect Instagram · demo account": a fictional @demo.brand with mock insights (46% are 25–34, 78% women; Reels reach 3.08× feed, on-hand photos get the most saves). Point at the chips: "Real · your screenshot" on the profile read, "Demo input" on the insights, "Engine" on the trust card. Say: the real account plugs into the same seam after App Review; we never show mock numbers under a real handle. Confirm, then "See this month's directions".
2. **Directions (0:40–1:05)**: first card "This month: November · demand 1.04× · ad cost 1.18×" with Spark's one line ("a little more shopping interest, ads cost more, so favor your own posts and creators") and the two sources (Census jewelry-store sales via FRED, Gupta Media Meta CPM). Say: a seasonal prior from public data, applied to this month only, never learned from one year of history. Next card "Where to post, and why": Reels and Naver blog lead, feed and Threads support, TikTok skip. Point at one "Your data" line (quotes 3.08× and 262 saves, numbers the engine computed) and one "Spark reasoning" line (industry judgment, labelled). Say: reasoned from the brand's own data and sourced statistics, not stress-tested. Then Spark ranks send_to_creators first. Point at the playbook lines under each card ("Week 1: look for small creators…", "Creators: small creators young women trust, soft clean look"): written for this brand, not generic advice. Say Spark proposes text and ranges only; the engine does all numbers. Select send_to_creators and pay_for_reach, click "Test these".
3. **Simulate (1:05–1:25)**: show the matrix and the adversary's shock list. Say 20 shock combinations are run against each plan, plus a do-nothing baseline.
4. **Verdict (1:25–1:50)**: pay_for_reach survived 15/20, send_to_creators 13/20, baseline 5/20. Flip assumption: seed_post_rate, threshold 0.88 (currently 0.60); read the check line (Spark: "check this week if enough creators will actually show your pieces to get from 60% to 88%"). Click "Commit to this plan" on the **send_to_creators** card (the owner follows Spark's #1 pick), because the Review screen reviews the seeding plan the brand ran in October. Click "Make it happen →".
5. **Launch (1:50–2:15)**: the plan became actions. Creators card: search terms, the piece to send, the DM; click Copy DM; below it the six accounts found for those search terms with engine stats and Spark's fit read. Calendar: 4 weeks with post ideas; click Download calendar. Say: for a plan with ads this screen also shows the exact Meta Marketing API request and creates the campaign paused in the owner's ad account; nothing is spent until they switch it on. Then "4 Review" in the nav.
6. **Review (2:15–2:35)**: October: 1 of 4 creators posted, seed_post_rate 0.6 → 0.49. Click Plan November.
7. **Directions, month 2 (2:35–2:55)**: the month card now says December (demand 2.23×, ad cost 1.27×); send_to_creators ranks last with a caution. If time allows, Test these on pay_for_reach + send_to_creators: ads 20/20, seeding 15/20 and it dies if "creators don't post + the seasonal lift does not come". Say: models at both ends, engine in the middle, a person before anything is sent.
