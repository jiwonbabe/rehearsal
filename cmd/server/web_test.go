package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var banned = regexp.MustCompile(`(?i)forecast|prediction`)

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
	pages, _ := filepath.Glob("web/*.html") // Task 5 switches this to the explicit five-page list
	js, _ := os.ReadFile("web/app.js")
	if !strings.Contains(string(js), "Demo brand · synthetic data") {
		t.Fatal("app.js must carry the badge")
	}
	for _, p := range pages {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("%s missing", p)
			continue
		}
		if !strings.Contains(string(raw), "app.js") || !strings.Contains(string(raw), "renderHeader(") {
			t.Errorf("%s does not load the shell", p)
		}
	}
}
