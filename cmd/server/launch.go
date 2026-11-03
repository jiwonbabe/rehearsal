package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	raw, err := spark.Playbook(map[string]any{"id": id}, nil, nil)
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

func metaEnv() (token, account string) {
	return os.Getenv("META_ACCESS_TOKEN"), strings.TrimPrefix(os.Getenv("META_AD_ACCOUNT_ID"), "act_")
}

// adRequest: the Meta Marketing API objects for this plan, both PAUSED. nil when the plan has no ads.
func adRequest(spec engine.LaunchSpec, plan engine.Plan, info engine.BrandInfo, title string) (campaign, adset map[string]any) {
	if plan.AdsUsd <= 0 {
		return nil, nil
	}
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
	// OUTCOME_TRAFFIC + LINK_CLICKS needs no pixel; the ad set ends at midnight after the last day so all days are bought.
	endDay, _ := time.Parse("2006-01-02", spec.EndDate)
	campaign = map[string]any{"name": "Rehearsal · " + title + " · " + spec.StartDate, "objective": "OUTCOME_TRAFFIC",
		"status": "PAUSED", "special_ad_categories": []string{}}
	adset = map[string]any{"name": fmt.Sprintf("%s · %s %s %s", title, country, info.Target["gender"], info.Target["age"]),
		"daily_budget": spec.AdsDailyCents, "billing_event": "IMPRESSIONS", "optimization_goal": "LINK_CLICKS",
		"bid_strategy": "LOWEST_COST_WITHOUT_CAP", "targeting": targeting,
		"start_time": spec.StartDate + "T00:00:00+0000", "end_time": endDay.AddDate(0, 0, 1).Format("2006-01-02") + "T00:00:00+0000", "status": "PAUSED"}
	return campaign, adset
}

func formOf(m map[string]any) url.Values {
	form := url.Values{}
	for k, v := range m {
		switch x := v.(type) {
		case string:
			form.Set(k, x)
		case int, float64:
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
	campaign, adset := adRequest(l.Spec, l.Plan, s.brand.Brand, l.Title)
	var adReq map[string]any
	if campaign != nil {
		adReq = map[string]any{"campaign": campaign, "adset": adset}
	}
	s.writeJSON(w, map[string]any{
		"direction":     map[string]string{"id": l.Plan.Direction, "title": l.Title},
		"spec":          l.Spec,
		"playbook":      l.Playbook,
		"ad_request":    adReq,
		"ads_connected": token != "" && account != "",
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
	raw, err := spark.CreatorFit(handle, brief, in.Images)
	if err != nil {
		writeError(w, 503, "no cached read for @"+handle+"; run `go run ./cmd/sparkcache` with MUSE_API_KEY set")
		return
	}
	s.writeJSON(w, spark.ValidateCreatorFit(raw))
}
