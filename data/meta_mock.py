"""Generate data/meta/bnd.jewelry/*.json: mock Meta data shaped like the Graph API replies, seeded.

Run:  python3 data/meta_mock.py
Live replacements: Instagram Graph API (account + media insights), Threads API, hashtag search,
business_discovery, Marketing API delivery_estimate. Every file carries "source": "mock".
"""
import json
import os
import random
from datetime import date, timedelta

SEED = 7
POST_TYPES = ["on_hand", "close_up", "flat_lay", "on_neck"]
OUT = os.path.join(os.path.dirname(__file__), "meta", "bnd.jewelry")
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
    rng = random.Random(SEED)
    dump("account", {"source": "mock", "username": "bnd.jewelry", "followers_count": 5561,
                     "follower_demographics": {
                         "age": {"18-24": 0.22, "25-34": 0.46, "35-44": 0.21, "45-54": 0.08, "55+": 0.03},
                         "gender": {"women": 0.78, "men": 0.19, "unknown": 0.03},
                         "city": {"Seoul": 0.41, "Gyeonggi": 0.18, "Busan": 0.07, "Incheon": 0.05, "Los Angeles": 0.04}}})
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
        {"hashtag": "#데일리주얼리", "top_media": posts(rng, 30, [0.5, 0.3, 0.15, 0.05], 420)},
        {"hashtag": "#everydaygoldjewelry", "top_media": posts(rng, 30, [0.55, 0.25, 0.12, 0.08], 610)}]})
    dump("peers", {"source": "mock", "data": [
        {"username": "mellow.gold.seoul", "followers_count": 24100, "media_count": 412, "recent": posts(rng, 12, [0.6, 0.25, 0.1, 0.05], 520)},
        {"username": "une.pieces", "followers_count": 11300, "media_count": 236, "recent": posts(rng, 12, [0.35, 0.45, 0.15, 0.05], 260)},
        {"username": "hazel.jewel.kr", "followers_count": 38400, "media_count": 690, "recent": posts(rng, 12, [0.5, 0.2, 0.2, 0.1], 740)}]})
    dump("creators", {"source": "mock", "data": [
        {"username": "soft.daily.yj", "followers_count": 6200, "biography": "일상, 커피, 작은 반지 ✨ 서울", "recent": posts(rng, 6, [0.6, 0.2, 0.2, 0], 310)},
        {"username": "minsol.outfit", "followers_count": 14800, "biography": "데일리룩 · 베이지 러버 · 협찬문의 DM", "recent": posts(rng, 6, [0.4, 0.3, 0.2, 0.1], 520)},
        {"username": "hana.handpoem", "followers_count": 3900, "biography": "hands, light, quiet mornings", "recent": posts(rng, 6, [0.8, 0.2, 0, 0], 260)},
        {"username": "goldie.seoul", "followers_count": 52000, "biography": "jewelry reviews · PR welcome", "recent": posts(rng, 6, [0.3, 0.5, 0.2, 0], 900)},
        {"username": "yeon.cafe.log", "followers_count": 9100, "biography": "카페 투어 · 필름 사진", "recent": posts(rng, 6, [0.2, 0.1, 0.6, 0.1], 280)},
        {"username": "ari.minimal", "followers_count": 21500, "biography": "minimal wardrobe · 30대 직장인", "recent": posts(rng, 6, [0.5, 0.3, 0.1, 0.1], 610)}]})
    dump("reach_estimate", {"source": "mock", "targeting": {"countries": ["KR"], "age_min": 25, "age_max": 34, "genders": ["women"]},
                            "data": {"instagram_reels": {"estimate_dau": 412000, "estimate_mau": 1260000},
                                     "instagram_feed": {"estimate_dau": 388000, "estimate_mau": 1310000},
                                     "instagram_stories": {"estimate_dau": 351000, "estimate_mau": 1190000},
                                     "facebook_feed": {"estimate_dau": 96000, "estimate_mau": 540000}}})
    print("mock written to", OUT)


if __name__ == "__main__":
    main()
