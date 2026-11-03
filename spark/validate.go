package spark

import (
	"errors"
	"math"
)

var validDirections = map[string]bool{
	"just_post": true, "send_to_creators": true, "pay_for_reach": true, "one_week_sale": true, "mix": true,
}

// pair returns a [lo, hi] from a 2-element numeric list (swapped if reversed), else def.
func pair(v any, def [2]float64) []float64 {
	list, ok := v.([]any)
	if !ok || len(list) != 2 {
		return def[:]
	}
	lo, okLo := list[0].(float64)
	hi, okHi := list[1].(float64)
	if !okLo || !okHi {
		return def[:]
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	return []float64{lo, hi}
}

// promoWeeks keeps nil and whole numbers 1-4; defaults to [nil].
func promoWeeks(v any) []*int {
	out := []*int{}
	list, _ := v.([]any)
	for _, item := range list {
		if item == nil {
			out = append(out, nil)
		} else if f, ok := item.(float64); ok && f == math.Trunc(f) && f >= 1 && f <= 4 {
			week := int(f)
			out = append(out, &week)
		}
	}
	if len(out) == 0 {
		return []*int{nil}
	}
	return out
}

func repairRanges(raw any, budget, productCost float64) map[string]any {
	in, _ := raw.(map[string]any)
	if in == nil {
		in = map[string]any{}
	}
	postsPerWeek := pair(in["posts_per_week"], [2]float64{3, 5})
	creatorCount := pair(in["creator_count"], [2]float64{0, 0})
	adsUsd := pair(in["ads_usd"], [2]float64{0, 0})
	promoDiscount := pair(in["promo_discount"], [2]float64{0, 0})

	adsUsd[1] = math.Min(adsUsd[1], budget)
	maxCreators := math.Max(0, math.Floor((budget-adsUsd[0])/productCost))
	creatorCount[1] = math.Min(creatorCount[1], maxCreators)
	adsUsd[0] = math.Min(adsUsd[0], adsUsd[1])
	creatorCount[0] = math.Min(creatorCount[0], creatorCount[1])

	mixIn, _ := in["post_mix"].(map[string]any)
	postMix := map[string][]float64{}
	for key, def := range map[string][2]float64{
		"on_hand": {0.4, 0.6}, "close_up": {0.2, 0.4}, "flat_lay": {0.1, 0.3}, "on_neck": {0, 0.1},
	} {
		postMix[key] = pair(mixIn[key], def)
	}

	return map[string]any{
		"posts_per_week": postsPerWeek, "creator_count": creatorCount, "ads_usd": adsUsd,
		"promo_discount": promoDiscount, "promo_week": promoWeeks(in["promo_week"]), "post_mix": postMix,
	}
}

// ValidateDirections returns the directions list with ranges repaired and clamped to budget.
func ValidateDirections(raw map[string]any, budget, productCost float64) ([]map[string]any, error) {
	list, _ := raw["directions"].([]any)
	out := []map[string]any{}
	for _, item := range list {
		in, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if id, _ := in["id"].(string); !validDirections[id] {
			continue
		}
		dir := map[string]any{}
		for k, v := range in {
			dir[k] = v
		}
		dir["ranges"] = repairRanges(in["ranges"], budget, productCost)
		out = append(out, dir)
	}
	if len(out) == 0 {
		return nil, errors.New("no valid directions")
	}
	return out, nil
}

// ParseShocks returns the codes list of each shock, in order. No validation:
// engine.CleanCombos drops unknown codes.
func ParseShocks(raw map[string]any) [][]string {
	list, _ := raw["shocks"].([]any)
	out := [][]string{}
	for _, item := range list {
		shock, _ := item.(map[string]any)
		codes, ok := shock["codes"].([]any)
		if !ok {
			continue
		}
		row := []string{}
		for _, c := range codes {
			if s, ok := c.(string); ok {
				row = append(row, s)
			}
		}
		out = append(out, row)
	}
	return out
}
