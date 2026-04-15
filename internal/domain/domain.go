package domain

import "time"

type WatchAddress struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Address   string    `db:"address"`
	Network   string    `db:"network"`
	Asset     string    `db:"asset"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type DepositStatus string

const (
	DepositStatusDetected  DepositStatus = "detected"
	DepositStatusConfirmed DepositStatus = "confirmed"
)

type Deposit struct {
	ID            int64         `db:"id"`
	UserID        int64         `db:"user_id"`
	Address       string        `db:"address"`
	Network       string        `db:"network"`
	Asset         string        `db:"asset"`
	TxHash        string        `db:"tx_hash"`
	LogIndex      int64         `db:"log_index"`
	Amount        string        `db:"amount"`
	BlockNumber   int64         `db:"block_number"`
	Confirmations int64         `db:"confirmations"`
	Status        DepositStatus `db:"status"`
	ConfirmedAt   *time.Time    `db:"confirmed_at"`
	OutboxSentAt  *time.Time    `db:"outbox_sent_at"`
	CreatedAt     time.Time     `db:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at"`
}

type Cursor struct {
	Network   string    `db:"network"`
	LastBlock int64     `db:"last_block"`
	LastTxAt  time.Time `db:"last_tx_at"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RegisterAddressEvent struct {
	UserID   int64  `json:"user_id"`
	Address  string `json:"address"`
	Network  string `json:"network"`
	Asset    string `json:"asset"`
	Source   string `json:"source,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type DepositConfirmedEvent struct {
	UserID        int64  `json:"user_id"`
	Address       string `json:"address"`
	Network       string `json:"network"`
	Asset         string `json:"asset"`
	TxHash        string `json:"tx_hash"`
	LogIndex      int64  `json:"log_index"`
	Amount        string `json:"amount"`
	BlockNumber   int64  `json:"block_number"`
	Confirmations int64  `json:"confirmations"`
}
