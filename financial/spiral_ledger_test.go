package financial

import (
	"os"
	"testing"
)

func TestSpiralLedger(t *testing.T) {
	path := "test_spiral_ledger.json"
	defer os.Remove(path)

	sl := NewSpiralLedger(path)
	if sl.Balance != 0 {
		t.Errorf("Expected balance 0, got %f", sl.Balance)
	}

	sl.RecordRevenue(150, "test")
	if sl.Balance != 150 {
		t.Errorf("Expected balance 150, got %f", sl.Balance)
	}

	if sl.TotalRevenue != 150 {
		t.Errorf("Expected total revenue 150, got %f", sl.TotalRevenue)
	}

	status := sl.GetStatus()
	if status["balance"] != 150.0 {
		t.Errorf("Expected status balance 150, got %v", status["balance"])
	}
}
