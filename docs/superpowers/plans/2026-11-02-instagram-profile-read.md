# Instagram profile read (screenshot → Spark → confirm) — plan

**Goal:** On Setup, the owner pastes their Instagram handle and drops a screenshot of their profile (or clicks "Use demo screenshot"); Muse Spark reads what they sell, the style, the post types and the likely customer; Setup pre-fills market, gender, age and price band; the owner confirms. Leaves a seam to replace the screenshot with Instagram Login (Graph API) after Meta app review.

**Spec/rules:** SPEC.md, CLAUDE.md. Models return labels and short sentences only — never a number. No live Spark call from the server (it unsets `MUSE_API_KEY`); the demo serves `spark/cache/profile-bnd.jewelry.json` made by `cmd/sparkcache`. Badge and banned words unchanged. Go stdlib only; vanilla JS.

**Verified facts:** `muse-spark-1.3` on `https://api.meta.ai/v1/chat/completions` accepts `{"type":"image_url","image_url":{"url":"data:image/png;base64,..."}}` content parts (base64 data URLs; remote URLs are fetched by Meta and may fail). With the demo screenshot and the prompt below it returned: market KR, women, 25-34, 200+, close_up most, confidence high.

## Files
- Create `prompts/profile.md`, `spark/cache/profile-bnd.jewelry.json` (via sparkcache), `web/demo-posts/bnd-profile.png` (already added).
- Modify `spark/client.go` (multimodal call + `Profile`), `spark/validate.go` (`ValidateProfile`), `spark/validate_test.go`, `cmd/server/handlers.go` + `main.go` (`POST /api/profile`), `cmd/server/main_test.go`, `cmd/sparkcache/main.go` (profile step), `web/index.html` (Step 1 card), `README.md` (Next + how it works), `DEMO.md` (Setup beat), `NOTES.md`.

## Task 1 — spark: prompt, multimodal call, Profile, ValidateProfile
- `prompts/profile.md` exact text:
```
You read a small brand's Instagram for Rehearsal. The owner is not a marketer. Below is the handle they typed and one or more screenshots of their profile or posts. Read the bio and the photo grid. Return ONLY this JSON. Labels and short sentences only; never a number, price, follower count or percentage.

Handle typed by the owner: {{HANDLE}}

{"handle": "<handle as shown, or the typed one>",
 "what_they_sell": "<one short sentence>",
 "style": "<one short sentence>",
 "market": "US|KR|other",
 "likely_customer": {"gender": "women|men|everyone", "age": "18-24|25-34|35-44", "price_band": "30-80|80-200|200+"},
 "post_types_seen": {"on_hand": "none|some|most", "close_up": "none|some|most", "flat_lay": "none|some|most", "on_neck": "none|some|most"},
 "best_post_type": "on_hand|close_up|flat_lay|on_neck",
 "confidence": "low|medium|high",
 "why": "<one sentence pointing at what you saw in the screenshot>"}
```
- `spark/client.go`: refactor the HTTP body building so `Call(name, prompt)` and a new `CallWithImages(name, prompt string, images []string)` share one function `callContent(name string, content []any)`; `images` are data URLs placed as `image_url` parts after the text part. Cache behaviour identical (by name). `Profile(handle string, images []string) (map[string]any, error)` fills `{{HANDLE}}` and calls `CallWithImages("profile-"+safe(handle), ...)` where `safe` lowercases, strips a leading `@`, and keeps only `[a-z0-9._-]`.
- `spark/validate.go`: `ValidateProfile(raw map[string]any) map[string]any` returns a map with exactly the keys above; enum fields outside their list become `""`; `post_types_seen` values outside `none|some|most` become `""`; sentences are trimmed to 200 runes; missing keys become `""`/empty map. Add `TestValidateProfile` (invalid enum dropped, long sentence trimmed, valid passes through unchanged) to `spark/validate_test.go`.

## Task 2 — server `POST /api/profile`
- Request `{"source": "screenshot", "handle": "bnd.jewelry", "images": ["data:image/png;base64,..."]}`. 400 when `source != "screenshot"` (message: `only source "screenshot" is available; Instagram Login comes after app review`), handle empty after `safe`, images empty or more than 8, or any image not starting with `data:image/` or longer than 4,000,000 bytes.
- Response: `ValidateProfile(raw)` plus `"cached": true|false` is NOT needed — just the validated map. When `spark.Profile` errors (no cache file and no key): 503 `{"error": "no cached read for @<handle>; run `go run ./cmd/sparkcache` with MUSE_API_KEY set"}`.
- Lock as the other handlers. Tests in `main_test.go`: 400 for bad source / empty handle / no images; 200 for `{"source":"screenshot","handle":"bnd.jewelry","images":["data:image/png;base64,AA=="]}` asserting `market`, `likely_customer.gender/age/price_band`, `best_post_type`, `confidence` are each in their allowed list (cache serves the answer; the image bytes do not matter); 503 for an unknown handle with `MUSE_API_KEY` unset (`t.Setenv`).

## Task 3 — sparkcache profile step + cache file
- `cmd/sparkcache`: after directions, read every `.png`/`.jpg`/`.jpeg` in `web/demo-posts/` (sorted), encode as data URLs, call `spark.Profile("bnd.jewelry", images)`, print `ok profile-bnd.jewelry: <market> <gender> <age> <price_band> <best_post_type> <confidence>` after `ValidateProfile`. The controller runs it with the key and commits `spark/cache/profile-bnd.jewelry.json`.

## Task 4 — Setup page Step 1 card (`web/index.html`)
- New first card "Step 1 · Your Instagram": text input (placeholder `@yourbrand or instagram.com/yourbrand`; the JS extracts the handle from a URL or `@handle`), a file input `accept="image/*" multiple` labelled "Drop a screenshot of your profile (or a few posts)", a "Use demo screenshot" button that loads `/demo-posts/bnd-profile.png` and sets the handle `bnd.jewelry`, and a "Read my feed" button. Before sending, images are resized in a canvas to max 1400 px on the long side and exported as JPEG quality 0.85 data URLs (the demo screenshot may be sent as-is if it is already under 1400 px wide).
- Result block "What we read from your feed · Muse Spark": `what_they_sell`, `style`, "Most of your photos: <best_post_type label>" using the labels on-hand / close-up / flat-lay / on-neck, "Likely customer: <gender>, <age>, $<price_band>, <market>", "Confidence: <confidence>", and `why` in muted text. Then the line: "Check these and fix anything wrong. This decides which directions we suggest." The Step 2 selects are set from the result (market, gender, age, price_band); the existing Save button is renamed "Confirm". A small honest caption under the card: "Demo: real @bnd.jewelry profile screenshot, synthetic sales data." Errors from the API show in the page's red line.
- Keep the Step 2 and budget/goal cards and the trust card as they are, below Step 1.

## Task 5 — docs
- README: in the architecture paragraph add the fourth Spark role (reads the profile screenshot → labels only → owner confirms). In "Next", first bullet: "Instagram Login (Graph API, `instagram_business_basic`) replaces the screenshot: `/api/profile` gains `source: "instagram"` and the server fetches recent media itself; same prompt, same validation. Needs Meta App Review (2–4 weeks)."
- DEMO.md Setup beat: "Click Use demo screenshot → Read my feed. Say: Spark read the real @bnd.jewelry profile: Korea, women 25–34, $200+, close-up photos lead. The owner confirms or fixes it, then Confirm." Keep timing.
- NOTES.md: one line on the handle-keyed profile cache (same handle → same cached read regardless of the image) and one on the screenshot/instagram seam.

## Verification
`gofmt -l . ; go vet ./... && go test -count=1 ./...` green; `go run ./cmd/sparkcache` (controller) prints the profile line; `curl -s -X POST localhost:PORT/api/profile -d '{"source":"screenshot","handle":"bnd.jewelry","images":["data:image/png;base64,AA=="]}'` returns the cached KR/women/25-34/200+ read.
