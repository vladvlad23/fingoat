package model

import (
	"math/big"
	"testing"
	"time"

	"github.com/fingoat/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func makeNumeric(s string) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(s)
	return n
}

func TestUserToResponse(t *testing.T) {
	t.Parallel()

	payDay := int32(15)
	u := db.User{
		ID:            1,
		Email:         "test@example.com",
		PasswordHash:  "hashed",
		MonthlyIncome: makeNumeric("5000.00"),
		PaymentDay:    &payDay,
		CreatedAt:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	resp := UserToResponse(u)

	if resp.ID != 1 {
		t.Errorf("ID = %d, want 1", resp.ID)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", resp.Email, "test@example.com")
	}
	if resp.MonthlyIncome != "5000.00" {
		t.Errorf("MonthlyIncome = %q, want %q", resp.MonthlyIncome, "5000.00")
	}
	if resp.PaymentDay == nil || *resp.PaymentDay != 15 {
		t.Errorf("PaymentDay = %v, want 15", resp.PaymentDay)
	}
}

func TestUserToResponse_ZeroIncome(t *testing.T) {
	t.Parallel()

	u := db.User{
		ID:    2,
		Email: "zero@example.com",
		// MonthlyIncome is zero value (not valid)
	}

	resp := UserToResponse(u)
	if resp.MonthlyIncome != "0.00" {
		t.Errorf("MonthlyIncome = %q, want %q", resp.MonthlyIncome, "0.00")
	}
	if resp.PaymentDay != nil {
		t.Errorf("PaymentDay = %v, want nil", resp.PaymentDay)
	}
}

func TestGoalToResponse(t *testing.T) {
	t.Parallel()

	deadline := pgtype.Date{Time: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), Valid: true}
	g := db.Goal{
		ID:            10,
		UserID:        1,
		Title:         "Vacation Fund",
		TargetAmount:  makeNumeric("3000.00"),
		CurrentAmount: makeNumeric("1500.50"),
		Deadline:      &deadline,
		Status:        "active",
		CreatedAt:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	resp := GoalToResponse(g)

	if resp.ID != 10 {
		t.Errorf("ID = %d, want 10", resp.ID)
	}
	if resp.Title != "Vacation Fund" {
		t.Errorf("Title = %q, want %q", resp.Title, "Vacation Fund")
	}
	if resp.TargetAmount != "3000.00" {
		t.Errorf("TargetAmount = %q, want %q", resp.TargetAmount, "3000.00")
	}
	if resp.CurrentAmount != "1500.50" {
		t.Errorf("CurrentAmount = %q, want %q", resp.CurrentAmount, "1500.50")
	}
	if resp.Deadline == nil || *resp.Deadline != "2025-12-31" {
		t.Errorf("Deadline = %v, want %q", resp.Deadline, "2025-12-31")
	}
	if resp.Status != "active" {
		t.Errorf("Status = %q, want %q", resp.Status, "active")
	}
}

func TestGoalToResponse_NoDeadline(t *testing.T) {
	t.Parallel()

	g := db.Goal{
		ID:            11,
		UserID:        1,
		Title:         "No Deadline Goal",
		TargetAmount:  makeNumeric("1000.00"),
		CurrentAmount: makeNumeric("0.00"),
		Deadline:      nil,
		Status:        "active",
	}

	resp := GoalToResponse(g)
	if resp.Deadline != nil {
		t.Errorf("Deadline = %v, want nil", resp.Deadline)
	}
}

func TestGoalToResponse_InvalidDeadline(t *testing.T) {
	t.Parallel()

	deadline := pgtype.Date{Valid: false}
	g := db.Goal{
		ID:       12,
		Deadline: &deadline,
	}

	resp := GoalToResponse(g)
	if resp.Deadline != nil {
		t.Errorf("Deadline = %v, want nil for invalid date", resp.Deadline)
	}
}

func TestTransactionToResponse(t *testing.T) {
	t.Parallel()

	goalID := int64(5)
	category := "savings"
	tx := db.Transaction{
		ID:        100,
		UserID:    1,
		GoalID:    &goalID,
		Title:     "Monthly Savings",
		Amount:    makeNumeric("500.00"),
		Type:      "income",
		Category:  &category,
		Date:      pgtype.Date{Time: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), Valid: true},
		CreatedAt: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
	}

	resp := TransactionToResponse(tx)

	if resp.ID != 100 {
		t.Errorf("ID = %d, want 100", resp.ID)
	}
	if resp.GoalID == nil || *resp.GoalID != 5 {
		t.Errorf("GoalID = %v, want 5", resp.GoalID)
	}
	if resp.Amount != "500.00" {
		t.Errorf("Amount = %q, want %q", resp.Amount, "500.00")
	}
	if resp.Date != "2025-06-15" {
		t.Errorf("Date = %q, want %q", resp.Date, "2025-06-15")
	}
	if resp.Category == nil || *resp.Category != "savings" {
		t.Errorf("Category = %v, want %q", resp.Category, "savings")
	}
}

func TestNumericToString_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    pgtype.Numeric
		want string
	}{
		{
			name: "zero value (invalid)",
			n:    pgtype.Numeric{},
			want: "0.00",
		},
		{
			name: "valid NaN",
			n:    pgtype.Numeric{Valid: true, NaN: true},
			want: "NaN",
		},
		{
			name: "valid zero",
			n:    pgtype.Numeric{Valid: true, Int: big.NewInt(0), Exp: 0},
			want: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := numericToString(tt.n)
			if got != tt.want {
				t.Errorf("numericToString() = %q, want %q", got, tt.want)
			}
		})
	}
}
