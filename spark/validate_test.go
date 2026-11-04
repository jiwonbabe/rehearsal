package spark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func rawDir(id string, ranges map[string]any) map[string]any {
	return map[string]any{"id": id, "title": id, "rank": 1.0, "ranges": ranges}
}

func TestValidateDirectionsClampsToBudget(t *testing.T) {
	raw := map[string]any{"directions": []any{
		rawDir("pay_for_reach", map[string]any{"ads_usd": []any{120.0, 900.0}, "creator_count": []any{0.0, 0.0}}),
		rawDir("send_to_creators", map[string]any{"ads_usd": []any{0.0, 60.0}, "creator_count": []any{4.0, 20.0}}),
	}}
	out, err := ValidateDirections(raw, 400, 45)
	if err != nil || len(out) != 2 {
		t.Fatalf("got %d dirs, err %v", len(out), err)
	}
	ads := out[0]["ranges"].(map[string]any)["ads_usd"].([]float64)
	if ads[0] != 120 || ads[1] != 400 {
		t.Errorf("ads_usd clamp: %v", ads)
	}
	creators := out[1]["ranges"].(map[string]any)["creator_count"].([]float64)
	if creators[0] != 4 || creators[1] != 8 { // floor((400-0)/45) = 8
		t.Errorf("creator_count clamp: %v", creators)
	}
}

func TestValidateDirectionsRepairsMalformedRanges(t *testing.T) {
	raw := map[string]any{"directions": []any{
		rawDir("just_post", map[string]any{"posts_per_week": []any{5.0}, "ads_usd": []any{60.0, 10.0}, "promo_discount": "x"}),
		map[string]any{"title": "no id"},
		rawDir("bogus_direction", map[string]any{}),
	}}
	out, err := ValidateDirections(raw, 400, 45)
	if err != nil || len(out) != 1 {
		t.Fatalf("got %d dirs, err %v", len(out), err)
	}
	r := out[0]["ranges"].(map[string]any)
	want := map[string][]float64{"posts_per_week": {3, 5}, "creator_count": {0, 0}, "ads_usd": {10, 60},
		"promo_discount": {0, 0}}
	for k, w := range want {
		got := r[k].([]float64)
		if got[0] != w[0] || got[1] != w[1] {
			t.Errorf("%s: got %v want %v", k, got, w)
		}
	}
	if pw := r["promo_week"].([]*int); len(pw) != 1 || pw[0] != nil {
		t.Errorf("promo_week default: %v", pw)
	}
	mix := r["post_mix"].(map[string][]float64)
	if mix["on_hand"][0] != 0.4 || mix["on_hand"][1] != 0.6 || mix["on_neck"][1] != 0.1 {
		t.Errorf("post_mix default: %v", mix)
	}
}

func TestValidateDirectionsEmpty(t *testing.T) {
	if _, err := ValidateDirections(map[string]any{"directions": []any{}}, 400, 45); err == nil {
		t.Error("want error for no directions")
	}
}

func TestParseShocks(t *testing.T) {
	raw := map[string]any{"shocks": []any{
		map[string]any{"codes": []any{"reach_drop", "bogus"}, "why": "x"},
		map[string]any{"why": "no codes"},
	}}
	got := ParseShocks(raw)
	if len(got) != 1 || len(got[0]) != 2 || got[0][0] != "reach_drop" {
		t.Errorf("got %v", got)
	}
}

func TestValidateCheck(t *testing.T) {
	allowed := []float64{0.86, 0.6, 86, 60}
	if got, ok := ValidateCheck(map[string]any{"check": "Ask the creators who posted if they will again; at 86% seeding wins, you are at 60%."}, allowed); !ok || got == "" {
		t.Errorf("valid line rejected")
	}
	if _, ok := ValidateCheck(map[string]any{"check": "Expect about 40 orders if 86% post."}, allowed); ok {
		t.Errorf("line with a foreign number accepted")
	}
	if _, ok := ValidateCheck(map[string]any{"check": ""}, allowed); ok {
		t.Errorf("empty accepted")
	}
}

func TestValidateDirectionsNormalizesPercentDiscount(t *testing.T) {
	raw := map[string]any{"directions": []any{
		rawDir("one_week_sale", map[string]any{"promo_discount": []any{15.0, 25.0}}),
	}}
	out, err := ValidateDirections(raw, 400, 45)
	if err != nil || len(out) != 1 {
		t.Fatalf("got %d dirs, err %v", len(out), err)
	}
	got := out[0]["ranges"].(map[string]any)["promo_discount"].([]float64)
	if got[0] != 0.15 || got[1] != 0.25 {
		t.Errorf("promo_discount: %v", got)
	}
}

func TestValidateProfile(t *testing.T) {
	long := strings.Repeat("é", 300)
	raw := map[string]any{
		"handle": "bnd.jewelry", "what_they_sell": long, "style": "minimal", "market": "JP",
		"likely_customer": map[string]any{"gender": "women", "age": "25-34", "price_band": "5000"},
		"post_types_seen": map[string]any{"on_hand": "some", "close_up": "most", "flat_lay": "lots"},
		"best_post_type":  "close_up", "confidence": "high",
		"success_metric": "revenue", "ambition": "boom", "why_goal": long,
	}
	out := ValidateProfile(raw)
	if n := len([]rune(out["what_they_sell"].(string))); n != 200 {
		t.Errorf("sentence not trimmed to 200 runes: %d", n)
	}
	if out["market"] != "" {
		t.Errorf("invalid market kept: %v", out["market"])
	}
	c := out["likely_customer"].(map[string]any)
	if c["gender"] != "women" || c["age"] != "25-34" || c["price_band"] != "" {
		t.Errorf("likely_customer: %v", c)
	}
	p := out["post_types_seen"].(map[string]any)
	if p["on_hand"] != "some" || p["close_up"] != "most" || p["flat_lay"] != "" || p["on_neck"] != "" {
		t.Errorf("post_types_seen: %v", p)
	}
	if out["best_post_type"] != "close_up" || out["confidence"] != "high" || out["style"] != "minimal" || out["handle"] != "bnd.jewelry" {
		t.Errorf("valid fields changed: %v", out)
	}
	if out["success_metric"] != "revenue" || out["ambition"] != "" || len([]rune(out["why_goal"].(string))) != 200 {
		t.Errorf("goal fields: %v %v", out["success_metric"], out["ambition"])
	}
	if len(ValidateProfile(nil)) != 12 {
		t.Errorf("empty input should still have all 12 keys")
	}
}

func TestValidatePlaybook(t *testing.T) {
	raw := map[string]any{
		"platform": "tiktok",
		"weeks": []any{
			map[string]any{"week": 1, "do": "Post the ring close-ups.", "post_idea": "Ring on a beige cloth"},
			map[string]any{"week": 2, "do": "Spend $120 on ads.", "post_idea": "Expect 30 orders"},
		},
		"ad":            map[string]any{"audience": "Women in Seoul", "creative": "Close-ups", "placements": []any{"instagram_feed", "tiktok_feed"}},
		"creator_brief": "not an object",
	}
	p := ValidatePlaybook(raw)
	if p["platform"] != "instagram" {
		t.Errorf("platform %v", p["platform"])
	}
	weeks := p["weeks"].([]map[string]any)
	if len(weeks) != 4 || weeks[0]["week"] != 1 || weeks[3]["week"] != 4 {
		t.Fatalf("weeks %v", weeks)
	}
	if weeks[1]["do"] != "" || weeks[1]["post_idea"] != "" {
		t.Errorf("numbers not dropped: %v", weeks[1])
	}
	if weeks[2]["do"] != "" {
		t.Errorf("padded week should be empty: %v", weeks[2])
	}
	ad := p["ad"].(map[string]any)
	if pl := ad["placements"].([]string); len(pl) != 1 || pl[0] != "instagram_feed" {
		t.Errorf("placements %v", pl)
	}
	if p["creator_brief"] != nil {
		t.Errorf("creator_brief should be nil, got %v", p["creator_brief"])
	}
	if noNumbers("Keep 20% for later") != "" || noNumbers("Post three times") == "" {
		t.Error("noNumbers")
	}
}

func TestValidateCreatorFit(t *testing.T) {
	p := ValidateCreatorFit(map[string]any{"handle": "@soft.daily", "fit": "very high", "size_band": "small",
		"why": "Posts hands with thin rings on 30% of days", "send": "the thin band"})
	if p["handle"] != "@soft.daily" || p["fit"] != "" || p["size_band"] != "small" || p["why"] != "" || p["send"] != "the thin band" {
		t.Errorf("%v", p)
	}
	if len(p) != 5 {
		t.Errorf("keys %v", p)
	}
}

func TestPlaybookIDCannotEscapeCache(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "secret.json"), []byte(`{"weeks":[]}`), 0o644)
	old := CacheDir
	CacheDir = filepath.Join(tmp, "cache")
	t.Cleanup(func() { CacheDir = old })
	t.Setenv("MUSE_API_KEY", "")
	if _, err := Playbook(map[string]any{"id": "/../secret"}, nil, nil, nil); err == nil {
		t.Fatal("traversal id read a file outside the cache dir")
	}
	if !KnownDirection("pay_for_reach") || KnownDirection("../x") {
		t.Error("KnownDirection")
	}
}

func TestValidateChannels(t *testing.T) {
	raw := map[string]any{"channels": []any{
		map[string]any{"channel": "instagram_reels", "verdict": "lead", "why_data": "Reels reach 3.08 times feed and get 262 saves.", "why_category": "Wearing shots build trust for gifts.", "how": "Post two Reels a week."},
		map[string]any{"channel": "threads", "verdict": "support", "why_data": "Threads reach 9999 people.", "why_category": "Spend 20% there.", "how": "Ask a question."},
		map[string]any{"channel": "myspace", "verdict": "lead", "why_data": "", "why_category": "", "how": ""},
		map[string]any{"channel": "tiktok", "verdict": "maybe", "why_data": "", "why_category": "", "how": ""},
	}}
	out := ValidateChannels(raw, []float64{3.08, 262, 46, 0.46}, []string{"25-34"})
	if len(out) != 2 {
		t.Fatalf("%v", out)
	}
	if out[0]["why_data"] == "" || out[1]["why_data"] != "" || out[1]["why_category"] != "" {
		t.Errorf("%v", out)
	}
	labels := []string{"25-34", "55+", "30s"}
	if !quotesAllowed("46% of followers are 25-34", []float64{46}, labels) || quotesAllowed("about 1,200 people", []float64{46}, labels) {
		t.Error("quotesAllowed")
	}
	if quotesAllowed("10-15 orders a week", nil, labels) || quotesAllowed("grow 12-15% a month", []float64{46}, labels) || quotesAllowed("in their 20s", nil, labels) {
		t.Error("ranges that are not labels must be rejected")
	}
	if !quotesAllowed("55+ buyers and 30s women", nil, labels) {
		t.Error("declared labels must pass")
	}
}

func TestValidateTranslation(t *testing.T) {
	en := map[string]any{"title": "Boost posts", "rank": 1, "weeks": []any{map[string]any{"do": "Post twice", "week": 1}}, "caution": nil, "list": []any{"a", "b"}}
	ko := map[string]any{"title": "게시물 밀어주기", "rank": 99, "weeks": []any{map[string]any{"do": "", "week": 2}}, "list": []any{"ㄱ"}}
	out := ValidateTranslation(en, ko).(map[string]any)
	if out["title"] != "게시물 밀어주기" || out["rank"] != 1 || out["caution"] != nil {
		t.Errorf("%v", out)
	}
	w := out["weeks"].([]any)[0].(map[string]any)
	if w["do"] != "Post twice" || w["week"] != 1 {
		t.Errorf("week %v", w)
	}
	if l := out["list"].([]any); len(l) != 2 || l[0] != "a" {
		t.Errorf("list %v", l)
	}
	bad := ValidateTranslation(map[string]any{"a": "262 saves", "b": "Our guess", "c": "seven days"},
		map[string]any{"a": "260개 저장", "b": "지난 예측", "c": "7일"}).(map[string]any)
	if bad["a"] != "262 saves" || bad["b"] != "Our guess" {
		t.Errorf("digits or banned words changed: %v", bad)
	}
	if bad["c"] != "7일" {
		t.Errorf("a word written as the same numeral with a Korean unit is fine: %v", bad["c"])
	}
	dates := ValidateTranslation(map[string]any{"n": "December lifts demand 2.23 times."}, map[string]any{"n": "12월에는 수요가 2.23배입니다."}).(map[string]any)
	if dates["n"] != "12월에는 수요가 2.23배입니다." {
		t.Errorf("month numerals are not figures: %v", dates["n"])
	}
	flat := map[string]string{}
	FlattenStrings(en, out, flat)
	if flat["Boost posts"] != "게시물 밀어주기" || len(flat) != 1 {
		t.Errorf("flat %v", flat)
	}
}

func TestSeasonNote(t *testing.T) {
	ok := map[string]any{"season_note": "December lifts demand 2.23 times but ads cost 1.27 times more."}
	bad := map[string]any{"season_note": "Expect 120 orders in December."}
	if SeasonNote(ok, []float64{2.23, 1.27}) == "" || SeasonNote(bad, []float64{2.23, 1.27}) != "" || SeasonNote(map[string]any{}, nil) != "" {
		t.Error("SeasonNote")
	}
}
