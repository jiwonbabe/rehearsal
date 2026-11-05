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
  'Connect Instagram · demo account': '인스타그램 연결 · 데모 계정', 'Demo account followers · mock Instagram insights': '데모 계정 팔로워 · 목 인스타그램 인사이트', 'Demo input': '데모 입력', 'Demo input · mock profile screenshot': '데모 입력 · 목 프로필 스크린샷', 'Real · your screenshot': '실제 · 내 스크린샷', 'Real · Instagram': '실제 · Instagram', 'Engine · synthetic brand data': '엔진 · 합성 브랜드 데이터', 'Engine · your data': '엔진 · 내 데이터',
  'Known-answer check · only possible on synthetic data': '정답 복원 검사 · 합성 데이터라서 가능', 'The demo data was generated with known rates. The engine never reads them; it learns from the weekly rows alone. Here is what it recovered.': '데모 데이터는 알려진 비율로 생성했습니다. 엔진은 그 비율을 읽지 않고 주간 데이터만 보고 배웁니다. 엔진이 복원한 값입니다.', 'Rate': '비율', 'Generator used': '생성에 쓴 값', 'Engine recovered': '엔진이 복원한 값', 'Miss': '오차', 'generator value scaled by average followers': '평균 팔로워로 보정한 생성값', 'October carries a planted creator no-show': '10월에 크리에이터 노쇼를 심어둠',
  'Engagement per on-hand post': '착용(손) 포스트당 반응', 'Engagement per close-up post': '클로즈업 포스트당 반응', 'Engagement per flat-lay post': '플랫레이 포스트당 반응', 'Engagement per on-neck post': '착용(목) 포스트당 반응', 'Orders per engagement': '반응당 주문', 'Sale week multiplier': '세일 주 배수', 'Creators who post after seeding': '시딩 후 게시하는 크리에이터 비율', 'Engagement per creator post': '크리에이터 포스트당 반응', 'Engagement per ad dollar': '광고 1달러당 반응',

  'Demo brand': '데모 브랜드', '52 weeks of synthetic data': '합성 데이터 52주', 'Step 2 · Who you sell to': '2단계 · 누구에게 파는지', 'Market': '시장', 'What you sell': '파는 것', 'Jewelry': '주얼리', 'Clothing (apparel and shoes)': '의류·신발', 'Beauty, health and personal care': '뷰티·건강·퍼스널케어', 'Furniture and home': '가구·홈', 'Electronics and appliances': '전자·가전', 'Sporting goods, hobby, music and books': '스포츠·취미·악기·도서', 'Food and drink': '식음료', 'Online store, other goods': '온라인 스토어·기타', 'United States': '미국', 'South Korea': '한국', 'Gender': '성별', 'Women': '여성', 'Men': '남성', 'Everyone': '모두', 'Age': '연령', 'Price band': '가격대',
  'Step 3 · What counts as success': '3단계 · 성공의 기준', 'Pick what a good month looks like.': '좋은 한 달이 어떤 모습인지 고르세요.', 'Count success in': '성공 기준 단위', 'orders': '주문', 'revenue': '매출', 'Orders goal per month': '월 주문 목표', 'Revenue goal per month ($)': '월 매출 목표($)', 'Budget per month ($)': '월 예산($)', 'Confirm': '확인',
  'Hold': '유지', 'Grow': '성장', 'Stretch': '도전', '10% more than Hold': '유지보다 10% 더', 'your best 4 weeks': '가장 좋았던 4주',
  'Engine trust check': '엔진 신뢰도 점검', 'Before you trust any plan: how well does the engine know your brand?': '플랜을 믿기 전에: 엔진이 내 브랜드를 얼마나 잘 아는가?', 'Typical miss': '보통 오차', 'how far off the engine was in a typical week': '보통 한 주에 엔진이 얼마나 빗나갔는지', 'Weeks it got right': '맞힌 주', 'its answer was within 20% of what really happened': '실제와 20% 이내로 맞힌 주',
  'Week': '주', 'Orders you got': '실제 주문', 'Engine said': '엔진 답', "Engine's range (low–high)": '엔진 범위(낮음–높음)', 'Result': '결과', 'inside': '범위 안', 'outside': '범위 밖', 'sale week': '세일 주',
  "The smaller the miss, the more the plans on the next pages are worth. The engine's sale-week guesses are its weakest, so sale plans carry a wider range.": '오차가 작을수록 다음 화면의 플랜이 믿을 만합니다. 엔진은 세일 주 추정이 가장 약해서 세일 플랜은 범위가 넓습니다.',
  "Known weakness: at least one sale week landed outside the engine's range.": '알려진 약점: 세일 주 중 적어도 하나가 엔진 범위 밖이었습니다.', "See this month's directions": '이번 달 방향 보기', 'Likely customer': '예상 고객', '(guessed from the photos, no prices seen)': '(사진으로 추측, 가격은 못 봄)', 'Demo screenshot loaded': '데모 스크린샷 불러옴', 'Connected · demo data': '연결됨 · 데모 데이터',
  // directions
  'What October taught us': '10월이 알려준 것', 'Test these': '이 방향들 테스트', 'Full playbook · Muse Spark': '전체 플레이북 · Muse Spark', 'Why for you': '왜 당신에게', 'Caution': '주의', 'Suggested first': '먼저 제안', 'Also fits': '이것도 맞음', 'Select': '선택', 'Selected': '선택됨',
  'Where to post, and why': '어디에 올릴지, 그리고 왜', 'Connect your Instagram on Setup to see where to post for your brand.': '설정에서 인스타그램을 연결하면 내 브랜드에 맞는 채널을 보여 드립니다.', 'Your data': '내 데이터', 'Spark reasoning': 'Spark 판단', 'Lead': '주력', 'Support': '보조', 'Skip': '보류', 'Instagram feed': 'Instagram 피드',
  'This month: pick one or two directions': '이번 달: 방향 한두 개 고르기', 'No data': '데이터 없음',
  // simulate
  'Pick a direction first.': '먼저 방향을 고르세요.', 'Back to directions': '방향으로 돌아가기', 'What this page shows': '이 화면이 보여주는 것',
  'Each row is one way to run a direction: how many posts a week, how many creators, how much on ads.': '각 행은 방향을 실행하는 한 가지 방식입니다: 주당 포스트 수, 크리에이터 수, 광고비.',
  'Each column is a bad month the adversary invented: reach drops, creators do not post, a rival runs a sale, and so on.': '각 열은 공격자가 만든 나쁜 달입니다: 도달 감소, 크리에이터 미게시, 경쟁사 세일 등.',
  'Hover any square to see which bad month it was and the orders and spend the engine computed.': '칸에 마우스를 올리면 어떤 나쁜 달이었는지, 엔진이 계산한 주문과 지출을 볼 수 있습니다.',
  'The version under your cursor': '커서 아래의 버전', "The 13 hidden weeks, in order · blue = your orders landed inside the engine's range, red = outside": '숨긴 13주, 순서대로 · 파랑 = 실제 주문이 엔진 범위 안, 빨강 = 밖', 'Both together, budget split': '함께 쓰기, 예산 나눔', 'Check': '확인', 'Have your own creator list? Paste it (CSV or JSON from a public-profile export, e.g. Bright Data)': '내 크리에이터 목록이 있나요? 붙여넣기(공개 프로필 내보내기 CSV/JSON, 예: Bright Data)', 'Columns: username, followers, and engagement_rate or avg_likes (biography optional). The list becomes the creators every plan is counted against and the list Launch shows. Nothing here is demo data.': '열: username, followers, 그리고 engagement_rate 또는 avg_likes(biography 선택). 이 목록이 모든 플랜의 크리에이터 집계 기준이 되고 Launch에 그대로 보입니다. 데모 데이터가 아닙니다.', 'Use these creators': '이 크리에이터들로', 'Real · your list': '실제 · 내 목록', 'Your imported creator list · real public profiles': '가져온 내 크리에이터 목록 · 실제 공개 프로필', 'Who to look for, and what they are worth': '누구를 찾고, 그 사람은 얼마나 되는가', 'Not rated': '평가 안 함', 'not rated': '평가 안 함', 'Direction': '방향', 'Posts per week': '주당 포스트', 'Photo mix': '사진 구성', 'Creators seeded': '시딩 크리에이터', 'Ad budget': '광고 예산', 'Sale': '세일', 'Month': '달', 'none': '없음', 'not set': '미지정', 'Just post': '그냥 포스팅', 'Pay for reach': '도달에 돈 쓰기', 'Send to creators': '크리에이터에게 보내기', 'One-week sale': '일주일 세일', 'What the planner (Muse Spark) sent to the engine: settings only, never a number for orders or revenue.': '플래너(Muse Spark)가 엔진에 보낸 것: 설정값만, 주문이나 매출 숫자는 없음.', 'See the verdict': '결론 보기', 'survived': '생존', 'died': '실패',
  // verdict
  'The best variant from each direction you picked, plus the no-spend baseline. Survived = hit your order goal in that many of 20 bad months without running out of budget early.': '고른 방향마다 최고 버전과 돈을 안 쓰는 기준선입니다. 생존 = 나쁜 달 20개 중 그만큼에서 예산을 일찍 소진하지 않고 주문 목표를 달성.',
  'Week 1': '1주차', 'Week 2': '2주차', 'Week 3': '3주차', 'Week 4': '4주차', 'Review opens next month': '다음 달 리뷰 시작', 'Your month, generated on commit': '확정하면 만들어지는 이번 달 계획', 'Make it happen →': '실행하기 →', 'Review October': '10월 리뷰', 'Just post · baseline': '그냥 포스팅 · 기준선', 'Verdict: no paid direction holds this month.': '결론: 이번 달엔 돈 쓰는 방향 중 버티는 게 없습니다.',
  'Commit to this plan': '이 플랜으로 확정', 'The assumption that flips the answer': '답을 뒤집는 가정', 'Pick two directions on the Directions page to see which assumption flips the answer.': '방향 화면에서 두 개를 고르면 어떤 가정이 답을 뒤집는지 보여 드립니다.', 'No creator post rate between 30% and 100% changes this ranking.': '크리에이터 게시율이 30%에서 100% 사이 어디여도 이 순위는 바뀌지 않습니다.', 'What to check this week · Muse Spark': '이번 주 확인할 것 · Muse Spark',
  // review
  'Expected range vs what happened': '예상 범위 vs 실제', 'Diagnosis': '진단', 'What the engine changed': '엔진이 바꾼 것', "Small steps. One month is evidence, not a verdict. Next month's directions are re-ranked with this.": '작은 걸음으로. 한 달은 증거이지 결론이 아닙니다. 다음 달 방향은 이걸 반영해 다시 순위를 매깁니다.', 'Creator posts': '크리에이터 게시', 'No rate changed.': '바뀐 비율이 없습니다.', 'Orders': '주문',
  'Engagement came in as expected but fewer of it turned into orders. Conversion was off.': '반응은 예상대로 왔지만 주문으로 이어진 비율이 낮았습니다. 전환이 어긋났습니다.', 'Orders were outside the range, but no single rate explains it. Watch next month.': '주문이 범위 밖이었지만 하나의 비율로 설명되지 않습니다. 다음 달을 지켜보세요.',
  // launch
  'Commit a plan on Verdict first.': '먼저 결론 화면에서 플랜을 확정하세요.', 'Back to verdict': '결론으로 돌아가기',
  "Nothing is sent live. The ads campaign is created paused in your Meta ad account; you switch it on in Ads Manager. Every number here comes from the engine; the words come from Muse Spark's playbook for your brand.": '아무것도 바로 집행되지 않습니다. 광고 캠페인은 Meta 광고 계정에 일시정지 상태로 만들어지고, 켜는 건 Ads Manager에서 직접 합니다. 여기 숫자는 전부 엔진에서, 문장은 Muse Spark의 브랜드 플레이북에서 옵니다.',
  'The exact request Rehearsal sends': 'Rehearsal이 보내는 요청 원문', 'Starts': '시작', 'Posts': '포스트', 'Also': '그리고', 'Post idea': '포스트 아이디어', 'Ads · Meta': '광고 · Meta', 'Who': '누구에게', 'Which posts': '어떤 포스트', 'Where': '어디에', 'Create paused campaign in Meta Ads Manager': 'Meta Ads Manager에 일시정지 캠페인 만들기', 'Open Ads Manager': 'Ads Manager 열기',
  'Not connected to a Meta ad account yet. This is the exact request Rehearsal would send. Set META_AD_ACCOUNT_ID and META_ACCESS_TOKEN on the server to send it.': '아직 Meta 광고 계정이 연결되지 않았습니다. 위가 Rehearsal이 보낼 요청 원문입니다. 서버에 META_AD_ACCOUNT_ID와 META_ACCESS_TOKEN을 설정하면 전송됩니다.',
  'Creators · seeding kit': '크리에이터 · 시딩 키트', 'Type these into Instagram search': '인스타그램 검색창에 이렇게 입력하세요', 'Send': '보낼 것', 'Ask': '부탁할 것', 'Copy DM': 'DM 복사', 'Copied': '복사됨', 'Found for your search terms · Instagram': '검색어로 찾은 계정 · Instagram', 'Found for your search terms · Instagram · demo data, fictional accounts': '검색어로 찾은 계정 · Instagram · 데모 데이터, 가상 계정', 'Public data': '공개 통계', 'Your data · demo': '내 데이터 · 데모', 'Found one? Check the fit': '찾았나요? 맞는지 확인', 'Does this creator fit?': '이 크리에이터, 맞을까요?', 'Reading…': '읽는 중…', 'Loading…': '불러오는 중…',
  'Calendar · 4 weeks': '캘린더 · 4주', 'Download calendar (.ics)': '캘린더 다운로드(.ics)', 'no Meta ad account connected': 'Meta 광고 계정 미연결', 'Meta ad account connected': 'Meta 광고 계정 연결됨', 'small': '소형', 'mid': '중형', 'large': '대형', 'high fit': '적합도 높음', 'medium fit': '적합도 보통', 'low fit': '적합도 낮음',
  // feedback and state copy
  'Pick one or two directions to test.': '테스트할 방향을 하나나 둘 고르세요.', 'Up to two directions at a time. Unselect one first.': '한 번에 두 개까지입니다. 하나를 먼저 해제하세요.', 'Show why, channel by channel': '채널별 이유 보기', 'Show the 13 hidden weeks': '숨긴 13주 보기', 'Show all nine rates': '비율 9개 전부 보기',
  'Not confirmed yet: the engine still uses the last confirmed values. Click Confirm before moving on.': '아직 확인 전입니다: 엔진은 마지막으로 확인한 값을 씁니다. 넘어가기 전에 확인을 누르세요.', 'Remembered from your last read': '지난 읽기 결과를 기억해 두었습니다', 'Confirmed': '확인됨',
  'Baseline committed: keep posting, spend nothing': '기준선 확정: 계속 올리고 돈은 안 씀', 'This is the do-nothing-new choice. Launch will only give you the posting calendar.': '새로 하는 건 없는 선택입니다. 실행 화면에는 포스팅 캘린더만 나옵니다.', 'Type the handle and add 1–8 screenshots of their profile or posts, then check.': '핸들을 적고 프로필이나 게시물 스크린샷 1–8장을 올린 뒤 확인하세요.', 'Could not build the calendar': '캘린더를 만들지 못했습니다',
  // shock catalog (engine text)
  'Reach drops 40% (algorithm change)': '도달 40% 감소(알고리즘 변경)', 'The seasonal lift does not come (an average month after all)': '시즌 효과가 오지 않음(결국 평범한 달)', 'Spark': 'Spark', 'November': '11월', 'December': '12월', 'January': '1월', 'February': '2월', 'March': '3월', 'April': '4월', 'May': '5월', 'June': '6월', 'July': '7월', 'August': '8월', 'September': '9월', 'October': '10월', 'Half of seeded creators never post': '시딩한 크리에이터 절반이 게시 안 함', 'A competitor runs a 20% sale': '경쟁사가 20% 세일', 'Ad cost per result up 60%': '광고 성과당 비용 60% 상승', 'Shipping delay in week 3': '3주차 배송 지연', 'One post goes viral': '게시물 하나가 터짐', "Sale week does worse than the engine's weakest estimate": '세일 주가 엔진의 가장 보수적인 추정보다 나쁨', 'never': '없음',
};

const WORDS = { 'on hand': '착용(손)', 'close up': '클로즈업', 'flat lay': '플랫레이', 'on neck': '착용(목)', 'on-hand': '착용(손)', 'close-up': '클로즈업', 'flat-lay': '플랫레이', 'on-neck': '착용(목)', women: '여성', men: '남성', everyone: '모두', 'or better': '이상', 'or worse': '이하', 'inside range': '범위 안', 'above range': '범위 위', 'below range': '범위 아래', high: '높음', medium: '보통', low: '낮음', unsure: '불확실' };
const w = s => WORDS[s] || s;
const LV = { ads: '광고', creators: '크리에이터', 'a sale': '세일' };
const mixKo = s => s.split(', ').map(p => p.replace(/^(\d+%) (.+)$/, (m, a, b) => `${w(b)} ${a}`)).join(', ');
const shock = s => s.split(' + ').map(x => KO[x] || x).join(' + ');
const planKo = s => s.replace(/^post (\S+)×\/wk \((.+?)\), (\d+) creators, (\$[\d,]+) ads(?:, (\d+%) off in week (\d+))?$/, (m, a, mix, c, ads, off, wk) => `주 ${a}회 포스트(${mixKo(mix)}), 크리에이터 ${c}명, 광고 ${ads}` + (off ? `, ${wk}주차 ${off} 할인` : ''));

// Text with numbers in it: the numbers are captured and put back untouched.
const KO_RX = [
  [/^We hid the last (\d+) weeks from the engine\. .* really got\.$/, (m, n) => `마지막 ${n}주를 엔진에게 숨겼습니다. 엔진은 그 앞의 주들로 비율을 배운 뒤, 숨긴 주들의 주문을(실제로 한 포스트·광고·크리에이터·세일 주를 주고) 답했고, 그 답을 실제 주문과 비교했습니다.`],
  [/^If the engine may also learn from those (\d+) weeks \(in-sample\), the miss is (\S+)% and it gets (\d+) of (\d+) right\. The holdout number above is the honest one\.$/, (m, n, e, a, b) => `엔진이 그 ${n}주도 보고 배우게 하면(인샘플) 오차 ${e}%, ${b}주 중 ${a}주 적중입니다. 위의 홀드아웃 숫자가 정직한 쪽입니다.`],
  [/^Connected · demo account @(.+) · mock numbers, not your real account$/, (m, h) => `연결됨 · 데모 계정 @${h} · 목 숫자, 실제 계정 아님`],
  [/^Reasoned from the demo account @(.+)'s mock Instagram and Threads data \((\d+) days; not your real numbers\) and sourced public statistics · not stress-tested$/, (m, h, d) => `데모 계정 @${h}의 목 인스타그램·스레드 데이터(${d}일, 실제 숫자 아님)와 출처 있는 공개 통계로 판단 · 스트레스 테스트 대상 아님`],
  [/^(\w+): pick one or two directions$/, (m, a) => `${KO[a] || a}: 방향 한두 개 고르기`], [/^Plan (\w+)$/, (m, a) => `${KO[a] || a} 계획하기`],
  [/^(\d+% [a-z ]+(?:, \d+% [a-z ]+)*)$/, (m, a) => mixKo(a)], [/^(\d+%) off in week (\d+)$/, (m, off, wk) => `${wk}주차 ${off} 할인`],
  [/^Your history used ads in (\d+) weeks?(.*?), creators in (\d+) weeks?(.*?), a sale in (\d+) weeks?(.*?)\. A lever used in fewer than 6 weeks is thin evidence; one never used is shown but not rated\.$/, (m, a, af, c, cf, s, sf) => { const f = x => x.includes('thin') ? '(얇음)' : x.includes('never') ? '(없음)' : ''; return `지난 기록에서 광고 ${a}주${f(af)}, 크리에이터 ${c}주${f(cf)}, 세일 ${s}주${f(sf)} 사용. 6주 미만이면 근거가 얇고, 한 번도 안 썼으면 보여주되 평가하지 않습니다.`; }],
  [/^Thin evidence: your history used (ads|creators|a sale) in only (\d+) weeks, so the \1 rate rests on a handful of observations\.$/, (m, l, n) => `근거 얇음: 지난 기록에서 ${LV[l]}을(를) ${n}주만 써서, 그 비율은 관측 몇 개에 기대고 있습니다.`],
  [/^Not rated: your history has no weeks with (.+?), so the engine cannot say what \1 would do for you\. Run a small test first; do not trust a number here\.$/, (m, l) => `평가 안 함: 지난 기록에 ${l.split(' or ').map(x => LV[x] || x).join('·')}을(를) 쓴 주가 없어서 엔진이 효과를 말할 수 없습니다. 먼저 작게 시험해 보세요. 여기 숫자는 믿지 마세요.`],
  [/^Your history never used (.+?), so the engine cannot rate this direction\. You can still test it, but its numbers will not mean anything until you have run a small trial\.$/, (m, l) => `지난 기록에 ${l.split(' or ').map(x => LV[x] || x).join('·')}이(가) 없어서 엔진이 이 방향을 평가할 수 없습니다. 테스트는 되지만, 작게 시험해 보기 전까지 숫자는 의미가 없습니다.`],
  [/^Creators counted \(followers × engagement rate\): (.+)\. Each bad month lands a different (\d+) of them\.$/, (m, a, n) => `집계된 크리에이터(팔로워 × 참여율): ${a}. 나쁜 달마다 이 중 다른 ${n}명을 만납니다.`], [/^expected engagement per post (\d+)$/, (m, n) => `포스트당 기대 참여 ${n}`],
  [/^In numbers: (\S+) creators( of (\S+) followers \(the creators found\))? · (\S+) posts expected at your post rate · (\S+) engagement per post · about (\S+) orders from creators in a plain month · (\S+) in pieces shipped$/, (m, n, _w, f, p, e, o, s) => `숫자로: 크리에이터 ${n}명${f ? `(찾은 계정 기준 팔로워 ${f})` : ''} · 게시율 기준 기대 포스트 ${p}개 · 포스트당 참여 ${e} · 평범한 달 기준 크리에이터발 주문 약 ${o}건 · 보낼 제품 원가 ${s}`],
  [/^(Small|Mid-size|Large) creators, (\S+) followers: (\S+) engagement per post ≈ (\S+) orders per post · about (\S+) orders per creator you contact, since not everyone posts$/, (m, b, f, e, o, c) => `${{Small: '소형', 'Mid-size': '중형', Large: '대형'}[b]} 크리에이터(팔로워 ${f}): 포스트당 참여 ${e} ≈ 포스트당 주문 ${o}건 · 연락한 크리에이터 1명당 약 ${c}건(모두 올리진 않으므로)`], [/^Search Instagram for: (.+)$/, (m, a) => `인스타그램에서 검색: ${a}`],
  [/^Your budget ships (\d+) pieces at cost\. Numbers from the creators found for your search terms and your own order rate; the stress test above is what survives when some do not post\.$/, (m, n) => `예산으로 원가 기준 ${n}개를 보낼 수 있습니다. 숫자는 검색어로 찾은 크리에이터와 내 주문 전환율에서 나왔고, 위의 스트레스 테스트는 일부가 안 올렸을 때 살아남는 결과입니다.`],
  [/^Using (\d+) imported creators · (.+)$/, (m, n, a) => `가져온 크리에이터 ${n}명 사용 · ${a}`], [/^Could not read the list: (.+)$/, (m, a) => `목록을 읽지 못했습니다: ${a}`],
  [/^Among (\S+) adults (\S+): (.+) \((.+), (\d{4})\)\. Your target: (.*)\.$/, (m, mk, band, list, src, y, tg) => `${mk} 성인 ${band}: ${list} (${src}, ${y}). 내 타겟: ${tg.split(' ').map(x => w(x)).join(' ')}.`],
  [/^The verdicts skip (\S+), but (\d+%) of (\S+) adults (\S+) use it \((.+), (\d{4})\)\. Check that call against your own numbers\.$/, (m, p, pct, mk, band, src, y) => `판정은 ${p}을(를) 보류했지만 ${mk} 성인 ${band}의 ${pct}가 씁니다(${src}, ${y}). 내 숫자로 그 판단을 확인하세요.`],
  [/^No verdict covers (\S+), used by (\d+%) of (\S+) adults (\S+) \((.+), (\d{4})\)\.$/, (m, p, pct, mk, band, src, y) => `${p}에 대한 판정이 없습니다. ${mk} 성인 ${band}의 ${pct}가 씁니다(${src}, ${y}).`],
  [/^Demo: @(\S+) is a fictional US brand \(mock profile screenshot\), synthetic sales data, mock Instagram insights shaped like the Graph API\. Instagram Login replaces the mock after Meta App Review\.$/, (m, h) => `데모: @${h}는 가상의 미국 브랜드(목 프로필 스크린샷), 합성 판매 데이터, Graph API 모양의 목 인스타그램 인사이트. Meta 앱 리뷰 뒤에는 Instagram 로그인이 목을 대체합니다.`],
  [/^vs just posting: \+(\d+) orders in a bad month for (\$[\d,]+) more · (\$[\d,.]+) per extra order \(cost per purchase\) · (\S+)× return on the extra spend(.*)$/, (m, n, sp, per, ro, rest) => `그냥 포스팅 대비: 나쁜 달 주문 +${n}건, ${sp} 더 들어서 · 추가 주문당 ${per}(구매당 비용) · 추가 지출 대비 ${ro}배 매출${rest ? ' ' + rest.trim() : ''}`],
  [/^Where the money goes: nothing\. Your (\$[\d,]+) budget stays whole\.$/, (m, b) => `돈이 가는 곳: 없음. 예산 ${b}가 그대로 남습니다.`],
  [/^Where the money goes: (.+) = (\$[\d,]+) · (.+)$/, (m, parts, tot, left) => { const p = parts.split(' + ').map(x => x.replace(/^(\$[\d,]+) on ads$/, '광고 $1').replace(/^(\$[\d,]+) on (\d+) pieces for creators$/, '크리에이터에게 보낼 제품 $2개 $1').replace(/^(\$[\d,]+) given up in sale discounts$/, '세일 할인으로 포기하는 매출 $1')).join(' + '); const l = left.replace(/^(\$[\d,]+) of your (\$[\d,]+) budget left$/, '예산 $2 중 $1 남음').replace(/^(\$[\d,]+) over your (\$[\d,]+) budget$/, '예산 $2를 $1 초과'); return `돈이 가는 곳: ${p} = ${tot} · ${l}`; }],
  [/^The best version dies in (\d+) of (\d+) bad months and survives (\d+)\.$/, (m, a, b, c) => `가장 좋은 버전은 나쁜 달 ${b}개 중 ${a}개에서 죽고 ${c}개에서 살아남습니다.`],
  [/^Picked by the adversary \(Muse Spark\)\. Below are the ones that beat it; the bar shows how many of the (\d+) versions? died\. Numbers match the columns on the left\.$/, (m, n) => `적대자(Muse Spark)가 고른 달입니다. 아래는 그 버전을 이긴 달이고, 막대는 ${n}개 버전 중 몇 개가 죽었는지입니다. 번호는 왼쪽 열 번호와 같습니다.`],
  [/^Show the other (\d+) that beat it$/, (m, n) => `그 버전을 이긴 나머지 ${n}개 보기`], [/^Show the (\d+) it survives$/, (m, n) => `살아남는 ${n}개 보기`],
  [/^Bad months · (.+)$/, (m, a) => `나쁜 달 · ${a}`],
  [/^Picked by the adversary \(Muse Spark\), deadliest first\. The bar is how many of the (\d+) versions? died\. Numbers match the columns on the left\.$/, (m, n) => `적대자(Muse Spark)가 고른 순서, 가장 치명적인 것부터. 막대는 ${n}개 버전 중 몇 개가 죽었는지입니다. 번호는 왼쪽 열 번호와 같습니다.`],
  [/^(\d+) of (\d+) died$/, (m, a, b) => `${b}개 중 ${a}개 사망`], [/^Show the other (\d+) bad months$/, (m, n) => `나머지 나쁜 달 ${n}개 보기`], [/^Happens in about (\d+%) of months$/, (m, p) => `약 ${p}의 달에 발생`],
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
  [/^October: (inside range|above range|below range)\. Here is what was off\.$/, (m, r) => `10월: ${w(r)}. 무엇이 어긋났는지 봅시다.`], [/^You ran (.+)\.$/, (m, p) => `실행한 플랜: ${planKo(p)}.`], [/^(\d+) vs (\d+)–(\d+)$/, (m, a, b, c) => `${a}건 vs ${b}–${c}건`], [/^(\d+) of (\d+)$/, (m, a, b) => `${b} 중 ${a}`],
  [/^Only (\d+) of (\d+) seeded creators posted, so the seeded engagement never arrived\. Your own posts and conversion held up\. This was a creator problem, not a product or audience problem\.$/, (m, a, b) => `시딩한 크리에이터 ${b}명 중 ${a}명만 게시해서 기대한 반응이 오지 않았습니다. 자체 포스트와 전환은 괜찮았습니다. 제품이나 고객 문제가 아니라 크리에이터 문제였습니다.`],
  [/^(\$[\d,]+) over (\d+) days · (\$[\d,]+)\/day · paused until you switch it on$/, (m, a, d, b) => `${d}일간 ${a} · 하루 ${b} · 켜기 전까지 일시정지`],
  [/^Meta estimate for this audience: (\S+) reachable on Reels, (\S+) on feed, (\S+) on Stories \(monthly\) · (live from your ad account|demo numbers)$/, (m, a, b, c, src) => `이 타겟에 대한 Meta 추정: 릴스 월 ${a}명, 피드 ${b}명, 스토리 ${c}명 도달 가능 · ${src === 'demo numbers' ? '데모 숫자' : '내 광고 계정에서 실시간'}`],
  [/^Created paused: campaign (.+), ad set (.+)\. $/, (m, a, b) => `일시정지 상태로 생성: 캠페인 ${a}, 광고세트 ${b}. `],
  [/^Ship (\d+) pieces \((\$[\d,]+) at cost\) · send (\d+) DMs, expect about half to answer$/, (m, n, c, d) => `${n}개 발송(원가 ${c}) · DM ${d}개 보내기, 절반쯤 답한다고 가정`],
  [/^([\d,]+) followers · (\w+) · ([\d.]+)% engagement · mostly (.+)$/, (m, f, b, e, t) => `팔로워 ${f} · ${KO[b] || b} · 참여율 ${e}% · 주로 ${w(t)}`], [/^(high|medium|low|unsure) fit · (.*) account$/, (m, f, b) => `적합도 ${w(f)} · ${KO[b] || b} 계정`],
  [/^ship (\d+) pieces$/, (m, n) => `${n}개 발송`], [/^(\$[\d,]+) ads$/, (m, a) => `광고 ${a}`], [/^(\d+%) off$/, (m, p) => `${p} 할인`], [/^(\d+) \((.+)\)$/, (m, n, mix) => `${n} (${mix.split(', ').map(x => x.replace(/^(\d+) (.+)$/, (mm, k, t) => `${w(t)} ${k}`)).join(', ')})`],
  [/^Launch: (.+)$/, (m, t) => `실행: ${SPARK[t] || KO[t] || t}`], [/^Starts (\S+) · (\d+) days · total (\$[\d,]+) · (.+)$/, (m, d, n, tot, c) => `시작 ${d} · ${n}일 · 총 ${tot} · ${KO[c] || c}`],
  [/^Connected · @(.+)$/, (m, h) => `연결됨 · @${h}`],
  [/^If (\d+) orders is your average-month goal, the (\w+)-sized goal is (\d+) \((\d+) × (\S+)\)\. Test against the month you are actually in:$/, (m, g, mo, sg, g2, d) => `${g}건이 평범한 달 기준 목표라면, ${KO[mo] || mo} 크기의 목표는 ${sg}건(${g2} × ${d})입니다. 실제 그 달 기준으로 시험하세요:`],
  [/^Use (\d+) for (\w+)$/, (m, n, mo) => `${KO[mo] || mo}엔 ${n}건으로`], [/^Use (\d+) and re-test$/, (m, n) => `${n}건으로 다시 테스트`],
  [/^Verdict: just posting already clears a (\d+)-order goal in (\d+) of (\d+) bad months\.$/, (m, g, a, b) => `결론: 그냥 포스팅만으로도 목표 ${g}건을 나쁜 달 ${b}개 중 ${a}개에서 넘깁니다.`],
  [/^(\w+) usually brings (\S+)× an average month, so (\d+) is a low bar\. Either keep the money, or raise the bar to a (\w+)-sized goal of (\d+) and test again\.$/, (m, mo, d, g, mo2, sg) => `${KO[mo] || mo}은 보통 평달의 ${d}배라서 ${g}건은 낮은 기준입니다. 돈을 아끼거나, 기준을 ${KO[mo2] || mo2} 크기인 ${sg}건으로 올려 다시 시험하세요.`],
  [/^Could not (read|connect): (.+)$/, (m, w, e) => `${w === 'read' ? '읽지' : '연결하지'} 못했습니다: ${e}`], [/^(\d+) of (\d+) rates recovered within 5%\. .*$/, (m, a, b) => `비율 ${b}개 중 ${a}개를 5% 이내로 복원했습니다. 벗어난 하나(시딩 후 게시하는 크리에이터)는 일부러 그렇습니다: 10월에 노쇼를 심어두었습니다.`],
  [/^Computed for a goal of (\d+) orders, a (\$[\d,]+) budget, (\w+)\. Change these on Setup and this page recomputes\.$/, (m, g, b, mo) => `목표 ${g}건, 예산 ${b}, ${KO[mo] || mo} 기준으로 계산. 설정에서 바꾸면 이 화면은 다시 계산됩니다.`],
  [/^vs just posting: \+(\d+) orders in a bad month for (\$[\d,]+) more · (\$[\d,.]+) per extra order( \(more than a piece sells for\))?$/, (m, n, s, per, over) => `그냥 포스팅 대비: 나쁜 달에 ${s} 더 쓰고 주문 +${n}건 · 추가 주문당 ${per}${over ? ' (제품 한 개 판매가보다 큼)' : ''}`], [/^vs just posting: no extra orders in a bad month for (\$[\d,]+) more$/, (m, s) => `그냥 포스팅 대비: ${s} 더 쓰고도 나쁜 달에 추가 주문 없음`],
  [/^Verdict: (.+) barely beats just posting this month\.$/, (m, t) => `결론: ${tx(t)}, 그냥 포스팅보다 아주 조금 낫습니다.`],
  [/^You committed this plan for (\w+)\. The app is now planning (\w+): run the stress test again and commit a (\w+) plan\.$/, (m, a, b, c) => `이 플랜은 ${KO[a] || a}용으로 확정했습니다. 앱은 지금 ${KO[b] || b}을 계획 중입니다: 스트레스 테스트를 다시 돌려 ${KO[c] || c} 플랜을 확정하세요.`],
  [/^October: (inside range|above range|below range)\. Here is what (went better than the engine thought|the engine got right|was off)\.$/, (m, r, w) => `10월: ${w(r)}. ${w === 'was off' ? '무엇이 어긋났는지 봅시다.' : w === 'the engine got right' ? '엔진이 맞힌 것을 봅시다.' : '엔진 생각보다 잘된 것을 봅시다.'}`],
  [/^This month: (\w+)$/, (m, n) => `이번 달: ${KO[n] || n}`], [/^demand (\S+)× · ad cost (\S+)× · 1\.00 = an average month$/, (m, d, a) => `수요 ${d}배 · 광고 단가 ${a}배 · 1.00 = 평범한 달`],
  [/^Seasonal prior from public data \((\w+) · (.+?)\), applied to the order rate and ad reach for this month only; .*$/, (m, mk, cat) => `공개 데이터 기반 계절 지수(${mk} · ${KO[cat] || cat}). 이번 달의 주문율과 광고 도달에만 적용하고, 공격자는 이 달이 평범하게 끝나는 경우를 시험합니다. 과거 이력에서 배운 값이 아닙니다: 1년치 데이터로는 계절을 분리할 수 없습니다.`], [/^Source \((\w+)\): (.+), (\d{4})$/, (m, f, a, y) => `출처(${f === 'demand' ? '수요' : '광고 단가'}): ${a}, ${y}`], [/^(\d+)% (on hand|close up|flat lay|on neck)$/, (m, p, t) => `${w(t)} ${p}%`],
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
