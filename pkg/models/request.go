//go:generate easyjson -all request.go

package models

import (
	"time"
)

type PaymentRequest struct {
	CorrelationId string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}
