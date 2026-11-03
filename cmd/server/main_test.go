package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rehearsal/engine"
	"rehearsal/spark"
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
	thr := fmt.Sprintf("%d%%", int(math.Round(out.Flip.Threshold*100)))
	cur := fmt.Sprintf("%d%%", int(math.Round(out.Flip.Current*100)))
	if !strings.Contains(out.Check, thr) || !strings.Contains(out.Check, cur) {
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

// firstDirectionID: a direction id that exists in the current month, so tests do not depend on cache content.
func firstDirectionID(t *testing.T, ts *httptest.Server) string {
	t.Helper()
	var out struct {
		Directions []struct {
			ID string `json:"id"`
		} `json:"directions"`
	}
	if code := call(t, ts, "GET", "/api/directions", "", &out); code != 200 || len(out.Directions) == 0 {
		t.Fatalf("directions: status %d, %d dirs", code, len(out.Directions))
	}
	return out.Directions[0].ID
}

func TestStressFallsBackWithoutAdversaryCache(t *testing.T) {
	ts := newTestServer(t)
	t.Setenv("MUSE_API_KEY", "")
	id := firstDirectionID(t, ts)
	dir := t.TempDir()
	files, _ := filepath.Glob(filepath.Join(spark.CacheDir, "*.json"))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.Base(f)), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old := spark.CacheDir
	spark.CacheDir = dir
	t.Cleanup(func() { spark.CacheDir = old })
	if err := os.Remove(filepath.Join(dir, "adversary-"+id+".json")); err != nil {
		t.Fatal(err)
	}
	var out stressResponse
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":["`+id+`"]}`, &out); code != 200 {
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
	id := firstDirectionID(t, ts)
	var out stressResponse
	if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":["`+id+`"]}`, &out); code != 200 {
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
	for _, path := range []string{"/api/review/apply", "/api/launch", "/api/launch/ads", "/api/launch/calendar"} {
		if code := call(t, ts, "GET", path, "", nil); code == 200 {
			t.Errorf("GET %s: %d", path, code)
		}
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

func TestBaselineIndependentOfClickOrder(t *testing.T) {
	ts := newTestServer(t)
	base := func(ids string) int {
		var out stressResponse
		if code := call(t, ts, "POST", "/api/stress", `{"direction_ids":`+ids+`}`, &out); code != 200 {
			t.Fatalf("status %d", code)
		}
		for _, v := range out.Verdict {
			if v.Plan.Direction == "baseline" || v.Plan.PlanID == "baseline" {
				return v.Survived
			}
		}
		t.Fatal("no baseline verdict")
		return -1
	}
	if a, b := base(`["pay_for_reach","send_to_creators"]`), base(`["send_to_creators","pay_for_reach"]`); a != b {
		t.Errorf("baseline survived %d vs %d", a, b)
	}
}

func TestProfileBadRequests(t *testing.T) {
	ts := newTestServer(t)
	img := `"images":["data:image/png;base64,AA=="]`
	for name, body := range map[string]string{
		"bad source":   `{"source":"instagram","handle":"bnd.jewelry",` + img + `}`,
		"empty handle": `{"source":"screenshot","handle":"@ !",` + img + `}`,
		"no images":    `{"source":"screenshot","handle":"bnd.jewelry","images":[]}`,
		"not an image": `{"source":"screenshot","handle":"bnd.jewelry","images":["http://x/y.png"]}`,
	} {
		if code := call(t, ts, "POST", "/api/profile", body, nil); code != 400 {
			t.Errorf("%s: status %d, want 400", name, code)
		}
	}
}

func TestProfileCached(t *testing.T) {
	ts := newTestServer(t)
	var out struct {
		Market       string            `json:"market"`
		Customer     map[string]string `json:"likely_customer"`
		BestPostType string            `json:"best_post_type"`
		Confidence   string            `json:"confidence"`
	}
	body := `{"source":"screenshot","handle":"bnd.jewelry","images":["data:image/png;base64,AA=="]}`
	if code := call(t, ts, "POST", "/api/profile", body, &out); code != 200 {
		t.Fatalf("status %d", code)
	}
	in := func(v string, list ...string) bool {
		for _, l := range list {
			if v == l {
				return true
			}
		}
		return false
	}
	if !in(out.Market, "US", "KR", "other") || !in(out.Customer["gender"], "women", "men", "everyone") ||
		!in(out.Customer["age"], "18-24", "25-34", "35-44") || !in(out.Customer["price_band"], "30-80", "80-200", "200+") ||
		!in(out.BestPostType, "on_hand", "close_up", "flat_lay", "on_neck") || !in(out.Confidence, "low", "medium", "high") {
		t.Errorf("labels not in allowed lists: %+v", out)
	}
}

func TestProfileNoCacheNoKey(t *testing.T) {
	t.Setenv("MUSE_API_KEY", "")
	ts := newTestServer(t)
	body := `{"source":"screenshot","handle":"nobody.here","images":["data:image/png;base64,AA=="]}`
	if code := call(t, ts, "POST", "/api/profile", body, nil); code != 503 {
		t.Errorf("status %d, want 503", code)
	}
	if _, err := os.Stat("spark/cache/profile-nobody.here.json"); err == nil {
		t.Error("cache file should not be created")
	}
}

func TestGoalsEndpoint(t *testing.T) {
	ts := newTestServer(t)
	var g engine.GoalSuggestions
	if code := call(t, ts, "GET", "/api/goals", "", &g); code != 200 {
		t.Fatalf("status %d", code)
	}
	if g.Hold <= 0 || g.Stretch < g.Hold || g.Grow < g.Hold || g.WeeksUsed != 13 {
		t.Errorf("goals: %+v", g)
	}
}

func launchPlan(t *testing.T, ts *httptest.Server) map[string]any {
	t.Helper()
	var st struct {
		Verdict []struct {
			Plan map[string]any `json:"plan"`
		} `json:"verdict"`
	}
	call(t, ts, "POST", "/api/stress", `{"direction_ids":["pay_for_reach"]}`, &st)
	for _, v := range st.Verdict {
		if v.Plan["direction"] == "pay_for_reach" {
			return v.Plan
		}
	}
	t.Fatal("no plan")
	return nil
}

func launchBody(t *testing.T, ts *httptest.Server) string {
	body, _ := json.Marshal(map[string]any{"plan": launchPlan(t, ts)})
	return string(body)
}

func TestDirectionsCarryPlaybook(t *testing.T) {
	ts := newTestServer(t)
	var out struct {
		Directions []struct {
			ID       string         `json:"id"`
			Playbook map[string]any `json:"playbook"`
		} `json:"directions"`
	}
	call(t, ts, "GET", "/api/directions", "", &out)
	for _, d := range out.Directions {
		if d.Playbook == nil {
			t.Errorf("%s: no playbook", d.ID)
			continue
		}
		if weeks, _ := d.Playbook["weeks"].([]any); len(weeks) != 4 {
			t.Errorf("%s: weeks %v", d.ID, d.Playbook["weeks"])
		}
	}
}

func TestLaunchSpecMatchesEngine(t *testing.T) {
	ts := newTestServer(t)
	plan := launchPlan(t, ts)
	body, _ := json.Marshal(map[string]any{"plan": plan})
	var out struct {
		Spec         engine.LaunchSpec         `json:"spec"`
		Playbook     map[string]any            `json:"playbook"`
		AdRequest    map[string]map[string]any `json:"ad_request"`
		AdsConnected bool                      `json:"ads_connected"`
	}
	if code := call(t, ts, "POST", "/api/launch", string(body), &out); code != 200 {
		t.Fatal(code)
	}
	var p engine.Plan
	via(plan, &p)
	s, _ := newServer("data/brand.json")
	want := engine.RunLaunch(p, s.brand)
	if out.Spec.StartDate != "2026-11-02" || out.Spec.TotalSpendUsd != want.TotalSpendUsd {
		t.Errorf("spec %+v", out.Spec)
	}
	if out.AdsConnected {
		t.Error("ads_connected without env")
	}
	if out.Playbook == nil {
		t.Error("no playbook")
	}
	if p.AdsUsd > 0 {
		if out.AdRequest["adset"]["daily_budget"] != float64(out.Spec.AdsDailyCents) {
			t.Errorf("adset %v", out.AdRequest["adset"])
		}
		if out.AdRequest["campaign"]["status"] != "PAUSED" {
			t.Error("campaign not paused")
		}
	}
}

func TestLaunchJustPost(t *testing.T) {
	ts := newTestServer(t)
	body := `{"plan":{"plan_id":"x","direction":"just_post","posts_per_week":4,"post_mix":{"on_hand":0.5,"close_up":0.3,"flat_lay":0.2,"on_neck":0},"creator_count":0,"ads_usd":0,"promo_week":null,"promo_discount":0}}`
	var out struct {
		Spec      engine.LaunchSpec `json:"spec"`
		AdRequest map[string]any    `json:"ad_request"`
	}
	if code := call(t, ts, "POST", "/api/launch", body, &out); code != 200 {
		t.Fatal(code)
	}
	if out.AdRequest != nil || len(out.Spec.Calendar) != 4 || out.Spec.TotalSpendUsd != 0 {
		t.Errorf("%+v %v", out.Spec, out.AdRequest)
	}
	if code := call(t, ts, "POST", "/api/launch", `{"plan":{}}`, nil); code != 400 {
		t.Errorf("empty plan: %d", code)
	}
}

func TestLaunchAdsDryRun(t *testing.T) {
	t.Setenv("META_ACCESS_TOKEN", "")
	t.Setenv("META_AD_ACCOUNT_ID", "")
	ts := newTestServer(t)
	var out struct {
		Sent    bool                      `json:"sent"`
		Reason  string                    `json:"reason"`
		Request map[string]map[string]any `json:"request"`
	}
	if code := call(t, ts, "POST", "/api/launch/ads", launchBody(t, ts), &out); code != 200 {
		t.Fatal(code)
	}
	if out.Sent || out.Reason != "no_credentials" || out.Request["adset"]["status"] != "PAUSED" {
		t.Errorf("%+v", out)
	}
}

func TestLaunchAdsLive(t *testing.T) {
	var paths []string
	graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		paths = append(paths, r.URL.Path)
		if r.Form.Get("status") != "PAUSED" || r.Form.Get("access_token") != "tok" {
			w.WriteHeader(400)
			w.Write([]byte(`{"error":{"message":"bad"}}`))
			return
		}
		fmt.Fprintf(w, `{"id":"%d"}`, len(paths))
	}))
	defer graph.Close()
	t.Setenv("META_GRAPH_URL", graph.URL)
	t.Setenv("META_ACCESS_TOKEN", "tok")
	t.Setenv("META_AD_ACCOUNT_ID", "123")
	ts := newTestServer(t)
	var out struct {
		Sent       bool   `json:"sent"`
		CampaignID string `json:"campaign_id"`
		AdsetID    string `json:"adset_id"`
		URL        string `json:"url"`
	}
	if code := call(t, ts, "POST", "/api/launch/ads", launchBody(t, ts), &out); code != 200 {
		t.Fatal(code)
	}
	if !out.Sent || out.CampaignID != "1" || out.AdsetID != "2" || !strings.Contains(out.URL, "act=123") {
		t.Errorf("%+v", out)
	}
	if len(paths) != 2 || paths[0] != "/act_123/campaigns" || paths[1] != "/act_123/adsets" {
		t.Errorf("paths %v", paths)
	}
}

func TestLaunchAdsLiveError(t *testing.T) {
	graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error":{"message":"Invalid token"}}`))
	}))
	defer graph.Close()
	t.Setenv("META_GRAPH_URL", graph.URL)
	t.Setenv("META_ACCESS_TOKEN", "tok")
	t.Setenv("META_AD_ACCOUNT_ID", "123")
	ts := newTestServer(t)
	resp, err := http.Post(ts.URL+"/api/launch/ads", "application/json", strings.NewReader(launchBody(t, ts)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out struct {
		Error string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode != 502 || !strings.Contains(out.Error, "Invalid token") {
		t.Errorf("%d %+v", resp.StatusCode, out)
	}
}

func TestLaunchCalendar(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Post(ts.URL+"/api/launch/calendar", "application/json", strings.NewReader(launchBody(t, ts)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	ics := string(raw)
	if resp.Header.Get("Content-Type") != "text/calendar; charset=utf-8" || strings.Count(ics, "BEGIN:VEVENT") < 4 ||
		!strings.Contains(ics, "DTSTART;VALUE=DATE:20261102") || !strings.HasSuffix(ics, "END:VCALENDAR\r\n") {
		t.Errorf("%s", ics)
	}
}
