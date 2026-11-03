// Package spark calls Muse Spark through the OpenAI-compatible Meta Model API and caches
// every response on disk. Cache files are named by call (directions-m1, adversary-<id>),
// committed to git, and hand-editable. It must not import rehearsal/engine (oracle isolation): it
// receives plain maps and returns parsed JSON.
//
// Env:
//
//	MUSE_API_KEY   required for live calls
//	MUSE_BASE_URL  default https://api.meta.ai/v1  (confirm on dev.meta.ai; override if different)
//	MUSE_MODEL     default muse-spark-1.3           (override to the exact model id the API lists)
package spark

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var CacheDir = "spark/cache"
var PromptDir = "prompts"

func loadPrompt(name string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(PromptDir, name))
	return string(raw), err
}

func fill(template string, values map[string]any) string {
	out := template
	for key, value := range values {
		var text string
		if s, ok := value.(string); ok {
			text = s
		} else {
			raw, _ := json.MarshalIndent(value, "", " ")
			text = string(raw)
		}
		out = strings.ReplaceAll(out, "{{"+key+"}}", text)
	}
	return out
}

func cachePath(name string) string {
	return filepath.Join(CacheDir, name+".json")
}

// Call returns parsed JSON from Spark. Cache first; live only with MUSE_API_KEY.
func Call(name, prompt string) (map[string]any, error) {
	if err := os.MkdirAll(CacheDir, 0o755); err != nil {
		return nil, err
	}
	path := cachePath(name)
	if raw, err := os.ReadFile(path); err == nil {
		var out map[string]any
		return out, json.Unmarshal(raw, &out)
	}
	key := os.Getenv("MUSE_API_KEY")
	if key == "" {
		return nil, errors.New("no cached response for " + name + " and no MUSE_API_KEY")
	}

	model := os.Getenv("MUSE_MODEL")
	if model == "" {
		model = "muse-spark-1.3"
	}
	base := os.Getenv("MUSE_BASE_URL")
	if base == "" {
		base = "https://api.meta.ai/v1"
	}
	body, _ := json.Marshal(map[string]any{
		"model":           model,
		"messages":        []map[string]string{{"role": "user", "content": prompt}},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.2,
	})
	req, err := http.NewRequest("POST", strings.TrimRight(base, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("spark %s: %s", resp.Status, string(raw))
	}
	var wrapper struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil || len(wrapper.Choices) == 0 {
		return nil, fmt.Errorf("spark: unexpected response: %s", string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(wrapper.Choices[0].Message.Content), &out); err != nil {
		return nil, fmt.Errorf("spark: content is not JSON: %v", err)
	}
	pretty, _ := json.MarshalIndent(out, "", " ")
	return out, os.WriteFile(path, pretty, 0o644)
}

// Directions: planner call for the given month (1 or 2). summary comes from engine.Summary; lastMonthBlock is "" on month one.
func Directions(month int, summary map[string]any, lastMonthBlock string) (map[string]any, error) {
	tpl, err := loadPrompt("directions.md")
	if err != nil {
		return nil, err
	}
	return Call(fmt.Sprintf("directions-m%d", month), fill(tpl, map[string]any{"SUMMARY_JSON": summary, "LAST_MONTH_BLOCK": lastMonthBlock}))
}

// Adversary: shock-picker call for one direction. priors come from engine.PriorProbabilities.
func Adversary(direction map[string]any, priors map[string]float64) (map[string]any, error) {
	tpl, err := loadPrompt("adversary.md")
	if err != nil {
		return nil, err
	}
	values := map[string]any{"DIRECTION_JSON": direction}
	for code, p := range priors {
		values["P_"+code] = fmt.Sprintf("%.2f", p)
	}
	id, _ := direction["id"].(string)
	return Call("adversary-"+id, fill(tpl, values))
}

// Check: one-sentence "what to verify this week" from the engine's flip result.
func Check(aID, bID string, flip map[string]any, aTitle, bTitle string) (map[string]any, error) {
	tpl, err := loadPrompt("check.md")
	if err != nil {
		return nil, err
	}
	return Call("check-"+aID+"-"+bID, fill(tpl, map[string]any{"FLIP_JSON": flip, "A_TITLE": aTitle, "B_TITLE": bTitle}))
}
