"""Generate data/brand.json: one synthetic brand, 52 weekly rows, seeded.

Run:  python data/synthetic.py
"""
import json
import os
import random
from datetime import date, timedelta

SEED = 7
POST_TYPES = ["on_hand", "close_up", "flat_lay", "on_neck"]

# "true" rates the generator uses. The engine must re-discover these from the data.
TRUE_ENG_PER_POST = {"on_hand": 205, "close_up": 150, "flat_lay": 95, "on_neck": 120}
TRUE_ORDER_PER_ENG = 0.0115
TRUE_PROMO_MULT = 2.0
TRUE_ENG_PER_CREATOR_POST = 300
TRUE_AD_ENG_PER_USD = 5.0
TRUE_SEED_POST_RATE = 0.75

BRAND = {
    "name": "Wrenfield Jewelry (demo)",
    "synthetic": True,
    "market": "US",
    "category": "jewelry",
    "target": {"gender": "women", "age": "25-34", "price_band": "80-200"},
    "avg_price_usd": 120,
    "product_cost_usd": 45,
    "goal_orders_per_month": 46,
    "budget_usd_per_month": 400,
}


def week_plan(rng, i):
    """Decide what the brand did in week i (0..51). Last 4 weeks = October 2026."""
    posts = {"on_hand": 1, "close_up": 1, "flat_lay": 1, "on_neck": 0}
    extra = rng.choice(["on_hand", "on_hand", "close_up", "flat_lay", "on_neck"])
    posts[extra] += 1
    if rng.random() < 0.3:
        posts[rng.choice(POST_TYPES)] += 1

    creators_seeded = 0
    if i in (18, 19):          # a seeding try in spring, decent post rate
        creators_seeded = 3
    if i in (48, 49):          # October 2026: 4 seeded, only 1 posted (the lesson)
        creators_seeded = 2

    ads_usd = 0
    if i in (10, 11, 22, 23, 30, 31, 32, 40, 50, 51):
        ads_usd = rng.choice([60, 90, 120])

    promo_discount = 0.0
    if i in (14, 36):          # two sale weeks in the year
        promo_discount = 0.15

    return posts, creators_seeded, ads_usd, promo_discount


def make_weeks(rng):
    weeks = []
    followers = 3900
    start = date(2025, 11, 3)
    for i in range(52):
        posts, creators_seeded, ads_usd, promo_discount = week_plan(rng, i)

        engagement = {}
        for t in POST_TYPES:
            noise = rng.uniform(0.85, 1.15)
            engagement[t] = int(posts[t] * TRUE_ENG_PER_POST[t] * noise * (followers / 4000))

        if i in (48, 49):
            creators_posted = 1 if i == 48 else 0
        else:
            creators_posted = sum(1 for _ in range(creators_seeded) if rng.random() < TRUE_SEED_POST_RATE)
        creator_engagement = int(creators_posted * TRUE_ENG_PER_CREATOR_POST * rng.uniform(0.8, 1.2))

        ad_engagement = int(ads_usd * TRUE_AD_ENG_PER_USD * rng.uniform(0.85, 1.15))

        total_eng = sum(engagement.values()) + creator_engagement + ad_engagement
        mult = TRUE_PROMO_MULT if promo_discount > 0 else 1.0
        orders = int(round(total_eng * TRUE_ORDER_PER_ENG * mult * rng.uniform(0.9, 1.1)))

        spend_usd = ads_usd + creators_seeded * BRAND["product_cost_usd"]
        spend_usd += int(orders * BRAND["avg_price_usd"] * promo_discount)

        weeks.append({
            "week": (start + timedelta(weeks=i)).isoformat(),
            "posts": posts,
            "engagement": engagement,
            "followers": followers,
            "creators_seeded": creators_seeded,
            "creators_posted": creators_posted,
            "creator_engagement": creator_engagement,
            "ads_usd": ads_usd,
            "ad_engagement": ad_engagement,
            "promo_discount": promo_discount,
            "orders": orders,
            "spend_usd": spend_usd,
        })
        followers += int(total_eng * 0.02) + rng.randint(5, 25)
    return weeks


def main():
    rng = random.Random(SEED)
    out = {"brand": BRAND, "weeks": make_weeks(rng),
           # The rates the generator used. The engine never reads these; the trust check compares its
           # calibrated rates against them so a reader can see the engine recover a known answer.
           "generator_truth": {"eng_per_post": TRUE_ENG_PER_POST, "order_per_engagement": TRUE_ORDER_PER_ENG,
                               "promo_multiplier": TRUE_PROMO_MULT, "seed_post_rate": TRUE_SEED_POST_RATE,
                               "eng_per_creator_post": TRUE_ENG_PER_CREATOR_POST, "ad_eng_per_usd": TRUE_AD_ENG_PER_USD,
                               "followers_base": 4000,
                               "note": "eng_per_post scales with followers/4000; October carries a planted creator no-show, so seed_post_rate in the data is below 0.75 on purpose."}}
    path = os.path.join(os.path.dirname(__file__), "brand.json")
    with open(path, "w") as f:
        json.dump(out, f, indent=1)
    total_orders = sum(w["orders"] for w in out["weeks"])
    print(f"wrote {path}: 52 weeks, {total_orders} orders")


if __name__ == "__main__":
    main()
