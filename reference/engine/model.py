"""Deterministic engine: calibrate rates from brand data, run a plan, judge survival.

No model calls here. No randomness except seeded sampling in sample.py.
"""
import json
import statistics

POST_TYPES = ["on_hand", "close_up", "flat_lay", "on_neck"]
HORIZON_WEEKS = 4


class Rates:
    def __init__(self):
        self.eng_per_post = {t: 0.0 for t in POST_TYPES}
        self.order_per_engagement = 0.0
        self.promo_multiplier = 1.0
        self.seed_post_rate = 0.0
        self.eng_per_creator_post = 0.0
        self.ad_eng_per_usd = 0.0

    def copy(self):
        r = Rates()
        r.eng_per_post = dict(self.eng_per_post)
        r.order_per_engagement = self.order_per_engagement
        r.promo_multiplier = self.promo_multiplier
        r.seed_post_rate = self.seed_post_rate
        r.eng_per_creator_post = self.eng_per_creator_post
        r.ad_eng_per_usd = self.ad_eng_per_usd
        return r

    def to_dict(self):
        return {
            "eng_per_post": self.eng_per_post,
            "order_per_engagement": self.order_per_engagement,
            "promo_multiplier": self.promo_multiplier,
            "seed_post_rate": self.seed_post_rate,
            "eng_per_creator_post": self.eng_per_creator_post,
            "ad_eng_per_usd": self.ad_eng_per_usd,
        }


class Plan:
    def __init__(self, plan_id, direction, posts_per_week, post_mix, creator_count,
                 ads_usd, promo_week=None, promo_discount=0.0):
        self.plan_id = plan_id
        self.direction = direction
        self.posts_per_week = posts_per_week
        self.post_mix = post_mix
        self.creator_count = creator_count
        self.ads_usd = ads_usd
        self.promo_week = promo_week
        self.promo_discount = promo_discount

    def to_dict(self):
        return {
            "plan_id": self.plan_id,
            "direction": self.direction,
            "posts_per_week": self.posts_per_week,
            "post_mix": self.post_mix,
            "creator_count": self.creator_count,
            "ads_usd": self.ads_usd,
            "promo_week": self.promo_week,
            "promo_discount": self.promo_discount,
        }

    @staticmethod
    def from_dict(d):
        return Plan(d["plan_id"], d["direction"], d["posts_per_week"], d["post_mix"],
                    d["creator_count"], d["ads_usd"], d.get("promo_week"), d.get("promo_discount", 0.0))


def load_brand(path="data/brand.json"):
    with open(path) as f:
        return json.load(f)


def safe_div(a, b):
    return a / b if b else 0.0


def calibrate(brand):
    """Rediscover the funnel rates from the weekly rows. Pure arithmetic."""
    weeks = brand["weeks"]
    rates = Rates()

    for t in POST_TYPES:
        posts = sum(w["posts"][t] for w in weeks)
        eng = sum(w["engagement"][t] for w in weeks)
        rates.eng_per_post[t] = safe_div(eng, posts)

    plain = [w for w in weeks if w["promo_discount"] == 0]
    promo = [w for w in weeks if w["promo_discount"] > 0]
    eng_plain = sum(total_engagement(w) for w in plain)
    rates.order_per_engagement = safe_div(sum(w["orders"] for w in plain), eng_plain)

    if promo:
        promo_rate = safe_div(sum(w["orders"] for w in promo), sum(total_engagement(w) for w in promo))
        rates.promo_multiplier = safe_div(promo_rate, rates.order_per_engagement) or 1.0

    seeded = sum(w["creators_seeded"] for w in weeks)
    posted = sum(w["creators_posted"] for w in weeks)
    rates.seed_post_rate = safe_div(posted, seeded)
    rates.eng_per_creator_post = safe_div(sum(w["creator_engagement"] for w in weeks), posted)

    ads = sum(w["ads_usd"] for w in weeks)
    rates.ad_eng_per_usd = safe_div(sum(w["ad_engagement"] for w in weeks), ads)
    return rates


def total_engagement(w):
    return sum(w["engagement"].values()) + w["creator_engagement"] + w["ad_engagement"]


def week_inputs_from_plan(plan, week_index):
    """What the brand does in one week under this plan (week_index 0..3)."""
    posts = {t: plan.posts_per_week * plan.post_mix.get(t, 0.0) for t in POST_TYPES}
    creators = plan.creator_count if week_index == 0 else 0
    ads_usd = plan.ads_usd / HORIZON_WEEKS
    discount = plan.promo_discount if plan.promo_week == week_index + 1 else 0.0
    return posts, creators, ads_usd, discount


def run_week(posts, creators_seeded, ads_usd, discount, rates, brand, mods):
    """One week of the funnel. `mods` is a dict of multipliers from shocks (see shocks.py)."""
    organic = sum(posts[t] * rates.eng_per_post[t] for t in POST_TYPES)
    organic *= mods.get("organic_mult", 1.0)

    post_rate = rates.seed_post_rate * mods.get("seed_post_rate_mult", 1.0)
    creators_posted = creators_seeded * post_rate
    creator_eng = creators_posted * rates.eng_per_creator_post

    ad_eng = ads_usd * rates.ad_eng_per_usd * mods.get("ad_eng_mult", 1.0)

    engagement = organic + creator_eng + ad_eng
    order_rate = rates.order_per_engagement * mods.get("order_rate_mult", 1.0)
    promo_mult = rates.promo_multiplier * mods.get("promo_mult_mult", 1.0) if discount > 0 else 1.0
    orders = engagement * order_rate * promo_mult

    spend = ads_usd + creators_seeded * brand["brand"]["product_cost_usd"]
    spend += orders * brand["brand"]["avg_price_usd"] * discount
    return {"engagement": engagement, "orders": orders, "spend_usd": spend}


def run_plan(plan, rates, brand, shocks=None, weeks=HORIZON_WEEKS):
    from engine.shocks import mods_for_week
    shocks = shocks or []
    rows = []
    for i in range(weeks):
        posts, creators, ads_usd, discount = week_inputs_from_plan(plan, i)
        mods = mods_for_week(shocks, i, rates)
        creators = int(round(creators * mods.get("creator_count_mult", 1.0)))
        rows.append(run_week(posts, creators, ads_usd, discount, rates, brand, mods))
    return {
        "plan_id": plan.plan_id,
        "weeks": rows,
        "orders": sum(r["orders"] for r in rows),
        "spend_usd": sum(r["spend_usd"] for r in rows),
        "engagement": sum(r["engagement"] for r in rows),
    }


def survived(result, brand):
    goal = brand["brand"]["goal_orders_per_month"]
    budget = brand["brand"]["budget_usd_per_month"]
    cumulative = 0.0
    for i, r in enumerate(result["weeks"]):
        cumulative += r["spend_usd"]
        if i < 2 and cumulative > budget:
            return False
    return result["orders"] >= goal and result["spend_usd"] <= budget * 1.1


def backtest(brand, rates, weeks_back=13):
    """Replay the last N weeks' actual inputs and compare engine orders to actual orders."""
    rows = []
    errors = []
    inside = 0
    for w in brand["weeks"][-weeks_back:]:
        r = run_week(w["posts"], w["creators_seeded"], w["ads_usd"], w["promo_discount"],
                     rates, brand, {})
        engine_orders = r["orders"]
        actual = w["orders"]
        err = safe_div(abs(engine_orders - actual), actual)
        errors.append(err)
        lo, hi = engine_orders * 0.8, engine_orders * 1.2
        inside += 1 if lo <= actual <= hi else 0
        rows.append({"week": w["week"], "engine": round(engine_orders, 1), "actual": actual,
                     "low": round(lo, 1), "high": round(hi, 1), "promo": w["promo_discount"] > 0})
    return {
        "rows": rows,
        "median_abs_pct_error": round(statistics.median(errors) * 100, 1) if errors else None,
        "weeks_inside_range": inside,
        "weeks_total": len(rows),
    }
