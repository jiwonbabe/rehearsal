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
	"path/filepath"
	"strings"
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
	imported         []engine.CreatorStat // creator list the owner pasted (a dataset export); becomes the creator pool
	seasonNote       string
	catalog          *engine.SeasonCatalog
	goalMonth        int // month the goal was last sized for via "Use N for <month>"; 0 = owner's own number
	reachLive        map[string]any
	stats            []engine.StatRow
	brandKey         string // jewelry | apparel | beauty
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

// demoBrands: the three synthetic demo brands the header can switch between, in display order.
var demoBrands = []struct{ Key, Path, Cache string }{
	{"jewelry", "data/brand.json", "spark/cache"},
	{"apparel", "data/brands/apparel.json", "spark/cache/apparel"},
	{"beauty", "data/brands/beauty.json", "spark/cache/beauty"},
}

func brandKeyOf(path string) string {
	for _, b := range demoBrands {
		if filepath.Clean(path) == b.Path {
			return b.Key
		}
	}
	return strings.TrimSuffix(filepath.Base(path), ".json")
}

// reset: become a fresh server for another brand (month 1, nothing connected, rates re-learned from its history).
func (s *server) reset(fresh *server) {
	s.brand, s.baseRates, s.rates, s.month = fresh.brand, fresh.baseRates, fresh.rates, 1
	s.lastReview, s.meta, s.evidence, s.metaHandle, s.imported = nil, nil, nil, "", nil
	s.seasonNote, s.catalog, s.goalMonth, s.reachLive, s.stats, s.brandKey = "", fresh.catalog, 0, nil, fresh.stats, fresh.brandKey
}

func (s *server) handleBrands(w http.ResponseWriter, r *http.Request) {
	out := []map[string]any{}
	for _, b := range demoBrands {
		if bb, err := engine.LoadBrand(b.Path); err == nil {
			out = append(out, map[string]any{"key": b.Key, "name": bb.Brand.Name, "category": bb.Brand.Category, "handle": bb.Brand.Handle, "current": b.Key == s.brandKey})
		}
	}
	s.writeJSON(w, out)
}

func (s *server) handleBrandSwitch(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Brand string `json:"brand"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	for _, b := range demoBrands {
		if b.Key != in.Brand {
			continue
		}
		fresh, err := newServer(b.Path)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		spark.CacheDir = b.Cache
		s.reset(fresh)
		s.writeJSON(w, s.brand.Brand)
		return
	}
	writeError(w, 400, "unknown brand")
}

func newServer(brandPath string) (*server, error) {
	brand, err := engine.LoadBrand(brandPath)
	if err != nil {
		return nil, err
	}
	catalog, _ := engine.LoadSeasonCatalog("data/seasonality.json") // optional: no table, no seasonal prior
	brand.Season = catalog.For(brand.Brand.Category)
	rates := engine.Calibrate(brand)
	stats, _ := engine.LoadStats("data/platform_stats.json") // optional: no table, no public-data lines
	return &server{brand: brand, baseRates: rates, rates: rates, month: 1, stats: stats, catalog: catalog, brandKey: brandKeyOf(brandPath)}, nil
}

func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/brand", s.locked(func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, s.brand.Brand) }))
	mux.HandleFunc("/api/brands", s.locked(s.handleBrands))
	mux.HandleFunc("POST /api/brand/switch", s.locked(s.handleBrandSwitch))
	mux.HandleFunc("/api/categories", s.locked(func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, s.categories()) }))
	mux.HandleFunc("/api/rates", s.locked(func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, s.rates) }))
	mux.HandleFunc("/api/backtest", s.locked(func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, engine.RunBacktest(s.brand, s.rates, 13))
	}))
	mux.HandleFunc("/api/summary", s.locked(func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, engine.Summary(s.brand, s.rates, engine.RunBacktest(s.brand, s.rates, 13)))
	}))
	mux.HandleFunc("/api/trust", s.locked(func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, map[string]any{"holdout": engine.RunHoldout(s.brand, 13), "in_sample": engine.RunBacktest(s.brand, s.rates, 13), "coverage": engine.LeverCoverage(s.brand),
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
	mux.HandleFunc("POST /api/month/next", s.locked(s.handleNextMonth))
	mux.HandleFunc("POST /api/launch", s.locked(s.handleLaunch))
	mux.HandleFunc("POST /api/launch/ads", s.locked(s.handleLaunchAds))
	mux.HandleFunc("POST /api/launch/calendar", s.locked(s.handleLaunchCalendar))
	mux.HandleFunc("POST /api/launch/creator", s.locked(s.handleCreatorFit))
	mux.HandleFunc("POST /api/meta/connect", s.locked(s.handleMetaConnect))
	mux.HandleFunc("POST /api/creators", s.locked(s.handleImportCreators))
	mux.HandleFunc("/api/evidence", s.locked(s.handleEvidence))
	mux.HandleFunc("/api/i18n/ko", s.locked(func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, spark.KoreanMap()) }))
	mux.HandleFunc("/setup.html", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusFound) })
	// Static files revalidate on every load (Last-Modified still gives 304s): a browser that kept an old app.js
	// next to a new page showed an empty screen.
	static := http.FileServer(http.Dir("web"))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		static.ServeHTTP(w, r)
	}))
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
	brandPath := "data/brand.json"
	if p := os.Getenv("BRAND"); p != "" { // another synthetic brand (data/brands/<key>.json) with its own Spark cache (spark/cache/<key>/)
		brandPath = p
		spark.CacheDir = filepath.Join("spark", "cache", strings.TrimSuffix(filepath.Base(p), ".json"))
	}
	s, err := newServer(brandPath)
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
