package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"rehearsal/engine"
	"rehearsal/spark"
)

type shockCombo struct {
	Codes []string `json:"codes"`
	Why   string   `json:"why"`
}

type directionStress struct {
	ID     string          `json:"id"`
	Title  string          `json:"title"`
	Plans  []engine.Plan   `json:"plans"`
	Shocks []shockCombo    `json:"shocks"`
	Stress engine.Stressed `json:"stress"`
}

type stressResponse struct {
	Directions []directionStress  `json:"directions"`
	Verdict    []engine.Verdict   `json:"verdict"`
	Flip       *engine.Flip       `json:"flip"`
	FlipA      string             `json:"flip_a"`
	FlipB      string             `json:"flip_b"`
	Catalog    map[string]string  `json:"catalog"`
	Priors     map[string]float64 `json:"priors"`
	Check      string             `json:"check"`
	Gaps       []engine.Gap       `json:"gaps"`
	Goal       float64            `json:"goal"`
	Budget     float64            `json:"budget"`
	AvgPrice   float64            `json:"avg_price_usd"`
	Month      int                `json:"month"`
	MonthName  string             `json:"month_name"`
	SeasonGoal int                `json:"season_goal"`
	Demand     float64            `json:"demand"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// via round-trips src through JSON into dst.
func via(src, dst any) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}

func (s *server) currentDirections() ([]engine.Direction, error) {
	block := ""
	if s.month == 2 && s.lastReview != nil {
		block = engine.LastMonthBlock(*s.lastReview)
	}
	summary := engine.Summary(s.brand, s.rates, engine.RunBacktest(s.brand, s.rates, 13))
	raw, err := spark.Directions(s.month, summary, block, engine.SeasonBlock(s.brand.Season, s.planMonth()))
	if err != nil {
		return nil, fmt.Errorf("directions cache missing or unreadable: %w", err)
	}
	s.seasonNote = spark.SeasonNote(raw, s.seasonNumbers())
	valid, err := spark.ValidateDirections(raw, s.brand.Brand.BudgetUsdPerMonth, s.brand.Brand.ProductCostUsd)
	if err != nil {
		return nil, err
	}
	var dirs []engine.Direction
	if err := via(valid, &dirs); err != nil {
		return nil, err
	}
	sort.SliceStable(dirs, func(i, j int) bool { return dirs[i].Rank < dirs[j].Rank })
	return dirs, nil
}

// planMonth: the calendar month being planned (November for demo month 1, December for month 2).
func (s *server) planMonth() int { return engine.PlanMonth(s.brand, s.month-1) }

// seasonNumbers: the two indexes the planner's season_note may quote.
func (s *server) seasonNumbers() []float64 {
	se := engine.SeasonFor(s.brand.Season, s.planMonth())
	return []float64{se.Demand, se.AdCost}
}

// seasonOut: what the pages show for the month being planned. Indexes from the table, note from Spark.
func (s *server) seasonOut() map[string]any {
	se := engine.SeasonFor(s.brand.Season, s.planMonth())
	out := map[string]any{"month": se.Month, "name": se.Name, "demand": se.Demand, "ad_cost": se.AdCost, "note": s.seasonNote, "applied": s.brand.Season != nil,
		"goal": s.brand.Brand.GoalOrdersPerMonth, "season_goal": engine.SeasonGoal(s.brand.Brand.GoalOrdersPerMonth, se.Demand)}
	if s.brand.Season != nil {
		out["sources"], out["table_note"], out["category"], out["market"] = s.brand.Season.Sources, s.brand.Season.Note, s.brand.Season.Category, s.brand.Season.Market
	}
	return out
}

func (s *server) handleDirections(w http.ResponseWriter, r *http.Request) {
	dirs, err := s.currentDirections()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	type directionOut struct {
		engine.Direction
		Playbook map[string]any `json:"playbook"`
	}
	var out []directionOut
	for _, d := range dirs {
		out = append(out, directionOut{d, playbookFor(d.ID)})
	}
	resp := struct {
		Month      int              `json:"month"`
		Directions []directionOut   `json:"directions"`
		LastMonth  *engine.Review   `json:"last_month"`
		Channels   []map[string]any `json:"channels"`
		Season     map[string]any   `json:"season"`
	}{s.month, out, s.lastReview, s.channels(), s.seasonOut()}
	s.writeJSON(w, resp)
}

func (s *server) handleSetup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Budget *float64           `json:"budget_usd_per_month"`
		Goal   *float64           `json:"goal_orders_per_month"`
		Target *map[string]string `json:"target"`
		Market *string            `json:"market"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if (in.Budget != nil && *in.Budget < 0) || (in.Goal != nil && *in.Goal < 0) {
		writeError(w, 400, "budget and goal must be 0 or more")
		return
	}
	b := &s.brand.Brand
	if in.Budget != nil {
		b.BudgetUsdPerMonth = *in.Budget
	}
	if in.Goal != nil {
		b.GoalOrdersPerMonth = *in.Goal
	}
	if in.Target != nil {
		if b.Target == nil {
			b.Target = map[string]string{}
		}
		for k, v := range *in.Target {
			b.Target[k] = v
		}
	}
	if in.Market != nil {
		b.Market = *in.Market
	}
	s.writeJSON(w, *b)
}

// combosFor asks the adversary for shock combos; falls back to the fixed list.
// Codes and whys are read in one pass so they stay paired.
func (s *server) combosFor(d engine.Direction) ([][]string, []string, error) {
	var dirMap map[string]any
	if err := via(d, &dirMap); err != nil {
		return nil, nil, err
	}
	raw, err := spark.Adversary(dirMap, engine.PriorProbabilities(s.rates))
	if err != nil {
		return engine.FallbackCombos, nil, nil
	}
	var combos [][]string
	var whys []string
	list, _ := raw["shocks"].([]any)
	for _, item := range list {
		shock, _ := item.(map[string]any)
		codes, _ := shock["codes"].([]any)
		var row []string
		for _, c := range codes {
			if code, ok := c.(string); ok {
				row = append(row, code)
			}
		}
		if len(row) == 0 {
			continue
		}
		why, _ := shock["why"].(string)
		combos = append(combos, row)
		whys = append(whys, why)
	}
	return combos, whys, nil
}

// alignShocks pairs each stressed combo with the adversary's why (matched by cleaned codes).
func alignShocks(st engine.Stressed, combos [][]string, whys []string) []shockCombo {
	whyOf := map[string]string{}
	for i, c := range combos {
		cleaned := engine.CleanCombos([][]string{c})
		if len(cleaned) == 0 || i >= len(whys) {
			continue
		}
		key := strings.Join(cleaned[0], ",")
		if _, seen := whyOf[key]; !seen {
			whyOf[key] = whys[i]
		}
	}
	out := []shockCombo{}
	for _, c := range st.Combos {
		out = append(out, shockCombo{Codes: c, Why: whyOf[strings.Join(c, ",")]})
	}
	return out
}

func pct(x float64) string { return fmt.Sprintf("%.0f%%", x*100) }

func checkLine(f *engine.Flip) string {
	dir := " or worse"
	if f.Threshold > f.Current {
		dir = " or better"
	}
	return "Check before committing: message the creators who did post last month and ask if they would do it again. Creators posting at " +
		pct(f.Threshold) + dir + " flips the answer; you are at " + pct(f.Current) + "."
}

// flipPair returns the first two non-baseline verdict plans: A and B of the flip.
func flipPair(verdict []engine.Verdict) (a, b engine.Plan) {
	var picked []engine.Plan
	for _, v := range verdict {
		if v.Plan.PlanID != "baseline" {
			picked = append(picked, v.Plan)
		}
	}
	if len(picked) < 2 {
		return a, b
	}
	return picked[0], picked[1]
}

func (s *server) flipAndCheck(verdict []engine.Verdict, titles map[string]string) (*engine.Flip, string) {
	a, b := flipPair(verdict)
	if b.PlanID == "" {
		return nil, ""
	}
	flip := engine.RunFlip(a, b, engine.FallbackCombos, s.rates, s.brand, "seed_post_rate", 0.3, 1.0, 0.02)
	if flip == nil {
		return nil, ""
	}
	var flipMap map[string]any
	if err := via(flip, &flipMap); err != nil {
		return flip, checkLine(flip)
	}
	raw, err := spark.Check(a.Direction, b.Direction, flipMap, titles[a.Direction], titles[b.Direction])
	if err == nil {
		allowed := []float64{flip.Threshold, flip.Current, flip.Threshold * 100, flip.Current * 100}
		if line, ok := spark.ValidateCheck(raw, allowed); ok {
			return flip, line
		}
	}
	return flip, checkLine(flip)
}

func (s *server) handleStress(w http.ResponseWriter, r *http.Request) {
	var in struct {
		DirectionIDs []string `json:"direction_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || len(in.DirectionIDs) == 0 {
		writeError(w, 400, "direction_ids must be a non-empty list")
		return
	}
	dirs, err := s.currentDirections()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	byID := map[string]engine.Direction{}
	for _, d := range dirs {
		byID[d.ID] = d
	}
	var picked []engine.Direction
	seen := map[string]bool{}
	for _, id := range in.DirectionIDs {
		d, ok := byID[id]
		if !ok {
			writeError(w, 400, "unknown direction id: "+id)
			return
		}
		if !seen[id] {
			seen[id] = true
			picked = append(picked, d)
		}
	}
	resp := stressResponse{Catalog: map[string]string{}, Priors: engine.PriorProbabilities(s.rates)}
	for code, sh := range engine.Catalog {
		resp.Catalog[code] = sh.Label
	}
	var allPlans []engine.Plan
	merged := engine.Stressed{Matrix: map[string][]engine.Cell{}}
	titles := map[string]string{}
	for _, d := range picked {
		combos, whys, err := s.combosFor(d)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		plans := engine.SampleVariants(d, 60, 7)
		for i := range plans {
			plans[i].Month = s.planMonth()
		}
		st := engine.Stress(plans, combos, s.rates, s.brand)
		resp.Directions = append(resp.Directions, directionStress{ID: d.ID, Title: d.Title, Plans: plans,
			Shocks: alignShocks(st, combos, whys), Stress: st})
		allPlans = append(allPlans, plans...)
		titles[d.ID] = d.Title
		for k, cells := range st.Matrix {
			merged.Matrix[k] = cells
		}
	}
	merged.Combos = engine.FallbackCombos // fixed catalog list: baseline must not depend on click order
	resp.Verdict = engine.RunVerdict(allPlans, merged, s.rates, s.brand)
	resp.Gaps = engine.BaselineGaps(resp.Verdict)
	resp.Goal, resp.Budget, resp.AvgPrice = s.brand.Brand.GoalOrdersPerMonth, s.brand.Brand.BudgetUsdPerMonth, s.brand.Brand.AvgPriceUsd
	se := engine.SeasonFor(s.brand.Season, s.planMonth())
	resp.Month, resp.MonthName = se.Month, se.Name
	resp.SeasonGoal, resp.Demand = engine.SeasonGoal(resp.Goal, se.Demand), se.Demand
	resp.Flip, resp.Check = s.flipAndCheck(resp.Verdict, titles)
	if resp.Flip != nil {
		fa, fb := flipPair(resp.Verdict)
		resp.FlipA, resp.FlipB = fa.Direction, fb.Direction
	}
	s.writeJSON(w, resp)
}

// reviewOf runs the review of the posted plan (default: the October plan) on the calibrated rates.
func (s *server) reviewOf(r *http.Request) (engine.Plan, engine.Review, error) {
	var in struct {
		Plan engine.Plan `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil && !errors.Is(err, io.EOF) {
		return engine.Plan{}, engine.Review{}, err
	}
	plan := in.Plan
	if plan.PlanID == "" {
		plan = engine.OctoberPlan()
	}
	return plan, engine.RunReview(s.brand, plan, s.baseRates), nil
}

func (s *server) handleReview(w http.ResponseWriter, r *http.Request) {
	plan, review, err := s.reviewOf(r)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	s.writeJSON(w, map[string]any{"committed": plan, "review": review})
}

func (s *server) handleReviewApply(w http.ResponseWriter, r *http.Request) {
	_, review, err := s.reviewOf(r)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	s.rates = review.NewRates
	s.month = 2
	s.lastReview = &review
	s.writeJSON(w, map[string]any{"month": 2, "rates": s.rates})
}

// handleProfile reads the owner's Instagram screenshots through Spark (cache-only in the demo).
func (s *server) handleProfile(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Source string   `json:"source"`
		Handle string   `json:"handle"`
		Images []string `json:"images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if in.Source != "screenshot" {
		writeError(w, 400, `only source "screenshot" is available; Instagram Login comes after app review`)
		return
	}
	handle := spark.SafeHandle(in.Handle)
	if handle == "" {
		writeError(w, 400, "handle is required")
		return
	}
	if msg := imagesError(in.Images); msg != "" {
		writeError(w, 400, msg)
		return
	}
	raw, err := spark.Profile(handle, in.Images)
	if err != nil {
		writeError(w, 503, "no cached read for @"+handle+"; run `go run ./cmd/sparkcache` with MUSE_API_KEY set")
		return
	}
	s.writeJSON(w, spark.ValidateProfile(raw))
}

// imagesError: "" when images is 1-8 data:image/ URLs under 4 MB each, else the message for a 400.
func imagesError(images []string) string {
	if len(images) == 0 || len(images) > 8 {
		return "send 1 to 8 screenshots"
	}
	for _, img := range images {
		if !strings.HasPrefix(img, "data:image/") || len(img) > 4_000_000 {
			return "each screenshot must be a data:image/ URL under 4 MB"
		}
	}
	return ""
}
