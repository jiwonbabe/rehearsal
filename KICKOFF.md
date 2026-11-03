# Claude Code 시작 프롬프트 (복사해서 첫 메시지로)

이 폴더는 Meta Global AI Developer Hackathon 제출용 "Rehearsal" 프로젝트야. 제출 마감 11/4(수) 6pm PT.
언어는 Go 1.22, stdlib만. 먼저 CLAUDE.md와 SPEC.md를 읽고 아래 순서대로. 각 단계 끝나면
`gofmt -l . && go vet ./... && go test ./...` 통과 확인하고 커밋한 뒤 다음으로.

## 현재 상태 (이미 돌아감)

- `data/synthetic.py` → `data/brand.json` (52주 합성 브랜드, 10월에 시딩 노쇼 심어둠)
- `engine/` Go 엔진 완성: Calibrate, RunPlan, shocks, Stress, RunVerdict, RunFlip, RunBacktest, RunReview, Summary.
  `reference/`의 Python 원본과 골든 픽스처(`reference/fixtures/*.json`)에 대해 `go test ./...` 전부 통과
- `spark/client.go`: OpenAI 호환 호출 + 디스크 캐시 (키 없으면 캐시만). MUSE_BASE_URL/MUSE_MODEL은
  dev.meta.ai 문서 보고 환경변수로 맞출 것 (기본값은 추정치)
- `cmd/server`: /api/brand, /api/rates, /api/backtest, /api/summary 동작. `go run ./cmd/server`
- `prompts/directions.md`, `prompts/adversary.md`
- `mockup/`: 화면 6개 HTML (디자인 그대로 따라도 됨)
- 확인된 숫자: 백테스트 중앙 오차 9.5% (13주 중 8주 범위 안), 광고 방향 13/20 생존, 시딩 12/20,
  flip 임계값 seed_post_rate 0.7 (현재 0.6 → 10월 리뷰 후 0.49)

## 할 일

1. Spark 호출 캐시: `/api/summary` 결과로 `spark.Directions()` 1회, 방향별 `spark.Adversary()` 1회씩.
   `spark/validate.go` 추가: ranges가 예산 안인지, 쇼크 코드가 카탈로그 안인지 검증, 틀리면 기본값으로
   대체. 캐시 파일(`spark/cache/*.json`)을 커밋. 호출이 안 되면 `prompts/`의 출력 예시 모양으로 캐시
   파일을 손으로 만들어 두고 NOTES.md에 적어.
2. `cmd/server`에 `/api/directions` (캐시 읽기), `POST /api/stress` ({direction_ids} → 방향별
   SampleVariants 60개, Stress, RunVerdict, RunFlip), `POST /api/review` ({plan} → RunReview, 서버의
   rates를 new_rates로 교체하는 `/api/review/apply`).
3. `web/` 화면 0–2 (Setup, Directions, Simulate). 상단 "Demo brand · synthetic data" 배지. 타깃
   프로필은 `/api/brand`에서 미리 채움.
4. 화면 3–4 (Verdict, Review). Review의 "Plan November" 버튼 → review/apply → Directions가 새
   rates로 재랭킹·재스트레스. **시딩 카드에 caution이 붙고 광고가 1위로 올라오는 게 눈에 보여야 함.**
   안 바뀌면 `data/synthetic.py` 상수를 조정하고 `python reference/dump_fixtures.py` 다시 → go test.
5. README.md (한 문단, 아키텍처: 모델은 양 끝·엔진은 가운데·사람은 마지막, 백테스트 수치, 오라클 격리,
   두 구현이 같은 숫자를 낸다는 것, 합성 데이터 명시, Spark 역할 2개, Next = SPEC §8) + DEMO.md
   (2분 30초 녹화 순서: Setup → Directions → Simulate → Verdict → Review → Directions 순위 변경. API
   라이브 호출 없음, 전부 캐시).

## 심사 규칙에서 온 추가 사항 (participant guide 기준)

- **첫 작업이 `git init` + 첫 커밋.** 규칙상 해커톤 기간(10/27–11/4)에 만든 것만 인정. 커밋 히스토리가 증거.
  README에 "Built Nov 2–4, 2026 during the hackathon" 명시. 외부 코드 재사용 없음.
- Spark 세 번째 역할 추가 (호출 1개): Verdict의 flip 결과(JSON)를 넘겨 "이번 달 확인할 것 한 줄"을 쓰게.
  숫자는 엔진 값 그대로 인용만. 프롬프트는 `prompts/check.md`로. 공격자의 `why`도 Simulate 화면에 표시.
- 배포: Go 바이너리를 Fly.io 또는 Railway에 올려 공개 URL 확보 (심사가 "live product"도 봄). 5번 단계에 포함.
- README에 "agent loop" (planner → engine → adversary → engine → verdict → human)를 그 단어로 명시.
  Best Agentic Use-Case 보너스 대상. 모델 호출은 Meta Model API 기본, Baseten의
  `meta-models/muse-spark-open-1026`를 같은 프롬프트로 폴백 (환경변수만 교체). 둘 다 README에 기록.
- 영상 3분 제한. 첫 20초: 실제 문제(언니) + 누가 이득인지 한 줄. "forecast" 금지는 그대로.

## 지킬 것

- 모델이 숫자를 내는 코드 쓰지 마. 숫자는 engine/에서만.
- "forecast", "prediction" 단어 금지. stress test / survived / dies if / flip assumption.
- 막히면 NOTES.md에 한 줄 적고 더 단순한 쪽으로. 나한테 물어보지 말고.
