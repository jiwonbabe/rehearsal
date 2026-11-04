package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"rehearsal/engine"
	"rehearsal/spark"
)

// playbookFor loads the cached executor brief for a direction id; nil when the cache is missing.
func playbookFor(id string) map[string]any {
	raw, err := spark.Playbook(map[string]any{"id": id}, nil, nil, nil)
	if err != nil {
		return nil
	}
	return spark.ValidatePlaybook(raw)
}

type launch struct {
	Plan     engine.Plan
	Title    string
	Spec     engine.LaunchSpec
	Playbook map[string]any
}

// launchOf reads {plan} from the request and resolves its direction title, spec and playbook.
func (s *server) launchOf(r *http.Request) (launch, error) {
	var in struct {
		Plan engine.Plan `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil && !errors.Is(err, io.EOF) {
		return launch{}, err
	}
	if in.Plan.PlanID == "" {
		return launch{}, errors.New("plan is required: commit one on Verdict first")
	}
	if !spark.KnownDirection(in.Plan.Direction) {
		return launch{}, errors.New("unknown direction: " + in.Plan.Direction)
	}
	budget, cost := s.brand.Brand.BudgetUsdPerMonth, s.brand.Brand.ProductCostUsd
	if in.Plan.AdsUsd < 0 || in.Plan.CreatorCount < 0 || in.Plan.AdsUsd+in.Plan.CreatorCount*cost > budget {
		return launch{}, fmt.Errorf("plan spends more than the $%.0f monthly budget", budget)
	}
	l := launch{Plan: in.Plan, Title: in.Plan.Direction, Spec: engine.RunLaunch(in.Plan, s.brand), Playbook: playbookFor(in.Plan.Direction)}
	if dirs, err := s.currentDirections(); err == nil {
		for _, d := range dirs {
			if d.ID == in.Plan.Direction {
				l.Title = d.Title
			}
		}
	}
	return l, nil
}

// today is the clock for the ads request only; tests replace it.
var today = time.Now

func metaEnv() (token, account string) {
	return os.Getenv("META_ACCESS_TOKEN"), strings.TrimPrefix(os.Getenv("META_AD_ACCOUNT_ID"), "act_")
}

// targetingFor: the Marketing API targeting spec for the confirmed customer (country, age band, gender).
func targetingFor(info engine.BrandInfo) (map[string]any, string) {
	country := info.Market
	if country != "US" && country != "KR" {
		country = "US"
	}
	ageMin, ageMax := 18, 65
	if parts := strings.Split(info.Target["age"], "-"); len(parts) == 2 {
		if lo, err := strconv.Atoi(parts[0]); err == nil {
			if hi, err := strconv.Atoi(parts[1]); err == nil {
				ageMin, ageMax = lo, hi
			}
		}
	}
	targeting := map[string]any{"geo_locations": map[string]any{"countries": []string{country}}, "age_min": ageMin, "age_max": ageMax}
	switch info.Target["gender"] {
	case "women":
		targeting["genders"] = []int{2}
	case "men":
		targeting["genders"] = []int{1}
	}
	return targeting, country
}

// adRequest: the Meta Marketing API objects for this plan, both PAUSED. nil when the plan has no ads.
func adRequest(spec engine.LaunchSpec, plan engine.Plan, info engine.BrandInfo, title string) (campaign, adset map[string]any) {
	if plan.AdsUsd <= 0 {
		return nil, nil
	}
	targeting, country := targetingFor(info)
	// Meta requires an explicit choice on Advantage+ audience expansion; 0 keeps the audience exactly as confirmed.
	targeting["targeting_automation"] = map[string]any{"advantage_audience": 0}
	// OUTCOME_TRAFFIC + LINK_CLICKS needs no pixel; the ad set ends at midnight after the last day so all days are bought.
	// Meta rejects a start in the past: a month already under way starts today and keeps its end; a month
	// entirely past starts today and keeps its length (the engine's dates come from the data; only this request looks at the clock).
	startDay, _ := time.Parse("2006-01-02", spec.StartDate)
	endDay, _ := time.Parse("2006-01-02", spec.EndDate)
	if now := today().UTC().Truncate(24 * time.Hour); startDay.Before(now) {
		if endDay.Before(now) {
			endDay = now.AddDate(0, 0, spec.Days-1)
		}
		startDay = now
	}
	// No campaign-level budget (the ad set carries it), so Meta requires is_adset_budget_sharing_enabled to be stated.
	campaign = map[string]any{"name": "Rehearsal · " + title + " · " + spec.StartDate, "objective": "OUTCOME_TRAFFIC",
		"status": "PAUSED", "special_ad_categories": []string{}, "is_adset_budget_sharing_enabled": false}
	adset = map[string]any{"name": fmt.Sprintf("%s · %s %s %s", title, country, info.Target["gender"], info.Target["age"]),
		"daily_budget": spec.AdsDailyCents, "billing_event": "IMPRESSIONS", "optimization_goal": "LINK_CLICKS",
		"bid_strategy": "LOWEST_COST_WITHOUT_CAP", "targeting": targeting,
		"start_time": startDay.Format("2006-01-02") + "T00:00:00+0000", "end_time": endDay.AddDate(0, 0, 1).Format("2006-01-02") + "T00:00:00+0000", "status": "PAUSED"}
	return campaign, adset
}

func formOf(m map[string]any) url.Values {
	form := url.Values{}
	for k, v := range m {
		switch x := v.(type) {
		case string:
			form.Set(k, x)
		case int, float64, bool:
			form.Set(k, fmt.Sprint(x))
		default:
			raw, _ := json.Marshal(x)
			form.Set(k, string(raw))
		}
	}
	return form
}

// graphPost creates one Marketing API object and returns its id; a non-2xx reply is an error carrying the body.
func graphPost(base, path string, form url.Values, token string) (string, error) {
	form.Set("access_token", token)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.PostForm(strings.TrimRight(base, "/")+path, form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("meta %s: %s", resp.Status, string(raw))
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.ID == "" {
		return "", fmt.Errorf("meta: unexpected reply: %s", string(raw))
	}
	return out.ID, nil
}

func (s *server) handleLaunch(w http.ResponseWriter, r *http.Request) {
	l, err := s.launchOf(r)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	token, account := metaEnv()
	reach, reachSource := s.reach()
	campaign, adset := adRequest(l.Spec, l.Plan, s.brand.Brand, l.Title)
	var adReq map[string]any
	if campaign != nil {
		adReq = map[string]any{"campaign": campaign, "adset": adset}
	}
	s.writeJSON(w, map[string]any{
		"direction":          map[string]string{"id": l.Plan.Direction, "title": l.Title},
		"spec":               l.Spec,
		"playbook":           l.Playbook,
		"ad_request":         adReq,
		"ads_connected":      token != "" && account != "",
		"current_month":      s.planMonth(),
		"current_month_name": engine.SeasonFor(nil, s.planMonth()).Name,
		"reach":              reach,
		"reach_source":       reachSource,
		"creators":           s.suggestedCreators(),
		"creator_math":       engine.CreatorMath(l.Plan.CreatorCount, s.rates, s.brand, l.Plan.Month),
		"creator_bands":      engine.CreatorBands(s.rates, s.brand, l.Plan.Month),
		"creators_source":    s.creatorSource(),
	})
}

func (s *server) handleLaunchAds(w http.ResponseWriter, r *http.Request) {
	l, err := s.launchOf(r)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	campaign, adset := adRequest(l.Spec, l.Plan, s.brand.Brand, l.Title)
	if campaign == nil {
		writeError(w, 400, "this plan has no ads")
		return
	}
	request := map[string]any{"campaign": campaign, "adset": adset}
	token, account := metaEnv()
	if token == "" || account == "" {
		s.writeJSON(w, map[string]any{"sent": false, "reason": "no_credentials", "request": request})
		return
	}
	base := os.Getenv("META_GRAPH_URL")
	if base == "" {
		base = "https://graph.facebook.com/v21.0"
	}
	campaignID, err := graphPost(base, "/act_"+account+"/campaigns", formOf(campaign), token)
	if err != nil {
		writeError(w, 502, err.Error())
		return
	}
	adset["campaign_id"] = campaignID
	adsetID, err := graphPost(base, "/act_"+account+"/adsets", formOf(adset), token)
	if err != nil {
		// The paused campaign exists; return its id so the owner can find or delete it instead of making another.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(502)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error(), "campaign_id": campaignID})
		return
	}
	s.writeJSON(w, map[string]any{"sent": true, "campaign_id": campaignID, "adset_id": adsetID,
		"url": "https://adsmanager.facebook.com/adsmanager/manage/campaigns?act=" + account})
}

func postsText(posts map[string]int) string {
	total := 0
	var parts []string
	for _, t := range engine.PostTypes {
		if n := posts[t]; n > 0 {
			total += n
			parts = append(parts, fmt.Sprintf("%d %s", n, strings.ReplaceAll(t, "_", " ")))
		}
	}
	return fmt.Sprintf("post %d× (%s)", total, strings.Join(parts, ", "))
}

// icsFor: one all-day event per calendar week, plus one on the start date when ads run.
func icsFor(spec engine.LaunchSpec, title string) string {
	var b strings.Builder
	line := func(s string) { b.WriteString(s + "\r\n") }
	date := func(d string) string { return strings.ReplaceAll(d, "-", "") }
	event := func(uid, day, summary string) {
		line("BEGIN:VEVENT")
		line("UID:" + uid + "@rehearsal")
		line("DTSTAMP:" + date(spec.StartDate) + "T000000Z")
		line("DTSTART;VALUE=DATE:" + date(day))
		line("SUMMARY:" + summary)
		line("END:VEVENT")
	}
	line("BEGIN:VCALENDAR")
	line("VERSION:2.0")
	line("PRODID:-//Rehearsal//EN")
	if spec.AdsUsd > 0 {
		event("rehearsal-ads-"+spec.StartDate, spec.StartDate, fmt.Sprintf("Ads start · $%.2f/day paused campaign · %s", spec.AdsDailyUsd, title))
	}
	for _, w := range spec.Calendar {
		parts := []string{postsText(w.Posts)}
		if w.Creators > 0 {
			parts = append(parts, fmt.Sprintf("ship %d pieces", w.Creators))
		}
		if w.AdsUsd > 0 {
			parts = append(parts, fmt.Sprintf("$%.0f ads", w.AdsUsd))
		}
		if w.PromoDiscount > 0 {
			parts = append(parts, fmt.Sprintf("%.0f%% off", w.PromoDiscount*100))
		}
		event(fmt.Sprintf("rehearsal-week-%d-%s", w.Week, spec.StartDate), w.StartDate,
			fmt.Sprintf("Week %d · %s · %s", w.Week, strings.Join(parts, " · "), title))
	}
	line("END:VCALENDAR")
	return b.String()
}

func (s *server) handleLaunchCalendar(w http.ResponseWriter, r *http.Request) {
	l, err := s.launchOf(r)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="rehearsal-`+l.Spec.StartDate+`.ics"`)
	io.WriteString(w, icsFor(l.Spec, l.Title))
}

// handleCreatorFit: {handle, images, direction?} -> labels on whether this creator fits the brief (Spark cache by handle).
func (s *server) handleCreatorFit(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Handle    string   `json:"handle"`
		Images    []string `json:"images"`
		Direction string   `json:"direction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	handle := spark.SafeHandle(in.Handle)
	if handle == "" {
		writeError(w, 400, "handle is required")
		return
	}
	if msg := imagesError(in.Images); msg != "" {
		writeError(w, 400, msg)
		return
	}
	var brief map[string]any
	if pb := playbookFor(in.Direction); pb != nil {
		brief, _ = pb["creator_brief"].(map[string]any)
	}
	raw, err := spark.CreatorFit(handle, brief, in.Images, nil)
	if err != nil {
		writeError(w, 503, "no cached read for @"+handle+"; run `go run ./cmd/sparkcache` with MUSE_API_KEY set")
		return
	}
	s.writeJSON(w, spark.ValidateCreatorFit(raw))
}

func (s *server) metaSource() string {
	if s.evidence == nil {
		return ""
	}
	return s.evidence.Source
}

var placements = []struct {
	Key  string
	Spec map[string]any
}{
	{"instagram_reels", map[string]any{"publisher_platforms": []string{"instagram"}, "instagram_positions": []string{"reels"}}},
	{"instagram_feed", map[string]any{"publisher_platforms": []string{"instagram"}, "instagram_positions": []string{"stream"}}},
	{"instagram_stories", map[string]any{"publisher_platforms": []string{"instagram"}, "instagram_positions": []string{"story"}}},
	{"facebook_feed", map[string]any{"publisher_platforms": []string{"facebook"}, "facebook_positions": []string{"feed"}}},
}

// metaReach: Marketing API delivery_estimate per placement for the confirmed customer. Audience size depends on
// targeting and the auction, not on the product, so any ad account gives the same numbers.
func metaReach(base, account, token string, info engine.BrandInfo) (map[string]any, error) {
	out := map[string]any{}
	client := &http.Client{Timeout: 30 * time.Second}
	for _, pl := range placements {
		targeting, _ := targetingFor(info)
		for k, v := range pl.Spec {
			targeting[k] = v
		}
		spec, _ := json.Marshal(targeting)
		q := url.Values{"optimization_goal": {"LINK_CLICKS"}, "targeting_spec": {string(spec)}, "access_token": {token}}
		resp, err := client.Get(strings.TrimRight(base, "/") + "/act_" + account + "/delivery_estimate?" + q.Encode())
		if err != nil {
			return nil, err
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("meta %s: %s", resp.Status, string(raw))
		}
		var body struct {
			Data []struct {
				DAU      float64  `json:"estimate_dau"`
				MAU      *float64 `json:"estimate_mau"`
				MAULower float64  `json:"estimate_mau_lower_bound"`
				MAUUpper float64  `json:"estimate_mau_upper_bound"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &body); err != nil || len(body.Data) == 0 {
			return nil, fmt.Errorf("meta: unexpected reply: %s", string(raw))
		}
		d := body.Data[0]
		mau := (d.MAULower + d.MAUUpper) / 2
		if d.MAU != nil {
			mau = *d.MAU
		}
		out[pl.Key] = map[string]any{"estimate_dau": d.DAU, "estimate_mau": mau}
	}
	return out, nil
}

// reach: Meta's delivery estimate for the target: live from the connected ad account when credentials are set
// (cached for the process), else the mock account's numbers, else nil. The second value names the source.
func (s *server) reach() (map[string]any, string) {
	if token, account := metaEnv(); token != "" && account != "" {
		if s.reachLive != nil {
			return s.reachLive, "meta"
		}
		base := os.Getenv("META_GRAPH_URL")
		if base == "" {
			base = "https://graph.facebook.com/v21.0"
		}
		if r, err := metaReach(base, account, token, s.brand.Brand); err == nil {
			s.reachLive = r
			return r, "meta"
		} else {
			log.Printf("meta delivery_estimate failed, using demo numbers: %v", err)
		}
	}
	if s.evidence == nil {
		return nil, ""
	}
	return s.evidence.Reach, "mock"
}

// suggestedCreators: creators found for the brief (business_discovery) with engine stats and, when cached, Spark's fit read.
// creatorSource: where the creator pool came from: "import" for a pasted list, else the Meta data's source.
func (s *server) creatorSource() string {
	if len(s.imported) > 0 {
		return "import"
	}
	return s.metaSource()
}

func (s *server) suggestedCreators() []map[string]any {
	pool := s.imported
	if len(pool) == 0 && s.evidence != nil {
		pool = s.evidence.Creators
	}
	if len(pool) == 0 {
		return nil
	}
	out := []map[string]any{}
	for _, c := range pool {
		row := map[string]any{"username": c.Username, "followers": c.Followers, "followers_band": c.FollowersBand,
			"engagement_rate": c.EngagementRate, "post_mix": c.PostMix, "biography": c.Biography, "fit": nil,
			"expected_engagement": engine.ExpectedEngagement(c)}
		if raw, err := spark.CreatorFit(c.Username, nil, nil, nil); err == nil {
			row["fit"] = spark.ValidateCreatorFit(raw)
		}
		out = append(out, row)
	}
	return out
}
