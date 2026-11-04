package engine

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const eps = 1e-6

func root() string { return filepath.Join("..") }

func loadJSON(t *testing.T, name string, into any) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root(), "reference", "fixtures", name+".json"))
	if err != nil {
		t.Fatalf("fixture %s: %v (run `python reference/dump_fixtures.py`)", name, err)
	}
	if err := json.Unmarshal(raw, into); err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
}

func brand(t *testing.T) *Brand {
	t.Helper()
	b, err := LoadBrand(filepath.Join(root(), "data", "brand.json"))
	if err != nil {
		t.Fatalf("brand: %v (run `python data/synthetic.py`)", err)
	}
	return b
}

func near(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > eps*math.Max(1, math.Abs(want)) {
		t.Errorf("%s: got %v want %v", label, got, want)
	}
}

func TestCalibrateMatchesReference(t *testing.T) {
	var want Rates
	loadJSON(t, "rates", &want)
	got := Calibrate(brand(t))
	for _, p := range PostTypes {
		near(t, "eng_per_post."+p, got.EngPerPost[p], want.EngPerPost[p])
	}
	near(t, "order_per_engagement", got.OrderPerEngagement, want.OrderPerEngagement)
	near(t, "promo_multiplier", got.PromoMultiplier, want.PromoMultiplier)
	near(t, "seed_post_rate", got.SeedPostRate, want.SeedPostRate)
	near(t, "eng_per_creator_post", got.EngPerCreatorPost, want.EngPerCreatorPost)
	near(t, "ad_eng_per_usd", got.AdEngPerUsd, want.AdEngPerUsd)
}

func TestBacktestMatchesReference(t *testing.T) {
	var want Backtest
	loadJSON(t, "backtest", &want)
	b := brand(t)
	got := RunBacktest(b, Calibrate(b), 13)
	if got.WeeksTotal != want.WeeksTotal || got.WeeksInsideRange != want.WeeksInsideRange {
		t.Fatalf("weeks: got %d/%d want %d/%d", got.WeeksInsideRange, got.WeeksTotal, want.WeeksInsideRange, want.WeeksTotal)
	}
	near(t, "median", got.MedianAbsPctError, want.MedianAbsPctError)
	for i := range want.Rows {
		near(t, "row engine "+want.Rows[i].Week, got.Rows[i].Engine, want.Rows[i].Engine)
	}
}

func TestRunPlanMatchesReference(t *testing.T) {
	var plans []Plan
	loadJSON(t, "plans", &plans)
	var want struct {
		Clean           Result `json:"clean"`
		Shocked         Result `json:"shocked"`
		CleanSurvived   bool   `json:"clean_survived"`
		ShockedSurvived bool   `json:"shocked_survived"`
	}
	loadJSON(t, "run_plan", &want)
	b := brand(t)
	r := Calibrate(b)

	clean := RunPlan(plans[0], r, b, nil)
	near(t, "clean orders", clean.Orders, want.Clean.Orders)
	near(t, "clean spend", clean.SpendUsd, want.Clean.SpendUsd)
	if Survived(clean, b) != want.CleanSurvived {
		t.Errorf("clean survived mismatch")
	}
	shocked := RunPlan(plans[25], r, b, []string{"reach_drop", "creator_noshow"})
	near(t, "shocked orders", shocked.Orders, want.Shocked.Orders)
	near(t, "shocked spend", shocked.SpendUsd, want.Shocked.SpendUsd)
	if Survived(shocked, b) != want.ShockedSurvived {
		t.Errorf("shocked survived mismatch")
	}
	again := RunPlan(plans[25], r, b, []string{"reach_drop", "creator_noshow"})
	if again.Orders != shocked.Orders {
		t.Errorf("not deterministic")
	}
}

func TestStressAndVerdictMatchReference(t *testing.T) {
	var plans []Plan
	loadJSON(t, "plans", &plans)
	var wantStress Stressed
	loadJSON(t, "stress", &wantStress)
	var wantVerdict []Verdict
	loadJSON(t, "verdict", &wantVerdict)
	b := brand(t)
	r := Calibrate(b)

	got := Stress(plans, FallbackCombos, r, b)
	for id, cells := range wantStress.Matrix {
		gc := got.Matrix[id]
		if len(gc) != len(cells) {
			t.Fatalf("%s: %d cells want %d", id, len(gc), len(cells))
		}
		for i := range cells {
			near(t, id+" orders", gc[i].Orders, cells[i].Orders)
			if gc[i].Survived != cells[i].Survived {
				t.Errorf("%s cell %d survived mismatch", id, i)
			}
		}
	}
	v := RunVerdict(plans, got, r, b)
	if len(v) != len(wantVerdict) {
		t.Fatalf("verdict len %d want %d", len(v), len(wantVerdict))
	}
	for i := range wantVerdict {
		if v[i].Plan.PlanID != wantVerdict[i].Plan.PlanID || v[i].Survived != wantVerdict[i].Survived {
			t.Errorf("verdict %d: got %s %d want %s %d", i, v[i].Plan.PlanID, v[i].Survived,
				wantVerdict[i].Plan.PlanID, wantVerdict[i].Survived)
		}
		near(t, "verdict spend", v[i].SpendUsd, wantVerdict[i].SpendUsd)
	}
}

func TestFlipMatchesReference(t *testing.T) {
	var plans []Plan
	loadJSON(t, "plans", &plans)
	var wantVerdict []Verdict
	loadJSON(t, "verdict", &wantVerdict)
	var want *Flip
	loadJSON(t, "flip", &want)
	b := brand(t)
	r := Calibrate(b)
	best := map[string]Plan{}
	for _, v := range wantVerdict {
		best[v.Plan.Direction] = v.Plan
	}
	got := RunFlip(best["pay_for_reach"], best["send_to_creators"], FallbackCombos, r, b, "seed_post_rate", 0.3, 1.0, 0.02)
	if (got == nil) != (want == nil) {
		t.Fatalf("flip nil mismatch: got %v want %v", got, want)
	}
	if got != nil {
		near(t, "threshold", got.Threshold, want.Threshold)
		if got.AWinsNow != want.AWinsNow {
			t.Errorf("a_wins_now mismatch")
		}
	}
}

func TestReviewMatchesReference(t *testing.T) {
	var want Review
	loadJSON(t, "review", &want)
	b := brand(t)
	r := Calibrate(b)
	committed := Plan{PlanID: "oct", Direction: "send_to_creators", PostsPerWeek: 3,
		PostMix: map[string]float64{"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}, CreatorCount: 4}
	got := RunReview(b, committed, r)
	near(t, "expected", got.ExpectedOrders, want.ExpectedOrders)
	if got.InsideRange != want.InsideRange {
		t.Errorf("inside_range mismatch")
	}
	if len(got.RateChanges) != len(want.RateChanges) {
		t.Fatalf("rate changes %d want %d", len(got.RateChanges), len(want.RateChanges))
	}
	for i := range want.RateChanges {
		if got.RateChanges[i].Rate != want.RateChanges[i].Rate {
			t.Errorf("rate change %d name", i)
		}
		near(t, "new rate", got.RateChanges[i].New, want.RateChanges[i].New)
	}
	near(t, "new seed_post_rate", got.NewRates.SeedPostRate, want.NewRates.SeedPostRate)
	if got.Diagnosis != want.Diagnosis {
		t.Errorf("diagnosis:\n got %q\nwant %q", got.Diagnosis, want.Diagnosis)
	}
}

func TestSampleVariantsStayInRanges(t *testing.T) {
	d := Direction{ID: "pay_for_reach", Ranges: Ranges{
		PostsPerWeek: []float64{3, 5}, CreatorCount: []float64{0, 0}, AdsUsd: []float64{120, 260},
		PromoWeek: []*int{nil}, PromoDiscount: []float64{0, 0},
		PostMix: map[string][]float64{"on_hand": {0.4, 0.6}, "close_up": {0.2, 0.4}, "flat_lay": {0.1, 0.3}, "on_neck": {0, 0.1}}}}
	a := SampleVariants(d, 30, 7)
	b := SampleVariants(d, 30, 7)
	for i, p := range a {
		if p.PostsPerWeek < 3 || p.PostsPerWeek > 5 || p.AdsUsd < 120 || p.AdsUsd > 260 || p.CreatorCount != 0 {
			t.Errorf("plan %d out of range: %+v", i, p)
		}
		sum := 0.0
		for _, v := range p.PostMix {
			sum += v
		}
		if math.Abs(sum-1) > 0.05 {
			t.Errorf("plan %d mix sums to %v", i, sum)
		}
		if b[i].AdsUsd != p.AdsUsd || b[i].PostsPerWeek != p.PostsPerWeek {
			t.Errorf("sampling not deterministic at %d", i)
		}
	}
}

func TestCleanCombos(t *testing.T) {
	out := CleanCombos([][]string{{"reach_drop", "nonsense"}, {"nothing"}, {"reach_drop", "nonsense"}})
	if len(out) != 1 || out[0][0] != "reach_drop" {
		t.Errorf("got %v", out)
	}
}

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

func TestGoalSuggestionsMatchReference(t *testing.T) {
	var want GoalSuggestions
	loadJSON(t, "goals", &want)
	got := RunGoalSuggestions(brand(t), 13)
	near(t, "avg_weekly_orders", got.AvgWeeklyOrders, want.AvgWeeklyOrders)
	near(t, "hold", got.Hold, want.Hold)
	near(t, "grow", got.Grow, want.Grow)
	near(t, "stretch", got.Stretch, want.Stretch)
	near(t, "avg_monthly_spend_usd", got.AvgMonthlySpendUsd, want.AvgMonthlySpendUsd)
	near(t, "avg_price_usd", got.AvgPriceUsd, want.AvgPriceUsd)
	near(t, "current_goal", got.CurrentGoal, want.CurrentGoal)
	near(t, "current_budget", got.CurrentBudget, want.CurrentBudget)
	if got.WeeksUsed != want.WeeksUsed {
		t.Errorf("weeks_used: got %d want %d", got.WeeksUsed, want.WeeksUsed)
	}
}

func TestLaunchMatchesReference(t *testing.T) {
	var want map[string]LaunchSpec
	loadJSON(t, "launch", &want)
	var verdict []Verdict
	loadJSON(t, "verdict", &verdict)
	var ads Plan
	for _, v := range verdict {
		if v.Plan.Direction == "pay_for_reach" {
			ads = v.Plan
			break
		}
	}
	b := brand(t)
	for name, p := range map[string]Plan{"ads": ads, "seeding": OctoberPlan()} {
		got, _ := json.Marshal(RunLaunch(p, b))
		exp, _ := json.Marshal(want[name])
		if string(got) != string(exp) {
			t.Errorf("%s:\n got %s\nwant %s", name, got, exp)
		}
	}
}

func TestLaunchPostsSumToWeekly(t *testing.T) {
	got := splitPosts(4, map[string]float64{"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0})
	if got["on_hand"]+got["close_up"]+got["flat_lay"]+got["on_neck"] != 4 || got["on_hand"] != 2 {
		t.Errorf("%v", got)
	}
	p := OctoberPlan()
	p.PostsPerWeek = 3.5
	w := RunLaunch(p, brand(t)).Calendar[0]
	if w.Posts["on_hand"]+w.Posts["close_up"]+w.Posts["flat_lay"]+w.Posts["on_neck"] != 4 {
		t.Errorf("%v", w.Posts)
	}
}

func TestEvidenceMatchesReference(t *testing.T) {
	var want Evidence
	loadJSON(t, "evidence", &want)
	m, err := LoadMeta(filepath.Join(root(), "data", "meta", "demo.brand"))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(RunEvidence(m))
	exp, _ := json.Marshal(want)
	if string(got) != string(exp) {
		t.Errorf("\n got %s\nwant %s", got, exp)
	}
}

func TestEvidenceNoFeedPosts(t *testing.T) {
	m, _ := LoadMeta(filepath.Join(root(), "data", "meta", "demo.brand"))
	var kept []MediaItem
	for _, p := range m.Media.Data {
		if p.MediaProductType != "FEED" {
			kept = append(kept, p)
		}
	}
	m.Media.Data = kept
	e := RunEvidence(m)
	if e.Formats[0].Posts != 0 || e.Formats[0].ReachVsFeed != 0 || e.Formats[1].ReachVsFeed != 0 {
		t.Errorf("%+v", e.Formats[:2])
	}
	if len(EvidenceNumbers(e, nil)) == 0 {
		t.Error("no numbers")
	}
}

func TestTruthCheckMatchesReference(t *testing.T) {
	var want []TruthRow
	loadJSON(t, "truth", &want)
	b := brand(t)
	got := RunTruthCheck(b, Calibrate(b))
	if len(got) != len(want) || len(got) != 9 {
		t.Fatalf("rows %d want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Rate != want[i].Rate {
			t.Errorf("row %d: %s want %s", i, got[i].Rate, want[i].Rate)
		}
		near(t, got[i].Rate+" true", got[i].True, want[i].True)
		near(t, got[i].Rate+" engine", got[i].Engine, want[i].Engine)
		near(t, got[i].Rate+" err", got[i].ErrPct, want[i].ErrPct)
	}
	b.Truth = nil
	if rows := RunTruthCheck(b, Calibrate(b)); len(rows) != 0 {
		t.Errorf("no truth → no rows, got %d", len(rows))
	}
}

func TestHoldoutMatchesReference(t *testing.T) {
	var want Backtest
	loadJSON(t, "holdout", &want)
	got := RunHoldout(brand(t), 13)
	near(t, "median", got.MedianAbsPctError, want.MedianAbsPctError)
	if got.WeeksInsideRange != want.WeeksInsideRange || got.WeeksTotal != 13 {
		t.Errorf("inside %d/%d want %d/%d", got.WeeksInsideRange, got.WeeksTotal, want.WeeksInsideRange, want.WeeksTotal)
	}
	if got.Rows[0].Week != want.Rows[0].Week {
		t.Errorf("first week %s want %s", got.Rows[0].Week, want.Rows[0].Week)
	}
}

func TestSeasonMatchesReference(t *testing.T) {
	var raw map[string]json.RawMessage
	loadJSON(t, "season", &raw)
	want := map[string]Result{}
	for _, k := range []string{"november", "december", "december_flat", "no_month"} {
		var r Result
		if err := json.Unmarshal(raw[k], &r); err != nil {
			t.Fatal(err)
		}
		want[k] = r
	}
	var plans []Plan
	loadJSON(t, "plans", &plans)
	b := brand(t)
	season, err := LoadSeasonality(filepath.Join(root(), "data", "seasonality.json"), "jewelry")
	if err != nil {
		t.Fatal(err)
	}
	b.Season = season
	r := Calibrate(b)
	nov, dec := plans[0], plans[0]
	nov.Month, dec.Month = 11, 12
	for name, got := range map[string]Result{"november": RunPlan(nov, r, b, nil), "december": RunPlan(dec, r, b, nil),
		"december_flat": RunPlan(dec, r, b, []string{"season_flat"}), "no_month": RunPlan(plans[0], r, b, nil)} {
		near(t, name+" orders", got.Orders, want[name].Orders)
		near(t, name+" spend", got.SpendUsd, want[name].SpendUsd)
		near(t, name+" engagement", got.Engagement, want[name].Engagement)
	}
	if want["december"].Orders <= want["november"].Orders || want["december_flat"].Orders >= want["december"].Orders {
		t.Errorf("season ordering: nov %v dec %v flat %v", want["november"].Orders, want["december"].Orders, want["december_flat"].Orders)
	}
	if s := SeasonFor(season, 11); s.Demand != 1.04 || s.AdCost != 1.18 {
		t.Errorf("november %+v", s)
	}
	if s := SeasonFor(nil, 11); s.Demand != 1 || s.AdCost != 1 {
		t.Errorf("nil table %+v", s)
	}
}

func TestBaselineGapsMatchReference(t *testing.T) {
	var want []Gap
	loadJSON(t, "gaps", &want)
	var verdict []Verdict
	loadJSON(t, "verdict", &verdict)
	got := BaselineGaps(verdict)
	if len(got) != len(want) || len(got) == 0 {
		t.Fatalf("gaps %d want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Direction != want[i].Direction || got[i].ExtraOrders != want[i].ExtraOrders {
			t.Errorf("%d: %+v want %+v", i, got[i], want[i])
		}
		near(t, got[i].Direction+" usd", got[i].UsdPerExtraOrder, want[i].UsdPerExtraOrder)
	}
	if len(BaselineGaps(verdict[:1])) != 0 {
		t.Error("no baseline → no gaps")
	}
}

func TestLaunchLaterMonthStartsLater(t *testing.T) {
	var want map[string]LaunchSpec
	loadJSON(t, "launch", &want)
	p := OctoberPlan()
	p.Month = 12
	got := RunLaunch(p, brand(t))
	if got.StartDate != want["december"].StartDate || got.StartDate != "2026-11-30" || got.Calendar[0].StartDate != "2026-11-30" {
		t.Errorf("december start %s want %s", got.StartDate, want["december"].StartDate)
	}
	p.Month = 11
	if s := RunLaunch(p, brand(t)).StartDate; s != "2026-11-02" {
		t.Errorf("november start %s", s)
	}
}

func TestSeasonGoalMatchesReference(t *testing.T) {
	var want map[string]any
	loadJSON(t, "season", &want)
	b := brand(t)
	if got := SeasonGoal(b.Brand.GoalOrdersPerMonth, 2.23); float64(got) != want["goal_december"].(float64) || got != 103 {
		t.Errorf("december goal %d want %v", got, want["goal_december"])
	}
	if got := SeasonGoal(b.Brand.GoalOrdersPerMonth, 1.04); float64(got) != want["goal_november"].(float64) {
		t.Errorf("november goal %d want %v", got, want["goal_november"])
	}
	if SeasonGoal(46, 1) != 46 {
		t.Error("average month keeps the goal")
	}
}

func TestSeasonCatalogPerCategory(t *testing.T) {
	c, err := LoadSeasonCatalog(filepath.Join(root(), "data", "seasonality.json"))
	if err != nil {
		t.Fatal(err)
	}
	if j := c.For("jewelry"); j == nil || j.Demand["12"] != 2.23 || j.Label != "Jewelry" || len(j.Sources) != 3 {
		t.Errorf("jewelry table: %+v", j)
	}
	if a := c.For("apparel"); a == nil || a.Demand["12"] != 1.57 || a.AdCost["12"] != 1.27 {
		t.Errorf("apparel table: %+v", a)
	}
	if c.For("nope") != nil || (*SeasonCatalog)(nil).For("jewelry") != nil {
		t.Error("unknown category or nil catalog must mean no prior")
	}
}

func TestLeverCoverageMatchesReference(t *testing.T) {
	var want struct {
		Coverage        []Coverage `json:"coverage"`
		NoAds           []Coverage `json:"no_ads"`
		UntestedAdsPlan []string   `json:"untested_ads_plan"`
		TestedAdsPlan   []string   `json:"tested_ads_plan"`
	}
	loadJSON(t, "coverage", &want)
	b := brand(t)
	if got := LeverCoverage(b); fmt.Sprint(got) != fmt.Sprint(want.Coverage) {
		t.Errorf("coverage %v want %v", got, want.Coverage)
	}
	noAds := *b
	noAds.Weeks = append([]Week{}, b.Weeks...)
	for i := range noAds.Weeks {
		noAds.Weeks[i].AdsUsd, noAds.Weeks[i].AdEngagement = 0, 0
	}
	cov := LeverCoverage(&noAds)
	if fmt.Sprint(cov) != fmt.Sprint(want.NoAds) {
		t.Errorf("no-ads coverage %v want %v", cov, want.NoAds)
	}
	plan := Plan{PlanID: "x", Direction: "pay_for_reach", PostsPerWeek: 4, PostMix: map[string]float64{"on_hand": 0.6, "close_up": 0.2, "flat_lay": 0.1, "on_neck": 0.1}, AdsUsd: 300}
	if got := Untested(plan, cov); fmt.Sprint(got) != fmt.Sprint(want.UntestedAdsPlan) || len(got) != 1 {
		t.Errorf("untested %v want %v", got, want.UntestedAdsPlan)
	}
	if got := Untested(plan, LeverCoverage(b)); len(got) != 0 || len(want.TestedAdsPlan) != 0 {
		t.Errorf("tested plan must have no untested lever: %v", got)
	}
}

func TestCreatorPoolMatchesReference(t *testing.T) {
	var want struct {
		Weights    []float64 `json:"weights"`
		Expected   []float64 `json:"expected"`
		Plan       Result    `json:"plan"`
		PlanCycles Result    `json:"plan_cycles"`
		PlanNoPool Result    `json:"plan_no_pool"`
		PlanOffset Result    `json:"plan_offset2"`
	}
	loadJSON(t, "creator_pool", &want)
	var ev Evidence
	loadJSON(t, "evidence", &ev)
	b := brand(t)
	r := Calibrate(b)
	w := CreatorWeights(ev.Creators)
	if len(w) != len(want.Weights) {
		t.Fatalf("weights %v want %v", w, want.Weights)
	}
	for i := range w {
		if math.Abs(w[i]-want.Weights[i]) > 1e-9 || ExpectedEngagement(ev.Creators[i]) != want.Expected[i] {
			t.Errorf("creator %d: weight %v want %v, expected %v want %v", i, w[i], want.Weights[i], ExpectedEngagement(ev.Creators[i]), want.Expected[i])
		}
	}
	mix := map[string]float64{"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}
	seed := Plan{PlanID: "seed", Direction: "send_to_creators", PostsPerWeek: 3, PostMix: mix, CreatorCount: 4}
	big := Plan{PlanID: "seed8", Direction: "send_to_creators", PostsPerWeek: 3, PostMix: mix, CreatorCount: 8}
	noPool := RunPlan(seed, r, b, nil)
	pb := *b
	pb.CreatorPool = ev.Creators
	withPool, cycles := RunPlan(seed, r, &pb, nil), RunPlan(big, r, &pb, nil)
	for _, c := range []struct {
		name      string
		got, want Result
	}{{"no pool", noPool, want.PlanNoPool}, {"pool", withPool, want.Plan}, {"cycles", cycles, want.PlanCycles}, {"offset 2", RunPlanAt(seed, r, &pb, nil, 2), want.PlanOffset}} {
		if math.Abs(c.got.Orders-c.want.Orders) > 1e-6 || math.Abs(c.got.SpendUsd-c.want.SpendUsd) > 1e-6 {
			t.Errorf("%s: orders %v want %v, spend %v want %v", c.name, c.got.Orders, c.want.Orders, c.got.SpendUsd, c.want.SpendUsd)
		}
	}
	if withPool.Orders == noPool.Orders {
		t.Error("the pool must change the creator plan's orders")
	}
	if RunBacktest(&pb, r, 13).MedianAbsPctError != RunBacktest(b, r, 13).MedianAbsPctError {
		t.Error("history must never use the pool")
	}
}
