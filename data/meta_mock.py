"""Generate data/meta/demo.brand/*.json: mock Meta data shaped like the Graph API replies, seeded.

Run:  python3 data/meta_mock.py
Live replacements: Instagram Graph API (account + media insights), Threads API, hashtag search,
business_discovery, Marketing API delivery_estimate. Every file carries "source": "mock"; the account and every peer/creator handle are fictional (demo.* prefix), so no real account's numbers are shown.
"""
import json
import os
import random
from datetime import date, timedelta

import sys

SEED = 7
POST_TYPES = ["on_hand", "close_up", "flat_lay", "on_neck"]
HANDLE = "demo.brand"
CITIES = {"Portland": 0.41, "Seattle": 0.18, "Los Angeles": 0.07, "San Francisco": 0.05, "New York": 0.04}
HASHTAGS = ["#dailygoldjewelry", "#everydaygoldjewelry"]
CREATORS = [("demo.soft.daily", 6200, "slow mornings, coffee, tiny rings ✨ Portland", [0.6, 0.2, 0.2, 0], 310),
            ("demo.minsol.outfit", 14800, "daily outfits · beige lover · collabs by DM", [0.4, 0.3, 0.2, 0.1], 520),
            ("demo.hana.hands", 3900, "hands, light, quiet mornings", [0.8, 0.2, 0, 0], 260),
            ("demo.goldie.reviews", 52000, "jewelry reviews · PR welcome", [0.3, 0.5, 0.2, 0], 900),
            ("demo.yeon.cafe", 9100, "cafe hopping · film photos", [0.2, 0.1, 0.6, 0.1], 280),
            ("demo.ari.minimal", 21500, "minimal wardrobe · thirties, office life", [0.5, 0.3, 0.1, 0.1], 610)]
PEERS = [("demo.mellow.gold", 24100, 412, [0.6, 0.25, 0.1, 0.05], 520), ("demo.une.pieces", 11300, 236, [0.35, 0.45, 0.15, 0.05], 260), ("demo.hazel.jewel", 38400, 690, [0.5, 0.2, 0.2, 0.1], 740)]
# Mock accounts for the other synthetic brands (data/brands/*.json): same shapes, their own handles, cities, tags and creators.
PROFILES = {
    "demo.apparel": {"SEED": 11, "CITIES": {"Austin": 0.33, "Dallas": 0.14, "Houston": 0.09, "Los Angeles": 0.06, "New York": 0.05},
                     "HASHTAGS": ["#slowfashion", "#capsulewardrobe"],
                     "CREATORS": [("demo.linen.days", 7400, "linen, coffee, slow weekends · Austin", [0.6, 0.2, 0.2, 0], 360),
                                  ("demo.june.wears", 18900, "outfit diaries · thrifted + small brands", [0.4, 0.3, 0.2, 0.1], 640),
                                  ("demo.pocket.fits", 4300, "fit checks from a tiny apartment", [0.8, 0.2, 0, 0], 300),
                                  ("demo.style.ledger", 61000, "honest clothing reviews · PR welcome", [0.3, 0.5, 0.2, 0], 1100),
                                  ("demo.marg.market", 8800, "farmers markets · film", [0.2, 0.1, 0.6, 0.1], 320),
                                  ("demo.ada.office", 23500, "workwear that is not boring", [0.5, 0.3, 0.1, 0.1], 700)],
                     "PEERS": [("demo.ferry.street", 31000, 480, [0.6, 0.25, 0.1, 0.05], 700), ("demo.loom.co", 12600, 310, [0.35, 0.45, 0.15, 0.05], 330), ("demo.north.knit", 44000, 820, [0.5, 0.2, 0.2, 0.1], 900)]},
    "demo.beauty": {"SEED": 23, "CITIES": {"Los Angeles": 0.29, "New York": 0.17, "Miami": 0.09, "Chicago": 0.06, "Atlanta": 0.05},
                    "HASHTAGS": ["#skincareroutine", "#cleanbeauty"],
                    "CREATORS": [("demo.glow.notes", 9800, "skin diaries · sensitive skin · LA", [0.6, 0.2, 0.2, 0], 520),
                                 ("demo.rae.routine", 27000, "morning routines · honest reviews", [0.4, 0.3, 0.2, 0.1], 980),
                                 ("demo.tiny.shelf", 5100, "small shelfie, big opinions", [0.8, 0.2, 0, 0], 390),
                                 ("demo.derm.diary", 88000, "esthetician · ingredient breakdowns · PR welcome", [0.3, 0.5, 0.2, 0], 1700),
                                 ("demo.sun.week", 12400, "beach days · SPF nerd", [0.2, 0.1, 0.6, 0.1], 450),
                                 ("demo.jo.minimal", 33000, "five-product routine, that's it", [0.5, 0.3, 0.1, 0.1], 1050)],
                    "PEERS": [("demo.pale.studio", 41000, 530, [0.6, 0.25, 0.1, 0.05], 1100), ("demo.lumi.lab", 15800, 290, [0.35, 0.45, 0.15, 0.05], 480), ("demo.soft.serum", 57000, 760, [0.5, 0.2, 0.2, 0.1], 1400)]},
}
OUT = os.path.join(os.path.dirname(__file__), "meta", HANDLE)
END = date(2026, 10, 26)

REACH = {"FEED": 1800, "REELS": 5400, "STORY": 900}
SAVES_RATE = {"on_hand": 0.052, "close_up": 0.038, "flat_lay": 0.021, "on_neck": 0.030}
TAPS_RATE = {"FEED": 0.011, "REELS": 0.016, "STORY": 0.024}


def dump(name, obj):
    os.makedirs(OUT, exist_ok=True)
    with open(os.path.join(OUT, name + ".json"), "w") as f:
        json.dump(obj, f, indent=1, ensure_ascii=False)


def posts(rng, n, weights, like_base):
    out = []
    for i in range(n):
        t = rng.choices(POST_TYPES, weights=weights)[0]
        mult = {"on_hand": 1.35, "close_up": 1.0, "flat_lay": 0.7, "on_neck": 0.85}[t]
        out.append({"id": f"{i+1}", "post_type": t,
                    "like_count": int(like_base * mult * rng.uniform(0.7, 1.3)),
                    "comments_count": int(like_base * mult * rng.uniform(0.02, 0.05))})
    return out


def main():
    global OUT
    if len(sys.argv) > 1:
        globals().update(PROFILES[sys.argv[1]])
        globals()["HANDLE"] = sys.argv[1]
        OUT = os.path.join(os.path.dirname(__file__), "meta", HANDLE)
    rng = random.Random(SEED)
    dump("account", {"source": "mock", "username": HANDLE, "followers_count": 5561,
                     "follower_demographics": {
                         "age": {"18-24": 0.22, "25-34": 0.46, "35-44": 0.21, "45-54": 0.08, "55+": 0.03},
                         "gender": {"women": 0.78, "men": 0.19, "unknown": 0.03},
                         "city": CITIES}})
    media = []
    kinds = ["FEED"] * 12 + ["REELS"] * 8 + ["STORY"] * 4
    rng.shuffle(kinds)
    for i, kind in enumerate(kinds):
        t = rng.choices(POST_TYPES, weights=[0.45, 0.3, 0.15, 0.1])[0]
        reach = int(REACH[kind] * {"on_hand": 1.25, "close_up": 1.0, "flat_lay": 0.75, "on_neck": 0.9}[t] * rng.uniform(0.75, 1.25))
        media.append({"id": f"1789{i:03d}", "timestamp": (END - timedelta(days=89 - i * 3)).isoformat() + "T09:00:00+0000",
                      "media_product_type": kind, "post_type": t,
                      "insights": {"reach": reach, "views": int(reach * rng.uniform(1.4, 2.2)),
                                   "saved": int(reach * SAVES_RATE[t] * rng.uniform(0.8, 1.2)),
                                   "shares": int(reach * 0.009 * rng.uniform(0.6, 1.4)),
                                   "likes": int(reach * 0.07 * rng.uniform(0.7, 1.3)),
                                   "comments": int(reach * 0.004 * rng.uniform(0.5, 1.5)),
                                   "link_taps": int(reach * TAPS_RATE[kind] * rng.uniform(0.7, 1.3))}})
    dump("media", {"source": "mock", "labelled_by": "spark-vision (mock labels)", "days": 90, "data": media})
    threads = []
    for i in range(10):
        views = int(2200 * rng.uniform(0.6, 1.6))
        threads.append({"id": f"th{i+1}", "timestamp": (END - timedelta(days=80 - i * 8)).isoformat() + "T12:00:00+0000",
                        "text_kind": rng.choice(["behind the scenes", "question to followers", "new piece", "styling tip"]),
                        "insights": {"views": views, "likes": int(views * 0.028), "replies": int(views * 0.009), "reposts": int(views * 0.003)}})
    dump("threads", {"source": "mock", "days": 90, "data": threads})
    dump("hashtags", {"source": "mock", "data": [
        {"hashtag": HASHTAGS[0], "top_media": posts(rng, 30, [0.5, 0.3, 0.15, 0.05], 420)},
        {"hashtag": HASHTAGS[1], "top_media": posts(rng, 30, [0.55, 0.25, 0.12, 0.08], 610)}]})
    dump("peers", {"source": "mock", "data": [
        {"username": u, "followers_count": f, "media_count": mc, "recent": posts(rng, 12, w, lb)} for u, f, mc, w, lb in PEERS]})
    dump("creators", {"source": "mock", "data": [
        {"username": u, "followers_count": f, "biography": b, "recent": posts(rng, 6, w, lb)} for u, f, b, w, lb in CREATORS]})
    dump("reach_estimate", {"source": "mock", "targeting": {"countries": ["US"], "age_min": 25, "age_max": 34, "genders": ["women"]},
                            "data": {"instagram_reels": {"estimate_dau": 412000, "estimate_mau": 1260000},
                                     "instagram_feed": {"estimate_dau": 388000, "estimate_mau": 1310000},
                                     "instagram_stories": {"estimate_dau": 351000, "estimate_mau": 1190000},
                                     "facebook_feed": {"estimate_dau": 96000, "estimate_mau": 540000}}})
    print("mock written to", OUT)


if __name__ == "__main__":
    main()
