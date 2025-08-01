//go:generate easyjson -all summary.go
package models

type SummaryResponse struct {
	Default  PaymentSummary `json:"default"`
	Fallback PaymentSummary `json:"fallback"`
}
