// Package engine is the deterministic core: calibrate rates from brand data, run a plan,
// judge survival. No model calls here. Same formulas as reference/engine (Python).
package engine

import (
	"encoding/json"
	"math"
	"os"
	"sort"
)

var PostTypes = []string{"on_hand", "close_up", "flat_lay", "on_neck"}

const HorizonWeeks = 4

type Brand struct {
	Brand  BrandInfo    `json:"brand"`
	Weeks  []Week       `json:"weeks"`
	Truth  *Truth       `json:"generator_truth,omitempty"`
	Season *Seasonality `json:"season,omitempty"` // set by cmd/ from data/seasonality.json; a planning prior only
}

type BrandInfo struct {
	Name               string            `json:"name"`
	Synthetic          bool              `json:"synthetic"`
	Market             string            `json:"market"`
	Category           string            `json:"category"` // key into data/seasonality.json categories
	Target             map[string]string `json:"target"`
	AvgPriceUsd        float64           `json:"avg_price_usd"`
	ProductCostUsd     float64           `json:"product_cost_usd"`
	GoalOrdersPerMonth float64           `json:"goal_orders_per_month"`
	BudgetUsdPerMonth  float64           `json:"budget_usd_per_month"`
}

type Week struct {
	Week              string             `json:"week"`
	Posts             map[string]float64 `json:"posts"`
	Engagement        map[string]float64 `json:"engagement"`
	Followers         float64            `json:"followers"`
	CreatorsSeeded    float64            `json:"creators_seeded"`
	CreatorsPosted    float64            `json:"creators_posted"`
	CreatorEngagement float64            `json:"creator_engagement"`
	AdsUsd            float64            `json:"ads_usd"`
	AdEngagement      float64            `json:"ad_engagement"`
	PromoDiscount     float64            `json:"promo_discount"`
	Orders            float64            `json:"orders"`
	SpendUsd          float64            `json:"spend_usd"`
}

type Rates struct {
	EngPerPost         map[string]float64 `json:"eng_per_post"`
	OrderPerEngagement float64            `json:"order_per_engagement"`
	PromoMultiplier    float64            `json:"promo_multiplier"`
	SeedPostRate       float64            `json:"seed_post_rate"`
	EngPerCreatorPost  float64            `json:"eng_per_creator_post"`
	AdEngPerUsd        float64            `json:"ad_eng_per_usd"`
}

func (r Rates) Copy() Rates {
	out := r
	out.EngPerPost = map[string]float64{}
	for k, v := range r.EngPerPost {
		out.EngPerPost[k] = v
	}
	return out
}

type Plan struct {
	PlanID        string             `json:"plan_id"`
	Direction     string             `json:"direction"`
	PostsPerWeek  float64            `json:"posts_per_week"`
	PostMix       map[string]float64 `json:"post_mix"`
	CreatorCount  float64            `json:"creator_count"`
	AdsUsd        float64            `json:"ads_usd"`
	PromoWeek     *int               `json:"promo_week"`
	PromoDiscount float64            `json:"promo_discount"`
	Month         int                `json:"month"` // calendar month being planned (1-12); 0 = no seasonal prior
}

type WeekResult struct {
	Engagement float64 `json:"engagement"`
	Orders     float64 `json:"orders"`
	SpendUsd   float64 `json:"spend_usd"`
}

type Result struct {
	PlanID     string       `json:"plan_id"`
	Weeks      []WeekResult `json:"weeks"`
	Orders     float64      `json:"orders"`
	SpendUsd   float64      `json:"spend_usd"`
	Engagement float64      `json:"engagement"`
}

func LoadBrand(path string) (*Brand, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var b Brand
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func totalEngagement(w Week) float64 {
	sum := 0.0
	for _, v := range w.Engagement {
		sum += v
	}
	return sum + w.CreatorEngagement + w.AdEngagement
}

// Calibrate rediscovers the funnel rates from the weekly rows. Pure arithmetic.
func Calibrate(b *Brand) Rates {
	r := Rates{EngPerPost: map[string]float64{}}
	for _, t := range PostTypes {
		posts, eng := 0.0, 0.0
		for _, w := range b.Weeks {
			posts += w.Posts[t]
			eng += w.Engagement[t]
		}
		r.EngPerPost[t] = safeDiv(eng, posts)
	}

	plainOrders, plainEng, promoOrders, promoEng := 0.0, 0.0, 0.0, 0.0
	hasPromo := false
	for _, w := range b.Weeks {
		if w.PromoDiscount == 0 {
			plainOrders += w.Orders
			plainEng += totalEngagement(w)
		} else {
			hasPromo = true
			promoOrders += w.Orders
			promoEng += totalEngagement(w)
		}
	}
	r.OrderPerEngagement = safeDiv(plainOrders, plainEng)
	r.PromoMultiplier = 1.0
	if hasPromo {
		promoRate := safeDiv(promoOrders, promoEng)
		if m := safeDiv(promoRate, r.OrderPerEngagement); m != 0 {
			r.PromoMultiplier = m
		}
	}

	seeded, posted, creatorEng, ads, adEng := 0.0, 0.0, 0.0, 0.0, 0.0
	for _, w := range b.Weeks {
		seeded += w.CreatorsSeeded
		posted += w.CreatorsPosted
		creatorEng += w.CreatorEngagement
		ads += w.AdsUsd
		adEng += w.AdEngagement
	}
	r.SeedPostRate = safeDiv(posted, seeded)
	r.EngPerCreatorPost = safeDiv(creatorEng, posted)
	r.AdEngPerUsd = safeDiv(adEng, ads)
	return r
}

// weekInputs: what the brand does in one week under this plan (weekIndex 0..3).
func weekInputs(p Plan, weekIndex int) (posts map[string]float64, creators, adsUsd, discount float64) {
	posts = map[string]float64{}
	for _, t := range PostTypes {
		posts[t] = p.PostsPerWeek * p.PostMix[t]
	}
	if weekIndex == 0 {
		creators = p.CreatorCount
	}
	adsUsd = p.AdsUsd / HorizonWeeks
	if p.PromoWeek != nil && *p.PromoWeek == weekIndex+1 {
		discount = p.PromoDiscount
	}
	return
}

// RunWeek: one week of the funnel. mods are multipliers from shocks (see shocks.go).
// RunWeek: one week of the funnel. month applies the seasonal prior (demand lift on the order rate,
// ad-cost index on ad engagement); 0 means none. Same formulas as reference/engine/model.py run_week.
func RunWeek(posts map[string]float64, creatorsSeeded, adsUsd, discount float64, r Rates, b *Brand, mods map[string]float64, month int) WeekResult {
	mult := func(key string) float64 {
		if v, ok := mods[key]; ok {
			return v
		}
		return 1.0
	}
	season := SeasonFor(b.Season, month)
	lift := 1.0 + (season.Demand-1.0)*mult("season_lift_mult")
	organic := 0.0
	for _, t := range PostTypes {
		organic += posts[t] * r.EngPerPost[t]
	}
	organic *= mult("organic_mult")

	postRate := r.SeedPostRate * mult("seed_post_rate_mult")
	creatorsPosted := creatorsSeeded * postRate
	creatorEng := creatorsPosted * r.EngPerCreatorPost

	adEng := adsUsd * r.AdEngPerUsd * mult("ad_eng_mult") / season.AdCost

	engagement := organic + creatorEng + adEng
	orderRate := r.OrderPerEngagement * mult("order_rate_mult") * lift
	promoMult := 1.0
	if discount > 0 {
		promoMult = r.PromoMultiplier * mult("promo_mult_mult")
	}
	orders := engagement * orderRate * promoMult

	spend := adsUsd + creatorsSeeded*b.Brand.ProductCostUsd
	spend += orders * b.Brand.AvgPriceUsd * discount
	return WeekResult{Engagement: engagement, Orders: orders, SpendUsd: spend}
}

func RunPlan(p Plan, r Rates, b *Brand, shocks []string) Result {
	out := Result{PlanID: p.PlanID}
	for i := 0; i < HorizonWeeks; i++ {
		posts, creators, adsUsd, discount := weekInputs(p, i)
		mods := ModsForWeek(shocks, i)
		if m, ok := mods["creator_count_mult"]; ok {
			creators = math.Round(creators * m)
		}
		wr := RunWeek(posts, creators, adsUsd, discount, r, b, mods, p.Month)
		out.Weeks = append(out.Weeks, wr)
		out.Orders += wr.Orders
		out.SpendUsd += wr.SpendUsd
		out.Engagement += wr.Engagement
	}
	return out
}

func Survived(res Result, b *Brand) bool {
	goal := b.Brand.GoalOrdersPerMonth
	budget := b.Brand.BudgetUsdPerMonth
	cumulative := 0.0
	for i, w := range res.Weeks {
		cumulative += w.SpendUsd
		if i < 2 && cumulative > budget {
			return false
		}
	}
	return res.Orders >= goal && res.SpendUsd <= budget*1.1
}

type BacktestRow struct {
	Week   string  `json:"week"`
	Engine float64 `json:"engine"`
	Actual float64 `json:"actual"`
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Promo  bool    `json:"promo"`
}

type Backtest struct {
	Rows              []BacktestRow `json:"rows"`
	MedianAbsPctError float64       `json:"median_abs_pct_error"`
	WeeksInsideRange  int           `json:"weeks_inside_range"`
	WeeksTotal        int           `json:"weeks_total"`
}

func round1(x float64) float64 { return math.Round(x*10) / 10 }

// RunBacktest replays the last N weeks' actual inputs and compares engine orders to actual.
func RunBacktest(b *Brand, r Rates, weeksBack int) Backtest {
	out := Backtest{}
	var errors []float64
	start := len(b.Weeks) - weeksBack
	if start < 0 {
		start = 0
	}
	for _, w := range b.Weeks[start:] {
		wr := RunWeek(w.Posts, w.CreatorsSeeded, w.AdsUsd, w.PromoDiscount, r, b, map[string]float64{}, 0)
		errors = append(errors, safeDiv(math.Abs(wr.Orders-w.Orders), w.Orders))
		lo, hi := wr.Orders*0.8, wr.Orders*1.2
		if lo <= w.Orders && w.Orders <= hi {
			out.WeeksInsideRange++
		}
		out.Rows = append(out.Rows, BacktestRow{Week: w.Week, Engine: round1(wr.Orders), Actual: w.Orders,
			Low: round1(lo), High: round1(hi), Promo: w.PromoDiscount > 0})
	}
	out.WeeksTotal = len(out.Rows)
	if len(errors) > 0 {
		sort.Float64s(errors)
		n := len(errors)
		med := errors[n/2]
		if n%2 == 0 {
			med = (errors[n/2-1] + errors[n/2]) / 2
		}
		out.MedianAbsPctError = round1(med * 100)
	}
	return out
}
