package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"rehearsal/engine"
	"rehearsal/spark"
)

// connectMeta loads one account's Meta data (mock folder today; Graph API when a token is given later).
func (s *server) connectMeta(source, handle string) (int, string) {
	handle = spark.SafeHandle(handle)
	if handle == "" {
		return 400, "handle is required"
	}
	switch source {
	case "mock":
		dir := filepath.Join("data", "meta", handle)
		if _, err := os.Stat(filepath.Join(dir, "account.json")); err != nil {
			return 404, "no demo data for @" + handle
		}
		m, err := engine.LoadMeta(dir)
		if err != nil {
			return 500, err.Error()
		}
		ev := engine.RunEvidence(m)
		s.meta, s.evidence, s.metaHandle = &m, &ev, handle
		return 200, ""
	case "instagram":
		return 400, "Instagram Login needs a token and Meta App Review; use source \"mock\" for the demo"
	}
	return 400, `source must be "mock" or "instagram"`
}

// channels: the cached strategist verdicts, validated against the numbers the evidence allows. nil before connect.
func (s *server) channels() []map[string]any {
	if s.evidence == nil {
		return nil
	}
	raw, err := spark.Channels(s.metaHandle, nil, nil, nil, nil)
	if err != nil {
		return nil
	}
	return spark.ValidateChannels(raw, engine.EvidenceNumbers(*s.evidence, s.stats))
}

func (s *server) evidenceOut() map[string]any {
	return map[string]any{"connected": s.evidence != nil, "handle": s.metaHandle, "evidence": s.evidence,
		"stats": s.stats, "channels": s.channels()}
}

func (s *server) handleMetaConnect(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Source string `json:"source"`
		Handle string `json:"handle"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if code, msg := s.connectMeta(in.Source, in.Handle); code != 200 {
		writeError(w, code, msg)
		return
	}
	s.writeJSON(w, s.evidenceOut())
}

func (s *server) handleEvidence(w http.ResponseWriter, r *http.Request) {
	if s.evidence == nil {
		writeError(w, 409, "connect an Instagram account first (POST /api/meta/connect)")
		return
	}
	s.writeJSON(w, s.evidenceOut())
}
