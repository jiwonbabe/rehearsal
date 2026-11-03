"""The short data summary the planner prompt receives. Numbers as context, never as output."""
from engine.model import POST_TYPES
from engine.review import actual_month


def summary(brand, rates, backtest_result):
    b = brand["brand"]
    month = actual_month(brand)
    return {
        "brand": {"market": b["market"], "target": b["target"], "avg_price_usd": b["avg_price_usd"],
                  "budget_usd_per_month": b["budget_usd_per_month"], "goal_orders_per_month": b["goal_orders_per_month"]},
        "engagement_per_post": {t: round(rates.eng_per_post[t]) for t in POST_TYPES},
        "best_post_type": max(POST_TYPES, key=lambda t: rates.eng_per_post[t]),
        "seed_post_rate": round(rates.seed_post_rate, 2),
        "eng_per_creator_post": round(rates.eng_per_creator_post),
        "ad_eng_per_usd": round(rates.ad_eng_per_usd, 1),
        "promo_multiplier": round(rates.promo_multiplier, 2),
        "last_month": {"orders": month["orders"], "spend_usd": month["spend_usd"],
                       "creators_seeded": month["creators_seeded"], "creators_posted": month["creators_posted"]},
        "backtest": {"median_abs_pct_error": backtest_result["median_abs_pct_error"],
                     "weeks_inside_range": backtest_result["weeks_inside_range"],
                     "weeks_total": backtest_result["weeks_total"]},
    }
