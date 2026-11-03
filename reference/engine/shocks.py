"""Shock catalog. Fixed effects; the adversary only picks combinations of these codes."""

CATALOG = {
    "reach_drop": {
        "label": "Reach drops 40% (algorithm change)",
        "weeks": [0, 1, 2, 3],
        "mods": {"organic_mult": 0.6},
    },
    "creator_noshow": {
        "label": "Half of seeded creators never post",
        "weeks": [0, 1, 2, 3],
        "mods": {"seed_post_rate_mult": 0.5},
    },
    "competitor_sale": {
        "label": "A competitor runs a 20% sale",
        "weeks": [1, 2],
        "mods": {"order_rate_mult": 0.7},
    },
    "cpm_spike": {
        "label": "Ad cost per result up 60%",
        "weeks": [0, 1, 2, 3],
        "mods": {"ad_eng_mult": 0.625},
    },
    "shipping_delay": {
        "label": "Shipping delay in week 3",
        "weeks": [2, 3],
        "mods": {"order_rate_mult": 0.85},
    },
    "viral_post": {
        "label": "One post goes viral",
        "weeks": [1],
        "mods": {"organic_mult": 3.0},
    },
    "promo_underestimate": {
        "label": "Sale week does worse than the engine's weakest estimate",
        "weeks": [0, 1, 2, 3],
        "mods": {"promo_mult_mult": 0.8},
    },
}

FALLBACK_COMBOS = [
    ["reach_drop"], ["creator_noshow"], ["competitor_sale"], ["cpm_spike"], ["shipping_delay"],
    ["viral_post"], ["promo_underestimate"],
    ["reach_drop", "creator_noshow"], ["reach_drop", "cpm_spike"], ["creator_noshow", "competitor_sale"],
    ["cpm_spike", "competitor_sale"], ["reach_drop", "shipping_delay"], ["creator_noshow", "shipping_delay"],
    ["competitor_sale", "promo_underestimate"], ["reach_drop", "competitor_sale"],
    ["reach_drop", "creator_noshow", "competitor_sale"], ["cpm_spike", "shipping_delay"],
    ["creator_noshow", "cpm_spike"], ["reach_drop", "promo_underestimate"], ["viral_post", "competitor_sale"],
]


def prior_probabilities(rates):
    """Starting weights, from the brand's own history where we have it."""
    return {
        "reach_drop": 0.22,
        "creator_noshow": round(max(0.1, 1.0 - rates.seed_post_rate), 2),
        "competitor_sale": 0.31,
        "cpm_spike": 0.18,
        "shipping_delay": 0.12,
        "viral_post": 0.08,
        "promo_underestimate": 0.30,
    }


def clean_combos(combos):
    """Drop unknown codes and empty combos; keep at most 20."""
    out = []
    for combo in combos:
        codes = [c for c in combo if c in CATALOG]
        if codes and codes not in out:
            out.append(codes)
    return out[:20]


def mods_for_week(shock_codes, week_index, rates):
    """Multiply the modifiers of every active shock for this week."""
    mods = {}
    for code in shock_codes:
        shock = CATALOG.get(code)
        if not shock or week_index not in shock["weeks"]:
            continue
        for key, value in shock["mods"].items():
            mods[key] = mods.get(key, 1.0) * value
    return mods
