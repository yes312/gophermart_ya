package models

import "time"

type User struct {
	Login string
	Hash  string
}

type OrderStatusNew struct {
	Number     string    `json:"order"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type OrderStatus struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current  float64 `json:"current"`
	Withdraw float64 `json:"withdrawn"`
}

type OrderSum struct {
	OrderNumber string  `json:"order"`
	Sum         float64 `json:"sum"`
}

type OrderUserID struct {
	OrderNumber string `json:"order"`
	UserID      string `json:"user_id"`
}

type Withdrawal struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type Billing struct {
	OrderNumber string    `json:"order"`
	Status      string    `json:"status"`
	Accrual     float64   `json:"accrual"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Time        time.Time `json:"time"`
}
