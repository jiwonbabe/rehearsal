# Rehearsal — Spark cache, API, UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish SPEC.md §7 steps 1–5 plus the KICKOFF.md judging additions: cached Spark calls with validation (three roles: planner, adversary, check line), the remaining API endpoints, the five web screens, the Review → Directions re-rank loop, README/DEMO, and a public deployment.

**Architecture:** `engine/` is done and golden-tested; do not change its formulas. `spark/` parses and validates model JSON into plain maps (no engine import). `cmd/server` is the only place that wires spark output into `engine.Direction` and holds in-memory month state (rates, reviewed flag). `web/` is five static HTML files plus one shared `app.js`/`app.css`; every number on screen comes from a JSON endpoint.

**Tech Stack:** Go 1.22 stdlib only (net/http, encoding/json, net/http/httptest for tests). Vanilla HTML/JS/CSS. Python 3 only for `data/synthetic.py` and `reference/dump_fixtures.py`.

**Spec:** `SPEC.md` (plus `CLAUDE.md` rules and `KICKOFF.md` task list). Read all three before any task.

## Global Constraints

- Go 1.22, stdlib only. `gofmt -l .` empty, `go vet ./...` clean, `go test ./...` green before every commit.
- `spark/` must not import `rehearsal/engine` or read `reference/`. `engine/` must not import `spark/`. Only `cmd/` packages import both.
- Models never produce order counts, revenue, or rates. Every number on screen comes from `engine/` via `/api/*`.
- Engine determinism: all sampling uses `rand.New(rand.NewSource(7))`; no `time.Now()` in `engine/`.
- Every page shows the badge text exactly: `Demo brand · synthetic data`.
- The strings `forecast` and `prediction` (any case) appear nowhere under `web/`, `prompts/`, `spark/cache/`, `cmd/`, `engine/`, `README.md`, `DEMO.md`. Use: stress test, survived, dies if, flip assumption. (The mockup copy "Not a forecast." must NOT be copied.)
- Horizon is 4 weeks. Mockup copy that says "12 weeks" or "week 6" is adapted to 4 weeks.
- Plan JSON has exactly the fields in SPEC §2. Direction JSON exactly SPEC §3. Shock JSON exactly SPEC §4.
- Never call Spark live during the demo; all responses come from `spark/cache/`.
- Commit after each task; one-line message saying what works now.
- Spark has exactly three roles: directions (call 1), adversary per direction (call 2), check line after the flip (call 3, `prompts/check.md`). The check line may only quote numbers that are in the flip JSON; anything else is replaced by the template line.
- Hackathon rule: only work committed between Oct 27 and Nov 4, 2026 counts. First commit is `git init` of the current state; README says "Built Nov 2–4, 2026 during the hackathon". No external code.
- When unsure, pick the simpler option and add one line to `NOTES.md`.

## Review Focus

1. `POST /api/stress` with an empty list or an id not in the current directions must return 400 with a message, never panic. (Test in Task 3.)
2. A direction whose adversary cache file is missing or unparsable must still stress-test, using `engine.FallbackCombos`. (Test in Task 3.)
3. A cached direction with a malformed range (missing, length 1, lo > hi, non-number) must be repaired to defaults so `engine.SampleVariants` never indexes out of range. (Test in Task 1.)
4. `POST /api/review/apply` called twice must leave rates at the once-blended values (0.49 for seed_post_rate), not blend again. (Test in Task 3.)
5. `POST /api/setup` with a budget lower than a direction's `ads_usd` range must clamp the range; a non-numeric budget must return 400. (Test in Task 3.)

---

### Task 1: Name-keyed Spark cache and `spark/validate.go`

**Files:**
- Modify: `spark/client.go` (`cachePath`, `Directions` signature)
- Create: `spark/validate.go`
- Test: `spark/validate_test.go`
- Create: `NOTES.md`, `.gitignore`

**Interfaces:**
- Consumes: nothing new.
- Produces:
  - `spark.Call(name, prompt string) (map[string]any, error)` — cache file is now `spark/cache/<name>.json` (no prompt hash).
  - `spark.Directions(month int, summary map[string]any, lastMonthBlock string) (map[string]any, error)` — cache name `directions-m1` / `directions-m2`.
  - `spark.Adversary(direction map[string]any, priors map[string]float64)` unchanged — cache name `adversary-<id>`.
  - `spark.ValidateDirections(raw map[string]any, budget, productCost float64) ([]map[string]any, error)` — returns the `directions` list with ranges repaired and clamped; drops entries without a string `id` in `{just_post, send_to_creators, pay_for_reach, one_week_sale, mix}`; error `"no valid directions"` if none remain.
  - `spark.ParseShocks(raw map[string]any) [][]string` — the `codes` lists from `shocks`, in order, no validation (engine.CleanCombos drops unknown codes).

- [ ] **Step 1: Init git and commit the current state**

```bash
cd /Users/jiwonbae/Projects/rehearsal
printf '.DS_Store\n__pycache__/\n*.pyc\n' > .gitignore
git init -q && git add -A && git commit -qm "engine matches reference fixtures; server serves brand/rates/backtest/summary"
```

- [ ] **Step 2: Write the failing tests in `spark/validate_test.go`**

```go
package spark

import "testing"

func rawDir(id string, ranges map[string]any) map[string]any {
	return map[string]any{"id": id, "title": id, "rank": 1.0, "ranges": ranges}
}

func TestValidateDirectionsClampsToBudget(t *testing.T) {
	raw := map[string]any{"directions": []any{
		rawDir("pay_for_reach", map[string]any{"ads_usd": []any{120.0, 900.0}, "creator_count": []any{0.0, 0.0}}),
		rawDir("send_to_creators", map[string]any{"ads_usd": []any{0.0, 60.0}, "creator_count": []any{4.0, 20.0}}),
	}}
	out, err := ValidateDirections(raw, 400, 45)
	if err != nil || len(out) != 2 {
		t.Fatalf("got %d dirs, err %v", len(out), err)
	}
	ads := out[0]["ranges"].(map[string]any)["ads_usd"].([]float64)
	if ads[0] != 120 || ads[1] != 400 {
		t.Errorf("ads_usd clamp: %v", ads)
	}
	creators := out[1]["ranges"].(map[string]any)["creator_count"].([]float64)
	if creators[0] != 4 || creators[1] != 8 { // floor((400-0)/45) = 8
		t.Errorf("creator_count clamp: %v", creators)
	}
}

func TestValidateDirectionsRepairsMalformedRanges(t *testing.T) {
	raw := map[string]any{"directions": []any{
		rawDir("just_post", map[string]any{"posts_per_week": []any{5.0}, "ads_usd": []any{60.0, 10.0}, "promo_discount": "x"}),
		map[string]any{"title": "no id"},
		rawDir("bogus_direction", map[string]any{}),
	}}
	out, err := ValidateDirections(raw, 400, 45)
	if err != nil || len(out) != 1 {
		t.Fatalf("got %d dirs, err %v", len(out), err)
	}
	r := out[0]["ranges"].(map[string]any)
	want := map[string][]float64{"posts_per_week": {3, 5}, "creator_count": {0, 0}, "ads_usd": {10, 60},
		"promo_discount": {0, 0}}
	for k, w := range want {
		got := r[k].([]float64)
		if got[0] != w[0] || got[1] != w[1] {
			t.Errorf("%s: got %v want %v", k, got, w)
		}
	}
	if pw := r["promo_week"].([]*int); len(pw) != 1 || pw[0] != nil {
		t.Errorf("promo_week default: %v", pw)
	}
	mix := r["post_mix"].(map[string][]float64)
	if mix["on_hand"][0] != 0.4 || mix["on_hand"][1] != 0.6 || mix["on_neck"][1] != 0.1 {
		t.Errorf("post_mix default: %v", mix)
	}
}

func TestValidateDirectionsEmpty(t *testing.T) {
	if _, err := ValidateDirections(map[string]any{"directions": []any{}}, 400, 45); err == nil {
		t.Error("want error for no directions")
	}
}

func TestParseShocks(t *testing.T) {
	raw := map[string]any{"shocks": []any{
		map[string]any{"codes": []any{"reach_drop", "bogus"}, "why": "x"},
		map[string]any{"why": "no codes"},
	}}
	got := ParseShocks(raw)
	if len(got) != 1 || len(got[0]) != 2 || got[0][0] != "reach_drop" {
		t.Errorf("got %v", got)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./spark/ -run 'TestValidate|TestParse' -v`
Expected: FAIL with `undefined: ValidateDirections`

- [ ] **Step 4: Implement `spark/validate.go`**

Defaults (used when a range is missing, not a 2-element numeric list, or non-numeric): `posts_per_week [3,5]`, `creator_count [0,0]`, `ads_usd [0,0]`, `promo_discount [0,0]`, `promo_week [nil]`, `post_mix {on_hand [0.4,0.6], close_up [0.2,0.4], flat_lay [0.1,0.3], on_neck [0,0.1]}`. A 2-element list with lo > hi is swapped, not defaulted. Each `post_mix` entry is validated the same way with its own default. `promo_week` keeps `nil` and integer entries 1–4, drops others, defaults to `[nil]` when empty. Clamps after repair: `ads_usd[1] = min(ads_usd[1], budget)`; `creator_count[1] = min(creator_count[1], floor((budget - ads_usd[0]) / productCost))` (floor at 0); then `lo = min(lo, hi)` for both. Output ranges use Go types `[]float64`, `[]*int`, `map[string][]float64` so the server can marshal them straight into `engine.Ranges`. Keep the direction's other keys untouched.

- [ ] **Step 5: Change the cache key and the Directions signature in `spark/client.go`**

`cachePath(name string) string` returns `filepath.Join(CacheDir, name+".json")` (drop sha256/hex imports). `Directions(month int, summary map[string]any, lastMonthBlock string)` calls `Call(fmt.Sprintf("directions-m%d", month), ...)`. Update the package comment: cache files are named by call, committed to git, and hand-editable.

- [ ] **Step 6: Run tests, vet, fmt**

Run: `gofmt -l . ; go vet ./... && go test ./...`
Expected: no gofmt output, all packages ok.

- [ ] **Step 7: Create `NOTES.md` and commit**

`NOTES.md` first lines:
```
# Notes (assumptions, one line each)
- Spark cache is keyed by call name (directions-m1, directions-m2, adversary-<id>), not by prompt hash: files are committed and hand-editable.
- Shock codes outside the catalog are dropped by engine.CleanCombos; spark.ParseShocks only extracts the lists.
```

```bash
git add -A && git commit -qm "spark: name-keyed cache, direction range validation with budget clamp"
```

---

### Task 2: Cache files for both months and every direction

**Files:**
- Modify: `engine/review.go` (add `LastMonthBlock`)
- Create: `cmd/sparkcache/main.go`
- Create: `spark/cache/directions-m1.json`, `spark/cache/directions-m2.json`, `spark/cache/adversary-{just_post,send_to_creators,pay_for_reach,one_week_sale}.json`
- Create: `prompts/check.md`; add `Check` to `spark/client.go` and `ValidateCheck` to `spark/validate.go`
- Test: `engine/engine_test.go` (one test added), `spark/cache_test.go`, `spark/validate_test.go` (one test added)

**Interfaces:**
- Consumes: `spark.Directions`, `spark.Adversary`, `spark.ValidateDirections`, `engine.Summary`, `engine.RunReview`, `engine.PriorProbabilities`.
- Produces:
  - `engine.OctoberPlan() Plan` — `{PlanID:"oct", Direction:"send_to_creators", PostsPerWeek:3, PostMix:{on_hand:.5, close_up:.3, flat_lay:.2, on_neck:0}, CreatorCount:4}` (the plan the brand actually ran; same as the review fixture).
  - `engine.LastMonthBlock(rv Review) string` — the `{{LAST_MONTH_BLOCK}}` text for the month-2 prompt.
  - `spark.Check(aID, bID string, flip map[string]any, aTitle, bTitle string) (map[string]any, error)` — call 3; prompt `prompts/check.md` filled with `{{FLIP_JSON}}`, `{{A_TITLE}}`, `{{B_TITLE}}`; cache name `check-<aID>-<bID>`.
  - `spark.ValidateCheck(raw map[string]any, allowed []float64) (string, bool)` — returns `raw["check"]` when it is a non-empty single-sentence string whose every number token (regex `\d+(\.\d+)?`) equals one of `allowed` (compared after rounding to 2 decimals); otherwise `("", false)`.
  - Cache files readable by `spark.Call`.

- [ ] **Step 1: Write the failing test for `LastMonthBlock` in `engine/engine_test.go`** (add `"strings"` to its imports)

```go
func TestLastMonthBlock(t *testing.T) {
	b := brand(t)
	rv := RunReview(b, OctoberPlan(), Calibrate(b))
	got := LastMonthBlock(rv)
	for _, want := range []string{"## Last month", "53 orders", "30.2", "40.8", "seed_post_rate 0.6 -> 0.49", rv.Diagnosis} {
		if !strings.Contains(got, want) {
			t.Errorf("block missing %q:\n%s", want, got)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./engine/ -run TestLastMonthBlock`
Expected: FAIL `undefined: OctoberPlan`

- [ ] **Step 3: Implement `OctoberPlan` and `LastMonthBlock` in `engine/review.go`**

Exact format (use `%g` for numbers):
```
## Last month (engine review; context only)

Actual: 53 orders, $360 spent, 1 of 4 seeded creators posted. Engine range was 30.2–40.8 orders.
Diagnosis: <rv.Diagnosis>
Rate changes: seed_post_rate 0.6 -> 0.49
```
With no rate changes the last line is `Rate changes: none`. Multiple changes are joined with `; `.

- [ ] **Step 4: Run engine tests**

Run: `go test ./engine/`
Expected: ok

- [ ] **Step 4b: Write `prompts/check.md`, `spark.Check`, `spark.ValidateCheck` and its test**

`prompts/check.md` (exact text):
```
You are the last voice inside Rehearsal before the owner commits money. The engine compared the two best plans and found the one assumption that flips the answer. Write ONE sentence telling the owner what to check this week before committing. Plain words, no jargon. Quote numbers only as they appear in the JSON below, as percentages; invent no other number. Return ONLY {"check": "<one sentence>"}.

Plan A: {{A_TITLE}}
Plan B: {{B_TITLE}}
Flip result from the engine (param, threshold, current, a_wins_now):
{{FLIP_JSON}}
```

Test in `spark/validate_test.go`:
```go
func TestValidateCheck(t *testing.T) {
	allowed := []float64{0.86, 0.6, 86, 60}
	if got, ok := ValidateCheck(map[string]any{"check": "Ask the creators who posted if they will again; at 86% seeding wins, you are at 60%."}, allowed); !ok || got == "" {
		t.Errorf("valid line rejected")
	}
	if _, ok := ValidateCheck(map[string]any{"check": "Expect about 40 orders if 86% post."}, allowed); ok {
		t.Errorf("line with a foreign number accepted")
	}
	if _, ok := ValidateCheck(map[string]any{"check": ""}, allowed); ok {
		t.Errorf("empty accepted")
	}
}
```
Run `go test ./spark/ -run TestValidateCheck` → FAIL undefined, implement, → PASS.

- [ ] **Step 5: Write `cmd/sparkcache/main.go`**

`package main`. Loads `data/brand.json`, calibrates, builds `summary := engine.Summary(brand, rates, engine.RunBacktest(brand, rates, 13))`. Then:
1. `spark.Directions(1, summary, "")`
2. `rv := engine.RunReview(brand, engine.OctoberPlan(), rates)`; `spark.Directions(2, engine.Summary(brand, rv.NewRates, backtest), engine.LastMonthBlock(rv))`
3. For every direction id present in either month's validated list: `spark.Adversary(dirMap, engine.PriorProbabilities(rates))`.
4. Flip check: `SampleVariants(60, 7)` + `Stress` with `engine.FallbackCombos` for `pay_for_reach` and `send_to_creators`, best plan of each via `RunVerdict`, `RunFlip(..., "seed_post_rate", 0.3, 1.0, 0.02)`, then `spark.Check("pay_for_reach", "send_to_creators", flipMap, aTitle, bTitle)` and `ValidateCheck` with `{threshold, current, threshold*100, current*100}`.
Each call result is validated (`ValidateDirections` with `brand.Brand.BudgetUsdPerMonth`, `ProductCostUsd`; `ParseShocks`) and printed as `ok <name>: N directions` / `N combos`. On error print the error and continue. Without `MUSE_API_KEY` and without cache files, it prints the errors from `spark.Call` and exits 1. Doc comment: `go run ./cmd/sparkcache` fills `spark/cache/`; run once, commit the files.

- [ ] **Step 6: Produce the cache files**

Run: `MUSE_API_KEY=... go run ./cmd/sparkcache` if a key is available (set `MUSE_BASE_URL`/`MUSE_MODEL` from dev.meta.ai docs first).

If there is no key (expected), write the six files by hand in the exact SPEC §3 / §4 shapes:

`directions-m1.json` — four directions, this order and rank:
| id | rank | caution | ranges (posts_per_week / creator_count / ads_usd / promo_week / promo_discount) |
|---|---|---|---|
| send_to_creators | 1 | null | [3,4] / [4,8] / [0,60] / [null] / [0,0] |
| pay_for_reach | 2 | null | [3,5] / [0,0] / [120,260] / [null] / [0,0] |
| just_post | 3 | "On your numbers this does not reach the order goal alone." | [4,5] / [0,0] / [0,0] / [null] / [0,0] |
| one_week_sale | 4 | "Your history has only two sale weeks, and the engine's sale-week estimate is its weakest." | [3,5] / [0,0] / [0,60] / [2,3] / [0.1,0.2] |

`post_mix` for all four: `{"on_hand":[0.4,0.6],"close_up":[0.2,0.4],"flat_lay":[0.1,0.3],"on_neck":[0,0.1]}`. `title`, `plain`, `why_for_you`, `you_would_do`: adapt the card copy from `mockup/Directions.dc.html` (short sentences, no jargon, `why_for_you` must name something in the summary: on-hand is the best post type; ad weeks landed inside range; last month's result).

`directions-m2.json` — same four, but `pay_for_reach` rank 1 and `send_to_creators` rank 2 with caution `"October post rate was 1 of 4. The engine now assumes 49% and attacks this direction with no-shows first."`; its `why_for_you` mentions October.

`check-pay_for_reach-send_to_creators.json` — `{"check": "<one sentence>"}`. Write it after Step 5 prints the flip numbers (`sparkcache` prints `flip threshold=X current=Y` before calling Check); the sentence tells the owner to message the October creators who did post and ask if they will post again, and quotes only those two numbers as percentages.

`adversary-<id>.json` — 20 combos each in the SPEC §4 shape, each with a one-sentence `why`. Start from `engine.FallbackCombos` and reorder so the direction's weak point comes first: `send_to_creators` → combos containing `creator_noshow` first; `pay_for_reach` → `cpm_spike` first; `one_week_sale` → `promo_underestimate` first; `just_post` → `reach_drop` first. Every file keeps at least 5 single-code combos and at least one combo with `viral_post`.

Add to `NOTES.md`: `- spark/cache/*.json were written by hand in the prompt output shape (no Muse API key at build time); cmd/sparkcache regenerates them when a key is set.` (Omit this line if the live call worked.)

- [ ] **Step 7: Write `spark/cache_test.go` and run it**

```go
func TestCacheFilesValidate(t *testing.T) {
	CacheDir, PromptDir = "cache", "../prompts"
	for month := 1; month <= 2; month++ {
		raw, err := Directions(month, map[string]any{}, "")
		if err != nil {
			t.Fatal(err)
		}
		dirs, err := ValidateDirections(raw, 400, 45)
		if err != nil || len(dirs) < 3 {
			t.Fatalf("month %d: %d dirs, %v", month, len(dirs), err)
		}
		for _, d := range dirs {
			raw, err := Adversary(d, map[string]float64{})
			if err != nil {
				t.Fatal(err)
			}
			if n := len(ParseShocks(raw)); n != 20 {
				t.Errorf("%s: %d combos", d["id"], n)
			}
		}
	}
}
```

Run: `go test ./spark/ -run TestCacheFilesValidate -v`
Expected: PASS (both months, four adversary files, 20 combos each).

- [ ] **Step 8: Commit**

```bash
gofmt -l . ; go vet ./... && go test ./... && git add -A && git commit -qm "spark cache for both months and four directions; cmd/sparkcache regenerates it"
```

---

### Task 3: Server endpoints `/api/directions`, `/api/stress`, `/api/review`, `/api/review/apply`, `/api/setup`

**Files:**
- Modify: `cmd/server/main.go` (split into `newServer`, `routes`, handlers; keep `main` small)
- Create: `cmd/server/handlers.go`
- Test: `cmd/server/main_test.go`

**Interfaces:**
- Consumes: Task 1 and 2 interfaces; `engine.SampleVariants(d, 60, 7)`, `engine.Stress`, `engine.RunVerdict`, `engine.RunFlip(a, b, engine.FallbackCombos, rates, brand, "seed_post_rate", 0.3, 1.0, 0.02)`, `engine.Catalog`, `engine.PriorProbabilities`.
- Produces (all JSON):
  - `GET /api/directions` → `{"month": 1|2, "directions": []engine.Direction (sorted by rank asc), "last_month": engine.Review|null}`.
  - `POST /api/setup` `{"budget_usd_per_month": n, "goal_orders_per_month": n, "target": {gender, age, price_band}, "market": s}` → updates `s.brand.Brand` in memory, returns the new `BrandInfo`. Decode into a struct with pointer fields (`*float64`, `*map[string]string`, `*string`) so missing fields keep their value; a decode error (non-numeric budget/goal) → 400.
  - `POST /api/stress` `{"direction_ids": [..]}` → `stressResponse`:
    ```go
    type shockCombo struct { Codes []string `json:"codes"`; Why string `json:"why"` }
    type directionStress struct {
        ID string `json:"id"`; Title string `json:"title"`
        Plans []engine.Plan `json:"plans"`; Shocks []shockCombo `json:"shocks"`
        Stress engine.Stressed `json:"stress"`
    }
    type stressResponse struct {
        Directions []directionStress `json:"directions"`
        Verdict []engine.Verdict `json:"verdict"`
        Flip *engine.Flip `json:"flip"`
        Catalog map[string]string `json:"catalog"`   // code -> engine.Catalog[code].Label
        Priors map[string]float64 `json:"priors"`
        Check string `json:"check"`                   // Spark call 3, or the template line
    }
    ```
    Each direction is stressed with its own 20 combos. `Verdict` = `engine.RunVerdict(allPlans, merged, rates, brand)` where `merged.Matrix` is the union of every direction's matrix and `merged.Combos` is the first direction's combos. `Flip` is non-nil only when two or more directions were picked: A = first non-baseline verdict plan, B = second non-baseline verdict plan. `shocks` carries the adversary's `why` for each combo in matrix column order (empty string when the engine added fallback combos). `Check`: when `Flip` is non-nil, `spark.Check(a.Direction, b.Direction, flipMap, aTitle, bTitle)` → `spark.ValidateCheck(raw, {threshold, current, threshold*100, current*100})`; on error or rejection use `checkLine(flip)` = `"Check before committing: message the creators who did post last month and ask if they would do it again. Creators posting at " + pct(threshold) + " or better flips the answer; you are at " + pct(current) + "."` With `Flip` nil, `Check` is `""`. 400 on empty list or unknown id.
  - `POST /api/review` body `{"plan": engine.Plan}` optional; an empty body (`io.EOF` from the decoder) or an empty `plan_id` uses `engine.OctoberPlan()` → `{"committed": engine.Plan, "review": engine.Review}` computed from `s.baseRates` (the calibrated rates, never the blended ones).
  - `POST /api/review/apply` → runs the same review on `s.baseRates`, sets `s.rates = review.NewRates`, `s.month = 2`, `s.lastReview = &review`; returns `{"month": 2, "rates": engine.Rates}`. Idempotent.
- Server state: `type server struct { brand *engine.Brand; baseRates, rates engine.Rates; month int; lastReview *engine.Review; mu sync.Mutex }`. `/api/rates`, `/api/backtest`, `/api/summary` keep working and use `s.rates`.

- [ ] **Step 1: Write the failing tests in `cmd/server/main_test.go`**

```go
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s, err := newServer("data/brand.json")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.routes())
	t.Cleanup(ts.Close)
	return ts
}

func call(t *testing.T, ts *httptest.Server, method, path, body string, into any) int {
	t.Helper()
	req, _ := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if into != nil && resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
			t.Fatal(err)
		}
	}
	return resp.StatusCode
}

func TestDirectionsMonthOne(t *testing.T) {
	ts := newTestServer(t)
	var out struct {
		Month      int                `json:"month"`
		Directions []map[string]any   `json:"directions"`
		LastMonth  *json.RawMessage   `json:"last_month"`
	}
	if code := call(t, ts, "GET", "/api/directions", "", &out); code != 200 {
		t.Fatalf("status %d", code)
	}
	if out.Month != 1 || len(out.Directions) < 3 || out.LastMonth != nil {
		t.Errorf("month %d, %d dirs, last_month %v", out.Month, len(out.Directions), out.LastMonth)
	}
	for i := 1; i < len(out.Directions); i++ {
		if out.Directions[i]["rank"].(float64) < out.Directions[i-1]["rank"].(float64) {
			t.Errorf("not sorted by rank")
		}
	}
}

func TestStressTwoDirections(t *testing.T) {
	ts := newTestServer(t)
	var out stressResponse
	body := `{"direction_ids":["pay_for_reach","send_to_creators"]}`
	if code := call(t, ts, "POST", "/api/stress", body, &out); code != 200 {
		t.Fatalf("status %d", code)
	}
	if len(out.Directions) != 2 || len(out.Directions[0].Plans) != 60 || len(out.Directions[0].Stress.Combos) != 20 {
		t.Fatalf("shape: %d dirs", len(out.Directions))
	}
	hasBaseline := false
	for _, v := range out.Verdict {
		hasBaseline = hasBaseline || v.Plan.PlanID == "baseline"
	}
	if len(out.Verdict) != 3 || !hasBaseline {
		t.Errorf("verdict should be 2 directions + baseline, got %d", len(out.Verdict))
	}
	if out.Flip == nil || out.Flip.Param != "seed_post_rate" {
		t.Errorf("flip: %+v", out.Flip)
	}
	if len(out.Catalog) != 7 || len(out.Priors) != 7 {
		t.Errorf("catalog/priors missing")
	}
	if out.Check == "" || !strings.Contains(out.Check, "%") {
		t.Errorf("check line: %q", out.Check)
	}
}

func TestStressRejectsBadIds(t *testing.T) {
	ts := newTestServer(t)
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":[]}`, nil); code != 400 {
		t.Errorf("empty: %d", code)
	}
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":["nope"]}`, nil); code != 400 {
		t.Errorf("unknown: %d", code)
	}
}

func TestStressFallsBackWithoutAdversaryCache(t *testing.T) {
	ts := newTestServer(t)
	t.Setenv("MUSE_API_KEY", "")
	os.Rename("spark/cache/adversary-just_post.json", "spark/cache/adversary-just_post.json.bak")
	defer os.Rename("spark/cache/adversary-just_post.json.bak", "spark/cache/adversary-just_post.json")
	var out stressResponse
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":["just_post"]}`, &out); code != 200 {
		t.Fatalf("status %d", code)
	}
	if len(out.Directions[0].Stress.Combos) != 20 || out.Flip != nil {
		t.Errorf("fallback combos %d, flip %v", len(out.Directions[0].Stress.Combos), out.Flip)
	}
}

func TestReviewApplyIsIdempotent(t *testing.T) {
	ts := newTestServer(t)
	var rv struct {
		Review struct {
			InsideRange bool `json:"inside_range"`
			NewRates    struct{ SeedPostRate float64 `json:"seed_post_rate"` } `json:"new_rates"`
		} `json:"review"`
	}
	if code := call(t, ts, "POST", "/api/review", "", &rv); code != 200 || rv.Review.InsideRange || rv.Review.NewRates.SeedPostRate != 0.49 {
		t.Fatalf("review: %d %+v", code, rv)
	}
	var applied struct {
		Month int `json:"month"`
		Rates struct{ SeedPostRate float64 `json:"seed_post_rate"` } `json:"rates"`
	}
	call(t, ts, "POST", "/api/review/apply", "", &applied)
	call(t, ts, "POST", "/api/review/apply", "", &applied)
	if applied.Month != 2 || applied.Rates.SeedPostRate != 0.49 {
		t.Errorf("after two applies: %+v", applied)
	}
	var dirs struct {
		Month     int             `json:"month"`
		LastMonth json.RawMessage `json:"last_month"`
	}
	call(t, ts, "GET", "/api/directions", "", &dirs)
	if dirs.Month != 2 || len(dirs.LastMonth) < 10 {
		t.Errorf("directions after apply: month %d", dirs.Month)
	}
}

func TestSetupClampsAndValidates(t *testing.T) {
	ts := newTestServer(t)
	if code := call(t, ts, "POST", "/api/setup", `{"budget_usd_per_month":"abc"}`, nil); code != 400 {
		t.Errorf("non-numeric budget: %d", code)
	}
	if code := call(t, ts, "POST", "/api/setup", `{"budget_usd_per_month":100}`, nil); code != 200 {
		t.Errorf("setup: %d", code)
	}
	var out struct{ Directions []struct{ Ranges struct{ AdsUsd []float64 `json:"ads_usd"` } `json:"ranges"` } `json:"directions"` }
	call(t, ts, "GET", "/api/directions", "", &out)
	for _, d := range out.Directions {
		if d.Ranges.AdsUsd[1] > 100 {
			t.Errorf("ads_usd not clamped to new budget: %v", d.Ranges.AdsUsd)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/server/ 2>&1 | head`
Expected: build FAIL `undefined: newServer`, `stressResponse`

- [ ] **Step 3: Refactor `cmd/server/main.go` and implement `cmd/server/handlers.go`**

`newServer(brandPath string) (*server, error)` loads the brand, sets `baseRates = rates = engine.Calibrate(brand)`, `month = 1`. `(s *server) routes() *http.ServeMux` registers every endpoint and `http.FileServer(http.Dir("web"))` at `/`. `main` keeps only the PORT handling and `ListenAndServe`.

Handlers in `handlers.go`:
- `currentDirections() ([]engine.Direction, error)`: `spark.Directions(s.month, summary, block)` where `block` is `""` in month 1 and `engine.LastMonthBlock(*s.lastReview)` in month 2 → `spark.ValidateDirections(raw, budget, productCost)` → `json.Marshal` then `json.Unmarshal` into `[]engine.Direction` → `sort.SliceStable` by `Rank`. On spark error respond 500 with the error text (the cache file is missing; say so).
- `combosFor(d engine.Direction) ([][]string, []string)`: `spark.Adversary(dirMap, engine.PriorProbabilities(s.rates))`; on error return `engine.FallbackCombos` with empty whys. Build `dirMap` from `d` via JSON round-trip.
- Stress handler: per picked id → `plans := engine.SampleVariants(d, 60, 7)`, `st := engine.Stress(plans, combos, s.rates, s.brand)`, align `shocks[i].Why` with `st.Combos[i]` by matching codes (CleanCombos may have dropped or reordered combos). Merge, verdict, flip as described in Interfaces.
- Review handlers as described. Lock `s.mu` in every handler that reads or writes server state.
- `writeError(w, code, msg)` writes `{"error": msg}` with the status code.

- [ ] **Step 4: Run tests, vet, fmt**

Run: `gofmt -l . ; go vet ./... && go test ./...`
Expected: all ok.

- [ ] **Step 5: Smoke-run the server**

Run: `go run ./cmd/server & sleep 1; curl -s -X POST localhost:8080/api/stress -d '{"direction_ids":["pay_for_reach","send_to_creators"]}' | head -c 400; kill %1`
Expected: JSON starting with `{"directions":[{"id":"pay_for_reach"`. Then `curl ... | python3 -c 'import json,sys; d=json.load(sys.stdin); print(d["flip"], d["check"])'` — if `check` is the template line, the hand-written `check-pay_for_reach-send_to_creators.json` quotes numbers that differ from this flip: fix the file so `ValidateCheck` accepts it, and re-run.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -qm "server: directions, stress, review, review/apply, setup endpoints with tests"
```

---

### Task 4: Web shell + screens 0–2 (Setup, Directions, Simulate)

**Files:**
- Create: `web/app.css`, `web/app.js`
- Replace: `web/index.html` (Setup)
- Create: `web/directions.html`, `web/simulate.html`
- Test: `cmd/server/web_test.go`

**Interfaces:**
- Consumes: `/api/brand`, `/api/backtest`, `/api/setup`, `/api/directions`, `/api/stress`.
- Produces (in `web/app.js`, all global functions):
  - `renderHeader(step int, subtitle string)` — inserts the mockup header at the top of `body`: logo + "Rehearsal" + subtitle, the five nav pills (`0 Setup` → `/`, `1 Directions` → `/directions.html`, `2 Simulate` → `/simulate.html`, `3 Verdict` → `/verdict.html`, `4 Review` → `/review.html`; active pill dark), and a badge pill with the exact text `Demo brand · synthetic data`.
  - `api(path, body)` — `fetch` JSON helper; `body` present → POST; throws `Error(json.error || status)` on non-2xx.
  - `getSelected()` / `setSelected(ids)` — `sessionStorage` key `rehearsal.selected`, JSON array of direction ids.
  - `money(n)` → `$` + rounded integer with thousands separator; `pct(x)` → `Math.round(x*100) + '%'`.
  - `shockText(codes, catalog)` → labels joined with ` + `; `[]` → `never`.
  - `planText(plan)` → one sentence: `post {posts_per_week}×/wk ({mix with on_hand first}), {creator_count} creators, {money(ads_usd)} ads[, {pct(promo_discount)} off in week {promo_week}]`.
- Styles: `web/app.css` carries the mockup's palette and the repeated card/pill/button rules (`body` font IBM Plex Sans from Google Fonts, background `#f3f4f1`, text `#17181c`, muted `#5c6066`, border `#d9dbd4`, accent `#2f5d8a`, cells `#2f5d8a` survived / `#c2410c` died). Page-specific layout stays inline, copied from the mockup.

- [ ] **Step 1: Write the failing tests in `cmd/server/web_test.go`**

```go
package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var banned = regexp.MustCompile(`(?i)forecast|prediction`)

func TestNoBannedWords(t *testing.T) {
	for _, glob := range []string{"web/*", "prompts/*", "spark/cache/*", "cmd/*/*.go", "engine/*.go", "README.md", "DEMO.md"} {
		files, _ := filepath.Glob(glob)
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue // this file carries the regex itself
			}
			raw, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			if loc := banned.FindIndex(raw); loc != nil {
				t.Errorf("%s contains banned word near byte %d", f, loc[0])
			}
		}
	}
}

func TestEveryPageHasBadgeAndShell(t *testing.T) {
	pages := []string{"index.html", "directions.html", "simulate.html", "verdict.html", "review.html"}
	js, _ := os.ReadFile("web/app.js")
	if !strings.Contains(string(js), "Demo brand · synthetic data") {
		t.Fatal("app.js must carry the badge")
	}
	for _, p := range pages {
		raw, err := os.ReadFile(filepath.Join("web", p))
		if err != nil {
			t.Errorf("%s missing", p)
			continue
		}
		if !strings.Contains(string(raw), "app.js") || !strings.Contains(string(raw), "renderHeader(") {
			t.Errorf("%s does not load the shell", p)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/server/ -run 'TestNoBanned|TestEveryPage' -v`
Expected: FAIL — `verdict.html missing`, `review.html missing` (and `index.html does not load the shell`). `TestNoBannedWords` passes now and guards every later task.

- [ ] **Step 3: Write `web/app.css` and `web/app.js`**

Implement the Interfaces above. Header markup: copy the first `<div style="display: flex; align-items: center; justify-content: space-between ...">` block of `mockup/Main.dc.html`, add the badge pill after the subtitle.

- [ ] **Step 4: Write `web/index.html` (Setup, layout from `mockup/Main.dc.html`)**

Sections, top to bottom:
1. Title "Tell us who you sell to. We'll do the marketing homework." Drop the Instagram-connect card (cut scope); replace with a card "Demo brand" showing `brand.name` and "52 weeks of synthetic data".
2. "Who you sell to" card: four `<select>`s (Market: US/KR; Gender: women/men/everyone; Age: 18-24/25-34/35-44; Price band: 30-80/80-200/200+) pre-filled from `/api/brand` (`market`, `target.gender`, `target.age`, `target.price_band`).
3. "Monthly budget and goal" card: two `<input type="number">` for budget and goal, pre-filled. A "Save" button POSTs `/api/setup` with all six values and re-renders.
4. "Engine trust check" card (backtest, from `/api/backtest`): the two numbers `Median weekly error {median_abs_pct_error}%` and `Weeks inside engine range {weeks_inside_range} / {weeks_total}`, then a table with columns Week · Actual · Engine · Range (`low–high`) · a dot (inside: accent, outside: `#c2410c`) and "sale week" when `promo`. Below: "Known weakness: the engine's sale-week estimate is its weakest" shown only if any `promo` row is outside range.
5. Button "See this month's directions" → `/directions.html`.

- [ ] **Step 5: Write `web/directions.html` (layout from `mockup/Directions.dc.html`)**

Load `/api/directions`. When `month === 2`, render the "What October taught us" block first: `last_month.diagnosis`, then `Orders {actual.orders} vs {range[0]}–{range[1]}`, then each rate change as `{rate} {old} → {new}`. Title "November: pick one or two directions" (month 1: "This month: pick one or two directions"). One card per direction in rank order: rank label (`Suggested first` for rank 1, `Also fits` otherwise), `title`, `plain`, a "Why for you" line, `you_would_do`, a range line built from `ranges` (`{money(ads_usd[0])}–{money(ads_usd[1])} ads`, `{creator_count[0]}–{creator_count[1]} creators`, `{posts_per_week[0]}–{posts_per_week[1]} posts/wk`), and a caution box when `caution` is non-null. Each card has a toggle button "Select" / "Selected" (max 2 selected; a third click is ignored). Footer line: `{n} direction(s) selected · the engine will build {n×60} variants and attack each with 20 bad months.` Button "Test these" (disabled until ≥1 selected) saves via `setSelected` and goes to `/simulate.html`.

- [ ] **Step 6: Write `web/simulate.html` (layout from `mockup/Stress.dc.html`)**

On load: `ids = getSelected()`; if empty, show "Pick a direction first" with a link back. POST `/api/stress`. Three columns as in the mockup:
1. Left "Planner · Muse Spark": `<pre>` with the JSON of the plan under the pointer (default: first plan). Caption: "Proposes parameters only. Never a number for orders or revenue."
2. Middle: one matrix block per direction: heading `{title} · 60 variants × 20 bad months`, legend (survived / died), then a grid of 60 rows × 20 cells (`div` cells 14×14px, colored by `survived`), each cell `title="{planText} · {shockText} · {orders} orders, {money(spend)}"`. Hovering a row updates the left `<pre>`. Row label = plan id suffix. Caption: `{plans×combos} runs · deterministic · same input, same output`.
3. Right "Adversary · Muse Spark": per direction, the 20 combos as `{shockText(codes)} · p {priors of first code}` with the `why` beneath (skip empty `why`).
Button "See the verdict" → `/verdict.html`.

- [ ] **Step 7: Run tests and look at the pages**

Run: `go test ./cmd/server/ -run 'TestNoBanned|TestEveryPage' -v` — still FAIL only for `verdict.html` / `review.html` missing (fixed in Task 5).
Run: `go run ./cmd/server` and open `http://localhost:8080/`, `/directions.html`, `/simulate.html`. Expected: badge on every page, backtest table with 13 rows, four direction cards, two matrix blocks after selecting two directions, no console errors.

- [ ] **Step 8: Commit**

```bash
gofmt -l . ; go vet ./... && git add -A && git commit -qm "web: shell, setup with backtest card, directions cards, simulate matrix"
```

---

### Task 5: Screens 3–4 (Verdict, Review) and the re-rank loop

**Files:**
- Create: `web/verdict.html`, `web/review.html`
- Modify: `NOTES.md`, maybe `data/synthetic.py` (only if the loop check fails)

**Interfaces:**
- Consumes: `/api/stress`, `/api/review`, `/api/review/apply`, `/api/directions`, and the `app.js` helpers from Task 4.
- Produces: the end-to-end demo loop.

- [ ] **Step 1: Write `web/verdict.html` (layout from `mockup/Survivors.dc.html`)**

POST `/api/stress` with `getSelected()` (deterministic, so it matches Simulate). Title: `Verdict: {title of verdict[0] direction} holds this month.` Subtitle: "The best variant from each direction you picked, plus the no-spend baseline. Survived = hit your order goal in that many of 20 bad months without running out of budget early." One card per `verdict` entry, in order: `{direction title or "Just post · baseline"} · survived {survived} / {total}`, `planText(plan)`, `Orders, bad month {orders_bad_month[0]}–{orders_bad_month[1]}`, `Spend {money(spend_usd)}`, `Dies if {shockText(dies_if, catalog)}`, button "Commit to this plan".

"The assumption that flips the answer" card: with `flip` non-null and `a_wins_now`: `If creators post at {pct(threshold)} or better, {B title} beats {A title}. At your current {pct(current)}, {A title} wins.` (swap A/B when `a_wins_now` is false). Below it, in its own box labelled "What to check this week · Muse Spark": the response's `check` string, verbatim. With `flip` null: "No creator post rate between 30% and 100% changes this ranking." and no check box.

On "Commit": render "Your month, generated on commit" under the cards from the plan fields: Week 1: `Post {posts_per_week}× ({mix})` + `Pick your 2 best on-hand posts and set a daily cap of {money(ads_usd/28)}` when `ads_usd > 0` + `DM and ship {creator_count} pieces` when `creator_count > 0`; Weeks 2–4: `Post {posts_per_week}×`, plus `Announce {pct(promo_discount)} off` on week `promo_week`; Week 4 ends with "review opens next month". Link "Review October" → `/review.html`.

- [ ] **Step 2: Write `web/review.html` (layout from `mockup/Review.dc.html`)**

POST `/api/review` with no body. Title: `October: {inside_range ? "inside range" : "below range"}. Here is what was off.` Line: `You ran {planText(committed)}.` Table "Expected range vs what happened": `Orders {actual.orders} vs {range[0]}–{range[1]}` (mark outside in `#c2410c`), `Creator posts {actual.creators_posted} of {actual.creators_seeded}`, `Spend {money(actual.spend_usd)}`. "Diagnosis" box: `review.diagnosis`. "What the engine changed": one row per `rate_changes` entry `{rate} {old} → {new}`; if none, "No rate changed." Footer: "Small steps. One month is evidence, not a verdict. Next month's directions are re-ranked with this." Button "Plan November" → POST `/api/review/apply`, then `sessionStorage.removeItem('rehearsal.selected')`, then go to `/directions.html`.

- [ ] **Step 3: Run the full test suite**

Run: `gofmt -l . ; go vet ./... && go test ./...`
Expected: all ok, including `TestEveryPageHasBadgeAndShell` and `TestNoBannedWords`.

- [ ] **Step 4: Verify the loop shows a visible ranking change**

Run the server, then in a second shell:
```bash
curl -s localhost:8080/api/directions | python3 -c 'import json,sys; print([(d["id"],d["rank"],bool(d["caution"])) for d in json.load(sys.stdin)["directions"]])'
curl -s -X POST localhost:8080/api/stress -d '{"direction_ids":["pay_for_reach","send_to_creators"]}' | python3 -c 'import json,sys; print([(v["plan"]["direction"],v["survived"]) for v in json.load(sys.stdin)["verdict"]])'
curl -s -X POST localhost:8080/api/review/apply > /dev/null
curl -s localhost:8080/api/directions | python3 -c 'import json,sys; print([(d["id"],d["rank"],bool(d["caution"])) for d in json.load(sys.stdin)["directions"]])'
curl -s -X POST localhost:8080/api/stress -d '{"direction_ids":["pay_for_reach","send_to_creators"]}' | python3 -c 'import json,sys; print([(v["plan"]["direction"],v["survived"]) for v in json.load(sys.stdin)["verdict"]])'
```
Expected: before apply `send_to_creators` rank 1 without caution; after apply `pay_for_reach` rank 1 and `send_to_creators` has a caution; after apply `send_to_creators` survivals are strictly lower than before apply and lower than `pay_for_reach`.

If the survival count does not drop: lower the October `creators_posted` or raise the seeded count in `data/synthetic.py` (keep `seed=7`), then `python3 data/synthetic.py && python3 reference/dump_fixtures.py && go test ./...`, and add a line to `NOTES.md` saying which constant changed and why.

- [ ] **Step 5: Click through in the browser**

Setup → Directions (select two) → Simulate → Verdict (commit) → Review (Plan November) → Directions. Expected: the "What October taught us" block appears, ads card is first, seeding card shows the caution. No banned words, badge on every page.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -qm "web: verdict with flip assumption and commit plan, review with re-rank loop"
```

---

### Task 6: README, DEMO.md, deployment

**Files:**
- Modify: `README.md`
- Create: `DEMO.md`, `Dockerfile`, `fly.toml`
- Modify: `NOTES.md`

- [ ] **Step 1: Write `README.md`**

First line under the title: `Built Nov 2–4, 2026 during the Meta Global AI Developer Hackathon. No external code; Go stdlib only.` Keep the existing run commands block. Add one paragraph covering, in this order: what Rehearsal does (one sentence from SPEC); the **agent loop**, written with those words: planner (Muse Spark) → engine → adversary (Muse Spark) → engine → verdict + check line (Muse Spark) → human commits; "models at both ends, engine in the middle, a person at the end"; oracle isolation (`spark/` cannot import `engine/`; only `cmd/` wires them); two implementations produce the same numbers (`reference/engine` Python golden fixtures, `go test ./...`); the backtest numbers from `/api/backtest` (median error and weeks inside range, read from the running server, not typed from memory); the demo brand is synthetic (`data/synthetic.py`, seed 7). A "Model access" block: default is the Meta Model API (`MUSE_API_KEY`, `MUSE_BASE_URL`, `MUSE_MODEL`); fallback is Baseten's OpenAI-compatible endpoint with `MUSE_MODEL=meta-models/muse-spark-open-1026` and `MUSE_BASE_URL` set to Baseten's base URL, same prompts, only env vars change; demo runs from `spark/cache/` with no live calls. A "Live" line with the public URL from Step 3. Then a "Next" list copied from SPEC §8.

- [ ] **Step 2: Write `DEMO.md`**

A recording script under 3:00. First 20 seconds, before any screen: the real problem in one line (a sister running a small jewelry brand spends a month's budget on a guess) and who benefits (owners with no marketer). Then one line per screen with what to say and what to click: Setup (badge, backtest card numbers) → Directions (seeding ranked first; select seeding + ads) → Simulate (matrix, planner JSON, adversary list) → Verdict (ads survived N/20, seeding M/20, flip assumption, commit) → Review (October 1 of 4 posted, seed_post_rate 0.6 → 0.49, Plan November) → Directions (ads first, seeding caution). State at the top: no live API calls, everything is served from `spark/cache/`; restart the server to reset to month 1.

- [ ] **Step 3: Dockerfile, fly.toml, deploy**

`Dockerfile` (multi-stage): `golang:1.22` builds `CGO_ENABLED=0 go build -o /rehearsal ./cmd/server`; final stage `gcr.io/distroless/static`, `WORKDIR /app`, copies the binary plus `data/`, `web/`, `prompts/`, `spark/cache/`, `CMD ["/rehearsal"]`. The server already honours `PORT`. `fly.toml`: app name `rehearsal`, `[http_service] internal_port = 8080, force_https = true, auto_stop_machines = true, min_machines_running = 0`.

Run: `docker build -t rehearsal . && docker run --rm -p 8080:8080 rehearsal` (skip if Docker is not installed; say so in NOTES.md) and curl `/api/directions`. Then `fly auth whoami`; if logged in, `fly launch --copy-config --no-deploy && fly deploy`, curl `https://rehearsal.fly.dev/api/backtest`, and put the URL in README. If not logged in, write the two commands at the top of `DEMO.md` under "Deploy" and add a NOTES.md line; do not block on it.

- [ ] **Step 4: Run everything once more and commit**

Run: `gofmt -l . ; go vet ./... && go test ./...`
Expected: all ok (`TestNoBannedWords` now also scans README.md and DEMO.md).

```bash
git add -A && git commit -qm "README agent loop and model access, DEMO script, Dockerfile and fly.toml"
```
