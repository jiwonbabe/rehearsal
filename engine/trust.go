package engine

import "math"

// Truth: the rates the synthetic generator used (data/synthetic.py). The engine never reads them to
// calibrate; RunTruthCheck only compares, so a reader can watch the engine recover a known answer.
type Truth struct {
	EngPerPost         map[string]float64 `json:"eng_per_post"`
	OrderPerEngagement float64            `json:"order_per_engagement"`
	PromoMultiplier    float64            `json:"promo_multiplier"`
	SeedPostRate       float64            `json:"seed_post_rate"`
	EngPerCreatorPost  float64            `json:"eng_per_creator_post"`
	AdEngPerUsd        float64            `json:"ad_eng_per_usd"`
	FollowersBase      float64            `json:"followers_base"`
	Note               string             `json:"note"`
}

type TruthRow struct {
	Rate   string  `json:"rate"`
	True   float64 `json:"true"`
	Engine float64 `json:"engine"`
	ErrPct float64 `json:"err_pct"`
	Note   string  `json:"note"`
}

// RunTruthCheck: engine-calibrated rates next to the generator's. Empty when the brand carries no truth.
func RunTruthCheck(b *Brand, r Rates) []TruthRow {
	if b.Truth == nil {
		return []TruthRow{}
	}
	t := b.Truth
	followers := 0.0
	for _, w := range b.Weeks {
		followers += w.Followers
	}
	avgFollowers := followers / float64(len(b.Weeks))
	var rows []TruthRow
	row := func(name string, truth, engine float64, note string) {
		err := 0.0
		if truth != 0 {
			err = roundTo(math.Abs(engine-truth)/truth*100, 1)
		}
		rows = append(rows, TruthRow{Rate: name, True: roundTo(truth, 4), Engine: roundTo(engine, 4), ErrPct: err, Note: note})
	}
	for _, pt := range PostTypes {
		row("eng_per_post."+pt, t.EngPerPost[pt]*avgFollowers/t.FollowersBase, r.EngPerPost[pt], "generator value scaled by average followers")
	}
	row("order_per_engagement", t.OrderPerEngagement, r.OrderPerEngagement, "")
	row("promo_multiplier", t.PromoMultiplier, r.PromoMultiplier, "")
	row("seed_post_rate", t.SeedPostRate, r.SeedPostRate, "October carries a planted creator no-show")
	row("eng_per_creator_post", t.EngPerCreatorPost, r.EngPerCreatorPost, "")
	row("ad_eng_per_usd", t.AdEngPerUsd, r.AdEngPerUsd, "")
	return rows
}

// RunHoldout: calibrate on everything before the last weeksBack weeks, then backtest those weeks the rates never saw.
func RunHoldout(b *Brand, weeksBack int) Backtest {
	pre := *b
	if len(b.Weeks) > weeksBack {
		pre.Weeks = b.Weeks[:len(b.Weeks)-weeksBack]
	}
	return RunBacktest(b, Calibrate(&pre), weeksBack)
}

// Coverage: how many of the brand's weeks actually used a lever. A lever with no history cannot be rated
// (Calibrate would read 0 ÷ 0 as "no effect"), and a thin one gets a warning next to its numbers.
type Coverage struct {
	Lever  string `json:"lever"`  // ads | creators | sale
	Weeks  int    `json:"weeks"`  // weeks in the history where the lever was used
	Status string `json:"status"` // none | thin | ok
}

const thinWeeks = 6 // fewer weeks than this and the rate rests on a handful of observations

func coverageStatus(weeks int) string {
	switch {
	case weeks == 0:
		return "none"
	case weeks < thinWeeks:
		return "thin"
	}
	return "ok"
}

func LeverCoverage(b *Brand) []Coverage {
	ads, creators, sale := 0, 0, 0
	for _, w := range b.Weeks {
		if w.AdsUsd > 0 {
			ads++
		}
		if w.CreatorsSeeded > 0 {
			creators++
		}
		if w.PromoDiscount > 0 {
			sale++
		}
	}
	return []Coverage{{"ads", ads, coverageStatus(ads)}, {"creators", creators, coverageStatus(creators)}, {"sale", sale, coverageStatus(sale)}}
}

// Untested: the levers a plan relies on that the history never used. Such a plan is shown, not rated.
func Untested(p Plan, cov []Coverage) []string {
	uses := map[string]bool{"ads": p.AdsUsd > 0, "creators": p.CreatorCount > 0, "sale": p.PromoWeek != nil}
	out := []string{}
	for _, c := range cov {
		if uses[c.Lever] && c.Status == "none" {
			out = append(out, c.Lever)
		}
	}
	return out
}
