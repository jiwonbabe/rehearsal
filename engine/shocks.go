package engine

// Shock catalog. Fixed effects; the adversary only picks combinations of these codes.
type Shock struct {
	Label string
	Weeks []int
	Mods  map[string]float64
}

var Catalog = map[string]Shock{
	"reach_drop":          {"Reach drops 40% (algorithm change)", []int{0, 1, 2, 3}, map[string]float64{"organic_mult": 0.6}},
	"creator_noshow":      {"Half of seeded creators never post", []int{0, 1, 2, 3}, map[string]float64{"seed_post_rate_mult": 0.5}},
	"competitor_sale":     {"A competitor runs a 20% sale", []int{1, 2}, map[string]float64{"order_rate_mult": 0.7}},
	"cpm_spike":           {"Ad cost per result up 60%", []int{0, 1, 2, 3}, map[string]float64{"ad_eng_mult": 0.625}},
	"shipping_delay":      {"Shipping delay in week 3", []int{2, 3}, map[string]float64{"order_rate_mult": 0.85}},
	"viral_post":          {"One post goes viral", []int{1}, map[string]float64{"organic_mult": 3.0}},
	"promo_underestimate": {"Sale week does worse than the engine's weakest estimate", []int{0, 1, 2, 3}, map[string]float64{"promo_mult_mult": 0.8}},
}

var FallbackCombos = [][]string{
	{"reach_drop"}, {"creator_noshow"}, {"competitor_sale"}, {"cpm_spike"}, {"shipping_delay"},
	{"viral_post"}, {"promo_underestimate"},
	{"reach_drop", "creator_noshow"}, {"reach_drop", "cpm_spike"}, {"creator_noshow", "competitor_sale"},
	{"cpm_spike", "competitor_sale"}, {"reach_drop", "shipping_delay"}, {"creator_noshow", "shipping_delay"},
	{"competitor_sale", "promo_underestimate"}, {"reach_drop", "competitor_sale"},
	{"reach_drop", "creator_noshow", "competitor_sale"}, {"cpm_spike", "shipping_delay"},
	{"creator_noshow", "cpm_spike"}, {"reach_drop", "promo_underestimate"}, {"viral_post", "competitor_sale"},
}

// PriorProbabilities: starting weights, from the brand's own history where we have it.
func PriorProbabilities(r Rates) map[string]float64 {
	noshow := 1.0 - r.SeedPostRate
	if noshow < 0.1 {
		noshow = 0.1
	}
	return map[string]float64{
		"reach_drop":          0.22,
		"creator_noshow":      float64(int(noshow*100+0.5)) / 100,
		"competitor_sale":     0.31,
		"cpm_spike":           0.18,
		"shipping_delay":      0.12,
		"viral_post":          0.08,
		"promo_underestimate": 0.30,
	}
}

func sameCombo(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CleanCombos drops unknown codes and empty or duplicate combos; keeps at most 20.
func CleanCombos(combos [][]string) [][]string {
	var out [][]string
	for _, combo := range combos {
		var codes []string
		for _, c := range combo {
			if _, ok := Catalog[c]; ok {
				codes = append(codes, c)
			}
		}
		if len(codes) == 0 {
			continue
		}
		dup := false
		for _, seen := range out {
			if sameCombo(seen, codes) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, codes)
		}
		if len(out) == 20 {
			break
		}
	}
	return out
}

// ModsForWeek multiplies the modifiers of every active shock for this week.
func ModsForWeek(codes []string, weekIndex int) map[string]float64 {
	mods := map[string]float64{}
	for _, code := range codes {
		shock, ok := Catalog[code]
		if !ok {
			continue
		}
		active := false
		for _, w := range shock.Weeks {
			if w == weekIndex {
				active = true
			}
		}
		if !active {
			continue
		}
		for key, value := range shock.Mods {
			if _, ok := mods[key]; !ok {
				mods[key] = 1.0
			}
			mods[key] *= value
		}
	}
	return mods
}
