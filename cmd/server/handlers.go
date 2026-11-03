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
	Catalog    map[string]string  `json:"catalog"`
	Priors     map[string]float64 `json:"priors"`
	Check      string             `json:"check"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// via round-trips src through JSON into dst.
func via(src, dst any) {
	raw, _ := json.Marshal(src)
	json.Unmarshal(raw, dst)
}

func (s *server) currentDirections() ([]engine.Direction, error) {
	block := ""
	if s.month == 2 && s.lastReview != nil {
		block = engine.LastMonthBlock(*s.lastReview)
	}
	summary := engine.Summary(s.brand, s.rates, engine.RunBacktest(s.brand, s.rates, 13))
	raw, err := spark.Directions(s.month, summary, block)
	if err != nil {
		return nil, fmt.Errorf("directions cache missing or unreadable: %w", err)
	}
	valid, err := spark.ValidateDirections(raw, s.brand.Brand.BudgetUsdPerMonth, s.brand.Brand.ProductCostUsd)
	if err != nil {
		return nil, err
	}
	var dirs []engine.Direction
	via(valid, &dirs)
	sort.SliceStable(dirs, func(i, j int) bool { return dirs[i].Rank < dirs[j].Rank })
	return dirs, nil
}

func (s *server) handleDirections(w http.ResponseWriter, r *http.Request) {
	dirs, err := s.currentDirections()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	resp := struct {
		Month      int                `json:"month"`
		Directions []engine.Direction `json:"directions"`
		LastMonth  *engine.Review     `json:"last_month"`
	}{s.month, dirs, s.lastReview}
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
	b := &s.brand.Brand
	if in.Budget != nil {
		b.BudgetUsdPerMonth = *in.Budget
	}
	if in.Goal != nil {
		b.GoalOrdersPerMonth = *in.Goal
	}
	if in.Target != nil {
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
func (s *server) combosFor(d engine.Direction) ([][]string, []string) {
	var dirMap map[string]any
	via(d, &dirMap)
	raw, err := spark.Adversary(dirMap, engine.PriorProbabilities(s.rates))
	if err != nil {
		return engine.FallbackCombos, nil
	}
	combos := spark.ParseShocks(raw)
	whys := make([]string, len(combos))
	list, _ := raw["shocks"].([]any)
	for i := range combos {
		if i < len(list) {
			item, _ := list[i].(map[string]any)
			whys[i], _ = item["why"].(string)
		}
	}
	return combos, whys
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
	return "Check before committing: message the creators who did post last month and ask if they would do it again. Creators posting at " +
		pct(f.Threshold) + " or better flips the answer; you are at " + pct(f.Current) + "."
}

// flipAndCheck: A and B are the first two non-baseline verdict plans.
func (s *server) flipAndCheck(verdict []engine.Verdict, titles map[string]string) (*engine.Flip, string) {
	var picked []engine.Plan
	for _, v := range verdict {
		if v.Plan.PlanID != "baseline" {
			picked = append(picked, v.Plan)
		}
	}
	if len(picked) < 2 {
		return nil, ""
	}
	a, b := picked[0], picked[1]
	flip := engine.RunFlip(a, b, engine.FallbackCombos, s.rates, s.brand, "seed_post_rate", 0.3, 1.0, 0.02)
	if flip == nil {
		return nil, ""
	}
	var flipMap map[string]any
	via(flip, &flipMap)
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
	resp := stressResponse{Catalog: map[string]string{}, Priors: engine.PriorProbabilities(s.rates)}
	for code, sh := range engine.Catalog {
		resp.Catalog[code] = sh.Label
	}
	var allPlans []engine.Plan
	merged := engine.Stressed{Matrix: map[string][]engine.Cell{}}
	titles := map[string]string{}
	for i, id := range in.DirectionIDs {
		d, ok := byID[id]
		if !ok {
			writeError(w, 400, "unknown direction id: "+id)
			return
		}
		combos, whys := s.combosFor(d)
		plans := engine.SampleVariants(d, 60, 7)
		st := engine.Stress(plans, combos, s.rates, s.brand)
		resp.Directions = append(resp.Directions, directionStress{ID: d.ID, Title: d.Title, Plans: plans,
			Shocks: alignShocks(st, combos, whys), Stress: st})
		allPlans = append(allPlans, plans...)
		titles[d.ID] = d.Title
		for k, cells := range st.Matrix {
			merged.Matrix[k] = cells
		}
		if i == 0 {
			merged.Combos = st.Combos
		}
	}
	resp.Verdict = engine.RunVerdict(allPlans, merged, s.rates, s.brand)
	resp.Flip, resp.Check = s.flipAndCheck(resp.Verdict, titles)
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
