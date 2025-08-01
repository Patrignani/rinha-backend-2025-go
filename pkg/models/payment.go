package models

import (
	"time"
)

type Health struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

type PaymentBasic struct {
	CorrelationId string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

type PaymentDb struct {
	CorrelationId string
	Amount        float64
	Fallback      bool
	CreatedAt     time.Time
}

type PaymentSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}
