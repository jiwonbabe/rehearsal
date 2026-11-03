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
