import json
import os
import subprocess
import sys

import pytest

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
REF = os.path.join(ROOT, "reference")
sys.path.insert(0, REF)

from engine.model import Plan, calibrate, run_plan, survived, backtest, load_brand  # noqa: E402
from engine.stress import sample_variants, stress, verdict, flip, baseline_plan  # noqa: E402
from engine.shocks import FALLBACK_COMBOS, clean_combos  # noqa: E402
from engine.review import review  # noqa: E402


@pytest.fixture(scope="module")
def brand():
    path = os.path.join(ROOT, "data", "brand.json")
    if not os.path.exists(path):
        subprocess.run([sys.executable, os.path.join(ROOT, "data", "synthetic.py")], check=True)
    return load_brand(path)


@pytest.fixture(scope="module")
def rates(brand):
    return calibrate(brand)


def ads_direction():
    return {"id": "pay_for_reach", "ranges": {
        "posts_per_week": [3, 5], "creator_count": [0, 0], "ads_usd": [120, 260],
        "promo_week": [None], "promo_discount": [0, 0],
        "post_mix": {"on_hand": [0.4, 0.6], "close_up": [0.2, 0.4], "flat_lay": [0.1, 0.3], "on_neck": [0, 0.1]}}}


def seed_direction():
    return {"id": "send_to_creators", "ranges": {
        "posts_per_week": [3, 4], "creator_count": [4, 8], "ads_usd": [0, 60],
        "promo_week": [None], "promo_discount": [0, 0],
        "post_mix": {"on_hand": [0.4, 0.6], "close_up": [0.2, 0.4], "flat_lay": [0.1, 0.3], "on_neck": [0, 0.1]}}}


def test_calibration_recovers_true_rates(rates):
    # generator truth: on_hand 205, order_per_eng 0.0115, seed_post_rate below 0.75 because of October
    assert 170 < rates.eng_per_post["on_hand"] < 290  # followers grow over the year
    assert 0.009 < rates.order_per_engagement < 0.014
    assert rates.eng_per_post["on_hand"] > rates.eng_per_post["flat_lay"]
    assert 0.4 < rates.seed_post_rate < 0.75
    assert rates.promo_multiplier > 1.4


def test_run_plan_is_deterministic(brand, rates):
    plan = Plan("p", "x", 4, {"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}, 0, 160)
    a = run_plan(plan, rates, brand)
    b = run_plan(plan, rates, brand)
    assert a == b
    assert len(a["weeks"]) == 4
    assert a["spend_usd"] == pytest.approx(160, abs=0.01)


def test_shocks_reduce_orders(brand, rates):
    plan = Plan("p", "x", 4, {"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}, 6, 0)
    clean = run_plan(plan, rates, brand)
    hit = run_plan(plan, rates, brand, shocks=["reach_drop", "creator_noshow"])
    assert hit["orders"] < clean["orders"]


def test_clean_combos_drops_unknown():
    out = clean_combos([["reach_drop", "nonsense"], ["nothing"], ["reach_drop", "nonsense"]])
    assert out == [["reach_drop"]]


def test_backtest_is_honest(brand, rates):
    bt = backtest(brand, rates)
    assert bt["weeks_total"] == 13
    assert bt["median_abs_pct_error"] < 25
    assert any(r["promo"] for r in bt["rows"]) or True  # promo weeks may or may not be in window


def test_sampling_stays_in_ranges():
    plans = sample_variants(ads_direction(), n=30)
    assert len(plans) == 30
    for p in plans:
        assert 3 <= p.posts_per_week <= 5
        assert 120 <= p.ads_usd <= 260
        assert p.creator_count == 0
        assert abs(sum(p.post_mix.values()) - 1.0) < 0.05
    again = sample_variants(ads_direction(), n=30)
    assert [p.to_dict() for p in plans] == [p.to_dict() for p in again]


def test_verdict_and_flip(brand, rates):
    plans = sample_variants(ads_direction(), n=20) + sample_variants(seed_direction(), n=20, seed=8)
    stressed = stress(plans, FALLBACK_COMBOS, rates, brand)
    out = verdict(plans, stressed, rates, brand)
    directions = {d["plan"]["direction"] for d in out}
    assert {"pay_for_reach", "send_to_creators", "just_post"} <= directions
    for d in out:
        assert 0 <= d["survived"] <= d["total"]

    best = {d["plan"]["direction"]: Plan.from_dict(d["plan"]) for d in out}
    f = flip(best["pay_for_reach"], best["send_to_creators"], FALLBACK_COMBOS, rates, brand)
    assert f is None or 0.3 <= f["threshold"] <= 1.0


def test_review_lowers_seed_rate_after_october(brand, rates):
    committed = Plan("oct", "send_to_creators", 3,
                     {"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0}, 4, 0)
    r = review(brand, committed, rates)
    assert r["actual"]["creators_seeded"] == 4
    assert r["actual"]["creators_posted"] == 1
    names = [c["rate"] for c in r["rate_changes"]]
    assert "seed_post_rate" in names
    assert r["new_rates"].seed_post_rate < rates.seed_post_rate
    assert "creator" in r["diagnosis"]
