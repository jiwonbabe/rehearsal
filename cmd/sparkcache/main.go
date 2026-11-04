// Command sparkcache fills spark/cache/ with Muse Spark responses for both months, every
// direction's adversary list, and the closing check line. `go run ./cmd/sparkcache` fills
// spark/cache/; run once, commit the files. Existing cache files are reused, not re-fetched.
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"rehearsal/engine"
	"rehearsal/spark"
)

var failed bool

func fail(name string, err error) {
	fmt.Printf("fail %s: %v\n", name, err)
	failed = true
}

// directionsFor calls and validates one month; nil on error.
func directionsFor(month int, summary map[string]any, block string, brand *engine.Brand) []map[string]any {
	name := fmt.Sprintf("directions-m%d", month)
	season := engine.SeasonBlock(brand.Season, engine.PlanMonth(brand, month-1))
	raw, err := spark.Directions(month, summary, block, season)
	if err != nil {
		fail(name, err)
		return nil
	}
	dirs, err := spark.ValidateDirections(raw, brand.Brand.BudgetUsdPerMonth, brand.Brand.ProductCostUsd)
	if err != nil {
		fail(name, err)
		return nil
	}
	fmt.Printf("ok %s: %d directions\n", name, len(dirs))
	return dirs
}

// toDirection converts a validated direction map into the engine type.
func toDirection(m map[string]any) engine.Direction {
	raw, _ := json.Marshal(m)
	var d engine.Direction
	json.Unmarshal(raw, &d)
	return d
}

func main() {
	brand, err := engine.LoadBrand("data/brand.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	brand.Season, _ = engine.LoadSeasonality("data/seasonality.json")
	rates := engine.Calibrate(brand)
	backtest := engine.RunBacktest(brand, rates, 13)

	// Month 1 sees the brand as it was before October, so the review has something to change.
	pre := *brand
	pre.Weeks = brand.Weeks[:len(brand.Weeks)-engine.HorizonWeeks]
	preRates := engine.Calibrate(&pre)
	m1 := directionsFor(1, engine.Summary(&pre, preRates, engine.RunBacktest(&pre, preRates, 13)), "", brand)
	rv := engine.RunReview(brand, engine.OctoberPlan(), rates)
	m2 := directionsFor(2, engine.Summary(brand, rv.NewRates, backtest), engine.LastMonthBlock(rv), brand)

	seen := map[string]bool{}
	for _, d := range append(append([]map[string]any{}, m1...), m2...) {
		id, _ := d["id"].(string)
		if seen[id] {
			continue
		}
		seen[id] = true
		raw, err := spark.Adversary(d, engine.PriorProbabilities(rates))
		if err != nil {
			fail("adversary-"+id, err)
			continue
		}
		fmt.Printf("ok adversary-%s: %d combos\n", id, len(spark.ParseShocks(raw)))
	}

	checkLine(m1, rates, brand)
	profileStep()
	channels := channelsStep(brand)
	playbookStep(append(append([]map[string]any{}, m1...), m2...), brand, channels)
	creatorStep()
	translateStep()
	if failed {
		os.Exit(1)
	}
}

// checkLine mirrors the server: each direction is stressed with its own adversary combos, the verdict picks the
// best plan per direction, and the flip between the two best plans runs on the fallback combos. Then call 3.
func checkLine(dirs []map[string]any, rates engine.Rates, brand *engine.Brand) {
	var allPlans []engine.Plan
	merged := engine.Stressed{Matrix: map[string][]engine.Cell{}, Combos: engine.FallbackCombos}
	titles := map[string]string{}
	for _, m := range dirs {
		d := toDirection(m)
		titles[d.ID] = d.Title
		if d.ID != "pay_for_reach" && d.ID != "send_to_creators" {
			continue
		}
		combos := engine.FallbackCombos
		if raw, err := spark.Adversary(m, engine.PriorProbabilities(rates)); err == nil {
			combos = spark.ParseShocks(raw)
		}
		plans := engine.SampleVariants(d, 60, 7)
		for i := range plans {
			plans[i].Month = engine.PlanMonth(brand, 0)
		}
		st := engine.Stress(plans, combos, rates, brand)
		for k, cells := range st.Matrix {
			merged.Matrix[k] = cells
		}
		allPlans = append(allPlans, plans...)
	}
	best := map[string]engine.Plan{}
	for _, v := range engine.RunVerdict(allPlans, merged, rates, brand) {
		best[v.Plan.Direction] = v.Plan
	}
	flip := engine.RunFlip(best["pay_for_reach"], best["send_to_creators"], engine.FallbackCombos, rates, brand, "seed_post_rate", 0.3, 1.0, 0.02)
	if flip == nil {
		fail("check", fmt.Errorf("no flip between pay_for_reach and send_to_creators"))
		return
	}
	fmt.Printf("flip threshold=%g current=%g\n", flip.Threshold, flip.Current)
	raw, _ := json.Marshal(flip)
	var flipMap map[string]any
	json.Unmarshal(raw, &flipMap)
	name := "check-pay_for_reach-send_to_creators"
	out, err := spark.Check("pay_for_reach", "send_to_creators", flipMap, titles["pay_for_reach"], titles["send_to_creators"])
	if err != nil {
		fail(name, err)
		return
	}
	allowed := []float64{flip.Threshold, flip.Current, flip.Threshold * 100, flip.Current * 100}
	if line, ok := spark.ValidateCheck(out, allowed); ok {
		fmt.Printf("ok %s: %s\n", name, line)
	} else {
		fail(name, fmt.Errorf("check line rejected"))
	}
}

// dataURLs reads image files matching glob into data: URLs, sorted by name.
func dataURLs(glob string) []string {
	files, _ := filepath.Glob(glob)
	sort.Strings(files)
	var images []string
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		mime := map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg"}[ext]
		raw, err := os.ReadFile(f)
		if mime == "" || err != nil {
			continue
		}
		images = append(images, "data:"+mime+";base64,"+base64.StdEncoding.EncodeToString(raw))
	}
	return images
}

// profileStep reads the demo screenshots in web/demo-posts/ into a profile read for bnd.jewelry.
func profileStep() {
	name := "profile-bnd.jewelry"
	images := dataURLs("web/demo-posts/*")
	raw, err := spark.Profile("bnd.jewelry", images)
	if err != nil {
		fail(name, err)
		return
	}
	p := spark.ValidateProfile(raw)
	c := p["likely_customer"].(map[string]any)
	fmt.Printf("ok %s: %v %v %v %v %v %v\n", name, p["market"], c["gender"], c["age"], c["price_band"], p["best_post_type"], p["confidence"])
}

// playbookStep writes the executor brief for every direction id, using the cached profile read.
func playbookStep(dirs []map[string]any, brand *engine.Brand, channels any) {
	rawProfile, err := spark.Profile("bnd.jewelry", nil)
	if err != nil {
		fail("playbook", err)
		return
	}
	profile := spark.ValidateProfile(rawProfile)
	basics := map[string]any{"name": brand.Brand.Name, "market": brand.Brand.Market,
		"target": brand.Brand.Target, "avg_price_usd": brand.Brand.AvgPriceUsd}
	seen := map[string]bool{}
	for _, d := range dirs {
		id, _ := d["id"].(string)
		if seen[id] {
			continue
		}
		seen[id] = true
		raw, err := spark.Playbook(d, profile, basics, channels)
		if err != nil {
			fail("playbook-"+id, err)
			continue
		}
		p := spark.ValidatePlaybook(raw)
		weeks := p["weeks"].([]map[string]any)
		if weeks[0]["do"] == "" {
			fail("playbook-"+id, fmt.Errorf("empty week 1"))
			continue
		}
		fmt.Printf("ok playbook-%s: week1=%q ad=%v creators=%v\n", id, weeks[0]["do"], p["ad"] != nil, p["creator_brief"] != nil)
	}
}

// creatorStep reads every creator in the mock business_discovery data, plus any web/demo-creators/<handle>.png,
// against the send_to_creators brief.
func creatorStep() {
	raw, err := spark.Playbook(map[string]any{"id": "send_to_creators"}, nil, nil, nil)
	if err != nil {
		fail("creator", err)
		return
	}
	brief, _ := spark.ValidatePlaybook(raw)["creator_brief"].(map[string]any)
	report := func(handle string, out map[string]any, err error) {
		if err != nil {
			fail("creator-"+handle, err)
			return
		}
		p := spark.ValidateCreatorFit(out)
		fmt.Printf("ok creator-%s: %v %v %q\n", handle, p["fit"], p["size_band"], p["why"])
	}
	if meta, err := engine.LoadMeta("data/meta/demo.brand"); err == nil {
		for _, c := range engine.RunEvidence(meta).Creators {
			out, err := spark.CreatorFit(c.Username, brief, nil, c)
			report(c.Username, out, err)
		}
	}
	files, _ := filepath.Glob("web/demo-creators/*")
	for _, f := range files {
		handle := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
		out, err := spark.CreatorFit(handle, brief, dataURLs(f), nil)
		report(handle, out, err)
	}
}

// channelsStep: evidence from the mock Meta data + sourced statistics -> channel verdicts for bnd.jewelry.
func channelsStep(brand *engine.Brand) any {
	meta, err := engine.LoadMeta("data/meta/demo.brand")
	if err != nil {
		fail("channels", err)
		return nil
	}
	stats, err := engine.LoadStats("data/platform_stats.json")
	if err != nil {
		fail("channels", err)
		return nil
	}
	ev := engine.RunEvidence(meta)
	rawProfile, err := spark.Profile("bnd.jewelry", nil)
	if err != nil {
		fail("channels", err)
		return nil
	}
	basics := map[string]any{"name": brand.Brand.Name, "market": brand.Brand.Market, "target": brand.Brand.Target, "avg_price_usd": brand.Brand.AvgPriceUsd}
	raw, err := spark.Channels("demo.brand", ev, spark.ValidateProfile(rawProfile), stats, basics)
	if err != nil {
		fail("channels", err)
		return nil
	}
	out := spark.ValidateChannels(raw, engine.EvidenceNumbers(ev, stats), engine.EvidenceLabels(ev, stats))
	if len(out) < 3 {
		fail("channels", fmt.Errorf("only %d valid channels", len(out)))
	}
	for _, c := range out {
		fmt.Printf("ok channels: %s=%s data=%q\n", c["channel"], c["verdict"], c["why_data"])
	}
	return out
}

// translateStep writes ko/<name>.json for every cached call that has none yet, or whose English file is newer.
func translateStep() {
	for _, name := range spark.CacheNames() {
		en, err := spark.Call(name, "")
		if err != nil {
			continue
		}
		enInfo, _ := os.Stat(filepath.Join(spark.CacheDir, name+".json"))
		koPath := filepath.Join(spark.CacheDir, "ko", name+".json")
		if koInfo, err := os.Stat(koPath); err == nil && enInfo != nil && enInfo.ModTime().After(koInfo.ModTime()) {
			os.Remove(koPath)
			fmt.Printf("stale ko/%s removed\n", name)
		}
		ko, err := spark.Translate(name, en)
		if err != nil {
			fail("ko/"+name, err)
			continue
		}
		flat := map[string]string{}
		spark.FlattenStrings(en, spark.ValidateTranslation(en, ko), flat)
		fmt.Printf("ok ko/%s: %d strings\n", name, len(flat))
	}
}
