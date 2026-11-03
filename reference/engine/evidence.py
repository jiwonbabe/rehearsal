"""Evidence from the brand's own Meta data: averages and shares only. Same formulas as engine/evidence.go."""
import json
import math
import os

from engine.model import POST_TYPES

FORMATS = ["feed", "reels", "stories", "threads"]
KIND = {"feed": "FEED", "reels": "REELS", "stories": "STORY"}


def load_meta(d):
    out = {}
    for name in ["account", "media", "threads", "hashtags", "peers", "creators", "reach_estimate"]:
        with open(os.path.join(d, name + ".json")) as f:
            out[name] = json.load(f)
    return out


def avg(xs):
    return int(math.floor(sum(xs) / len(xs) + 0.5)) if xs else 0


def share2(n, d):
    return round(n / d, 2) if d else 0


def top(m):
    k = max(sorted(m), key=lambda k: m[k])
    return k, round(m[k], 2)


def post_mix(items):
    n = len(items)
    return {t: share2(sum(1 for p in items if p["post_type"] == t), n) for t in POST_TYPES}


def evidence(m):
    demo = m["account"]["follower_demographics"]
    top_age, top_age_share = top(demo["age"])
    top_city, top_city_share = top(demo["city"])
    audience = {"followers": m["account"]["followers_count"], "top_age": top_age, "top_age_share": top_age_share,
                "women_share": round(demo["gender"].get("women", 0), 2), "top_city": top_city, "top_city_share": top_city_share}
    media = m["media"]["data"]
    feed_avg = avg([p["insights"]["reach"] for p in media if p["media_product_type"] == "FEED"])
    formats = []
    for f in FORMATS:
        if f == "threads":
            rows = m["threads"]["data"]
            reach = avg([p["insights"]["views"] for p in rows])
            saves = avg([p["insights"]["replies"] for p in rows])
            taps = 0
        else:
            rows = [p for p in media if p["media_product_type"] == KIND[f]]
            reach = avg([p["insights"]["reach"] for p in rows])
            saves = avg([p["insights"]["saved"] for p in rows])
            taps = avg([p["insights"]["link_taps"] for p in rows])
        formats.append({"format": f, "posts": len(rows), "avg_reach": reach, "avg_saves": saves,
                        "link_taps_per_post": taps, "reach_vs_feed": share2(reach, feed_avg)})
    post_types = []
    for t in POST_TYPES:
        rows = [p for p in media if p["post_type"] == t and p["media_product_type"] != "STORY"]
        post_types.append({"post_type": t, "posts": len(rows), "avg_reach": avg([p["insights"]["reach"] for p in rows]),
                           "avg_saves": avg([p["insights"]["saved"] for p in rows])})
    best = max(post_types, key=lambda r: (r["avg_saves"], -POST_TYPES.index(r["post_type"])))["post_type"]
    hashtags = [{"hashtag": h["hashtag"], "share_by_type": post_mix(h["top_media"]),
                 "avg_likes_by_type": {t: avg([p["like_count"] for p in h["top_media"] if p["post_type"] == t]) for t in POST_TYPES}}
                for h in m["hashtags"]["data"]]
    peers = [{"username": p["username"], "followers": p["followers_count"], "post_mix": post_mix(p["recent"]),
              "avg_likes": avg([x["like_count"] for x in p["recent"]])} for p in m["peers"]["data"]]
    creators = []
    for c in m["creators"]["data"]:
        likes = avg([x["like_count"] for x in c["recent"]])
        band = "small" if c["followers_count"] < 10000 else "mid" if c["followers_count"] < 100000 else "large"
        creators.append({"username": c["username"], "followers": c["followers_count"], "followers_band": band,
                         "engagement_rate": round(likes / c["followers_count"], 3), "post_mix": post_mix(c["recent"]),
                         "biography": c["biography"]})
    return {"source": m["account"].get("source", ""), "days": m["media"].get("days", 0), "audience": audience,
            "formats": formats, "post_types": post_types, "best_type": best,
            "category": {"hashtags": hashtags, "peers": peers}, "creators": creators, "reach": m["reach_estimate"]["data"]}
