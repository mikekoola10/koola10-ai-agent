package financial

import (
	"encoding/json"
	"os"
	"sync"
	"time"
	"koola10/tools"
)

type Holding struct {
	Symbol       string    `json:"symbol"`
	Shares       float64   `json:"shares"`
	AverageCost  float64   `json:"average_cost"`
	CurrentPrice float64   `json:"current_price"`
	PnL          float64   `json:"pnl"`
	LastUpdated  time.Time `json:"last_updated"`
}

type PortfolioState struct {
	Holdings       map[string]Holding `json:"holdings"`
	CashBalance    float64            `json:"cash_balance"`
	TotalValue     float64            `json:"total_value"`
	BenchmarkValue float64            `json:"benchmark_value"` // e.g. S&P 500 comparison
	LastUpdated    time.Time          `json:"last_updated"`
}

type PortfolioManager struct {
	State       PortfolioState
	storagePath string
	mu          sync.RWMutex
}

func NewPortfolioManager(path string) *PortfolioManager {
	pm := &PortfolioManager{
		storagePath: path,
		State: PortfolioState{
			Holdings:    make(map[string]Holding),
			CashBalance: 100000.0, // Initial paper trading balance
			LastUpdated: time.Now(),
		},
	}
	pm.load()
	return pm
}

func (pm *PortfolioManager) load() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	data, err := os.ReadFile(pm.storagePath)
	if err == nil {
		json.Unmarshal(data, &pm.State)
	}
	if pm.State.Holdings == nil {
		pm.State.Holdings = make(map[string]Holding)
	}
}

func (pm *PortfolioManager) save() {
	data, _ := json.MarshalIndent(pm.State, "", "  ")
	os.WriteFile(pm.storagePath, data, 0644)
}

func (pm *PortfolioManager) AddHolding(symbol string, shares float64, cost float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	h, ok := pm.State.Holdings[symbol]
	if ok {
		totalCost := (h.Shares * h.AverageCost) + (shares * cost)
		h.Shares += shares
		h.AverageCost = totalCost / h.Shares
	} else {
		h = Holding{
			Symbol:      symbol,
			Shares:      shares,
			AverageCost: cost,
		}
	}
	h.LastUpdated = time.Now()
	pm.State.Holdings[symbol] = h
	pm.State.CashBalance -= shares * cost
	pm.save()
}

func (pm *PortfolioManager) UpdatePrices() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	totalValue := pm.State.CashBalance
	for symbol, h := range pm.State.Holdings {
		// Call marketDataTool to get current price
		res := tools.RunTool("market_data", map[string]interface{}{
			"action": "get_stock_price",
			"symbol": symbol,
		})
		if !res.Success {
			// Try crypto if stock fails
			res = tools.RunTool("market_data", map[string]interface{}{
				"action": "get_crypto_price",
				"symbol": symbol,
			})
		}

		if res.Success {
			data := res.Data.(map[string]interface{})
			price, _ := data["price"].(float64)
			h.CurrentPrice = price
			h.PnL = (h.CurrentPrice - h.AverageCost) * h.Shares
			h.LastUpdated = time.Now()
			pm.State.Holdings[symbol] = h
		}
		totalValue += h.Shares * h.CurrentPrice
	}
	pm.State.TotalValue = totalValue
	pm.State.LastUpdated = time.Now()
	pm.save()
}

func (pm *PortfolioManager) GetPortfolioSummary() PortfolioState {
	pm.UpdatePrices()
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.State
}
