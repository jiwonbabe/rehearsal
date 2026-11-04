package main

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"rehearsal/engine"
)

// handleImportCreators: POST /api/creators {text} where text is a JSON array or a CSV of public creator profiles
// (for example a Bright Data Instagram export). Columns by any of these names: username|account|handle,
// followers|followers_count, engagement_rate|avg_engagement, avg_likes|likes, biography|bio. The list becomes the
// creator pool that plans are counted against and the list Launch shows; the owner's own data, never shown as demo.
func (s *server) handleImportCreators(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	creators, err := parseCreators(in.Text)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	s.imported = creators
	s.brand.CreatorPool = creators
	s.writeJSON(w, map[string]any{"count": len(creators), "source": "import", "creators": creators})
}

var creatorCols = map[string][]string{
	"username":  {"username", "account", "handle", "user", "profile_name", "name"},
	"followers": {"followers", "followers_count", "follower_count"},
	"rate":      {"engagement_rate", "avg_engagement", "engagement"},
	"likes":     {"avg_likes", "average_likes", "likes", "avg_like_count"},
	"bio":       {"biography", "bio"},
}

func parseCreators(text string) ([]engine.CreatorStat, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errNoCreators
	}
	rows := []map[string]string{}
	if strings.HasPrefix(text, "[") {
		var raw []map[string]any
		if err := json.Unmarshal([]byte(text), &raw); err != nil {
			return nil, err
		}
		for _, m := range raw {
			row := map[string]string{}
			for k, v := range m {
				row[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(fmtAny(v))
			}
			rows = append(rows, row)
		}
	} else {
		rd := csv.NewReader(strings.NewReader(text))
		rd.FieldsPerRecord = -1
		recs, err := rd.ReadAll()
		if err != nil {
			return nil, err
		}
		if len(recs) < 2 {
			return nil, errNoCreators
		}
		head := recs[0]
		for _, rec := range recs[1:] {
			row := map[string]string{}
			for i, h := range head {
				if i < len(rec) {
					row[strings.ToLower(strings.TrimSpace(h))] = strings.TrimSpace(rec[i])
				}
			}
			rows = append(rows, row)
		}
	}
	pick := func(row map[string]string, key string) string {
		for _, c := range creatorCols[key] {
			if v, ok := row[c]; ok && v != "" {
				return v
			}
		}
		return ""
	}
	out := []engine.CreatorStat{}
	for _, row := range rows {
		name := strings.TrimPrefix(pick(row, "username"), "@")
		followers := num(pick(row, "followers"))
		if name == "" || followers <= 0 {
			continue
		}
		rate := num(pick(row, "rate"))
		if rate > 1 { // a percent, not a fraction
			rate /= 100
		}
		if rate == 0 {
			if likes := num(pick(row, "likes")); likes > 0 {
				rate = likes / followers
			}
		}
		if rate <= 0 {
			continue
		}
		out = append(out, engine.ImportedCreator(name, followers, rate, pick(row, "bio")))
		if len(out) == 200 {
			break
		}
	}
	if len(out) == 0 {
		return nil, errNoCreators
	}
	return out, nil
}

var errNoCreators = errorString("no creators found: need username, followers and engagement_rate or avg_likes columns")

type errorString string

func (e errorString) Error() string { return string(e) }

func fmtAny(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	raw, _ := json.Marshal(v)
	return string(raw)
}

// num: "12,300", "12.3k", "1.2m", "4.5%" → a number.
func num(s string) float64 {
	s = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), ",", ""))
	s = strings.TrimSuffix(s, "%")
	mult := 1.0
	if strings.HasSuffix(s, "k") {
		mult, s = 1000, strings.TrimSuffix(s, "k")
	} else if strings.HasSuffix(s, "m") {
		mult, s = 1e6, strings.TrimSuffix(s, "m")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f * mult
}
