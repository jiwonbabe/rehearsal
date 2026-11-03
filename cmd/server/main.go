// Rehearsal server: JSON API over the engine + static pages. Start: go run ./cmd/server
//
// Endpoints (all GET unless noted, all JSON):
//
//	/api/brand           brand info + target + budget
//	/api/rates           calibrated rates
//	/api/backtest        engine vs actual, last 13 weeks
//	/api/summary         the block the planner prompt gets
//	/api/directions      cached Spark directions (see spark/)         [TODO step 1]
//	/api/stress  POST    {direction_ids:[...]} -> plans, matrix, verdict, flip   [TODO step 2]
//	/api/review  POST    {plan} -> review + new rates                   [TODO step 3]
//
// Pages are served from web/ (copy layouts from mockup/).
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"rehearsal/engine"
)

type server struct {
	brand *engine.Brand
	rates engine.Rates
}

func (s *server) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func main() {
	brand, err := engine.LoadBrand("data/brand.json")
	if err != nil {
		log.Fatalf("load brand: %v (run `python data/synthetic.py`)", err)
	}
	s := &server{brand: brand, rates: engine.Calibrate(brand)}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/brand", func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, brand.Brand) })
	mux.HandleFunc("/api/rates", func(w http.ResponseWriter, r *http.Request) { s.writeJSON(w, s.rates) })
	mux.HandleFunc("/api/backtest", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, engine.RunBacktest(brand, s.rates, 13))
	})
	mux.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, engine.Summary(brand, s.rates, engine.RunBacktest(brand, s.rates, 13)))
	})
	mux.Handle("/", http.FileServer(http.Dir("web")))

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("Rehearsal on http://localhost%s  (brand: %s, synthetic=%v)", addr, brand.Brand.Name, brand.Brand.Synthetic)
	log.Fatal(http.ListenAndServe(addr, mux))
}
