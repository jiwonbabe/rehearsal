"""Launch spec: a committed plan as dates, daily ad budget, creator kit numbers and a 4-week calendar."""
import datetime
import math

from engine.model import POST_TYPES, HORIZON_WEEKS


def half_up(x):
    return int(math.floor(x + 0.5))


def split_posts(total, mix):
    """Largest remainder: integer posts per type summing to total; ties by POST_TYPES order."""
    quotas = {t: total * mix.get(t, 0.0) for t in POST_TYPES}
    out = {t: int(math.floor(quotas[t])) for t in POST_TYPES}
    left = total - sum(out.values())
    order = sorted(POST_TYPES, key=lambda t: (-(quotas[t] - out[t]), POST_TYPES.index(t)))
    for t in order[:max(left, 0)]:
        out[t] += 1
    return out


def launch_spec(plan, brand):
    last = datetime.date.fromisoformat(brand["weeks"][-1]["week"])
    start = last + datetime.timedelta(days=7)
    if plan.month:
        # a plan for a later month starts 4-week blocks later (the data's weeks are Monday-based)
        k = (plan.month - start.month) % 12
        start = start + datetime.timedelta(days=28 * k)
    days = HORIZON_WEEKS * 7
    end = start + datetime.timedelta(days=days - 1)
    count = int(plan.creator_count)
    ship = count * brand["brand"]["product_cost_usd"]
    posts = half_up(plan.posts_per_week)
    calendar = []
    for w in range(HORIZON_WEEKS):
        calendar.append({
            "week": w + 1,
            "start_date": (start + datetime.timedelta(days=7 * w)).isoformat(),
            "posts": split_posts(posts, plan.post_mix),
            "creators": count if w == 0 else 0,
            "ads_usd": round(plan.ads_usd / HORIZON_WEEKS, 2),
            "promo_discount": plan.promo_discount if plan.promo_week == w + 1 else 0,
        })
    return {
        "start_date": start.isoformat(), "end_date": end.isoformat(), "days": days,
        "ads_usd": plan.ads_usd,
        "ads_daily_usd": round(plan.ads_usd / days, 2),
        "ads_daily_cents": half_up(plan.ads_usd / days * 100),
        "creator_count": count, "creator_ship_usd": ship, "creator_dms": count * 2,
        "total_spend_usd": plan.ads_usd + ship,
        "calendar": calendar,
    }


def creator_math(count, rates, brand, month=0):
    """What seeding N creators means in numbers: the brand's calibrated rates and the found creators' audience."""
    import math
    from engine.model import expected_engagement, season_for
    n = int(math.floor(count + 0.5))
    pool = brand.get("creator_pool") or []
    e = [expected_engagement(c) for c in pool]
    f = [c["followers"] for c in pool]
    posts = n * rates.seed_post_rate
    eng = posts * rates.eng_per_creator_post
    demand, _ = season_for(brand, month)
    return {"creators": n,
            "followers_min": min(f) if f else 0, "followers_max": max(f) if f else 0,
            "eng_per_post_min": min(e) if e else round(rates.eng_per_creator_post),
            "eng_per_post_max": max(e) if e else round(rates.eng_per_creator_post),
            "expected_posts": round(posts, 1), "expected_engagement": round(eng),
            "expected_orders": round(eng * rates.order_per_engagement * demand, 1),
            "ship_usd": n * brand["brand"]["product_cost_usd"]}


def creator_bands(rates, brand, month=0):
    """What one creator of each size band is worth, from the creators found in that band."""
    import math
    from engine.model import expected_engagement, season_for
    demand, _ = season_for(brand, month)
    order_per_eng = rates.order_per_engagement * demand
    cost = brand["brand"]["product_cost_usd"]
    pieces = int(math.floor(brand["brand"]["budget_usd_per_month"] / cost)) if cost > 0 else 0
    out = []
    for band in ["small", "mid", "large"]:
        cs = [c for c in (brand.get("creator_pool") or []) if c["followers_band"] == band]
        if not cs:
            continue
        e = [expected_engagement(c) for c in cs]
        f = [c["followers"] for c in cs]
        out.append({"band": band, "count": len(cs), "followers_min": min(f), "followers_max": max(f),
                    "eng_per_post_min": min(e), "eng_per_post_max": max(e),
                    "orders_per_post_min": round(min(e) * order_per_eng, 1), "orders_per_post_max": round(max(e) * order_per_eng, 1),
                    "orders_per_contact": round(sum(e) / len(e) * order_per_eng * rates.seed_post_rate, 1),
                    "pieces_for_budget": pieces})
    return out


def spend_breakdown(plan, rates, brand):
    """Where a plan's money goes in a plain month: ads, pieces shipped to creators, discounts given up."""
    import math
    from engine.model import run_plan
    clean = run_plan(plan, rates, brand)
    n = int(math.floor(plan.creator_count + 0.5))
    ship = n * brand["brand"]["product_cost_usd"]
    discount = max(0.0, clean["spend_usd"] - plan.ads_usd - ship)
    rd = lambda x: float(math.floor(x + 0.5))
    ads, shipr, disc = rd(plan.ads_usd), rd(ship), rd(discount)
    budget = brand["brand"]["budget_usd_per_month"]
    total = ads + shipr + disc
    return {"ads_usd": ads, "creator_count": n, "creator_ship_usd": shipr, "discount_usd": disc, "total_usd": total,
            "budget_usd": budget, "left_usd": budget - total}
