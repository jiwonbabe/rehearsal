// Rehearsal server: JSON API over the engine + static pages. Start: go run ./cmd/server
//
// Endpoints (all GET unless noted, all JSON):
//
//	/api/brand           brand info + target + budget
//	/api/rates           calibrated rates
//	/api/backtest        engine vs actual, last 13 weeks (in-sample)
//	/api/trust           holdout backtest (calibrated on the first 39 weeks), in-sample backtest, generator-truth recovery
//	/api/summary         the block the planner prompt gets
//	/api/directions      Spark directions for the current month, sorted by rank
//	/api/setup   POST    {budget_usd_per_month, goal_orders_per_month, target, market}
//	/api/profile POST    {source:"screenshot", handle, images:[data URLs]} -> labels read from the profile
//	/api/stress  POST    {direction_ids:[...]} -> plans, matrix, verdict, flip, check
//	/api/review  POST    {plan} (optional) -> committed plan + review + new rates
//	/api/review/apply POST  adopt the reviewed rates, move to month 2
//	/api/launch  POST    {plan} -> spec (engine), playbook (Spark text), paused Meta ads request, ads_connected
//	/api/launch/ads POST {plan} -> creates the PAUSED campaign + ad set when META_AD_ACCOUNT_ID/META_ACCESS_TOKEN are set; dry-run otherwise
//	/api/launch/calendar POST {plan} -> .ics with one event per week
//	/api/launch/creator POST {handle, images, direction} -> fit labels for one creator (cached by handle)
//	/api/meta/connect POST {source:"mock"|"instagram", handle} -> evidence from the account's Meta data + channel verdicts
//	/api/evidence        evidence, sourced statistics, channel verdicts (409 before connect)
//	/api/i18n/ko         every Spark sentence paired with its Korean translation (spark/cache/ko/)
//
// Pages are served from web/ (copy layouts from mockup/).
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"rehearsal/engine"
	"rehearsal/spark"
)

type server struct {
	brand            *engine.Brand
	baseRates, rates engine.Rates
	month            int
	lastReview       *engine.Review
	meta             *engine.MetaData
	evidence         *engine.Evidence
	metaHandle       string
	seasonNote       string
	reachLive        map[string]any
	stats            []engine.StatRow
	mu               sync.Mutex
}

func (s *server) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func newServer(brandPath string) (*server, error) {
	brand, err := engine.LoadBrand(brandPath)
	if err != nil {
		return nil, err
	}
	brand.Season, _ = engine.LoadSeasonality("data/seasonality.json") // optional: no table, no seasonal prior
	rates := engine.Calibrate(brand)
	stats, _ := engine.LoadStats("data/platform_stats.json") // optional: no table, no public-data lines
	return &server{brand: brand, baseRates: rates, rates: rates, month: 1, stats: stats}, nil
}

func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/brand", s.locked(func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, s.brand.Brand) }))
	mux.HandleFunc("/api/rates", s.locked(func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, s.rates) }))
	mux.HandleFunc("/api/backtest", s.locked(func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, engine.RunBacktest(s.brand, s.rates, 13))
	}))
	mux.HandleFunc("/api/summary", s.locked(func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, engine.Summary(s.brand, s.rates, engine.RunBacktest(s.brand, s.rates, 13)))
	}))
	mux.HandleFunc("/api/trust", s.locked(func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, map[string]any{"holdout": engine.RunHoldout(s.brand, 13), "in_sample": engine.RunBacktest(s.brand, s.rates, 13),
			"truth": engine.RunTruthCheck(s.brand, s.rates), "synthetic": s.brand.Brand.Synthetic})
	}))
	mux.HandleFunc("/api/goals", s.locked(func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, engine.RunGoalSuggestions(s.brand, 13))
	}))
	mux.HandleFunc("/api/directions", s.locked(s.handleDirections))
	mux.HandleFunc("POST /api/setup", s.locked(s.handleSetup))
	mux.HandleFunc("POST /api/profile", s.locked(s.handleProfile))
	mux.HandleFunc("POST /api/stress", s.locked(s.handleStress))
	mux.HandleFunc("POST /api/review", s.locked(s.handleReview))
	mux.HandleFunc("POST /api/review/apply", s.locked(s.handleReviewApply))
	mux.HandleFunc("POST /api/launch", s.locked(s.handleLaunch))
	mux.HandleFunc("POST /api/launch/ads", s.locked(s.handleLaunchAds))
	mux.HandleFunc("POST /api/launch/calendar", s.locked(s.handleLaunchCalendar))
	mux.HandleFunc("POST /api/launch/creator", s.locked(s.handleCreatorFit))
	mux.HandleFunc("POST /api/meta/connect", s.locked(s.handleMetaConnect))
	mux.HandleFunc("/api/evidence", s.locked(s.handleEvidence))
	mux.HandleFunc("/api/i18n/ko", s.locked(func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, spark.KoreanMap()) }))
	mux.HandleFunc("/setup.html", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusFound) })
	mux.Handle("/", http.FileServer(http.Dir("web")))
	return mux
}

// locked serializes every handler that reads or writes server state.
func (s *server) locked(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		h(w, r)
	}
}

func main() {
	os.Unsetenv("MUSE_API_KEY") // demo is cache-only; cmd/sparkcache is the only live caller
	s, err := newServer("data/brand.json")
	if err != nil {
		log.Fatalf("load brand: %v (run `python data/synthetic.py`)", err)
	}
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("Rehearsal on http://localhost%s  (brand: %s, synthetic=%v)", addr, s.brand.Brand.Name, s.brand.Brand.Synthetic)
	log.Fatal(http.ListenAndServe(addr, s.routes()))
}
