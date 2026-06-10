package sterling

import (
	"log"
	"time"
)

type Forecast struct {
	Date    time.Time `json:"date"`
	Balance float64   `json:"balance"`
}

// ForecastOpsFund provides a simple linear prediction for the operations fund balance.
func (cf *CashFlow) ForecastOpsFund(days int) ([]Forecast, error) {
	current := cf.ledger.GetOperationsFund()

	// Calculate average daily change from recent transactions (last 30 days)
	avgDailyChange := cf.getAvgDailyChange()

	forecasts := make([]Forecast, days)
	for i := 0; i < days; i++ {
		newBalance := current + (avgDailyChange * float64(i+1))
		forecasts[i] = Forecast{
			Date:    time.Now().AddDate(0, 0, i+1),
			Balance: newBalance,
		}
	}
	return forecasts, nil
}

func (cf *CashFlow) getAvgDailyChange() float64 {
	txs := cf.ledger.GetRecentTransactions(24 * 30) // 30 days
	if len(txs) == 0 {
		return 0
	}

	var totalChange float64
	for _, tx := range txs {
		if tx.Type == "revenue" {
			// Only 30% goes to ops fund
			totalChange += tx.Amount * 0.30
		} else if tx.Type == "cost" {
			// Check if it was an operational cost
			if len(tx.Description) >= 13 && tx.Description[:13] == "Auto Bill Pay" {
				totalChange -= tx.Amount
			}
		}
	}
	return totalChange / 30.0
}

func (cf *CashFlow) CheckForecastAndWarn() {
	forecasts, err := cf.ForecastOpsFund(30)
	if err != nil || len(forecasts) == 0 {
		return
	}

	minBalance := forecasts[len(forecasts)-1].Balance
	if minBalance < 0 {
		log.Printf("[Sterling] WARNING: Forecasted ops fund deficit in 30 days: $%.2f", minBalance)
	} else if minBalance < 50 {
		log.Printf("[Sterling] CAUTION: Forecasted ops fund low ($%.2f) in 30 days", minBalance)
	}
}

const IdleThresholdDays = 45

func (cf *CashFlow) CancelIdleSubscriptions(tracker *UsageTracker) {
	idleServices := tracker.GetIdleServices(IdleThresholdDays)
	cf.mu.Lock()
	defer cf.mu.Unlock()

	for _, svc := range idleServices {
		for i, bill := range cf.bills {
			if bill.Vendor == svc && !bill.Paid {
				log.Printf("[Sterling] Auto-cancelling idle subscription: %s", svc)

				if bill.CardID != "" {
					err := cf.agentCardClient.BlockCard(bill.CardID)
					if err != nil {
						log.Printf("[Sterling] Failed to block card for %s: %v", svc, err)
					}
				}

				// Remove recurring bill
				cf.bills = append(cf.bills[:i], cf.bills[i+1:]...)
				cf.save()

				cf.ledger.RecordTransaction("Auto-cancelled idle subscription: "+svc, 0, "info", "Idle for 45+ days")
				break
			}
		}
	}
}
