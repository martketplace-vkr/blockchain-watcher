package pg

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/martketplace-vkr/blockchain-watcher/internal/domain"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertWatchAddress(ctx context.Context, addr domain.WatchAddress) error {
	query := `
		insert into blockchain_watcher.watch_address (
			user_id,
			address,
			network,
			asset
		) values (
			:user_id,
			:address,
			:network,
			:asset
		)
		on conflict (network, address) do update
		set
			user_id = excluded.user_id,
			asset = excluded.asset,
			updated_at = now()
	`

	_, err := r.db.NamedExecContext(ctx, query, addr)
	return err
}

func (r *Repository) GetWatchAddress(ctx context.Context, network, address string) (*domain.WatchAddress, error) {
	query := `
		select
			id,
			user_id,
			address,
			network,
			asset,
			created_at,
			updated_at
		from blockchain_watcher.watch_address
		where network = $1
			and address = $2
	`

	var result domain.WatchAddress
	if err := r.db.GetContext(ctx, &result, query, network, address); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Repository) GetCursor(ctx context.Context, network string) (*domain.Cursor, error) {
	query := `
		select
			network,
			last_block,
			last_tx_at,
			created_at,
			updated_at
		from blockchain_watcher.cursor
		where network = $1
	`

	var cursor domain.Cursor
	if err := r.db.GetContext(ctx, &cursor, query, network); err != nil {
		return nil, err
	}

	return &cursor, nil
}

func (r *Repository) UpsertCursor(ctx context.Context, network string, lastBlock int64, lastTxAt time.Time) error {
	query := `
		insert into blockchain_watcher.cursor (
			network,
			last_block,
			last_tx_at
		) values ($1, $2, $3)
		on conflict (network) do update
		set
			last_block = excluded.last_block,
			last_tx_at = excluded.last_tx_at,
			updated_at = now()
	`

	_, err := r.db.ExecContext(ctx, query, network, lastBlock, lastTxAt)
	return err
}

func (r *Repository) GetDeposit(ctx context.Context, network, txHash string, logIndex int64) (*domain.Deposit, error) {
	query := `
		select
			id,
			user_id,
			address,
			network,
			asset,
			tx_hash,
			log_index,
			amount,
			block_number,
			confirmations,
			status,
			confirmed_at,
			outbox_sent_at,
			created_at,
			updated_at
		from blockchain_watcher.deposit
		where network = $1
			and tx_hash = $2
			and log_index = $3
	`

	var deposit domain.Deposit
	if err := r.db.GetContext(ctx, &deposit, query, network, txHash, logIndex); err != nil {
		return nil, err
	}

	return &deposit, nil
}

func (r *Repository) CreateDeposit(ctx context.Context, deposit domain.Deposit) error {
	query := `
		insert into blockchain_watcher.deposit (
			user_id,
			address,
			network,
			asset,
			tx_hash,
			log_index,
			amount,
			block_number,
			confirmations,
			status,
			confirmed_at,
			outbox_sent_at
		) values (
			:user_id,
			:address,
			:network,
			:asset,
			:tx_hash,
			:log_index,
			:amount,
			:block_number,
			:confirmations,
			:status,
			:confirmed_at,
			:outbox_sent_at
		)
		on conflict (network, tx_hash, log_index) do nothing
	`

	_, err := r.db.NamedExecContext(ctx, query, deposit)
	return err
}

func (r *Repository) UpdateDepositProgress(
	ctx context.Context,
	network, txHash string,
	logIndex int64,
	confirmations int64,
	status domain.DepositStatus,
	confirmedAt *time.Time,
) error {
	query := `
		update blockchain_watcher.deposit
		set
			confirmations = $4,
			status = $5,
			confirmed_at = coalesce($6, confirmed_at),
			updated_at = now()
		where network = $1
			and tx_hash = $2
			and log_index = $3
	`

	_, err := r.db.ExecContext(ctx, query, network, txHash, logIndex, confirmations, status, confirmedAt)
	return err
}

func (r *Repository) MarkDepositOutboxSent(ctx context.Context, network, txHash string, logIndex int64, sentAt time.Time) error {
	query := `
		update blockchain_watcher.deposit
		set
			outbox_sent_at = $4,
			updated_at = now()
		where network = $1
			and tx_hash = $2
			and log_index = $3
	`

	_, err := r.db.ExecContext(ctx, query, network, txHash, logIndex, sentAt)
	return err
}

func IsNotFound(err error) bool {
	return err == sql.ErrNoRows
}
