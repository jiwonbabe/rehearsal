package engine

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Blend: new = 0.7*old + 0.3*observed. One month is evidence, not a verdict.
const Blend = 0.3

type ActualMonth struct {
	Orders             float64 `json:"orders"`
	SpendUsd           float64 `json:"spend_usd"`
	Engagement         float64 `json:"engagement"`
	CreatorsSeeded     float64 `json:"creators_seeded"`
	CreatorsPosted     float64 `json:"creators_posted"`
	SeedPostRate       float64 `json:"seed_post_rate"`
	OrderPerEngagement float64 `json:"order_per_engagement"`
}

func LastMonth(b *Brand, weeks int) ActualMonth {
	start := len(b.Weeks) - weeks
	if start < 0 {
		start = 0
	}
	var m ActualMonth
	for _, w := range b.Weeks[start:] {
		m.Orders += w.Orders
		m.SpendUsd += w.SpendUsd
		m.Engagement += totalEngagement(w)
		m.CreatorsSeeded += w.CreatorsSeeded
		m.CreatorsPosted += w.CreatorsPosted
	}
	m.SeedPostRate = safeDiv(m.CreatorsPosted, m.CreatorsSeeded)
	m.OrderPerEngagement = safeDiv(m.Orders, m.Engagement)
	return m
}

type RateChange struct {
	Rate string  `json:"rate"`
	Old  float64 `json:"old"`
	New  float64 `json:"new"`
}

type Review struct {
	Actual         ActualMonth  `json:"actual"`
	ExpectedOrders float64      `json:"expected_orders"`
	Range          []float64    `json:"range"`
	InsideRange    bool         `json:"inside_range"`
	Diagnosis      string       `json:"diagnosis"`
	RateChanges    []RateChange `json:"rate_changes"`
	NewRates       Rates        `json:"new_rates"`
}

// roundTo rounds on the exact binary value, like Python's round(), so 0.495 (stored
// slightly below) rounds to 0.49 rather than 0.50.
func roundTo(x float64, places int) float64 {
	v, _ := strconv.ParseFloat(strconv.FormatFloat(x, 'f', places, 64), 64)
	return v
}

// RunReview compares what happened to what the committed plan expected and updates rates.
func RunReview(b *Brand, committed Plan, r Rates) Review {
	actual := LastMonth(b, HorizonWeeks)
	expected := RunPlan(committed, r, b, nil)
	lo, hi := expected.Orders*0.85, expected.Orders*1.15
	inside := lo <= actual.Orders && actual.Orders <= hi

	out := Review{Actual: actual, ExpectedOrders: round1(expected.Orders), Range: []float64{round1(lo), round1(hi)},
		InsideRange: inside, RateChanges: []RateChange{}, NewRates: r.Copy()}

	if actual.CreatorsSeeded > 0 && math.Abs(actual.SeedPostRate-r.SeedPostRate) > 0.1 {
		out.NewRates.SeedPostRate = roundTo((1-Blend)*r.SeedPostRate+Blend*actual.SeedPostRate, 2)
		out.RateChanges = append(out.RateChanges, RateChange{"seed_post_rate", roundTo(r.SeedPostRate, 2), out.NewRates.SeedPostRate})
	}
	if math.Abs(safeDiv(actual.OrderPerEngagement, r.OrderPerEngagement)-1) > 0.15 {
		out.NewRates.OrderPerEngagement = roundTo((1-Blend)*r.OrderPerEngagement+Blend*actual.OrderPerEngagement, 5)
		out.RateChanges = append(out.RateChanges, RateChange{"order_per_engagement", roundTo(r.OrderPerEngagement, 4), roundTo(out.NewRates.OrderPerEngagement, 4)})
	}
	out.Diagnosis = diagnosis(actual, inside, out.RateChanges)
	return out
}

// diagnosis: one sentence from a template. Not a model.
func diagnosis(actual ActualMonth, inside bool, changes []RateChange) string {
	if inside {
		return "Orders landed inside the expected range. No rate changed."
	}
	seed, conv := false, false
	for _, c := range changes {
		if c.Rate == "seed_post_rate" {
			seed = true
		}
		if c.Rate == "order_per_engagement" {
			conv = true
		}
	}
	if seed && !conv {
		return fmt.Sprintf("Only %d of %d seeded creators posted, so seeded engagement never arrived and orders followed. "+
			"Your own conversion held. This was a creator problem, not a product or audience problem.",
			int(actual.CreatorsPosted), int(actual.CreatorsSeeded))
	}
	if conv {
		return "Engagement came in as expected but fewer of it turned into orders. Conversion was off."
	}
	return "Orders were outside the range, but no single rate explains it. Watch next month."
}

// Summary is the short data block the planner prompt receives. Numbers as context, never as output.
func Summary(b *Brand, r Rates, bt Backtest) map[string]any {
	month := LastMonth(b, HorizonWeeks)
	engPerPost := map[string]float64{}
	best := PostTypes[0]
	for _, t := range PostTypes {
		engPerPost[t] = math.Round(r.EngPerPost[t])
		if r.EngPerPost[t] > r.EngPerPost[best] {
			best = t
		}
	}
	return map[string]any{
		"brand": map[string]any{"market": b.Brand.Market, "target": b.Brand.Target, "avg_price_usd": b.Brand.AvgPriceUsd,
			"budget_usd_per_month": b.Brand.BudgetUsdPerMonth, "goal_orders_per_month": b.Brand.GoalOrdersPerMonth},
		"engagement_per_post":  engPerPost,
		"best_post_type":       best,
		"seed_post_rate":       roundTo(r.SeedPostRate, 2),
		"eng_per_creator_post": math.Round(r.EngPerCreatorPost),
		"ad_eng_per_usd":       roundTo(r.AdEngPerUsd, 1),
		"promo_multiplier":     roundTo(r.PromoMultiplier, 2),
		"last_month": map[string]any{"orders": month.Orders, "spend_usd": month.SpendUsd,
			"creators_seeded": month.CreatorsSeeded, "creators_posted": month.CreatorsPosted},
		"backtest": map[string]any{"median_abs_pct_error": bt.MedianAbsPctError,
			"weeks_inside_range": bt.WeeksInsideRange, "weeks_total": bt.WeeksTotal},
	}
}

// OctoberPlan is the plan the brand actually ran last month (same as the review fixture).
func OctoberPlan() Plan {
	return Plan{PlanID: "oct", Direction: "send_to_creators", PostsPerWeek: 3,
		PostMix: map[string]float64{"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}, CreatorCount: 4}
}

// LastMonthBlock is the {{LAST_MONTH_BLOCK}} text for the month-2 planner prompt.
func LastMonthBlock(rv Review) string {
	changes := []string{}
	for _, c := range rv.RateChanges {
		changes = append(changes, fmt.Sprintf("%s %g -> %g", c.Rate, c.Old, c.New))
	}
	rates := "none"
	if len(changes) > 0 {
		rates = strings.Join(changes, "; ")
	}
	return fmt.Sprintf("## Last month (engine review; context only)\n\n"+
		"Actual: %g orders, $%g spent, %g of %g seeded creators posted. Engine range was %g\u2013%g orders.\n"+
		"Diagnosis: %s\nRate changes: %s",
		rv.Actual.Orders, rv.Actual.SpendUsd, rv.Actual.CreatorsPosted, rv.Actual.CreatorsSeeded,
		rv.Range[0], rv.Range[1], rv.Diagnosis, rates)
}
