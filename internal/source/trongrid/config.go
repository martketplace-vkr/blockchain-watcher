package trongrid

import "time"

type Config struct {
	BaseURL         string `validate:"required"`
	APIKey          string
	ContractAddress string        `validate:"required"`
	Timeout         time.Duration `validate:"required"`
	PageLimit       int           `validate:"required"`
}
