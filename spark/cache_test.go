package spark

import "testing"

func TestCacheFilesValidate(t *testing.T) {
	CacheDir, PromptDir = "cache", "../prompts"
	for month := 1; month <= 2; month++ {
		raw, err := Directions(month, map[string]any{}, "", "")
		if err != nil {
			t.Fatal(err)
		}
		dirs, err := ValidateDirections(raw, 400, 45)
		if err != nil || len(dirs) < 3 {
			t.Fatalf("month %d: %d dirs, %v", month, len(dirs), err)
		}
		for _, d := range dirs {
			raw, err := Adversary(d, map[string]float64{})
			if err != nil {
				t.Fatal(err)
			}
			if n := len(ParseShocks(raw)); n != 20 {
				t.Errorf("%s: %d combos", d["id"], n)
			}
			pb, err := Playbook(d, map[string]any{}, map[string]any{}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if ValidatePlaybook(pb)["weeks"].([]map[string]any)[0]["do"] == "" {
				t.Errorf("%s: empty playbook", d["id"])
			}
		}
	}
}
