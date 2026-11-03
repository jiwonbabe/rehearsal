package engine

// GoalSuggestions: candidate monthly order goals from the last N weeks of the brand's own data.
type GoalSuggestions struct {
	AvgWeeklyOrders    float64 `json:"avg_weekly_orders"`
	Hold               float64 `json:"hold"`
	Grow               float64 `json:"grow"`
	Stretch            float64 `json:"stretch"`
	AvgMonthlySpendUsd float64 `json:"avg_monthly_spend_usd"`
	AvgPriceUsd        float64 `json:"avg_price_usd"`
	CurrentGoal        float64 `json:"current_goal"`
	CurrentBudget      float64 `json:"current_budget"`
	WeeksUsed          int     `json:"weeks_used"`
}

func RunGoalSuggestions(b *Brand, weeksBack int) GoalSuggestions {
	start := len(b.Weeks) - weeksBack
	if start < 0 {
		start = 0
	}
	rows := b.Weeks[start:]
	n := float64(len(rows))
	orders, spend := 0.0, 0.0
	for _, w := range rows {
		orders += w.Orders
		spend += w.SpendUsd
	}
	g := GoalSuggestions{AvgPriceUsd: b.Brand.AvgPriceUsd, CurrentGoal: b.Brand.GoalOrdersPerMonth,
		CurrentBudget: b.Brand.BudgetUsdPerMonth, WeeksUsed: len(rows)}
	if len(rows) == 0 {
		return g
	}
	g.AvgWeeklyOrders = round1(orders / n)
	g.Hold = roundTo(orders/n*4, 0)
	g.Grow = roundTo(g.Hold*1.10, 0)
	g.AvgMonthlySpendUsd = roundTo(spend/n*4, 0)
	// Stretch: best 4 consecutive weeks (all weeks if fewer than 4).
	k := 4
	if len(rows) < k {
		k = len(rows)
	}
	for i := 0; i+k <= len(rows); i++ {
		sum := 0.0
		for _, w := range rows[i : i+k] {
			sum += w.Orders
		}
		if sum > g.Stretch {
			g.Stretch = sum
		}
	}
	return g
}
