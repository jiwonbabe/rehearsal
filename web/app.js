// Shared shell for every page. All numbers shown come from the API; this file only formats.
function renderHeader(step, subtitle) {
  const nav = [['0 Setup', '/'], ['1 Directions', '/directions.html'], ['2 Simulate', '/simulate.html'], ['3 Verdict', '/verdict.html'], ['4 Review', '/review.html'], ['5 Launch', '/launch.html']];
  const pills = nav.map(([t, h], i) => i === step
    ? `<div class="pill active">${t}</div>` : `<a class="pill" href="${h}">${t}</a>`).join('');
  const el = document.createElement('div');
  el.style.cssText = 'display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:12px;padding:20px 40px;border-bottom:1px solid #d9dbd4;background:#fff';
  el.innerHTML = `<div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
    <div style="width:28px;height:28px;border-radius:6px;background:#17181c;display:flex;align-items:center;justify-content:center"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#f3f4f1" stroke-width="2.2" aria-hidden="true"><path d="M4 18 L10 10 L14 14 L20 6"></path></svg></div>
    <div style="font-weight:600;font-size:16px">Rehearsal</div>
    <div style="font-size:13px;color:#5c6066;padding-left:12px;border-left:1px solid #d9dbd4">${subtitle}</div>
    <div class="badge">Demo brand · synthetic data</div></div>
    <div style="display:flex;gap:6px;font-size:13px;align-items:center">${pills}<button class="pill" style="border:1px solid #d9dbd4;background:#fff;cursor:pointer;font:inherit;font-size:12px" onclick="setLang(getLang() === 'ko' ? 'en' : 'ko')">${getLang() === 'ko' ? 'EN' : '한국어'}</button></div>`;
  document.body.prepend(el);
}

async function api(path, body) {
  if (typeof I18N_READY !== 'undefined') await I18N_READY;
  const res = await fetch(path, body === undefined ? undefined
    : { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
  let json = null;
  try { json = await res.json(); } catch (e) { /* non-JSON body */ }
  if (!res.ok) { const err = new Error((json && json.error) || String(res.status)); if (json && json.campaign_id) err.campaign_id = json.campaign_id; throw err; }
  return json;
}

function getSelected() {
  try { return JSON.parse(sessionStorage.getItem('rehearsal.selected')) || []; } catch (e) { return []; }
}
function setSelected(ids) { sessionStorage.setItem('rehearsal.selected', JSON.stringify(ids)); }
// The committed plan travels with the title and the month it was committed for, so Launch can say
// when the app has since moved on to another month.
function getCommitted() {
  try { const c = JSON.parse(sessionStorage.getItem('rehearsal.committed')); if (!c) return null; return c.plan ? c : { plan: c, title: c.direction, month: c.month }; } catch (e) { return null; }
}
function setCommitted(plan, title, month, monthName) { sessionStorage.setItem('rehearsal.committed', JSON.stringify({ plan, title, month, month_name: monthName })); }
function stateLine(goal, budget, monthName) { return `Computed for a goal of ${goal} orders, a ${money(budget)} budget, ${monthName}. Change these on Setup and this page recomputes.`; }

// prov: a small chip saying where a block's numbers come from. kind: real | demo | engine | spark
function prov(kind, text) {
  const look = { real: ['#e8f3ea', '#2f6f3e'], demo: ['#fbf3dc', '#8a6d1f'], engine: ['#e4edf6', '#2f5d8a'], spark: ['#f3f4f1', '#5c6066'] }[kind] || ['#f3f4f1', '#5c6066'];
  const b = document.createElement('span');
  b.textContent = text;
  b.style.cssText = `display:inline-block;padding:2px 8px;border-radius:10px;font-size:11px;font-weight:600;background:${look[0]};color:${look[1]};white-space:nowrap;margin-left:8px;vertical-align:middle`;
  return b;
}
function money(n) { return '$' + Math.round(n).toLocaleString('en-US'); }
function pct(x) { return Math.round(x * 100) + '%'; }
function shockText(codes, catalog) {
  return codes.length ? codes.map(c => (catalog && catalog[c]) || c).join(' + ') : 'never';
}
function planText(plan) {
  const m = plan.post_mix || {};
  const keys = Object.keys(m).sort((a, b) => (a === 'on_hand' ? -1 : b === 'on_hand' ? 1 : 0));
  const mix = keys.map(k => `${pct(m[k])} ${k.replace('_', ' ')}`).join(', ');
  let s = `post ${plan.posts_per_week}×/wk (${mix}), ${plan.creator_count} creators, ${money(plan.ads_usd)} ads`;
  if (plan.promo_week != null) s += `, ${pct(plan.promo_discount)} off in week ${plan.promo_week}`;
  return s;
}

function showError(e) {
  let el = document.getElementById('err');
  if (!el) { el = document.createElement('div'); el.id = 'err'; el.className = 'err'; document.querySelector('.page').prepend(el); }
  el.textContent = e.message || String(e);
}
