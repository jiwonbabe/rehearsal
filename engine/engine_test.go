package engine

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
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
