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
