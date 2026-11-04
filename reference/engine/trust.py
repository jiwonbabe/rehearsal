"""Trust checks on synthetic data: recover the generator's rates, and a holdout backtest."""
from engine.model import POST_TYPES, calibrate, backtest


def truth_check(brand, rates):
    """Engine-calibrated rates next to the generator's. Only meaningful when brand carries generator_truth."""
    truth = brand.get("generator_truth")
    if not truth:
        return []
    weeks = brand["weeks"]
    avg_followers = sum(w["followers"] for w in weeks) / len(weeks)
    rows = []

    def row(name, true, engine, note=""):
        err = round(abs(engine - true) / true * 100, 1) if true else 0
        rows.append({"rate": name, "true": round(true, 4), "engine": round(engine, 4), "err_pct": err, "note": note})

    for t in POST_TYPES:
        row("eng_per_post." + t, truth["eng_per_post"][t] * avg_followers / truth["followers_base"], rates.eng_per_post[t],
            "generator value scaled by average followers")
    row("order_per_engagement", truth["order_per_engagement"], rates.order_per_engagement)
    row("promo_multiplier", truth["promo_multiplier"], rates.promo_multiplier)
    row("seed_post_rate", truth["seed_post_rate"], rates.seed_post_rate, "October carries a planted creator no-show")
    row("eng_per_creator_post", truth["eng_per_creator_post"], rates.eng_per_creator_post)
    row("ad_eng_per_usd", truth["ad_eng_per_usd"], rates.ad_eng_per_usd)
    return rows


def holdout(brand, weeks_back=13):
    """Calibrate on everything before the last N weeks, then backtest those N weeks the rates never saw."""
    pre = dict(brand)
    pre["weeks"] = brand["weeks"][:-weeks_back]
    return backtest(brand, calibrate(pre), weeks_back)


THIN_WEEKS = 6


def coverage_status(weeks):
    return "none" if weeks == 0 else "thin" if weeks < THIN_WEEKS else "ok"


def lever_coverage(brand):
    """Weeks in the history where each lever (ads, creators, sale) was used, with none / thin / ok."""
    weeks = brand["weeks"]
    counts = {"ads": sum(1 for w in weeks if w["ads_usd"] > 0),
              "creators": sum(1 for w in weeks if w["creators_seeded"] > 0),
              "sale": sum(1 for w in weeks if w["promo_discount"] > 0)}
    return [{"lever": k, "weeks": n, "status": coverage_status(n)} for k, n in counts.items()]


def untested(plan, coverage):
    """Levers a plan relies on that the history never used."""
    uses = {"ads": plan.ads_usd > 0, "creators": plan.creator_count > 0, "sale": plan.promo_week is not None}
    return [c["lever"] for c in coverage if uses[c["lever"]] and c["status"] == "none"]
