package handler

import (
	"math/big"
	"net/http"
	"time"

	"github.com/fingoat/api/internal/config"
	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/middleware"
	"github.com/fingoat/api/internal/stores"
	"github.com/jackc/pgx/v5/pgtype"
)

// SummaryHandler handles GET /api/summary.
type SummaryHandler struct {
	store  stores.SummaryStore
	config *config.Config
}

// NewSummaryHandler constructs a SummaryHandler.
func NewSummaryHandler(q stores.SummaryStore, cfg *config.Config) *SummaryHandler {
	return &SummaryHandler{store: q, config: cfg}
}

// summaryPeriod is the period returned in the response.
type summaryPeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// categoryTotal represents aggregated amounts for one category.
type categoryTotal struct {
	Category      string `json:"category"`
	TotalExpenses string `json:"totalExpenses"`
	TotalIncome   string `json:"totalIncome"`
}

// goalProgressItem holds progress info for a single goal.
type goalProgressItem struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	TargetAmount  string  `json:"targetAmount"`
	CurrentAmount string  `json:"currentAmount"`
	ProgressPct   float64 `json:"progressPct"`
}

// summaryResponse is the full response body for GET /api/summary.
type summaryResponse struct {
	Period        summaryPeriod      `json:"period"`
	TotalIncome   string             `json:"totalIncome"`
	TotalExpenses string             `json:"totalExpenses"`
	Net           string             `json:"net"`
	ByCategory    []categoryTotal    `json:"byCategory"`
	GoalProgress  []goalProgressItem `json:"goalProgress"`
}

// GetSummary handles GET /api/summary.
func (h *SummaryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	from, to := currentMonthBounds()

	if v := r.URL.Query().Get("from"); v != "" {
		d, ok := parseDate(v)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid from date, use YYYY-MM-DD")
			return
		}
		from = d
	}
	if v := r.URL.Query().Get("to"); v != "" {
		d, ok := parseDate(v)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid to date, use YYYY-MM-DD")
			return
		}
		to = d
	}

	rows, err := h.store.GetTransactionSummary(r.Context(), db.GetTransactionSummaryParams{
		UserID: userID,
		From:   &from,
		To:     &to,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch transaction summary")
		return
	}

	goals, err := h.store.ListGoalsByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch goals")
		return
	}

	resp := buildSummaryResponse(from, to, rows, goals)
	writeJSON(w, http.StatusOK, resp)
}

// currentMonthBounds returns pgtype.Date values for the first and last day of
// the current month in UTC.
func currentMonthBounds() (pgtype.Date, pgtype.Date) {
	now := time.Now().UTC()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Last day = first day of next month minus one day.
	last := first.AddDate(0, 1, -1)

	var fromD, toD pgtype.Date
	_ = fromD.Scan(first.Format("2006-01-02"))
	_ = toD.Scan(last.Format("2006-01-02"))
	return fromD, toD
}

// numericToBigRat converts a pgtype.Numeric to a *big.Rat.
// Returns zero if the value is invalid or NaN.
func numericToBigRat(n pgtype.Numeric) *big.Rat {
	if !n.Valid || n.NaN || n.Int == nil {
		return new(big.Rat)
	}
	// value = Int * 10^Exp
	rat := new(big.Rat).SetInt(n.Int)
	if n.Exp >= 0 {
		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)
		rat.Mul(rat, new(big.Rat).SetInt(scale))
	} else {
		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)
		rat.Quo(rat, new(big.Rat).SetInt(scale))
	}
	return rat
}

// ratToString formats a *big.Rat as a decimal string with 2 decimal places.
func ratToString(r *big.Rat) string {
	// Use big.Float for formatting.
	f, _ := new(big.Float).SetPrec(128).SetRat(r).Float64()
	return formatMoney(f)
}

// formatMoney formats a float64 as a money string with 2 decimal places.
func formatMoney(f float64) string {
	// Use strconv-style formatting via big.Float to avoid importing fmt for %.2f
	// but fmt is already available via writeJSON. Use Sprintf here.
	return formatFloat2(f)
}

// formatFloat2 formats f with exactly 2 decimal places.
func formatFloat2(f float64) string {
	// Build the string manually to avoid a circular import with fmt.
	// We rely on strconv.AppendFloat under the hood via big.Float.Text.
	bf := new(big.Float).SetPrec(64).SetFloat64(f)
	return bf.Text('f', 2)
}

// buildSummaryResponse assembles the summaryResponse from raw DB rows and goals.
func buildSummaryResponse(from, to pgtype.Date, rows []db.TransactionSummaryRow, goals []db.Goal) summaryResponse {
	// Accumulate totals per category. Key: category string.
	type catAccum struct {
		income  *big.Rat
		expense *big.Rat
	}
	byCategory := make(map[string]*catAccum)
	totalIncome := new(big.Rat)
	totalExpenses := new(big.Rat)

	for _, row := range rows {
		rat := numericToBigRat(row.Total)
		cat := row.Category

		if _, ok := byCategory[cat]; !ok {
			byCategory[cat] = &catAccum{income: new(big.Rat), expense: new(big.Rat)}
		}

		switch row.Type {
		case "income":
			totalIncome.Add(totalIncome, rat)
			byCategory[cat].income.Add(byCategory[cat].income, rat)
		case "expense":
			totalExpenses.Add(totalExpenses, rat)
			byCategory[cat].expense.Add(byCategory[cat].expense, rat)
		}
	}

	// Build ordered category slice. Preserve insertion order via a separate
	// slice to keep output deterministic.
	var catOrder []string
	seen := make(map[string]bool)
	for _, row := range rows {
		if !seen[row.Category] {
			catOrder = append(catOrder, row.Category)
			seen[row.Category] = true
		}
	}

	cats := make([]categoryTotal, 0, len(catOrder))
	for _, cat := range catOrder {
		acc := byCategory[cat]
		cats = append(cats, categoryTotal{
			Category:      cat,
			TotalExpenses: ratToString(acc.expense),
			TotalIncome:   ratToString(acc.income),
		})
	}
	if cats == nil {
		cats = []categoryTotal{}
	}

	// net = income - expenses
	net := new(big.Rat).Sub(totalIncome, totalExpenses)

	// Goal progress
	gps := make([]goalProgressItem, 0, len(goals))
	for _, g := range goals {
		target := numericToBigRat(g.TargetAmount)
		current := numericToBigRat(g.CurrentAmount)

		var pct float64
		if target.Sign() > 0 {
			ratio := new(big.Rat).Quo(current, target)
			f, _ := ratio.Float64()
			pct = f * 100
		}

		gps = append(gps, goalProgressItem{
			ID:            g.ID,
			Title:         g.Title,
			TargetAmount:  ratToString(target),
			CurrentAmount: ratToString(current),
			ProgressPct:   pct,
		})
	}

	fromStr := from.Time.Format("2006-01-02")
	toStr := to.Time.Format("2006-01-02")

	return summaryResponse{
		Period:        summaryPeriod{From: fromStr, To: toStr},
		TotalIncome:   ratToString(totalIncome),
		TotalExpenses: ratToString(totalExpenses),
		Net:           ratToString(net),
		ByCategory:    cats,
		GoalProgress:  gps,
	}
}
