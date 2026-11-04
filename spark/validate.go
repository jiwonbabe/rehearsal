package spark

import (
	"errors"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var validDirections = map[string]bool{
	"just_post": true, "send_to_creators": true, "pay_for_reach": true, "one_week_sale": true, "mix": true,
}

// KnownDirection reports whether id is one of the catalog directions.
func KnownDirection(id string) bool { return validDirections[id] }

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
	if promoDiscount[1] > 1 { // percent, e.g. [15, 25]
		promoDiscount[0], promoDiscount[1] = promoDiscount[0]/100, promoDiscount[1]/100
	}

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

var numberToken = regexp.MustCompile(`\d+(\.\d+)?`)
var sentenceBreak = regexp.MustCompile(`[.!?]\s+\S`)

func round2(x float64) float64 { return math.Round(x*100) / 100 }

// ValidateCheck returns raw["check"] if it is one non-empty sentence whose every number is in allowed.
func ValidateCheck(raw map[string]any, allowed []float64) (string, bool) {
	text, _ := raw["check"].(string)
	text = strings.TrimSpace(text)
	if text == "" || sentenceBreak.MatchString(text) {
		return "", false
	}
	for _, token := range numberToken.FindAllString(text, -1) {
		n, _ := strconv.ParseFloat(token, 64)
		found := false
		for _, a := range allowed {
			if round2(a) == round2(n) {
				found = true
			}
		}
		if !found {
			return "", false
		}
	}
	return text, true
}

func oneOf(v any, allowed ...string) string {
	s, _ := v.(string)
	for _, a := range allowed {
		if s == a {
			return s
		}
	}
	return ""
}

func sentence(v any) string {
	text, _ := v.(string)
	r := []rune(strings.TrimSpace(text))
	if len(r) > 200 {
		r = r[:200]
	}
	return string(r)
}

// ValidateProfile returns exactly the profile keys; out-of-list labels become "".
func ValidateProfile(raw map[string]any) map[string]any {
	cust, _ := raw["likely_customer"].(map[string]any)
	seen, _ := raw["post_types_seen"].(map[string]any)
	posts := map[string]any{}
	for _, k := range []string{"on_hand", "close_up", "flat_lay", "on_neck"} {
		posts[k] = oneOf(seen[k], "none", "some", "most")
	}
	return map[string]any{
		"handle": sentence(raw["handle"]), "what_they_sell": sentence(raw["what_they_sell"]),
		"style": sentence(raw["style"]), "market": oneOf(raw["market"], "US", "KR", "other"),
		"likely_customer": map[string]any{
			"gender":     oneOf(cust["gender"], "women", "men", "everyone"),
			"age":        oneOf(cust["age"], "18-24", "25-34", "35-44"),
			"price_band": oneOf(cust["price_band"], "30-80", "80-200", "200+"),
		},
		"post_types_seen": posts,
		"best_post_type":  oneOf(raw["best_post_type"], "on_hand", "close_up", "flat_lay", "on_neck"),
		"confidence":      oneOf(raw["confidence"], "low", "medium", "high"), "why": sentence(raw["why"]),
		"success_metric": oneOf(raw["success_metric"], "orders", "revenue"),
		"ambition":       oneOf(raw["ambition"], "hold", "grow", "stretch"), "why_goal": sentence(raw["why_goal"]),
	}
}

var numberRe = regexp.MustCompile(`(?i)\$\d|\d+\s*(orders|revenue|%)`)

// noNumbers drops a playbook sentence that carries a money, order, revenue or percent figure.
func noNumbers(s string) string {
	if numberRe.MatchString(s) {
		return ""
	}
	return s
}

func clean(v any) string { return noNumbers(sentence(v)) }

// ValidatePlaybook returns exactly platform, weeks (always 4), ad, creator_brief; ad and
// creator_brief are nil unless Spark returned objects. Strings with figures are dropped.
func ValidatePlaybook(raw map[string]any) map[string]any {
	weeks := make([]map[string]any, 4)
	rawWeeks, _ := raw["weeks"].([]any)
	for i := range weeks {
		w := map[string]any{}
		if i < len(rawWeeks) {
			w, _ = rawWeeks[i].(map[string]any)
		}
		weeks[i] = map[string]any{"week": i + 1, "do": clean(w["do"]), "post_idea": clean(w["post_idea"])}
	}
	out := map[string]any{"platform": "instagram", "weeks": weeks, "ad": nil, "creator_brief": nil}
	if ad, ok := raw["ad"].(map[string]any); ok {
		placements := []string{}
		list, _ := ad["placements"].([]any)
		for _, p := range list {
			if s := oneOf(p, "instagram_feed", "instagram_reels", "instagram_stories"); s != "" {
				placements = append(placements, s)
			}
		}
		out["ad"] = map[string]any{"audience": clean(ad["audience"]), "creative": clean(ad["creative"]), "placements": placements}
	}
	if cb, ok := raw["creator_brief"].(map[string]any); ok {
		find := []string{}
		list, _ := cb["find_them"].([]any)
		for _, f := range list {
			if s := clean(f); s != "" && len(find) < 5 {
				find = append(find, s)
			}
		}
		out["creator_brief"] = map[string]any{"who": clean(cb["who"]), "find_them": find,
			"send": clean(cb["send"]), "ask": clean(cb["ask"]), "dm": clean(cb["dm"])}
	}
	return out
}

// ValidateCreatorFit returns exactly handle, fit, size_band, why, send; out-of-list labels become "".
func ValidateCreatorFit(raw map[string]any) map[string]any {
	return map[string]any{
		"handle": sentence(raw["handle"]), "fit": oneOf(raw["fit"], "high", "medium", "low"),
		"size_band": oneOf(raw["size_band"], "small", "mid", "large"),
		"why":       clean(raw["why"]), "send": clean(raw["send"]),
	}
}

var figureRe = regexp.MustCompile(`\d[\d,]*(?:\.\d+)?`)

// figures: the numbers written in s, after removing the declared labels (age bands like 25-34, 55+, 30s).
func figures(s string, labels []string) []float64 {
	for _, l := range labels {
		s = strings.ReplaceAll(s, l, " ")
	}
	var out []float64
	for _, f := range figureRe.FindAllString(s, -1) {
		if v, err := strconv.ParseFloat(strings.ReplaceAll(f, ",", ""), 64); err == nil {
			out = append(out, round2(v))
		}
	}
	return out
}

// quotesAllowed: every figure in s (declared labels aside) is one of allowed (2dp).
func quotesAllowed(s string, allowed []float64, labels []string) bool {
	ok := map[float64]bool{}
	for _, a := range allowed {
		ok[round2(a)] = true
	}
	for _, v := range figures(s, labels) {
		if !ok[v] {
			return false
		}
	}
	return true
}

// ValidateChannels keeps entries with a catalog channel and verdict; a why_data line quoting a figure
// outside allowed becomes "". why_category and how must carry no figures at all (labels aside).
func ValidateChannels(raw map[string]any, allowed []float64, labels []string) []map[string]any {
	out := []map[string]any{}
	list, _ := raw["channels"].([]any)
	for _, item := range list {
		c, _ := item.(map[string]any)
		channel := oneOf(c["channel"], "instagram_reels", "instagram_feed", "instagram_stories", "threads", "facebook", "tiktok", "x")
		verdict := oneOf(c["verdict"], "lead", "support", "skip")
		if channel == "" || verdict == "" {
			continue
		}
		why := sentence(c["why_data"])
		if !quotesAllowed(why, allowed, labels) {
			why = ""
		}
		noFigures := func(v any) string {
			t := sentence(v)
			if !quotesAllowed(t, nil, labels) {
				return ""
			}
			return t
		}
		out = append(out, map[string]any{"channel": channel, "verdict": verdict, "why_data": why,
			"why_category": noFigures(c["why_category"]), "how": noFigures(c["how"])})
		if len(out) == 6 {
			break
		}
	}
	return out
}

// ValidateTranslation returns en's shape with ko's strings where ko has a non-empty string at the same place;
// everything else (numbers, nulls, missing keys, mismatched lists) stays as in en.
func ValidateTranslation(en, ko any) any {
	switch e := en.(type) {
	case map[string]any:
		k, _ := ko.(map[string]any)
		out := map[string]any{}
		for key, v := range e {
			out[key] = ValidateTranslation(v, k[key])
		}
		return out
	case []any:
		k, _ := ko.([]any)
		out := make([]any, len(e))
		for i, v := range e {
			if len(k) == len(e) {
				out[i] = ValidateTranslation(v, k[i])
			} else {
				out[i] = ValidateTranslation(v, nil)
			}
		}
		return out
	case string:
		if s, ok := ko.(string); ok && strings.TrimSpace(s) != "" && !bannedRe.MatchString(s) && sameFigures(e, s) {
			return strings.TrimSpace(s)
		}
		return e
	}
	return en
}

var bannedRe = regexp.MustCompile(`(?i)forecast|prediction|예측`)

var koDateRe = regexp.MustCompile(`\d{1,2}(월|일|주차|시)`) // Korean writes month and day names as numerals (December → 12월)

// sameFigures: the translation carries exactly the numbers the English had (as a multiset); Korean date numerals are not figures.
func sameFigures(en, ko string) bool {
	a, b := figures(en, nil), figures(koDateRe.ReplaceAllString(ko, " "), nil)
	if len(a) != len(b) {
		return false
	}
	sort.Float64s(a)
	sort.Float64s(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// FlattenStrings records every en string whose translated counterpart differs, as into[en] = ko.
func FlattenStrings(en, ko any, into map[string]string) {
	switch e := en.(type) {
	case map[string]any:
		k, _ := ko.(map[string]any)
		for key, v := range e {
			FlattenStrings(v, k[key], into)
		}
	case []any:
		k, _ := ko.([]any)
		for i, v := range e {
			if i < len(k) {
				FlattenStrings(v, k[i], into)
			}
		}
	case string:
		if s, ok := ko.(string); ok && s != e && e != "" {
			into[e] = s
		}
	}
}

// SeasonNote: the planner's one-line reading of the month; "" unless every figure it quotes is allowed.
func SeasonNote(raw map[string]any, allowed []float64) string {
	note := sentence(raw["season_note"])
	if !quotesAllowed(note, allowed, nil) {
		return ""
	}
	return note
}
