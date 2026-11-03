"""Month review: compare what happened to what the committed plan expected, update rates."""
from engine.model import run_plan, safe_div

BLEND = 0.3   # new = 0.7 * old + 0.3 * observed. One month is evidence, not a verdict.


def actual_month(brand, weeks=4):
    rows = brand["weeks"][-weeks:]
    seeded = sum(w["creators_seeded"] for w in rows)
    posted = sum(w["creators_posted"] for w in rows)
    organic = sum(sum(w["engagement"].values()) for w in rows)
    return {
        "orders": sum(w["orders"] for w in rows),
        "spend_usd": sum(w["spend_usd"] for w in rows),
        "engagement": organic + sum(w["creator_engagement"] + w["ad_engagement"] for w in rows),
        "creators_seeded": seeded,
        "creators_posted": posted,
        "seed_post_rate": safe_div(posted, seeded),
        "order_per_engagement": safe_div(sum(w["orders"] for w in rows),
                                         organic + sum(w["creator_engagement"] + w["ad_engagement"] for w in rows)),
    }


def review(brand, committed_plan, rates):
    actual = actual_month(brand)
    expected = run_plan(committed_plan, rates, brand)
    lo, hi = expected["orders"] * 0.85, expected["orders"] * 1.15
    inside = lo <= actual["orders"] <= hi

    changes = []
    new = rates.copy()
    if actual["creators_seeded"] > 0 and abs(actual["seed_post_rate"] - rates.seed_post_rate) > 0.1:
        new.seed_post_rate = round((1 - BLEND) * rates.seed_post_rate + BLEND * actual["seed_post_rate"], 2)
        changes.append({"rate": "seed_post_rate", "old": round(rates.seed_post_rate, 2), "new": new.seed_post_rate})
    if abs(safe_div(actual["order_per_engagement"], rates.order_per_engagement) - 1) > 0.15:
        new.order_per_engagement = round((1 - BLEND) * rates.order_per_engagement + BLEND * actual["order_per_engagement"], 5)
        changes.append({"rate": "order_per_engagement", "old": round(rates.order_per_engagement, 4),
                        "new": round(new.order_per_engagement, 4)})

    return {
        "actual": actual,
        "expected_orders": round(expected["orders"], 1),
        "range": [round(lo, 1), round(hi, 1)],
        "inside_range": inside,
        "diagnosis": diagnosis(actual, rates, inside, changes),
        "rate_changes": changes,
        "new_rates": new,
    }


def diagnosis(actual, rates, inside, changes):
    """One sentence from a template. Not a model."""
    if inside:
        return "Orders landed inside the expected range. No rate changed."
    names = [c["rate"] for c in changes]
    if "seed_post_rate" in names and "order_per_engagement" not in names:
        return (f"Only {actual['creators_posted']} of {actual['creators_seeded']} seeded creators posted, "
                f"so seeded engagement never arrived and orders followed. Your own conversion held. "
                f"This was a creator problem, not a product or audience problem.")
    if "order_per_engagement" in names:
        return "Engagement came in as expected but fewer of it turned into orders. Conversion was off."
    return "Orders were outside the range, but no single rate explains it. Watch next month."
