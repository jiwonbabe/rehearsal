"""Write golden fixtures from the Python reference engine. The Go engine must match them.

Run from repo root:  python reference/dump_fixtures.py
"""
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, os.path.join(ROOT, "reference"))

from engine.model import Plan, calibrate, run_plan, survived, backtest, load_brand, season_goal  # noqa: E402
from engine.stress import sample_variants, stress, verdict, flip, baseline_plan, baseline_gaps  # noqa: E402
from engine.shocks import FALLBACK_COMBOS  # noqa: E402
from engine.review import review  # noqa: E402
from engine.goals import goal_suggestions  # noqa: E402
from engine.launch import launch_spec  # noqa: E402
from engine.evidence import load_meta, evidence  # noqa: E402
from engine.trust import lever_coverage, untested, truth_check, holdout  # noqa: E402

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
    season_brand = dict(brand)
    with open(os.path.join(ROOT, "data", "seasonality.json")) as f:
        catalog = json.load(f)
    season_brand["season"] = {"demand": catalog["categories"][brand["brand"]["category"]]["demand"], "ad_cost": catalog["ad_cost"]}
    nov = Plan.from_dict(dict(plans[0].to_dict(), month=11))
    dec = Plan.from_dict(dict(plans[0].to_dict(), month=12))
    dump("season", {"november": run_plan(nov, rates, season_brand), "december": run_plan(dec, rates, season_brand),
                    "december_flat": run_plan(dec, rates, season_brand, shocks=["season_flat"]),
                    "no_month": run_plan(plans[0], rates, season_brand),
                    "goal_november": season_goal(brand["brand"]["goal_orders_per_month"], season_brand["season"]["demand"]["11"]),
                    "goal_december": season_goal(brand["brand"]["goal_orders_per_month"], season_brand["season"]["demand"]["12"])})
    dump("truth", truth_check(brand, rates))
    dump("holdout", holdout(brand, 13))
    no_ads = dict(brand, weeks=[dict(w, ads_usd=0, ad_engagement=0) for w in brand["weeks"]])
    ads_plan = Plan("x", "pay_for_reach", 4, {"on_hand": 0.6, "close_up": 0.2, "flat_lay": 0.1, "on_neck": 0.1}, 0, 300)
    dump("coverage", {"coverage": lever_coverage(brand), "no_ads": lever_coverage(no_ads),
                      "untested_ads_plan": untested(ads_plan, lever_coverage(no_ads)), "tested_ads_plan": untested(ads_plan, lever_coverage(brand))})
    ev = evidence(load_meta(os.path.join(ROOT, "data", "meta", "demo.brand")))
    dump("evidence", ev)
    pool_brand = dict(brand, creator_pool=ev["creators"])
    seed_plan = Plan("seed", "send_to_creators", 3, {"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}, 4, 0)
    big_plan = Plan("seed8", "send_to_creators", 3, {"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}, 8, 0)
    from engine.model import creator_weights, expected_engagement
    dump("creator_pool", {"weights": creator_weights(ev["creators"]), "expected": [expected_engagement(c) for c in ev["creators"]],
                          "plan": run_plan(seed_plan, rates, pool_brand), "plan_cycles": run_plan(big_plan, rates, pool_brand),
                          "plan_no_pool": run_plan(seed_plan, rates, brand), "plan_offset2": run_plan(seed_plan, rates, pool_brand, offset=2)})
    dec = Plan.from_dict(dict(committed.to_dict(), month=12))
    dump("launch", {"ads": launch_spec(best["pay_for_reach"], brand), "seeding": launch_spec(committed, brand), "december": launch_spec(dec, brand)})
    dump("gaps", baseline_gaps(out))
    r["new_rates"] = r["new_rates"].to_dict()
    dump("review", r)
    dump("goals", goal_suggestions(brand))
    print("fixtures written to", OUT)


if __name__ == "__main__":
    main()
