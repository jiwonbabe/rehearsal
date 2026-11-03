// Rehearsal server: JSON API over the engine + static pages. Start: go run ./cmd/server
//
// Endpoints (all GET unless noted, all JSON):
//
//	/api/brand           brand info + target + budget
//	/api/rates           calibrated rates
//	/api/backtest        engine vs actual, last 13 weeks
//	/api/summary         the block the planner prompt gets
//	/api/directions      Spark directions for the current month, sorted by rank
//	/api/setup   POST    {budget_usd_per_month, goal_orders_per_month, target, market}
//	/api/stress  POST    {direction_ids:[...]} -> plans, matrix, verdict, flip, check
//	/api/review  POST    {plan} (optional) -> committed plan + review + new rates
//	/api/review/apply POST  adopt the reviewed rates, move to month 2
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
)

type server struct {
	brand            *engine.Brand
	baseRates, rates engine.Rates
	month            int
	lastReview       *engine.Review
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
	rates := engine.Calibrate(brand)
	return &server{brand: brand, baseRates: rates, rates: rates, month: 1}, nil
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
	mux.HandleFunc("/api/directions", s.locked(s.handleDirections))
	mux.HandleFunc("POST /api/setup", s.locked(s.handleSetup))
	mux.HandleFunc("POST /api/stress", s.locked(s.handleStress))
	mux.HandleFunc("POST /api/review", s.locked(s.handleReview))
	mux.HandleFunc("POST /api/review/apply", s.locked(s.handleReviewApply))
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
