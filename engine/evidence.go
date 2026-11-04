package engine

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Evidence: averages and shares from the brand's own Meta data. Same formulas as reference/engine/evidence.py.
// Spark may quote these numbers (and the public statistics) and nothing else.

type Audience struct {
	Followers    float64 `json:"followers"`
	TopAge       string  `json:"top_age"`
	TopAgeShare  float64 `json:"top_age_share"`
	WomenShare   float64 `json:"women_share"`
	TopCity      string  `json:"top_city"`
	TopCityShare float64 `json:"top_city_share"`
}

type FormatStat struct {
	Format          string  `json:"format"`
	Posts           int     `json:"posts"`
	AvgReach        int     `json:"avg_reach"`
	AvgSaves        int     `json:"avg_saves"`
	LinkTapsPerPost int     `json:"link_taps_per_post"`
	ReachVsFeed     float64 `json:"reach_vs_feed"`
}

type PostTypeStat struct {
	PostType string `json:"post_type"`
	Posts    int    `json:"posts"`
	AvgReach int    `json:"avg_reach"`
	AvgSaves int    `json:"avg_saves"`
}

type HashtagStat struct {
	Hashtag        string             `json:"hashtag"`
	ShareByType    map[string]float64 `json:"share_by_type"`
	AvgLikesByType map[string]int     `json:"avg_likes_by_type"`
}

type PeerStat struct {
	Username  string             `json:"username"`
	Followers float64            `json:"followers"`
	PostMix   map[string]float64 `json:"post_mix"`
	AvgLikes  int                `json:"avg_likes"`
}

type CreatorStat struct {
	Username       string             `json:"username"`
	Followers      float64            `json:"followers"`
	FollowersBand  string             `json:"followers_band"`
	EngagementRate float64            `json:"engagement_rate"`
	PostMix        map[string]float64 `json:"post_mix"`
	Biography      string             `json:"biography"`
}

type Category struct {
	Hashtags []HashtagStat `json:"hashtags"`
	Peers    []PeerStat    `json:"peers"`
}

type Evidence struct {
	Source    string         `json:"source"`
	Days      int            `json:"days"`
	Audience  Audience       `json:"audience"`
	Formats   []FormatStat   `json:"formats"`
	PostTypes []PostTypeStat `json:"post_types"`
	BestType  string         `json:"best_type"`
	Category  Category       `json:"category"`
	Creators  []CreatorStat  `json:"creators"`
	Reach     map[string]any `json:"reach"`
}

var formatKinds = []struct{ Format, Kind string }{{"feed", "FEED"}, {"reels", "REELS"}, {"stories", "STORY"}, {"threads", ""}}

func avgInt(xs []float64) int {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return halfUp(sum / float64(len(xs)))
}

func share2(n, d float64) float64 {
	if d == 0 {
		return 0
	}
	return roundTo(n/d, 2)
}

// topKey: the key with the largest value; ties go to the alphabetically first key.
func topKey(m map[string]float64) (string, float64) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	best, bestV := "", math.Inf(-1)
	for _, k := range keys {
		if m[k] > bestV {
			best, bestV = k, m[k]
		}
	}
	return best, roundTo(bestV, 2)
}

func postMix(items []LabelledPost) map[string]float64 {
	out := map[string]float64{}
	for _, t := range PostTypes {
		n := 0.0
		for _, p := range items {
			if p.PostType == t {
				n++
			}
		}
		out[t] = share2(n, float64(len(items)))
	}
	return out
}

func RunEvidence(m MetaData) Evidence {
	demo := m.Account.FollowerDemographics
	topAge, topAgeShare := topKey(demo["age"])
	topCity, topCityShare := topKey(demo["city"])
	e := Evidence{Source: m.Account.Source, Days: m.Media.Days,
		Audience: Audience{Followers: m.Account.FollowersCount, TopAge: topAge, TopAgeShare: topAgeShare,
			WomenShare: roundTo(demo["gender"]["women"], 2), TopCity: topCity, TopCityShare: topCityShare},
		Reach: m.Reach.Data}
	var feedReach []float64
	for _, p := range m.Media.Data {
		if p.MediaProductType == "FEED" {
			feedReach = append(feedReach, p.Insights.Reach)
		}
	}
	feedAvg := float64(avgInt(feedReach))
	for _, fk := range formatKinds {
		var reach, saves, taps []float64
		if fk.Format == "threads" {
			for _, p := range m.Threads.Data {
				reach, saves = append(reach, p.Insights.Views), append(saves, p.Insights.Replies)
			}
		} else {
			for _, p := range m.Media.Data {
				if p.MediaProductType == fk.Kind {
					reach, saves, taps = append(reach, p.Insights.Reach), append(saves, p.Insights.Saved), append(taps, p.Insights.LinkTaps)
				}
			}
		}
		r := avgInt(reach)
		e.Formats = append(e.Formats, FormatStat{Format: fk.Format, Posts: len(reach), AvgReach: r, AvgSaves: avgInt(saves),
			LinkTapsPerPost: avgInt(taps), ReachVsFeed: share2(float64(r), feedAvg)})
	}
	bestSaves := -1
	for _, t := range PostTypes {
		var reach, saves []float64
		for _, p := range m.Media.Data {
			if p.PostType == t && p.MediaProductType != "STORY" {
				reach, saves = append(reach, p.Insights.Reach), append(saves, p.Insights.Saved)
			}
		}
		st := PostTypeStat{PostType: t, Posts: len(reach), AvgReach: avgInt(reach), AvgSaves: avgInt(saves)}
		e.PostTypes = append(e.PostTypes, st)
		if st.AvgSaves > bestSaves {
			bestSaves, e.BestType = st.AvgSaves, t
		}
	}
	e.Category.Hashtags = []HashtagStat{}
	for _, h := range m.Hashtags.Data {
		likes := map[string]int{}
		for _, t := range PostTypes {
			var xs []float64
			for _, p := range h.TopMedia {
				if p.PostType == t {
					xs = append(xs, p.LikeCount)
				}
			}
			likes[t] = avgInt(xs)
		}
		e.Category.Hashtags = append(e.Category.Hashtags, HashtagStat{Hashtag: h.Hashtag, ShareByType: postMix(h.TopMedia), AvgLikesByType: likes})
	}
	e.Category.Peers = []PeerStat{}
	for _, p := range m.Peers.Data {
		e.Category.Peers = append(e.Category.Peers, PeerStat{Username: p.Username, Followers: p.FollowersCount, PostMix: postMix(p.Recent), AvgLikes: avgInt(likesOf(p.Recent))})
	}
	e.Creators = []CreatorStat{}
	for _, c := range m.Creators.Data {
		cs := ImportedCreator(c.Username, c.FollowersCount, float64(avgInt(likesOf(c.Recent)))/c.FollowersCount, c.Biography)
		cs.PostMix = postMix(c.Recent)
		e.Creators = append(e.Creators, cs)
	}
	return e
}

func likesOf(items []LabelledPost) []float64 {
	var xs []float64
	for _, p := range items {
		xs = append(xs, p.LikeCount)
	}
	return xs
}

// StatRow: one sourced public statistic (data/platform_stats.json).
type StatRow struct {
	Market   string  `json:"market"`
	AgeBand  string  `json:"age_band"`
	Platform string  `json:"platform"`
	Share    float64 `json:"share"`
	Source   string  `json:"source"`
	URL      string  `json:"url"`
	Year     int     `json:"year"`
	Note     string  `json:"note"`
}

// EvidenceNumbers: every number a Spark line may quote, in plain and percent forms.
func EvidenceNumbers(e Evidence, stats []StatRow) []float64 {
	var out []float64
	add := func(xs ...float64) {
		for _, x := range xs {
			out = append(out, x)
		}
	}
	pct := func(xs ...float64) {
		for _, x := range xs {
			add(x, roundTo(x*100, 2))
		}
	}
	add(float64(e.Days), e.Audience.Followers)
	pct(e.Audience.TopAgeShare, e.Audience.WomenShare, e.Audience.TopCityShare)
	for _, f := range e.Formats {
		add(float64(f.Posts), float64(f.AvgReach), float64(f.AvgSaves), float64(f.LinkTapsPerPost), f.ReachVsFeed)
	}
	for _, p := range e.PostTypes {
		add(float64(p.Posts), float64(p.AvgReach), float64(p.AvgSaves))
	}
	for _, h := range e.Category.Hashtags {
		for _, t := range PostTypes {
			pct(h.ShareByType[t])
			add(float64(h.AvgLikesByType[t]))
		}
	}
	for _, p := range e.Category.Peers {
		add(p.Followers, float64(p.AvgLikes))
		for _, t := range PostTypes {
			pct(p.PostMix[t])
		}
	}
	for _, c := range e.Creators {
		add(c.Followers)
		pct(c.EngagementRate)
	}
	for _, v := range e.Reach {
		if m, ok := v.(map[string]any); ok {
			for _, x := range m {
				if f, ok := x.(float64); ok {
					add(f)
				}
			}
		}
	}
	for _, s := range stats {
		pct(s.Share)
	}
	return out
}

// EvidenceLabels: age-band labels a Spark line may write without them counting as figures (25-34, 55+, 30s).
func EvidenceLabels(e Evidence, stats []StatRow) []string {
	seen := map[string]bool{}
	var out []string
	add := func(band string) {
		if band == "" || seen[band] {
			return
		}
		seen[band] = true
		out = append(out, band)
		if len(band) >= 2 && strings.Contains(band, "-") {
			out = append(out, band[:1]+"0s")
		}
	}
	add(e.Audience.TopAge)
	for _, s := range stats {
		add(s.AgeBand)
	}
	return out
}

// FollowersBand: small (<10k), mid (<100k), large.
func FollowersBand(followers float64) string {
	switch {
	case followers < 10000:
		return "small"
	case followers < 100000:
		return "mid"
	}
	return "large"
}

// ImportedCreator: a creator from any public-profile source (business_discovery, a dataset export): username,
// followers, engagement rate (average likes ÷ followers), biography. Post types are unknown until Spark reads them.
func ImportedCreator(username string, followers, rate float64, bio string) CreatorStat {
	return CreatorStat{Username: username, Followers: followers, FollowersBand: FollowersBand(followers),
		EngagementRate: roundTo(rate, 3), PostMix: map[string]float64{}, Biography: bio}
}

// pewBand: the Setup age band -> the age band the public statistics use.
func pewBand(age string) string {
	switch age {
	case "18-24", "25-34":
		return "18-29"
	case "35-44":
		return "30-49"
	case "45-54", "55+":
		return "50-64"
	}
	return age
}

// TargetPlatforms: the public-statistics rows for the brand's market and target age, most used first.
func TargetPlatforms(stats []StatRow, market, age string) []StatRow {
	band := pewBand(age)
	out := []StatRow{}
	for _, r := range stats {
		if r.Market == market && r.AgeBand == band {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Share > out[j].Share })
	return out
}

// ChannelChecks: where the channel verdicts (a model's judgment) disagree with public data: a platform that at
// least half of the target uses but the verdicts skip or never mention. One sentence each; the owner decides.
func ChannelChecks(channels []map[string]any, rows []StatRow, market, age string) []string {
	verdict := map[string]string{} // platform -> best verdict among its channels
	rank := map[string]int{"lead": 3, "support": 2, "skip": 1}
	for _, c := range channels {
		ch, _ := c["channel"].(string)
		v, _ := c["verdict"].(string)
		p := strings.SplitN(ch, "_", 2)[0]
		if rank[v] > rank[verdict[p]] {
			verdict[p] = v
		}
	}
	names := map[string]string{"instagram": "Instagram", "tiktok": "TikTok", "facebook": "Facebook", "x": "X", "threads": "Threads"}
	out := []string{}
	for _, r := range rows {
		if r.Share < 0.5 {
			continue
		}
		name := names[r.Platform]
		if name == "" {
			name = r.Platform
		}
		pct := fmt.Sprintf("%d%%", int(math.Floor(r.Share*100+0.5)))
		switch verdict[r.Platform] {
		case "skip":
			out = append(out, fmt.Sprintf("The verdicts skip %s, but %s of %s adults %s use it (%s, %d). Check that call against your own numbers.", name, pct, market, r.AgeBand, shortSource(r.Source), r.Year))
		case "":
			out = append(out, fmt.Sprintf("No verdict covers %s, used by %s of %s adults %s (%s, %d).", name, pct, market, r.AgeBand, shortSource(r.Source), r.Year))
		}
	}
	return out
}

func shortSource(s string) string {
	if i := strings.Index(s, ","); i > 0 {
		return s[:i]
	}
	return s
}
