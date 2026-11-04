"""Sample variants inside a direction, run the matrix, pick survivors, find the flip."""
import random

from engine.model import POST_TYPES, Plan, run_plan, survived
from engine.shocks import clean_combos, FALLBACK_COMBOS


def pick(rng, lo, hi, step=None):
    if lo == hi:
        return lo
    if step:
        n = int(round((hi - lo) / step))
        return lo + step * rng.randint(0, n)
    return rng.randint(int(lo), int(hi))


def sample_variants(direction, n=60, seed=7):
    """Plans inside the direction's ranges. Deterministic for a given seed."""
    rng = random.Random(seed)
    ranges = direction["ranges"]
    plans = []
    for i in range(n):
        mix = {}
        for t in POST_TYPES:
            lo, hi = ranges["post_mix"].get(t, [0, 0])
            mix[t] = pick(rng, lo, hi, 0.1)
        total = sum(mix.values()) or 1.0
        mix = {t: round(v / total, 2) for t, v in mix.items()}

        promo_choices = ranges.get("promo_week", [None])
        promo_week = rng.choice(promo_choices)
        promo_lo, promo_hi = ranges.get("promo_discount", [0, 0])
        promo_discount = pick(rng, promo_lo, promo_hi, 0.05) if promo_week else 0.0

        plans.append(Plan(
            plan_id=f"{direction['id']}-{i:03d}",
            direction=direction["id"],
            posts_per_week=pick(rng, *ranges["posts_per_week"]),
            post_mix=mix,
            creator_count=pick(rng, *ranges["creator_count"]),
            ads_usd=pick(rng, ranges["ads_usd"][0], ranges["ads_usd"][1], 20),
            promo_week=promo_week,
            promo_discount=promo_discount,
        ))
    return plans


def baseline_plan():
    return Plan("baseline", "just_post", 5,
                {"on_hand": 0.5, "close_up": 0.3, "flat_lay": 0.2, "on_neck": 0.0}, 0, 0)


def stress(plans, shock_combos, rates, brand):
    """matrix[plan_id] = list of cells, one per combo."""
    combos = clean_combos(shock_combos) or FALLBACK_COMBOS
    matrix = {}
    for plan in plans:
        cells = []
        for combo in combos:
            result = run_plan(plan, rates, brand, shocks=combo)
            cells.append({
                "shocks": combo,
                "orders": round(result["orders"], 1),
                "spend_usd": round(result["spend_usd"], 1),
                "survived": survived(result, brand),
            })
        matrix[plan.plan_id] = cells
    return {"combos": combos, "matrix": matrix}


def score(plan, cells):
    """Sort key: more survivals, then lower spend."""
    survivals = sum(1 for c in cells if c["survived"])
    spend = sum(c["spend_usd"] for c in cells) / len(cells)
    return (survivals, -spend)


def verdict(plans, stressed, rates, brand):
    """Best plan per direction + the baseline, with the shock that kills each."""
    by_direction = {}
    for plan in plans:
        cells = stressed["matrix"][plan.plan_id]
        current = by_direction.get(plan.direction)
        if current is None or score(plan, cells) > score(current["plan"], current["cells"]):
            by_direction[plan.direction] = {"plan": plan, "cells": cells}

    out = []
    for direction, best in by_direction.items():
        out.append(describe(best["plan"], best["cells"], rates, brand))

    base = baseline_plan()
    if plans:
        base.month = plans[0].month  # the baseline lives in the same month as the plans it is compared with
    base_cells = stress([base], stressed["combos"], rates, brand)["matrix"]["baseline"]
    out.append(describe(base, base_cells, rates, brand))
    out.sort(key=lambda d: (-d["survived"], d["spend_usd"]))
    return out


def describe(plan, cells, rates, brand):
    survivals = sum(1 for c in cells if c["survived"])
    dead = [c for c in cells if not c["survived"]]
    dies_if = dead[0]["shocks"] if dead else []
    orders = sorted(c["orders"] for c in cells)
    clean = run_plan(plan, rates, brand)
    return {
        "plan": plan.to_dict(),
        "survived": survivals,
        "total": len(cells),
        "orders_bad_month": [orders[len(orders) // 5], orders[-1 - len(orders) // 5]],
        "spend_usd": round(clean["spend_usd"]),
        "dies_if": dies_if,
    }


def flip(plan_a, plan_b, combos, rates, brand, param="seed_post_rate", lo=0.3, hi=1.0, step=0.02):
    """Sweep one rate; return the value where B starts beating A, or None."""
    def wins(a, b, r):
        sa = score(a, stress([a], combos, r, brand)["matrix"][a.plan_id])
        sb = score(b, stress([b], combos, r, brand)["matrix"][b.plan_id])
        return sa >= sb

    base = wins(plan_a, plan_b, rates)
    value = lo
    while value <= hi + 1e-9:
        r = rates.copy()
        setattr(r, param, value)
        if wins(plan_a, plan_b, r) != base:
            return {"param": param, "threshold": round(value, 2), "current": round(getattr(rates, param), 2),
                    "a_wins_now": base}
        value += step
    return None


def baseline_gaps(verdict):
    """What each paid plan buys over the baseline: extra orders in a bad month (midpoints) and the price of each."""
    base = next((v for v in verdict if v["plan"]["plan_id"] == "baseline"), None)
    if base is None:
        return []
    mid = lambda v: (v["orders_bad_month"][0] + v["orders_bad_month"][1]) / 2
    out = []
    for v in verdict:
        if v["plan"]["plan_id"] == "baseline":
            continue
        extra = int(round(mid(v) - mid(base)))
        spend = v["spend_usd"] - base["spend_usd"]
        out.append({"direction": v["plan"]["direction"], "plan_id": v["plan"]["plan_id"], "extra_orders": extra,
                    "extra_spend_usd": spend, "usd_per_extra_order": round(spend / extra, 2) if extra > 0 else 0})
    return out
