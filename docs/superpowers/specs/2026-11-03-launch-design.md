# Launch: brand-specific playbooks and one-click execution — design

Date: 2026-11-03. Deadline: Nov 4, 2026, 6pm PT. Decided by Claude without a brainstorm
round (the user asked for a plan directly); every choice below is Claude's call unless marked.

## Intent

Today the four directions read like generic advice ("post 3–5 times, seed creators, boost
posts"). The owner could get that from any chatbot. Two things change:

1. **Every direction carries a playbook written for THIS brand**: what to post where, which
   posts to put money behind, which creators to look for and what to DM them, week by week.
   Spark writes it from the profile read (style, likely customer, best post type) and the
   direction. Numbers still come from the engine.
2. **The committed plan becomes actions, not a to-do list.** A new screen, Launch, turns the
   stress-tested plan into three things the owner can act on with one click: a paused Meta
   ads campaign in their ad account, a creator seeding kit (search terms, DM text, what to
   send), and a 4-week posting calendar they can download.

The pitch line for judges: planner → engine → adversary → engine → verdict → human →
**executor**. The agent loop now ends in the owner's ad account and calendar, not in prose.

## Goals

- Directions page shows, per card, a concrete week-1 action and the ad audience or creator
  profile in the owner's words, so no card could have been written for another brand.
- Verdict "Commit" leads to Launch. Launch shows the exact Meta Marketing API request the
  server will send (campaign + ad set, both `PAUSED`), and sends it when credentials exist.
- Creator kit: who to look for, 3–5 search terms/hashtags, which piece to send, the ask, a
  DM the owner can copy. Count, ship cost and DM count come from the engine.
- Calendar: posts per week per type (engine) with one post idea per week (Spark), downloadable
  as `.ics`.
- Everything works offline from `spark/cache/` during the demo. No live Spark call from the
  server (unchanged rule).

## Non-goals (this round)

- Creating the ad creative itself. The owner picks the boosted posts in Ads Manager. The
  campaign and ad set are created paused with the targeting and budget filled in.
- Discovering creators automatically. Instagram has no public search API; the kit gives
  search terms and, if time allows, a "does this creator fit?" read of a pasted screenshot
  (same seam as the profile read).
- TikTok or any platform other than Instagram/Facebook placements.
- Changing the directions prompt, ranges, or the cached directions/adversary/check files.
  The flip (0.66 vs 0.60) stays as it is. The playbook is a separate Spark call per direction.

## Anti-goals

- Any number on screen that Spark produced. Playbook strings are validated: a string that
  contains `$<digit>` or `<digit> orders|revenue|%` is dropped.
- Spending money without the owner. The server only ever creates `PAUSED` objects; the
  Launch screen says so above the button.
- Hidden live calls. Ads go live only when `META_AD_ACCOUNT_ID` and `META_ACCESS_TOKEN` are
  set; otherwise the button returns the request it would have sent and says so.

## Constraints (unchanged from CLAUDE.md)

Go 1.22 stdlib only; deterministic engine (no `time.Now`; dates derive from the last data
week); oracle isolation (`spark/` never imports `engine`, `engine/` never imports `spark`;
only `cmd/*` wire them); every engine formula mirrored in `reference/engine` with a fixture;
"forecast"/"prediction" never appear; the "Demo brand · synthetic data" badge on every page.

## How it is built (decided)

### Spark role 5: executor brief (`prompts/playbook.md`, cache `playbook-<direction id>.json`)

Input: the validated direction JSON, the profile read JSON (`profile-bnd.jewelry`), and
brand basics (name, market, target, avg price band). Output, labels and short sentences only:

```
{"platform": "instagram",
 "weeks": [{"week": 1, "do": "<1-2 sentences of concrete work>", "post_idea": "<one post, tied to the brand's style and best post type>"}, ... exactly 4],
 "ad": {"audience": "<who and where, plain words>", "creative": "<which existing posts to put money behind and why>",
        "placements": ["instagram_feed","instagram_reels","instagram_stories"]} | null,
 "creator_brief": {"who": "<size and look of the creator>", "find_them": ["<hashtag or search term>", ... 3-5],
        "send": "<which piece>", "ask": "<what you ask for>", "dm": "<DM text, under 60 words>"} | null}
```

`ad` is null unless the direction has ads; `creator_brief` is null unless it has creators.
`ValidatePlaybook` enforces exactly 4 weeks, placements from the enum, `find_them` capped at
5, sentence trims (reuse `sentence`), and the no-number rule above. Month 2 reuses the same
playbook file by direction id (the direction's own `caution` already carries last month).

### Engine: `RunLaunch(plan, brand) LaunchSpec` (Go + `reference/engine/launch.py` + fixture)

- `start_date` = Monday after the last data week (`2026-11-02` on demo data); `end_date` =
  start + 27 days; `days` = 28.
- `ads_daily_usd` = round2(ads_usd / 28); `ads_daily_cents` = round(ads_usd / 28 × 100).
- `creator_count` = plan creators; `creator_ship_usd` = count × product cost;
  `creator_dms` = count × 2 (assume half reply; NOTES.md).
- `total_spend_usd` = ads_usd + creator_ship_usd.
- `calendar`: 4 weeks, each `{week, start_date, posts: {type: int}, creators, ads_usd,
  promo_discount}`. Posts per week = round(posts_per_week), split across types by largest
  remainder of `posts × mix` (ties by `PostTypes` order). Creators in week 1 only. Ads =
  round2(ads_usd / 4). Promo discount on the promo week.

### Server (`cmd/server/launch.go`)

- `POST /api/launch {plan}` → `{direction: {id, title}, spec: LaunchSpec, playbook,
  ad_request: {campaign, adset}, ads_connected: bool}`. The plan is the committed plan the
  browser holds from Verdict (`sessionStorage rehearsal.committed`).
- `POST /api/launch/ads {plan}` → with credentials: creates campaign then ad set on
  `https://graph.facebook.com/v21.0/act_<id>/...` (form-encoded, stdlib `net/http`), both
  `status=PAUSED`, returns `{sent: true, campaign_id, adset_id, url}`. Without credentials:
  `{sent: false, reason: "no_credentials", request}`. `META_GRAPH_URL` overrides the base for
  tests. Campaign: `objective=OUTCOME_SALES`, `special_ad_categories=[]`. Ad set:
  `daily_budget=<cents>`, `billing_event=IMPRESSIONS`, `optimization_goal=LINK_CLICKS`,
  `bid_strategy=LOWEST_COST_WITHOUT_CAP`, `targeting={geo_locations:{countries:[market]},
  age_min, age_max, genders}` (women→[2], men→[1], everyone→omitted; market `other`→`US`),
  `start_time`/`end_time` at 00:00Z.
- `POST /api/launch/calendar {plan}` → `text/calendar`, one all-day `VEVENT` per week
  ("Week N: post 3× (2 on hand, 1 close up) · ship 4 pieces") plus one on `start_date` for
  the ad campaign when ads > 0.
- `MUSE_API_KEY` stays unset in `main()`; the Meta env vars are read at request time.

### Web

- Nav gets `5 Launch`. Verdict "Commit" stores the plan and links to `/launch.html`.
- Directions cards show "Week 1: <do>" and "<ad audience>" or "<creator who>" from the
  playbook under the existing text, with a "Full playbook" toggle listing all four weeks.
- `launch.html`: a line "Nothing is sent live. The campaign is created paused; you switch
  it on in Ads Manager." then three cards: **Ads** (audience, creative, numbers, request
  preview, button "Create paused campaign in Meta Ads Manager", result line), **Creators**
  (who, search chips, send, ask, DM textarea + Copy, numbers), **Calendar** (table of 4
  weeks with post idea, button "Download calendar (.ics)").

### Optional if time remains: creator fit read

`prompts/creator.md`, `spark.CreatorFit(handle, brief, images)`, cache `creator-<handle>`,
`ValidateCreatorFit` → `{handle, fit: high|medium|low, size_band: small|mid|large, why,
send}`. `POST /api/launch/creator {handle, images}`. Demo needs screenshots the user drops
into `web/demo-creators/<handle>.png`; without them this part is skipped and documented.

## Left to the builder

Layout details inside the three cards (copy `.card`, `.btn`, `.btn.ghost` styles), exact
ICS wording, the order of keys in the request preview.
