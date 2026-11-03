"""Write golden fixtures from the Python reference engine. The Go engine must match them.

Run from repo root:  python reference/dump_fixtures.py
"""
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, os.path.join(ROOT, "reference"))

from engine.model import Plan, calibrate, run_plan, survived, backtest, load_brand  # noqa: E402
from engine.stress import sample_variants, stress, verdict, flip, baseline_plan  # noqa: E402
from engine.shocks import FALLBACK_COMBOS  # noqa: E402
from engine.review import review  # noqa: E402
from engine.goals import goal_suggestions  # noqa: E402
from engine.launch import launch_spec  # noqa: E402
from engine.evidence import load_meta, evidence  # noqa: E402

OUT = os.path.join(ROOT, "reference", "fixtures")

ADS = {"id": "pay_for_reach", "ranges": {
    "posts_per_week": [3, 5], "creator_count": [0, 0], "ads_usd": [120, 260],
    "promo_week": [None], "promo_discount": [0, 0],
    "post_mix": {"on_hand": [0.4, 0.6], "close_up": [0.2, 0.4], "flat_lay": [0.1, 0.3], "on_neck": [0, 0.1]}}}
SEEDING = {"id": "send_to_creators", "ranges": {
    "posts_per_week": [3, 4], "creator_count": [4, 8], "ads_usd": [0, 60],
    "promo_week": [None], "promo_discount": [0, 0],
    "post_mix": {"on_hand": [0.4, 0.6], "close_up": [0.2, 0.4], "flat_lay": [0.1, 0.3], "on_neck": [0, 0.1]}}}


def dump(name, obj):
    os.makedirs(OUT, exist_ok=True)
    with open(os.path.join(OUT, name + ".json"), "w") as f:
        json.dump(obj, f, indent=1)


def main():
    brand = load_brand(os.path.join(ROOT, "data", "brand.json"))
    rates = calibrate(brand)
    dump("rates", rates.to_dict())
    dump("backtest", backtest(brand, rates))

    plans = sample_variants(ADS, n=20) + sample_variants(SEEDING, n=20, seed=8)
    dump("plans", [p.to_dict() for p in plans])

    single = run_plan(plans[0], rates, brand)
    shocked = run_plan(plans[25], rates, brand, shocks=["reach_drop", "creator_noshow"])
    dump("run_plan", {"clean": single, "shocked": shocked,
                      "clean_survived": survived(single, brand), "shocked_survived": survived(shocked, brand)})

    stressed = stress(plans, FALLBACK_COMBOS, rates, brand)
    dump("stress", stressed)
    out = verdict(plans, stressed, rates, brand)
    dump("verdict", out)

    best = {d["plan"]["direction"]: Plan.from_dict(d["plan"]) for d in out}
    dump("flip", flip(best["pay_for_reach"], best["send_to_creators"], FALLBACK_COMBOS, rates, brand))

    committed = Plan("oct", "send_to_creators", 3,
                     {"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}, 4, 0)
    r = review(brand, committed, rates)
    dump("evidence", evidence(load_meta(os.path.join(ROOT, "data", "meta", "bnd.jewelry"))))
    dump("launch", {"ads": launch_spec(best["pay_for_reach"], brand), "seeding": launch_spec(committed, brand)})
    r["new_rates"] = r["new_rates"].to_dict()
    dump("review", r)
    dump("goals", goal_suggestions(brand))
    print("fixtures written to", OUT)


if __name__ == "__main__":
    main()
