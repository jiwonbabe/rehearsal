package engine

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// Direction ranges, as the planner (Spark) returns them. See SPEC.md §3.
type Ranges struct {
	PostsPerWeek  []float64            `json:"posts_per_week"`
	CreatorCount  []float64            `json:"creator_count"`
	AdsUsd        []float64            `json:"ads_usd"`
	PromoWeek     []*int               `json:"promo_week"`
	PromoDiscount []float64            `json:"promo_discount"`
	PostMix       map[string][]float64 `json:"post_mix"`
}

type Direction struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Plain      string  `json:"plain"`
	WhyForYou  string  `json:"why_for_you"`
	YouWouldDo string  `json:"you_would_do"`
	Rank       int     `json:"rank"`
	Caution    *string `json:"caution"`
	Ranges     Ranges  `json:"ranges"`
}

func pick(rng *rand.Rand, lo, hi, step float64) float64 {
	if lo == hi {
		return lo
	}
	if step > 0 {
		n := int(math.Round((hi - lo) / step))
		return lo + step*float64(rng.Intn(n+1))
	}
	return float64(int(lo) + rng.Intn(int(hi)-int(lo)+1))
}

// SampleVariants: plans inside the direction's ranges. Deterministic for a given seed.
// (Sequence differs from the Python reference; the reference fixtures carry their own plans.)
func SampleVariants(d Direction, n int, seed int64) []Plan {
	rng := rand.New(rand.NewSource(seed))
	var plans []Plan
	for i := 0; i < n; i++ {
		mix := map[string]float64{}
		total := 0.0
		for _, t := range PostTypes {
			lohi, ok := d.Ranges.PostMix[t]
			if !ok {
				lohi = []float64{0, 0}
			}
			mix[t] = pick(rng, lohi[0], lohi[1], 0.1)
			total += mix[t]
		}
		if total == 0 {
			total = 1
		}
		for t := range mix {
			mix[t] = math.Round(mix[t]/total*100) / 100
		}

		var promoWeek *int
		if len(d.Ranges.PromoWeek) > 0 {
			promoWeek = d.Ranges.PromoWeek[rng.Intn(len(d.Ranges.PromoWeek))]
		}
		promoDiscount := 0.0
		if promoWeek != nil && len(d.Ranges.PromoDiscount) == 2 {
			promoDiscount = pick(rng, d.Ranges.PromoDiscount[0], d.Ranges.PromoDiscount[1], 0.05)
		}

		plans = append(plans, Plan{
			PlanID:        fmt.Sprintf("%s-%03d", d.ID, i),
			Direction:     d.ID,
			PostsPerWeek:  pick(rng, d.Ranges.PostsPerWeek[0], d.Ranges.PostsPerWeek[1], 0),
			PostMix:       mix,
			CreatorCount:  pick(rng, d.Ranges.CreatorCount[0], d.Ranges.CreatorCount[1], 0),
			AdsUsd:        pick(rng, d.Ranges.AdsUsd[0], d.Ranges.AdsUsd[1], 20),
			PromoWeek:     promoWeek,
			PromoDiscount: promoDiscount,
		})
	}
	return plans
}

func BaselinePlan() Plan {
	return Plan{PlanID: "baseline", Direction: "just_post", PostsPerWeek: 5,
		PostMix: map[string]float64{"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}}
}

type Cell struct {
	Shocks   []string `json:"shocks"`
	Orders   float64  `json:"orders"`
	SpendUsd float64  `json:"spend_usd"`
	Survived bool     `json:"survived"`
}

type Stressed struct {
	Combos [][]string        `json:"combos"`
	Matrix map[string][]Cell `json:"matrix"`
}

// Stress runs every plan through every shock combo.
func Stress(plans []Plan, combos [][]string, r Rates, b *Brand) Stressed {
	combos = CleanCombos(combos)
	if len(combos) == 0 {
		combos = FallbackCombos
	}
	out := Stressed{Combos: combos, Matrix: map[string][]Cell{}}
	for _, p := range plans {
		var cells []Cell
		for ci, combo := range combos {
			res := RunPlanAt(p, r, b, combo, ci)
			cells = append(cells, Cell{Shocks: combo, Orders: round1(res.Orders), SpendUsd: round1(res.SpendUsd),
				Survived: Survived(res, b)})
		}
		out.Matrix[p.PlanID] = cells
	}
	return out
}

// score: more survivals, then lower average spend. Returns (survivals, -avgSpend).
func score(cells []Cell) (int, float64) {
	survivals, spend := 0, 0.0
	for _, c := range cells {
		if c.Survived {
			survivals++
		}
		spend += c.SpendUsd
	}
	return survivals, -spend / float64(len(cells))
}

func better(aS int, aSp float64, bS int, bSp float64) bool {
	if aS != bS {
		return aS > bS
	}
	return aSp > bSp
}

type Verdict struct {
	Plan           Plan      `json:"plan"`
	Survived       int       `json:"survived"`
	Total          int       `json:"total"`
	OrdersBadMonth []float64 `json:"orders_bad_month"`
	SpendUsd       float64   `json:"spend_usd"`
	DiesIf         []string  `json:"dies_if"`
}

func describe(p Plan, cells []Cell, r Rates, b *Brand) Verdict {
	survivals, _ := score(cells)
	var diesIf []string
	var orders []float64
	for _, c := range cells {
		if !c.Survived && diesIf == nil {
			diesIf = c.Shocks
		}
		orders = append(orders, c.Orders)
	}
	if diesIf == nil {
		diesIf = []string{}
	}
	sort.Float64s(orders)
	k := len(orders) / 5
	clean := RunPlan(p, r, b, nil)
	return Verdict{Plan: p, Survived: survivals, Total: len(cells),
		OrdersBadMonth: []float64{orders[k], orders[len(orders)-1-k]},
		SpendUsd:       math.Round(clean.SpendUsd), DiesIf: diesIf}
}

// RunVerdict: best plan per direction plus the no-spend baseline, sorted by survivals then spend.
func RunVerdict(plans []Plan, st Stressed, r Rates, b *Brand) []Verdict {
	best := map[string]Plan{}
	order := []string{}
	for _, p := range plans {
		cells := st.Matrix[p.PlanID]
		cur, ok := best[p.Direction]
		if !ok {
			best[p.Direction] = p
			order = append(order, p.Direction)
			continue
		}
		aS, aSp := score(cells)
		bS, bSp := score(st.Matrix[cur.PlanID])
		if better(aS, aSp, bS, bSp) {
			best[p.Direction] = p
		}
	}
	var out []Verdict
	for _, d := range order {
		out = append(out, describe(best[d], st.Matrix[best[d].PlanID], r, b))
	}
	base := BaselinePlan()
	if len(plans) > 0 {
		base.Month = plans[0].Month // the baseline lives in the same month as the plans it is compared with
	}
	baseCells := Stress([]Plan{base}, st.Combos, r, b).Matrix["baseline"]
	out = append(out, describe(base, baseCells, r, b))
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Survived != out[j].Survived {
			return out[i].Survived > out[j].Survived
		}
		return out[i].SpendUsd < out[j].SpendUsd
	})
	return out
}

type Flip struct {
	Param     string  `json:"param"`
	Threshold float64 `json:"threshold"`
	Current   float64 `json:"current"`
	AWinsNow  bool    `json:"a_wins_now"`
}

func setRate(r *Rates, param string, v float64) {
	switch param {
	case "seed_post_rate":
		r.SeedPostRate = v
	case "order_per_engagement":
		r.OrderPerEngagement = v
	case "ad_eng_per_usd":
		r.AdEngPerUsd = v
	case "eng_per_creator_post":
		r.EngPerCreatorPost = v
	}
}

func getRate(r Rates, param string) float64 {
	switch param {
	case "seed_post_rate":
		return r.SeedPostRate
	case "order_per_engagement":
		return r.OrderPerEngagement
	case "ad_eng_per_usd":
		return r.AdEngPerUsd
	case "eng_per_creator_post":
		return r.EngPerCreatorPost
	}
	return 0
}

// RunFlip sweeps one rate and returns the value where B starts beating A, or nil.
func RunFlip(a, b Plan, combos [][]string, r Rates, brand *Brand, param string, lo, hi, step float64) *Flip {
	wins := func(rr Rates) bool {
		aS, aSp := score(Stress([]Plan{a}, combos, rr, brand).Matrix[a.PlanID])
		bS, bSp := score(Stress([]Plan{b}, combos, rr, brand).Matrix[b.PlanID])
		return !better(bS, bSp, aS, aSp)
	}
	base := wins(r)
	for v := lo; v <= hi+1e-9; v += step {
		rr := r.Copy()
		setRate(&rr, param, v)
		if wins(rr) != base {
			return &Flip{Param: param, Threshold: math.Round(v*100) / 100,
				Current: math.Round(getRate(r, param)*100) / 100, AWinsNow: base}
		}
	}
	return nil
}

// Gap: what a paid plan buys over the no-spend baseline in a bad month (range midpoints), and the price of each extra order.
type Gap struct {
	Direction        string  `json:"direction"`
	PlanID           string  `json:"plan_id"`
	ExtraOrders      int     `json:"extra_orders"`
	ExtraSpendUsd    float64 `json:"extra_spend_usd"`
	UsdPerExtraOrder float64 `json:"usd_per_extra_order"`
}

func BaselineGaps(verdict []Verdict) []Gap {
	var base *Verdict
	for i := range verdict {
		if verdict[i].Plan.PlanID == "baseline" {
			base = &verdict[i]
		}
	}
	out := []Gap{}
	if base == nil {
		return out
	}
	mid := func(v Verdict) float64 { return (v.OrdersBadMonth[0] + v.OrdersBadMonth[1]) / 2 }
	for _, v := range verdict {
		if v.Plan.PlanID == "baseline" {
			continue
		}
		extra := halfUp(mid(v) - mid(*base))
		if mid(v)-mid(*base) < 0 {
			extra = -halfUp(mid(*base) - mid(v))
		}
		spend := v.SpendUsd - base.SpendUsd
		g := Gap{Direction: v.Plan.Direction, PlanID: v.Plan.PlanID, ExtraOrders: extra, ExtraSpendUsd: spend}
		if extra > 0 {
			g.UsdPerExtraOrder = roundTo(spend/float64(extra), 2)
		}
		out = append(out, g)
	}
	return out
}
