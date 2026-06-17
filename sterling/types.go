package sterling

type Ledger interface {
	RecordTransaction(description string, amount float64, txType string, notes string) (string, error)
}

type VaultEntry struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Notes       string  `json:"notes"`
}
