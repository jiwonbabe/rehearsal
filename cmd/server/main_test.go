package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"rehearsal/engine"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s, err := newServer("data/brand.json")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.routes())
	t.Cleanup(ts.Close)
	return ts
}

func call(t *testing.T, ts *httptest.Server, method, path, body string, into any) int {
	t.Helper()
	req, _ := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if into != nil && resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
			t.Fatal(err)
		}
	}
	return resp.StatusCode
}

func TestDirectionsMonthOne(t *testing.T) {
	ts := newTestServer(t)
	var out struct {
		Month      int              `json:"month"`
		Directions []map[string]any `json:"directions"`
		LastMonth  *json.RawMessage `json:"last_month"`
	}
	if code := call(t, ts, "GET", "/api/directions", "", &out); code != 200 {
		t.Fatalf("status %d", code)
	}
	if out.Month != 1 || len(out.Directions) < 3 || out.LastMonth != nil {
		t.Errorf("month %d, %d dirs, last_month %v", out.Month, len(out.Directions), out.LastMonth)
	}
	for i := 1; i < len(out.Directions); i++ {
		if out.Directions[i]["rank"].(float64) < out.Directions[i-1]["rank"].(float64) {
			t.Errorf("not sorted by rank")
		}
	}
}

func TestStressTwoDirections(t *testing.T) {
	ts := newTestServer(t)
	var out stressResponse
	body := `{"direction_ids":["pay_for_reach","send_to_creators"]}`
	if code := call(t, ts, "POST", "/api/stress", body, &out); code != 200 {
		t.Fatalf("status %d", code)
	}
	if len(out.Directions) != 2 || len(out.Directions[0].Plans) != 60 || len(out.Directions[0].Stress.Combos) != 20 {
		t.Fatalf("shape: %d dirs", len(out.Directions))
	}
	hasBaseline := false
	for _, v := range out.Verdict {
		hasBaseline = hasBaseline || v.Plan.PlanID == "baseline"
	}
	if len(out.Verdict) != 3 || !hasBaseline {
		t.Errorf("verdict should be 2 directions + baseline, got %d", len(out.Verdict))
	}
	if out.FlipA != "pay_for_reach" || out.FlipB != "send_to_creators" {
		t.Errorf("flip ids: %q %q", out.FlipA, out.FlipB)
	}
	if out.Flip == nil || out.Flip.Param != "seed_post_rate" {
		t.Errorf("flip: %+v", out.Flip)
	}
	if len(out.Catalog) != 7 || len(out.Priors) != 7 {
		t.Errorf("catalog/priors missing")
	}
	if !strings.Contains(out.Check, "68%") || !strings.Contains(out.Check, "60%") {
		t.Errorf("check line: %q", out.Check)
	}
}

func TestStressRejectsBadIds(t *testing.T) {
	ts := newTestServer(t)
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":[]}`, nil); code != 400 {
		t.Errorf("empty: %d", code)
	}
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":["nope"]}`, nil); code != 400 {
		t.Errorf("unknown: %d", code)
	}
}

func TestStressFallsBackWithoutAdversaryCache(t *testing.T) {
	ts := newTestServer(t)
	t.Setenv("MUSE_API_KEY", "")
	if err := os.Rename("spark/cache/adversary-just_post.json", "spark/cache/adversary-just_post.json.bak"); err != nil {
		t.Fatal(err)
	}
	defer os.Rename("spark/cache/adversary-just_post.json.bak", "spark/cache/adversary-just_post.json")
	var out stressResponse
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":["just_post"]}`, &out); code != 200 {
		t.Fatalf("status %d", code)
	}
	if len(out.Directions[0].Stress.Combos) != 20 || out.Flip != nil {
		t.Errorf("fallback combos %d, flip %v", len(out.Directions[0].Stress.Combos), out.Flip)
	}
	if !reflect.DeepEqual(out.Directions[0].Stress.Combos, engine.FallbackCombos) {
		t.Errorf("combos are not engine.FallbackCombos")
	}
	for i, sc := range out.Directions[0].Shocks {
		if sc.Why != "" {
			t.Errorf("shock %d has why %q in fallback", i, sc.Why)
		}
	}
}

func TestStressAlignsWhyWithCombos(t *testing.T) {
	ts := newTestServer(t)
	var out stressResponse
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":["just_post"]}`, &out); code != 200 {
		t.Fatalf("status %d", code)
	}
	d := out.Directions[0]
	if len(d.Shocks) != len(d.Stress.Combos) {
		t.Fatalf("%d shocks, %d combos", len(d.Shocks), len(d.Stress.Combos))
	}
	for i, sc := range d.Shocks {
		if !reflect.DeepEqual(sc.Codes, d.Stress.Combos[i]) || sc.Why == "" {
			t.Errorf("shock %d: %+v vs combo %v", i, sc, d.Stress.Combos[i])
		}
	}
}

func TestMutatingEndpointsRejectGet(t *testing.T) {
	ts := newTestServer(t)
	if code := call(t, ts, "GET", "/api/review/apply", "", nil); code == 200 {
		t.Errorf("GET apply: %d", code)
	}
}

func TestReviewApplyIsIdempotent(t *testing.T) {
	ts := newTestServer(t)
	var rv struct {
		Review struct {
			InsideRange bool `json:"inside_range"`
			NewRates    struct {
				SeedPostRate float64 `json:"seed_post_rate"`
			} `json:"new_rates"`
		} `json:"review"`
	}
	if code := call(t, ts, "POST", "/api/review", "", &rv); code != 200 || rv.Review.InsideRange || rv.Review.NewRates.SeedPostRate != 0.49 {
		t.Fatalf("review: %d %+v", code, rv)
	}
	var applied struct {
		Month int `json:"month"`
		Rates struct {
			SeedPostRate float64 `json:"seed_post_rate"`
		} `json:"rates"`
	}
	call(t, ts, "POST", "/api/review/apply", "", &applied)
	call(t, ts, "POST", "/api/review/apply", "", &applied)
	if applied.Month != 2 || applied.Rates.SeedPostRate != 0.49 {
		t.Errorf("after two applies: %+v", applied)
	}
	var dirs struct {
		Month     int             `json:"month"`
		LastMonth json.RawMessage `json:"last_month"`
	}
	call(t, ts, "GET", "/api/directions", "", &dirs)
	if dirs.Month != 2 || len(dirs.LastMonth) < 10 {
		t.Errorf("directions after apply: month %d", dirs.Month)
	}
}

func TestSetupClampsAndValidates(t *testing.T) {
	ts := newTestServer(t)
	if code := call(t, ts, "POST", "/api/setup", `{"budget_usd_per_month":"abc"}`, nil); code != 400 {
		t.Errorf("non-numeric budget: %d", code)
	}
	if code := call(t, ts, "POST", "/api/setup", `{"budget_usd_per_month":100}`, nil); code != 200 {
		t.Errorf("setup: %d", code)
	}
	var out struct {
		Directions []struct {
			Ranges struct {
				AdsUsd []float64 `json:"ads_usd"`
			} `json:"ranges"`
		} `json:"directions"`
	}
	call(t, ts, "GET", "/api/directions", "", &out)
	for _, d := range out.Directions {
		if d.Ranges.AdsUsd[1] > 100 {
			t.Errorf("ads_usd not clamped to new budget: %v", d.Ranges.AdsUsd)
		}
	}
}
