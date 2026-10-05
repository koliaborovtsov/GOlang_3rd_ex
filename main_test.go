package main

import (
	"strings"
	"testing"
)

func TestAddTransactionEnforcesBudget(t *testing.T) {
	ledger := NewLedger()
	ledger.SetBudget(Budget{Category: "еда", Limit: 5})

	for _, tx := range []Transaction{
		{ID: "1", Category: "еда", Amount: 2},
		{ID: "2", Category: "еда", Amount: 3},
	} {
		if err := ledger.AddTransaction(tx); err != nil {
			t.Fatalf("AddTransaction(%+v) returned error: %v", tx, err)
		}
	}

	if err := ledger.AddTransaction(Transaction{ID: "3", Category: "еда", Amount: 0.01}); err == nil {
		t.Fatal("expected transaction exceeding the budget to be rejected")
	}

	if got := len(ledger.Transactions()); got != 2 {
		t.Fatalf("expected 2 accepted transactions, got %d", got)
	}
}

func TestAddTransactionRejectsInvalidValues(t *testing.T) {
	ledger := NewLedger()

	for _, tx := range []Transaction{
		{ID: "1", Amount: 1},
		{ID: "2", Category: "еда", Amount: -1},
	} {
		if err := ledger.AddTransaction(tx); err == nil {
			t.Errorf("expected invalid transaction %+v to be rejected", tx)
		}
	}

	if got := len(ledger.Transactions()); got != 0 {
		t.Fatalf("invalid transactions were stored: got %d", got)
	}
}

func TestAddTransactionAllowsCategoryWithoutBudget(t *testing.T) {
	ledger := NewLedger()
	tx := Transaction{ID: "1", Category: "развлечения", Amount: 100000}

	if err := ledger.AddTransaction(tx); err != nil {
		t.Fatalf("AddTransaction returned error for category without a budget: %v", err)
	}
	if got := len(ledger.Transactions()); got != 1 {
		t.Fatalf("expected transaction to be stored, got %d transactions", got)
	}
}

func TestLoadBudgets(t *testing.T) {
	ledger := NewLedger()
	input := `[{"category":"еда","limit":10,"period":"месяц"}]`

	if err := ledger.LoadBudgets(strings.NewReader(input)); err != nil {
		t.Fatalf("LoadBudgets returned error: %v", err)
	}
	if err := ledger.AddTransaction(Transaction{Category: "еда", Amount: 10}); err != nil {
		t.Fatalf("loaded budget was not applied: %v", err)
	}
	if err := ledger.AddTransaction(Transaction{Category: "еда", Amount: 0.01}); err == nil {
		t.Fatal("expected loaded budget to reject an over-limit transaction")
	}
}

func TestLoadBudgetsRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "malformed JSON", input: `not-json`},
		{name: "empty category", input: `[{"category":"","limit":10}]`},
		{name: "negative limit", input: `[{"category":"еда","limit":-1}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ledger := NewLedger()
			if err := ledger.LoadBudgets(strings.NewReader(tt.input)); err == nil {
				t.Fatal("expected invalid budget input to be rejected")
			}
		})
	}
}
