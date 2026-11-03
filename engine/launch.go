package engine

import (
	"math"
	"sort"
	"time"
)

// LaunchWeek: one calendar week of the committed plan.
type LaunchWeek struct {
	Week          int            `json:"week"`
	StartDate     string         `json:"start_date"`
	Posts         map[string]int `json:"posts"`
	Creators      int            `json:"creators"`
	AdsUsd        float64        `json:"ads_usd"`
	PromoDiscount float64        `json:"promo_discount"`
}

// LaunchSpec: the committed plan as dates, daily ad budget, creator kit numbers and a calendar.
// Dates derive from the last data week, never from the clock. Same formulas as reference/engine/launch.py.
type LaunchSpec struct {
	StartDate      string       `json:"start_date"`
	EndDate        string       `json:"end_date"`
	Days           int          `json:"days"`
	AdsUsd         float64      `json:"ads_usd"`
	AdsDailyUsd    float64      `json:"ads_daily_usd"`
	AdsDailyCents  int          `json:"ads_daily_cents"`
	CreatorCount   int          `json:"creator_count"`
	CreatorShipUsd float64      `json:"creator_ship_usd"`
	CreatorDMs     int          `json:"creator_dms"`
	TotalSpendUsd  float64      `json:"total_spend_usd"`
	Calendar       []LaunchWeek `json:"calendar"`
}

func halfUp(x float64) int { return int(math.Floor(x + 0.5)) }

// splitPosts: integer posts per type summing to total, by largest remainder; ties keep PostTypes order.
func splitPosts(total int, mix map[string]float64) map[string]int {
	out := map[string]int{}
	rem := map[string]float64{}
	left := total
	for _, t := range PostTypes {
		q := float64(total) * mix[t]
		out[t] = int(math.Floor(q))
		rem[t] = q - math.Floor(q)
		left -= out[t]
	}
	order := append([]string{}, PostTypes...)
	sort.SliceStable(order, func(i, j int) bool { return rem[order[i]] > rem[order[j]] })
	for i := 0; i < left && i < len(order); i++ {
		out[order[i]]++
	}
	return out
}

func RunLaunch(p Plan, b *Brand) LaunchSpec {
	last, _ := time.Parse("2006-01-02", b.Weeks[len(b.Weeks)-1].Week)
	start := last.AddDate(0, 0, 7)
	days := HorizonWeeks * 7
	count := int(p.CreatorCount)
	ship := float64(count) * b.Brand.ProductCostUsd
	posts := halfUp(p.PostsPerWeek)
	var cal []LaunchWeek
	for w := 0; w < HorizonWeeks; w++ {
		week := LaunchWeek{Week: w + 1, StartDate: start.AddDate(0, 0, 7*w).Format("2006-01-02"),
			Posts: splitPosts(posts, p.PostMix), AdsUsd: roundTo(p.AdsUsd/HorizonWeeks, 2)}
		if w == 0 {
			week.Creators = count
		}
		if p.PromoWeek != nil && *p.PromoWeek == w+1 {
			week.PromoDiscount = p.PromoDiscount
		}
		cal = append(cal, week)
	}
	return LaunchSpec{
		StartDate: start.Format("2006-01-02"), EndDate: start.AddDate(0, 0, days-1).Format("2006-01-02"), Days: days,
		AdsUsd: p.AdsUsd, AdsDailyUsd: roundTo(p.AdsUsd/float64(days), 2), AdsDailyCents: halfUp(p.AdsUsd / float64(days) * 100),
		CreatorCount: count, CreatorShipUsd: ship, CreatorDMs: count * 2,
		TotalSpendUsd: p.AdsUsd + ship, Calendar: cal,
	}
}
