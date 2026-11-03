# Success goal suggestions — plan (approved in chat)

**Goal:** Setup proposes "what counts as success" before the owner types anything: the engine computes candidate goals from the last 13 weeks; Spark (inside the existing profile read) picks the metric and the ambition and says why; the owner adjusts numbers and confirms. Rule: models return labels and sentences only; every number comes from `engine/`.

## Task 1 — engine `GoalSuggestions` (Go + Python reference + fixture)
- `engine/goals.go`: 
  ```go
  type GoalSuggestions struct {
      AvgWeeklyOrders    float64 `json:"avg_weekly_orders"`
      Hold               float64 `json:"hold"`     // round(avg_weekly_orders * 4)
      Grow               float64 `json:"grow"`     // round(hold * 1.10)
      Stretch            float64 `json:"stretch"`  // max sum of orders over any 4 consecutive weeks inside the window
      AvgMonthlySpendUsd float64 `json:"avg_monthly_spend_usd"` // round(sum(spend_usd)/weeks * 4)
      AvgPriceUsd        float64 `json:"avg_price_usd"`
      CurrentGoal        float64 `json:"current_goal"`
      CurrentBudget      float64 `json:"current_budget"`
      WeeksUsed          int     `json:"weeks_used"`
  }
  func RunGoalSuggestions(b *Brand, weeksBack int) GoalSuggestions
  ```
  Window = last `weeksBack` weeks (13 from the server, like the backtest). `avg_weekly_orders` rounded to 1 decimal with `round1`; integer roundings use `roundTo(x, 0)` (Python `round` parity). If fewer than 4 weeks, Stretch = sum of all.
- `reference/engine/goals.py`: `goal_suggestions(brand, weeks_back=13)` with the same keys and Python `round`. Add `dump("goals", goal_suggestions(brand))` to `reference/dump_fixtures.py`, run it, and add `TestGoalSuggestionsMatchReference` in `engine/engine_test.go` comparing every field with `near`.

## Task 2 — profile prompt carries the proposal
- `prompts/profile.md`: add to the JSON `"success_metric": "orders|revenue"`, `"ambition": "hold|grow|stretch"`, `"why_goal": "<one sentence, no numbers, saying what success means for this brand this month and why that ambition>"`, and one rule line: `Pick "orders" unless the brand clearly sells big-ticket items where revenue matters more than count; pick "hold" when the feed looks steady, "grow" when there is visible momentum, "stretch" only when the feed shows a launch or season.`
- `spark/validate.go` `ValidateProfile`: enums for the two new fields, 200-rune trim for `why_goal`; test cases added.
- The controller regenerates `spark/cache/profile-bnd.jewelry.json` live afterwards; until then add the three fields to the existing cache file by hand with values `orders`, `hold`, and a why sentence, so tests pass.

## Task 3 — server `GET /api/goals`
- Returns `engine.RunGoalSuggestions(s.brand, 13)` under the lock. Test: 200, `hold > 0`, `stretch >= hold`, `grow >= hold`, `weeks_used == 13`.

## Task 4 — Setup "Step 3 · What counts as success" (`web/index.html`)
- Replace the budget/goal card with: heading "Step 3 · What counts as success"; a proposal line (from the profile result when present): "Suggested: {ambition label} on {metric label} — {why_goal}" in muted text, else "Pick what a good month looks like."
- Three selectable cards (radio behaviour) fed by `GET /api/goals`: **Hold** "{hold} orders · about what you did lately ({avg_weekly_orders} a week)", **Grow** "{grow} orders · 10% more than Hold", **Stretch** "{stretch} orders · your best 4 weeks". A unit toggle "orders / revenue": in revenue mode the same three cards show `money(n × avg_price_usd)` with the orders in small text; the engine still tests orders.
- Below: the goal input (number, editable; picking a card writes its value in the current unit) and the budget input with caption "You spent about {money(avg_monthly_spend_usd)} a month lately." Preselect the card named by the profile `ambition` (after a feed read) else the one closest to `current_goal`. Preset the unit from `success_metric` when present.
- "Confirm" posts `/api/setup` with `goal_orders_per_month` (revenue mode: goal ÷ avg_price_usd, rounded) and `budget_usd_per_month`, then re-renders.
- No new numbers computed in JS beyond unit conversion and formatting.

## Task 5 — docs
- DEMO.md Setup beat: after the feed read, "Spark suggests Hold on orders; the three numbers come from the engine's last 13 weeks; Confirm."
- README Next: "Followers and inquiries as success metrics (the engine models orders only today)."
- NOTES.md: one line: success metric revenue is display-only; the engine tests orders.

## Verification
`gofmt -l . ; go vet ./... && go test -count=1 ./...` green; `curl localhost:PORT/api/goals` shows hold/grow/stretch; page script passes `node --check`.
