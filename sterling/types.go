package sterling

type Ledger interface {
	RecordTransaction(description string, amount float64, txType string, notes string) (string, error)
}
