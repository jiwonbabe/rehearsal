"""Goal suggestions from the last N weeks: hold / grow / stretch monthly order goals."""


def goal_suggestions(brand, weeks_back=13):
    rows = brand["weeks"][-weeks_back:]
    n = len(rows)
    orders = sum(w["orders"] for w in rows)
    spend = sum(w["spend_usd"] for w in rows)
    b = brand["brand"]
    out = {"avg_weekly_orders": 0, "hold": 0, "grow": 0, "stretch": 0, "avg_monthly_spend_usd": 0,
           "avg_price_usd": b["avg_price_usd"], "current_goal": b["goal_orders_per_month"],
           "current_budget": b["budget_usd_per_month"], "weeks_used": n}
    if n == 0:
        return out
    hold = round(orders / n * 4)
    k = min(4, n)
    out.update(avg_weekly_orders=round(orders / n, 1), hold=hold, grow=round(hold * 1.10),
               stretch=max(sum(w["orders"] for w in rows[i:i + k]) for i in range(n - k + 1)),
               avg_monthly_spend_usd=round(spend / n * 4))
    return out
