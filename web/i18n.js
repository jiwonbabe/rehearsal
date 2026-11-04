// Korean for the whole site. Static copy lives in KO (exact text) and KO_RX (text with numbers);
// Spark sentences come from /api/i18n/ko (spark/cache/ko/, written by Spark once, validated by the server).
// Numbers are never changed: only the words around them. Default language is English.
const LANG_KEY = 'rehearsal.lang';
function getLang() { try { return localStorage.getItem(LANG_KEY) === 'ko' ? 'ko' : 'en'; } catch (e) { return 'en'; } }
function setLang(l) { try { localStorage.setItem(LANG_KEY, l); } catch (e) { /* private window */ } location.reload(); }

const KO = {
  // shell
  'Demo brand · synthetic data': '데모 브랜드 · 합성 데이터', '0 Setup': '0 설정', '1 Directions': '1 방향', '2 Simulate': '2 스트레스 테스트', '3 Verdict': '3 결론', '4 Review': '4 리뷰', '5 Launch': '5 실행',
  'Setup · once': '설정 · 한 번만', 'Directions · this month': '방향 · 이번 달', 'Simulate · stress test': '스트레스 테스트', 'Verdict': '결론', 'Review · October': '리뷰 · 10월', 'Launch': '실행',
  // setup
  "Tell us who you sell to. We'll do the marketing homework.": '누구에게 파는지만 알려 주세요. 마케팅 숙제는 저희가 합니다.',
  'Confirm who your customer is, set a monthly budget, and from then on you only pick a direction each month.': '고객이 누구인지 확인하고 월 예산을 정하면, 그다음부터는 매달 방향 하나만 고르면 됩니다.',
  'Step 1 · Your Instagram': '1단계 · 내 인스타그램', 'Drop a screenshot of your profile (or a few posts)': '프로필(또는 게시물 몇 장) 스크린샷을 올려 주세요', 'Use demo screenshot': '데모 스크린샷 사용', 'Read my feed': '내 피드 읽기',
  'What we read from your feed · Muse Spark': '피드에서 읽은 것 · Muse Spark', 'Check these and fix anything wrong. This decides which directions we suggest.': '틀린 것이 있으면 고쳐 주세요. 이 내용이 제안할 방향을 정합니다.',
  'Connect Instagram · demo account': '인스타그램 연결 · 데모 계정', 'Demo account followers · mock Instagram insights': '데모 계정 팔로워 · 목 인스타그램 인사이트', 'Demo input': '데모 입력', 'Real · your screenshot': '실제 · 내 스크린샷', 'Real · Instagram': '실제 · Instagram', 'Engine · synthetic brand data': '엔진 · 합성 브랜드 데이터', 'Engine · your data': '엔진 · 내 데이터',
  'Known-answer check · only possible on synthetic data': '정답 복원 검사 · 합성 데이터라서 가능', 'The demo data was generated with known rates. The engine never reads them; it learns from the weekly rows alone. Here is what it recovered.': '데모 데이터는 알려진 비율로 생성했습니다. 엔진은 그 비율을 읽지 않고 주간 데이터만 보고 배웁니다. 엔진이 복원한 값입니다.', 'Rate': '비율', 'Generator used': '생성에 쓴 값', 'Engine recovered': '엔진이 복원한 값', 'Miss': '오차', 'generator value scaled by average followers': '평균 팔로워로 보정한 생성값', 'October carries a planted creator no-show': '10월에 크리에이터 노쇼를 심어둠',
  'Engagement per on-hand post': '착용(손) 포스트당 반응', 'Engagement per close-up post': '클로즈업 포스트당 반응', 'Engagement per flat-lay post': '플랫레이 포스트당 반응', 'Engagement per on-neck post': '착용(목) 포스트당 반응', 'Orders per engagement': '반응당 주문', 'Sale week multiplier': '세일 주 배수', 'Creators who post after seeding': '시딩 후 게시하는 크리에이터 비율', 'Engagement per creator post': '크리에이터 포스트당 반응', 'Engagement per ad dollar': '광고 1달러당 반응',
  'Demo: real @bnd.jewelry profile screenshot, synthetic sales data, mock Instagram insights shaped like the Graph API. Instagram Login replaces the mock after Meta App Review.': '데모: 실제 @bnd.jewelry 프로필 스크린샷, 합성 판매 데이터, Graph API 모양의 목 인스타그램 인사이트. Meta 앱 심사 후 Instagram Login이 목 데이터를 대체합니다.',
  'Demo brand': '데모 브랜드', '52 weeks of synthetic data': '합성 데이터 52주', 'Step 2 · Who you sell to': '2단계 · 누구에게 파는지', 'Market': '시장', 'United States': '미국', 'South Korea': '한국', 'Gender': '성별', 'Women': '여성', 'Men': '남성', 'Everyone': '모두', 'Age': '연령', 'Price band': '가격대',
  'Step 3 · What counts as success': '3단계 · 성공의 기준', 'Pick what a good month looks like.': '좋은 한 달이 어떤 모습인지 고르세요.', 'Count success in': '성공 기준 단위', 'orders': '주문', 'revenue': '매출', 'Orders goal per month': '월 주문 목표', 'Revenue goal per month ($)': '월 매출 목표($)', 'Budget per month ($)': '월 예산($)', 'Confirm': '확인',
  'Hold': '유지', 'Grow': '성장', 'Stretch': '도전', '10% more than Hold': '유지보다 10% 더', 'your best 4 weeks': '가장 좋았던 4주',
  'Engine trust check': '엔진 신뢰도 점검', 'Before you trust any plan: how well does the engine know your brand?': '플랜을 믿기 전에: 엔진이 내 브랜드를 얼마나 잘 아는가?', 'Typical miss': '보통 오차', 'how far off the engine was in a typical week': '보통 한 주에 엔진이 얼마나 빗나갔는지', 'Weeks it got right': '맞힌 주', 'its answer was within 20% of what really happened': '실제와 20% 이내로 맞힌 주',
  'Week': '주', 'Orders you got': '실제 주문', 'Engine said': '엔진 답', "Engine's range (low–high)": '엔진 범위(낮음–높음)', 'Result': '결과', 'inside': '범위 안', 'outside': '범위 밖', 'sale week': '세일 주',
  "The smaller the miss, the more the plans on the next pages are worth. The engine's sale-week guesses are its weakest, so sale plans carry a wider range.": '오차가 작을수록 다음 화면의 플랜이 믿을 만합니다. 엔진은 세일 주 추정이 가장 약해서 세일 플랜은 범위가 넓습니다.',
  "Known weakness: at least one sale week landed outside the engine's range.": '알려진 약점: 세일 주 중 적어도 하나가 엔진 범위 밖이었습니다.', "See this month's directions": '이번 달 방향 보기', 'Likely customer': '예상 고객', '(guessed from the photos, no prices seen)': '(사진으로 추측, 가격은 못 봄)', 'Demo screenshot loaded': '데모 스크린샷 불러옴', 'Connected · demo data': '연결됨 · 데모 데이터',
  // directions
  'What October taught us': '10월이 알려준 것', 'Test these': '이 방향들 테스트', 'Full playbook · Muse Spark': '전체 플레이북 · Muse Spark', 'Why for you': '왜 당신에게', 'Caution': '주의', 'Suggested first': '먼저 제안', 'Also fits': '이것도 맞음', 'Select': '선택', 'Selected': '선택됨',
  'Where to post, and why': '어디에 올릴지, 그리고 왜', 'Connect your Instagram on Setup to see where to post for your brand.': '설정에서 인스타그램을 연결하면 내 브랜드에 맞는 채널을 보여 드립니다.', 'Your data': '내 데이터', 'Spark reasoning': 'Spark 판단', 'Lead': '주력', 'Support': '보조', 'Skip': '보류', 'Naver blog': '네이버 블로그', 'Instagram feed': 'Instagram 피드',
  'This month: pick one or two directions': '이번 달: 방향 한두 개 고르기', 'November: pick one or two directions': '11월: 방향 한두 개 고르기',
  // simulate
  'Pick a direction first.': '먼저 방향을 고르세요.', 'Back to directions': '방향으로 돌아가기', 'What this page shows': '이 화면이 보여주는 것',
  'Each row is one way to run a direction: how many posts a week, how many creators, how much on ads.': '각 행은 방향을 실행하는 한 가지 방식입니다: 주당 포스트 수, 크리에이터 수, 광고비.',
  'Each column is a bad month the adversary invented: reach drops, creators do not post, a rival runs a sale, and so on.': '각 열은 공격자가 만든 나쁜 달입니다: 도달 감소, 크리에이터 미게시, 경쟁사 세일 등.',
  'Hover any square to see which bad month it was and the orders and spend the engine computed.': '칸에 마우스를 올리면 어떤 나쁜 달이었는지, 엔진이 계산한 주문과 지출을 볼 수 있습니다.',
  'The version under your cursor': '커서 아래의 버전', 'What the planner (Muse Spark) sent to the engine: settings only, never a number for orders or revenue.': '플래너(Muse Spark)가 엔진에 보낸 것: 설정값만, 주문이나 매출 숫자는 없음.', 'See the verdict': '결론 보기', 'survived': '생존', 'died': '실패',
  // verdict
  'The best variant from each direction you picked, plus the no-spend baseline. Survived = hit your order goal in that many of 20 bad months without running out of budget early.': '고른 방향마다 최고 버전과 돈을 안 쓰는 기준선입니다. 생존 = 나쁜 달 20개 중 그만큼에서 예산을 일찍 소진하지 않고 주문 목표를 달성.',
  'Week 1': '1주차', 'Week 2': '2주차', 'Week 3': '3주차', 'Week 4': '4주차', 'Review opens next month': '다음 달 리뷰 시작', 'Your month, generated on commit': '확정하면 만들어지는 이번 달 계획', 'Make it happen →': '실행하기 →', 'Review October': '10월 리뷰', 'Just post · baseline': '그냥 포스팅 · 기준선', 'Verdict: no paid direction holds this month.': '결론: 이번 달엔 돈 쓰는 방향 중 버티는 게 없습니다.',
  'Commit to this plan': '이 플랜으로 확정', 'The assumption that flips the answer': '답을 뒤집는 가정', 'Pick two directions on the Directions page to see which assumption flips the answer.': '방향 화면에서 두 개를 고르면 어떤 가정이 답을 뒤집는지 보여 드립니다.', 'No creator post rate between 30% and 100% changes this ranking.': '크리에이터 게시율이 30%에서 100% 사이 어디여도 이 순위는 바뀌지 않습니다.', 'What to check this week · Muse Spark': '이번 주 확인할 것 · Muse Spark',
  // review
  'Expected range vs what happened': '예상 범위 vs 실제', 'Diagnosis': '진단', 'What the engine changed': '엔진이 바꾼 것', "Small steps. One month is evidence, not a verdict. Next month's directions are re-ranked with this.": '작은 걸음으로. 한 달은 증거이지 결론이 아닙니다. 다음 달 방향은 이걸 반영해 다시 순위를 매깁니다.', 'Plan November': '11월 계획하기', 'Creator posts': '크리에이터 게시', 'No rate changed.': '바뀐 비율이 없습니다.', 'Orders': '주문',
  'Engagement came in as expected but fewer of it turned into orders. Conversion was off.': '반응은 예상대로 왔지만 주문으로 이어진 비율이 낮았습니다. 전환이 어긋났습니다.', 'Orders were outside the range, but no single rate explains it. Watch next month.': '주문이 범위 밖이었지만 하나의 비율로 설명되지 않습니다. 다음 달을 지켜보세요.',
  // launch
  'Commit a plan on Verdict first.': '먼저 결론 화면에서 플랜을 확정하세요.', 'Back to verdict': '결론으로 돌아가기',
  "Nothing is sent live. The ads campaign is created paused in your Meta ad account; you switch it on in Ads Manager. Every number here comes from the engine; the words come from Muse Spark's playbook for your brand.": '아무것도 바로 집행되지 않습니다. 광고 캠페인은 Meta 광고 계정에 일시정지 상태로 만들어지고, 켜는 건 Ads Manager에서 직접 합니다. 여기 숫자는 전부 엔진에서, 문장은 Muse Spark의 브랜드 플레이북에서 옵니다.',
  'The exact request Rehearsal sends': 'Rehearsal이 보내는 요청 원문', 'Starts': '시작', 'Posts': '포스트', 'Also': '그리고', 'Post idea': '포스트 아이디어', 'Ads · Meta': '광고 · Meta', 'Who': '누구에게', 'Which posts': '어떤 포스트', 'Where': '어디에', 'Create paused campaign in Meta Ads Manager': 'Meta Ads Manager에 일시정지 캠페인 만들기', 'Open Ads Manager': 'Ads Manager 열기',
  'Not connected to a Meta ad account yet. This is the exact request Rehearsal would send. Set META_AD_ACCOUNT_ID and META_ACCESS_TOKEN on the server to send it.': '아직 Meta 광고 계정이 연결되지 않았습니다. 위가 Rehearsal이 보낼 요청 원문입니다. 서버에 META_AD_ACCOUNT_ID와 META_ACCESS_TOKEN을 설정하면 전송됩니다.',
  'Creators · seeding kit': '크리에이터 · 시딩 키트', 'Type these into Instagram search': '인스타그램 검색창에 이렇게 입력하세요', 'Send': '보낼 것', 'Ask': '부탁할 것', 'Copy DM': 'DM 복사', 'Copied': '복사됨', 'Found for your search terms · Instagram': '검색어로 찾은 계정 · Instagram', 'Found for your search terms · Instagram · demo data, fictional accounts': '검색어로 찾은 계정 · Instagram · 데모 데이터, 가상 계정', 'Public data': '공개 통계', 'Your data · demo': '내 데이터 · 데모', 'Found one? Check the fit': '찾았나요? 맞는지 확인', 'Does this creator fit?': '이 크리에이터, 맞을까요?', 'Reading…': '읽는 중…', 'Loading…': '불러오는 중…',
  'Calendar · 4 weeks': '캘린더 · 4주', 'Download calendar (.ics)': '캘린더 다운로드(.ics)', 'no Meta ad account connected': 'Meta 광고 계정 미연결', 'Meta ad account connected': 'Meta 광고 계정 연결됨', 'small': '소형', 'mid': '중형', 'large': '대형', 'high fit': '적합도 높음', 'medium fit': '적합도 보통', 'low fit': '적합도 낮음',
  // shock catalog (engine text)
  'Reach drops 40% (algorithm change)': '도달 40% 감소(알고리즘 변경)', 'The seasonal lift does not come (an average month after all)': '시즌 효과가 오지 않음(결국 평범한 달)', 'Spark': 'Spark', 'November': '11월', 'December': '12월', 'January': '1월', 'February': '2월', 'March': '3월', 'April': '4월', 'May': '5월', 'June': '6월', 'July': '7월', 'August': '8월', 'September': '9월', 'October': '10월', 'Half of seeded creators never post': '시딩한 크리에이터 절반이 게시 안 함', 'A competitor runs a 20% sale': '경쟁사가 20% 세일', 'Ad cost per result up 60%': '광고 성과당 비용 60% 상승', 'Shipping delay in week 3': '3주차 배송 지연', 'One post goes viral': '게시물 하나가 터짐', "Sale week does worse than the engine's weakest estimate": '세일 주가 엔진의 가장 보수적인 추정보다 나쁨', 'never': '없음',
};

const WORDS = { 'on hand': '착용(손)', 'close up': '클로즈업', 'flat lay': '플랫레이', 'on neck': '착용(목)', 'on-hand': '착용(손)', 'close-up': '클로즈업', 'flat-lay': '플랫레이', 'on-neck': '착용(목)', women: '여성', men: '남성', everyone: '모두', 'or better': '이상', 'or worse': '이하', 'inside range': '범위 안', 'above range': '범위 위', 'below range': '범위 아래', high: '높음', medium: '보통', low: '낮음', unsure: '불확실' };
const w = s => WORDS[s] || s;
const mixKo = s => s.split(', ').map(p => p.replace(/^(\d+%) (.+)$/, (m, a, b) => `${w(b)} ${a}`)).join(', ');
const shock = s => s.split(' + ').map(x => KO[x] || x).join(' + ');
const planKo = s => s.replace(/^post (\S+)×\/wk \((.+?)\), (\d+) creators, (\$[\d,]+) ads(?:, (\d+%) off in week (\d+))?$/, (m, a, mix, c, ads, off, wk) => `주 ${a}회 포스트(${mixKo(mix)}), 크리에이터 ${c}명, 광고 ${ads}` + (off ? `, ${wk}주차 ${off} 할인` : ''));

// Text with numbers in it: the numbers are captured and put back untouched.
const KO_RX = [
  [/^We hid the last (\d+) weeks from the engine\. .* really got\.$/, (m, n) => `마지막 ${n}주를 엔진에게 숨겼습니다. 엔진은 그 앞의 주들로 비율을 배운 뒤, 숨긴 주들의 주문을(실제로 한 포스트·광고·크리에이터·세일 주를 주고) 답했고, 그 답을 실제 주문과 비교했습니다.`],
  [/^If the engine may also learn from those (\d+) weeks \(in-sample\), the miss is (\S+)% and it gets (\d+) of (\d+) right\. The holdout number above is the honest one\.$/, (m, n, e, a, b) => `엔진이 그 ${n}주도 보고 배우게 하면(인샘플) 오차 ${e}%, ${b}주 중 ${a}주 적중입니다. 위의 홀드아웃 숫자가 정직한 쪽입니다.`],
  [/^Connected · demo account @(.+) · mock numbers, not your real account$/, (m, h) => `연결됨 · 데모 계정 @${h} · 목 숫자, 실제 계정 아님`],
  [/^Reasoned from the demo account @(.+)'s mock Instagram and Threads data \((\d+) days; not your real numbers\) and sourced public statistics · not stress-tested$/, (m, h, d) => `데모 계정 @${h}의 목 인스타그램·스레드 데이터(${d}일, 실제 숫자 아님)와 출처 있는 공개 통계로 판단 · 스트레스 테스트 대상 아님`],
  [/^Most of your photos: (.+)$/, (m, a) => `사진 대부분: ${w(a)}`], [/^Confidence: (.+)$/, (m, a) => `확신도: ${w(a)}`],
  [/^(\S+) followers · (\d+%) are (\S+) · (\d+%) women · (\d+%) in (.+)$/, (m, f, a, band, wm, c, city) => `팔로워 ${f}명 · ${band}가 ${a} · 여성 ${wm} · ${city} ${c}`],
  [/^Last (\d+) days: Reels reach (.+)× your feed posts · (.+) get the most saves · Threads (\d+) replies per post$/, (m, d, r, t, th) => `최근 ${d}일: 릴스 도달이 피드의 ${r}배 · 저장은 ${w(t)}이 가장 많음 · 스레드 글당 답글 ${th}개`],
  [/^about what you did lately \((.+) a week\)$/, (m, a) => `최근 수준 유지(주 ${a}건)`], [/^(\S+) orders · (.+)$/, (m, n, rest) => `${n}건 · ${tr(rest)}`], [/^(\S+) orders$/, (m, n) => `${n}건`],
  [/^You spent about (\$[\d,]+) a month lately\.$/, (m, a) => `최근 한 달 평균 약 ${a} 썼습니다.`], [/^Suggested: (\w+) on (\w+) — (.+)$/, (m, a, u, why) => `제안: ${KO[u] || u} 기준 ${KO[a] || a} — ${tx(why)}`],
  [/^(\$[\d,]+)–(\$[\d,]+) ads · (\d+)–(\d+) creators · (\d+)–(\d+) posts\/wk$/, (m, a, b, c, d, e, f) => `광고 ${a}–${b} · 크리에이터 ${c}–${d}명 · 주 ${e}–${f}회 포스트`],
  [/^You would do: (.+)$/, (m, a) => `할 일: ${tx(a)}`], [/^Week (\d+):$/, (m, n) => `${n}주차:`], [/^Week (\d+)$/, (m, n) => `${n}주차`], [/^Post idea: (.+)$/, (m, a) => `포스트 아이디어: ${tx(a)}`], [/^Week 1: (.+)$/, (m, a) => `1주차: ${tx(a)}`], [/^Ads: (.+)$/, (m, a) => `광고: ${tx(a)}`], [/^Creators: (.+)$/, (m, a) => `크리에이터: ${tx(a)}`],
  [/^(\d+) directions? selected · the engine will build (\d+) variants and attack each with 20 bad months\.$/, (m, n, v) => `${n}개 방향 선택 · 엔진이 ${v}개 변형을 만들어 각각 나쁜 달 20개로 공격합니다.`],
  [/^Reasoned from @(.+)'s Instagram and Threads data \((\d+) days(, demo data)?\) and sourced public statistics · not stress-tested$/, (m, h, d, demo) => `@${h}의 인스타그램·스레드 데이터(${d}일${demo ? ', 데모 데이터' : ''})와 출처 있는 공개 통계로 판단 · 스트레스 테스트 대상 아님`], [/^Source: (.+), (\d{4})$/, (m, a, y) => `출처: ${a}, ${y}`],
  [/^Orders (\d+) vs (\d+)–(\d+)$/, (m, a, b, c) => `주문 ${a}건 vs 예상 ${b}–${c}건`],
  [/^Blue square: that version still reached your goal of (\d+) orders without running past the (\$[\d,]+) budget\. Red square: it did not\.$/, (m, g, b) => `파란 칸: 그 버전이 예산 ${b}를 넘기지 않고 목표 ${g}건을 달성. 빨간 칸: 실패.`],
  [/^(\d+) ways to run it × (\d+) bad months$/, (m, a, b) => `실행 방식 ${a}개 × 나쁜 달 ${b}개`], [/^Best version survived (\d+) of (\d+) bad months$/, (m, a, b) => `최고 버전은 나쁜 달 ${b}개 중 ${a}개에서 살아남음`],
  [/^(\d+) runs · deterministic · same input, same output$/, (m, n) => `${n}회 실행 · 결정적 · 같은 입력이면 같은 출력`], [/^Bad months, chosen by the adversary \(Muse Spark\) · (.+)$/, (m, t) => `공격자(Muse Spark)가 고른 나쁜 달 · ${tx(t)}`],
  [/^(\d+)\. (.+)$/, (m, n, s) => `${n}. ${shock(s)}`], [/^happens in about (\d+%) of months$/, (m, p) => `달의 약 ${p}에서 발생`],
  [/^Post (\S+)× \((.+)\)$/, (m, n, mix) => `포스트 ${n}회(${mixKo(mix)})`], [/^Post (\S+)×$/, (m, n) => `포스트 ${n}회`], [/^Pick your 2 best on-hand posts and set a daily cap of (\$[\d.,]+)$/, (m, a) => `가장 좋은 착용샷 2개를 골라 일 한도 ${a} 설정`], [/^DM and ship (\d+) pieces$/, (m, n) => `DM 보내고 ${n}개 발송`], [/^Announce (\d+%) off$/, (m, p) => `${p} 할인 공지`],
  [/^Verdict: (.+) holds this month\.$/, (m, t) => `결론: ${tx(t)}, 이번 달 버팁니다.`], [/^(.+) · survived (\d+) \/ (\d+)$/, (m, t, a, b) => `${tx(t)} · ${b}번 중 ${a}번 생존`], [/^Orders, bad month (\d+)–(\d+)$/, (m, a, b) => `나쁜 달 주문 ${a}–${b}건`], [/^Spend (\$[\d,]+)$/, (m, a) => `지출 ${a}`], [/^Dies if (.+)$/, (m, s) => `죽는 조건: ${shock(s)}`],
  [/^If creators post at (\d+%) (or better|or worse), (.+) beats (.+)\. At your current (\d+%), (.+) wins\.$/, (m, p, dir, a, b, cur, win) => `크리에이터 게시율이 ${p} ${w(dir)}이면 ${tx(a)}이(가) ${tx(b)}을(를) 이깁니다. 현재 ${cur}에서는 ${tx(win)}이(가) 이깁니다.`],
  [/^October: (inside range|above range|below range)\. Here is what was off\.$/, (m, r) => `10월: ${w(r)}. 무엇이 어긋났는지 봅시다.`], [/^You ran (.+)\.$/, (m, p) => `실행한 플랜: ${planKo(p)}.`], [/^(\d+) vs (\d+)–(\d+)$/, (m, a, b, c) => `${a}건 vs ${b}–${c}건`], [/^(\d+) of (\d+)$/, (m, a, b) => `${b}명 중 ${a}명`],
  [/^Only (\d+) of (\d+) seeded creators posted, so the seeded engagement never arrived\. Your own posts and conversion held up\. This was a creator problem, not a product or audience problem\.$/, (m, a, b) => `시딩한 크리에이터 ${b}명 중 ${a}명만 게시해서 기대한 반응이 오지 않았습니다. 자체 포스트와 전환은 괜찮았습니다. 제품이나 고객 문제가 아니라 크리에이터 문제였습니다.`],
  [/^(\$[\d,]+) over (\d+) days · (\$[\d,]+)\/day · paused until you switch it on$/, (m, a, d, b) => `${d}일간 ${a} · 하루 ${b} · 켜기 전까지 일시정지`],
  [/^Meta estimate for this audience: (\S+) reachable on Reels, (\S+) on feed, (\S+) on Stories \(monthly\)$/, (m, a, b, c) => `이 타겟에 대한 Meta 추정: 릴스 월 ${a}명, 피드 ${b}명, 스토리 ${c}명 도달 가능`],
  [/^Created paused: campaign (.+), ad set (.+)\. $/, (m, a, b) => `일시정지 상태로 생성: 캠페인 ${a}, 광고세트 ${b}. `],
  [/^Ship (\d+) pieces \((\$[\d,]+) at cost\) · send (\d+) DMs, expect about half to answer$/, (m, n, c, d) => `${n}개 발송(원가 ${c}) · DM ${d}개 보내기, 절반쯤 답한다고 가정`],
  [/^([\d,]+) followers · (\w+) · ([\d.]+)% engagement · mostly (.+)$/, (m, f, b, e, t) => `팔로워 ${f} · ${KO[b] || b} · 참여율 ${e}% · 주로 ${w(t)}`], [/^(high|medium|low|unsure) fit · (.*) account$/, (m, f, b) => `적합도 ${w(f)} · ${KO[b] || b} 계정`],
  [/^ship (\d+) pieces$/, (m, n) => `${n}개 발송`], [/^(\$[\d,]+) ads$/, (m, a) => `광고 ${a}`], [/^(\d+%) off$/, (m, p) => `${p} 할인`], [/^(\d+) \((.+)\)$/, (m, n, mix) => `${n} (${mix.split(', ').map(x => x.replace(/^(\d+) (.+)$/, (mm, k, t) => `${w(t)} ${k}`)).join(', ')})`],
  [/^Launch: (.+)$/, (m, t) => `실행: ${tx(t)}`], [/^Starts (\S+) · (\d+) days · total (\$[\d,]+) · (.+)$/, (m, d, n, tot, c) => `시작 ${d} · ${n}일 · 총 ${tot} · ${KO[c] || c}`],
  [/^Connected · @(.+)$/, (m, h) => `연결됨 · @${h}`],
  [/^This month: (\w+)$/, (m, n) => `이번 달: ${KO[n] || n}`], [/^demand (\S+)× · ad cost (\S+)× · 1\.00 = an average month$/, (m, d, a) => `수요 ${d}배 · 광고 단가 ${a}배 · 1.00 = 평범한 달`],
  [/^Seasonal prior from public data \((\w+) (\w+)\), applied to the order rate and ad reach for this month only; .*$/, (m, mk, cat) => `공개 데이터 기반 계절 지수(${mk} ${cat}). 이번 달의 주문율과 광고 도달에만 적용하고, 공격자는 이 달이 평범하게 끝나는 경우를 시험합니다. 과거 이력에서 배운 값이 아닙니다: 1년치 데이터로는 계절을 분리할 수 없습니다.`], [/^Source \((\w+)\): (.+), (\d{4})$/, (m, f, a, y) => `출처(${f === 'demand' ? '수요' : '광고 단가'}): ${a}, ${y}`], [/^(\d+)% (on hand|close up|flat lay|on neck)$/, (m, p, t) => `${w(t)} ${p}%`],
];

let SPARK = {};
let I18N_READY = Promise.resolve();
function tx(s) { return SPARK[s] || s; }
function tr(s) {
  const k = s.trim();
  if (!k) return s;
  const hit = KO[k] || SPARK[k];
  if (hit) return s.replace(k, hit);
  for (const [rx, fn] of KO_RX) { const m = rx.exec(k); if (m) return s.replace(k, fn(...m)); }
  const i = k.indexOf(': ');
  if (i > 0) { const a = k.slice(0, i), b = k.slice(i + 2); const ta = KO[a], tb = KO[b] || SPARK[b]; if (ta || tb) return s.replace(k, `${ta || a}: ${tb || b}`); }
  return s;
}

const done = new WeakSet();
function translateNode(n) {
  if (n.nodeType === 3) {
    if (done.has(n) || !/[A-Za-z]/.test(n.nodeValue)) return;
    const out = tr(n.nodeValue);
    if (out !== n.nodeValue) { n.nodeValue = out; done.add(n); }
    return;
  }
  if (n.nodeType !== 1 || n.tagName === 'SCRIPT' || n.tagName === 'STYLE' || n.tagName === 'PRE' || n.tagName === 'TEXTAREA') return;
  if (n.placeholder) n.placeholder = tr(n.placeholder);
  n.childNodes.forEach(translateNode);
}
function translateDOM() { translateNode(document.body); }

if (getLang() === 'ko') {
  document.documentElement.lang = 'ko';
  new MutationObserver(muts => muts.forEach(m => m.addedNodes.forEach(translateNode))).observe(document.documentElement, { childList: true, subtree: true });
  document.addEventListener('DOMContentLoaded', translateDOM);
  I18N_READY = fetch('/api/i18n/ko').then(r => r.json()).then(m => { SPARK = m || {}; translateDOM(); }).catch(() => {});
}
