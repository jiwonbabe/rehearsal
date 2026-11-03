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
	raw, err := spark.Directions(month, summary, block)
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
	playbookStep(append(append([]map[string]any{}, m1...), m2...), brand)
	if failed {
		os.Exit(1)
	}
}

// checkLine runs the flip search between the best pay_for_reach and send_to_creators plans, then call 3.
func checkLine(dirs []map[string]any, rates engine.Rates, brand *engine.Brand) {
	var plans []engine.Plan
	titles := map[string]string{}
	for _, m := range dirs {
		d := toDirection(m)
		titles[d.ID] = d.Title
		if d.ID == "pay_for_reach" || d.ID == "send_to_creators" {
			plans = append(plans, engine.SampleVariants(d, 60, 7)...)
		}
	}
	st := engine.Stress(plans, engine.FallbackCombos, rates, brand)
	best := map[string]engine.Plan{}
	for _, v := range engine.RunVerdict(plans, st, rates, brand) {
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

// profileStep reads the demo screenshots in web/demo-posts/ into a profile read for bnd.jewelry.
func profileStep() {
	name := "profile-bnd.jewelry"
	files, _ := filepath.Glob("web/demo-posts/*")
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
func playbookStep(dirs []map[string]any, brand *engine.Brand) {
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
		raw, err := spark.Playbook(d, profile, basics)
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
