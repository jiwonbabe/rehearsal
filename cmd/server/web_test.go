package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var banned = regexp.MustCompile(`(?i)forecast|prediction|예측`)

func TestNoBannedWords(t *testing.T) {
	for _, glob := range []string{"web/*", "prompts/*", "spark/cache/*", "cmd/*/*.go", "engine/*.go", "README.md", "DEMO.md"} {
		files, _ := filepath.Glob(glob)
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue // this file carries the regex itself
			}
			raw, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			if loc := banned.FindIndex(raw); loc != nil {
				t.Errorf("%s contains banned word near byte %d", f, loc[0])
			}
		}
	}
}

func TestEveryPageHasBadgeAndShell(t *testing.T) {
	pages := []string{"web/index.html", "web/directions.html", "web/simulate.html", "web/verdict.html", "web/review.html", "web/launch.html"}
	js, _ := os.ReadFile("web/app.js")
	if !strings.Contains(string(js), "Demo brand · synthetic data") {
		t.Fatal("app.js must carry the badge")
	}
	ko, _ := os.ReadFile("web/i18n.js")
	if !strings.Contains(string(ko), "데모 브랜드 · 합성 데이터") {
		t.Fatal("i18n.js must carry the Korean badge")
	}
	for _, p := range pages {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("%s missing", p)
			continue
		}
		if !strings.Contains(string(raw), "app.js") || !strings.Contains(string(raw), "i18n.js") || !strings.Contains(string(raw), "renderHeader(") {
			t.Errorf("%s does not load the shell", p)
		}
	}
}
