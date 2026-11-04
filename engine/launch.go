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
	if p.Month > 0 {
		// a plan for a later month starts 4-week blocks later (the data's weeks are Monday-based)
		k := (p.Month - int(start.Month()) + 12) % 12
		start = start.AddDate(0, 0, 28*k)
	}
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

// CreatorPlanMath: what seeding N creators means in numbers an owner can check before committing. Every figure is
// the brand's own calibrated rate or the found creators' audience; no model writes any of it.
type CreatorPlanMath struct {
	Creators           int     `json:"creators"`
	FollowersMin       float64 `json:"followers_min"`       // from the creator pool; 0 without one
	FollowersMax       float64 `json:"followers_max"`       //
	EngPerPostMin      float64 `json:"eng_per_post_min"`    // expected engagement on one creator post (followers × their rate)
	EngPerPostMax      float64 `json:"eng_per_post_max"`    // without a pool both equal the history's engagement per creator post
	ExpectedPosts      float64 `json:"expected_posts"`      // creators × the history's post rate
	ExpectedEngagement float64 `json:"expected_engagement"` // expected posts × engagement per creator post (calibrated mean)
	ExpectedOrders     float64 `json:"expected_orders"`     // × order rate × the month's demand index, no shocks
	ShipUsd            float64 `json:"ship_usd"`            // creators × product cost
}

func CreatorMath(count float64, r Rates, b *Brand, month int) CreatorPlanMath {
	n := halfUp(count)
	m := CreatorPlanMath{Creators: n, EngPerPostMin: math.Round(r.EngPerCreatorPost), EngPerPostMax: math.Round(r.EngPerCreatorPost)}
	for i, c := range b.CreatorPool {
		e := ExpectedEngagement(c)
		if i == 0 || c.Followers < m.FollowersMin {
			m.FollowersMin = c.Followers
		}
		if i == 0 || c.Followers > m.FollowersMax {
			m.FollowersMax = c.Followers
		}
		if i == 0 || e < m.EngPerPostMin {
			m.EngPerPostMin = e
		}
		if i == 0 || e > m.EngPerPostMax {
			m.EngPerPostMax = e
		}
	}
	posts := float64(n) * r.SeedPostRate
	eng := posts * r.EngPerCreatorPost
	m.ExpectedPosts, m.ExpectedEngagement = round1(posts), math.Round(eng)
	m.ExpectedOrders = round1(eng * r.OrderPerEngagement * SeasonFor(b.Season, month).Demand)
	m.ShipUsd = float64(n) * b.Brand.ProductCostUsd
	return m
}

// CreatorBand: what one creator of a size band is worth, from the creators found in that band: expected engagement
// on a post (followers × their rate) and the orders that engagement brings at the brand's own order rate, for the
// month being planned. PerContact multiplies by the history's post rate (not everyone you send to posts).
type CreatorBand struct {
	Band             string  `json:"band"` // small (<10k) | mid (<100k) | large
	Count            int     `json:"count"`
	FollowersMin     float64 `json:"followers_min"`
	FollowersMax     float64 `json:"followers_max"`
	EngPerPostMin    float64 `json:"eng_per_post_min"`
	EngPerPostMax    float64 `json:"eng_per_post_max"`
	OrdersPerPostMin float64 `json:"orders_per_post_min"`
	OrdersPerPostMax float64 `json:"orders_per_post_max"`
	OrdersPerContact float64 `json:"orders_per_contact"` // mean orders per creator contacted (× post rate)
	PiecesForBudget  int     `json:"pieces_for_budget"`  // how many creators the budget ships to, at product cost
}

var bandOrder = []string{"small", "mid", "large"}

func CreatorBands(r Rates, b *Brand, month int) []CreatorBand {
	demand := SeasonFor(b.Season, month).Demand
	orderPerEng := r.OrderPerEngagement * demand
	pieces := 0
	if b.Brand.ProductCostUsd > 0 {
		pieces = int(math.Floor(b.Brand.BudgetUsdPerMonth / b.Brand.ProductCostUsd))
	}
	out := []CreatorBand{}
	for _, band := range bandOrder {
		cb := CreatorBand{Band: band, PiecesForBudget: pieces}
		sum := 0.0
		for _, c := range b.CreatorPool {
			if c.FollowersBand != band {
				continue
			}
			e := ExpectedEngagement(c)
			if cb.Count == 0 || c.Followers < cb.FollowersMin {
				cb.FollowersMin = c.Followers
			}
			if cb.Count == 0 || c.Followers > cb.FollowersMax {
				cb.FollowersMax = c.Followers
			}
			if cb.Count == 0 || e < cb.EngPerPostMin {
				cb.EngPerPostMin = e
			}
			if cb.Count == 0 || e > cb.EngPerPostMax {
				cb.EngPerPostMax = e
			}
			sum += e
			cb.Count++
		}
		if cb.Count == 0 {
			continue
		}
		cb.OrdersPerPostMin, cb.OrdersPerPostMax = round1(cb.EngPerPostMin*orderPerEng), round1(cb.EngPerPostMax*orderPerEng)
		cb.OrdersPerContact = round1(sum / float64(cb.Count) * orderPerEng * r.SeedPostRate)
		out = append(out, cb)
	}
	return out
}

// Breakdown: where a plan's money goes in a plain month, against the monthly budget. Ads and pieces shipped to
// creators are what the owner pays; discounts are revenue given up on the orders the plan brings.
type Breakdown struct {
	AdsUsd         float64 `json:"ads_usd"`
	CreatorCount   int     `json:"creator_count"`
	CreatorShipUsd float64 `json:"creator_ship_usd"`
	DiscountUsd    float64 `json:"discount_usd"`
	TotalUsd       float64 `json:"total_usd"`
	BudgetUsd      float64 `json:"budget_usd"`
	LeftUsd        float64 `json:"left_usd"` // budget minus total; negative = over budget
}

func SpendBreakdown(p Plan, r Rates, b *Brand) Breakdown {
	clean := RunPlan(p, r, b, nil)
	n := halfUp(p.CreatorCount)
	ship := float64(n) * b.Brand.ProductCostUsd
	discount := math.Max(0, clean.SpendUsd-p.AdsUsd-ship)
	bd := Breakdown{AdsUsd: math.Round(p.AdsUsd), CreatorCount: n, CreatorShipUsd: math.Round(ship), DiscountUsd: math.Round(discount),
		BudgetUsd: b.Brand.BudgetUsdPerMonth}
	bd.TotalUsd = bd.AdsUsd + bd.CreatorShipUsd + bd.DiscountUsd
	bd.LeftUsd = bd.BudgetUsd - bd.TotalUsd
	return bd
}

// CombinePlans: the two plans' levers together in one plan ("mix") that fits the monthly budget. Posting is the
// busier of the two, ads and creators the larger of each, the sale from whichever has one; when ads plus pieces for
// creators would pass the budget both are scaled down by the same factor (ads to the dollar ten, creators down).
func CombinePlans(a, b Plan, budget, productCost float64) Plan {
	p := a
	p.PlanID, p.Direction = "mix-001", "mix"
	if b.PostsPerWeek > p.PostsPerWeek {
		p.PostsPerWeek = b.PostsPerWeek
	}
	p.AdsUsd = math.Max(a.AdsUsd, b.AdsUsd)
	p.CreatorCount = math.Max(a.CreatorCount, b.CreatorCount)
	if p.PromoWeek == nil && b.PromoWeek != nil {
		p.PromoWeek, p.PromoDiscount = b.PromoWeek, b.PromoDiscount
	}
	if cost := p.AdsUsd + p.CreatorCount*productCost; cost > budget && cost > 0 {
		k := budget / cost
		p.AdsUsd = math.Floor(p.AdsUsd*k/10) * 10
		p.CreatorCount = math.Floor(p.CreatorCount * k)
		for p.AdsUsd+p.CreatorCount*productCost > budget && p.CreatorCount > 0 {
			p.CreatorCount--
		}
	}
	return p
}
