package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Seasonality: monthly demand and ad-cost indexes relative to an average month (mean 1.0), with sources.
// Applied as a prior for the month being planned; never used to calibrate or backtest.
type Seasonality struct {
	Market   string             `json:"market"`
	Category string             `json:"category"`
	Label    string             `json:"label"`
	Demand   map[string]float64 `json:"demand"`
	AdCost   map[string]float64 `json:"ad_cost"`
	Sources  []map[string]any   `json:"sources"`
	Note     string             `json:"note"`
}

type Season struct {
	Month  int     `json:"month"`
	Name   string  `json:"name"`
	Demand float64 `json:"demand"`
	AdCost float64 `json:"ad_cost"`
}

// SeasonCatalog: data/seasonality.json. One demand table per category (Census monthly retail sales by kind of
// business), one ad-cost table shared by every category.
type SeasonCatalog struct {
	Market       string                    `json:"market"`
	Categories   map[string]SeasonCategory `json:"categories"`
	AdCost       map[string]float64        `json:"ad_cost"`
	AdCostSource map[string]any            `json:"ad_cost_source"`
	Note         string                    `json:"note"`
}

type SeasonCategory struct {
	Label   string             `json:"label"`
	Demand  map[string]float64 `json:"demand"`
	Sources []map[string]any   `json:"sources"`
}

func LoadSeasonCatalog(path string) (*SeasonCatalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c SeasonCatalog
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// For: the active table for a category; nil when the catalog or the category is unknown (no seasonal prior).
func (c *SeasonCatalog) For(category string) *Seasonality {
	if c == nil {
		return nil
	}
	cat, ok := c.Categories[category]
	if !ok {
		return nil
	}
	return &Seasonality{Market: c.Market, Category: category, Label: cat.Label, Demand: cat.Demand, AdCost: c.AdCost,
		Sources: append(append([]map[string]any{}, cat.Sources...), c.AdCostSource), Note: c.Note}
}

// LoadSeasonality: the catalog's table for one category.
func LoadSeasonality(path, category string) (*Seasonality, error) {
	c, err := LoadSeasonCatalog(path)
	if err != nil {
		return nil, err
	}
	return c.For(category), nil
}

// SeasonFor: the indexes for a calendar month (1-12); 1.0 / 1.0 when the table is nil or month is 0.
func SeasonFor(t *Seasonality, month int) Season {
	s := Season{Month: month, Demand: 1, AdCost: 1}
	if month >= 1 && month <= 12 {
		s.Name = time.Month(month).String()
	}
	if t == nil || month == 0 {
		return s
	}
	key := fmt.Sprint(month)
	if v, ok := t.Demand[key]; ok {
		s.Demand = v
	}
	if v, ok := t.AdCost[key]; ok {
		s.AdCost = v
	}
	return s
}

// PlanMonth: the calendar month that starts the week after the last data week, plus monthsAhead.
func PlanMonth(b *Brand, monthsAhead int) int {
	last, err := time.Parse("2006-01-02", b.Weeks[len(b.Weeks)-1].Week)
	if err != nil {
		return 0
	}
	return int(last.AddDate(0, monthsAhead, 7).Month())
}

// SeasonBlock: the prompt text for the planner. Numbers come from the table; Spark may only quote them.
func SeasonBlock(t *Seasonality, month int) string {
	s := SeasonFor(t, month)
	if t == nil || month == 0 {
		return "No month given."
	}
	return fmt.Sprintf("Month: %s. Demand index %.2f (1.00 = an average month; %s %s, public data). Ad cost index %.2f (1.00 = average month). "+
		"The engine multiplies the order rate by the demand index and divides ad reach by the ad cost index; the adversary can take the seasonal lift away.",
		s.Name, s.Demand, t.Market, t.Label, s.AdCost)
}

// SeasonGoal: a month-sized goal, the average-month goal scaled by the month's demand index (half-up).
func SeasonGoal(goal, demand float64) int { return halfUp(goal * demand) }
