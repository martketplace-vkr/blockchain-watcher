package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/martketplace-vkr/blockchain-watcher/internal/app/cmp/processor"
	"github.com/martketplace-vkr/blockchain-watcher/internal/domain"
	"github.com/martketplace-vkr/blockchain-watcher/internal/repository/pg"
	mocksource "github.com/martketplace-vkr/blockchain-watcher/internal/source/mock"
	"github.com/martketplace-vkr/pkg/inbox/dto"
)

const eventDepositConfirmed = "crypto_deposit_confirmed"

type provider interface {
	LatestBlock(ctx context.Context) (int64, error)
	Transfers(ctx context.Context, fromBlock, toBlock int64) ([]mocksource.ChainTransfer, error)
}

type outbox interface {
	Send(ctx context.Context, eventType string, payload any) error
}

type Service struct {
	repository *pg.Repository
	provider   provider
	outbox     outbox
	cfg        processor.Config
}

func New(repository *pg.Repository, provider provider, outbox outbox, cfg processor.Config) *Service {
	return &Service{
		repository: repository,
		provider:   provider,
		outbox:     outbox,
		cfg:        cfg,
	}
}

func (s *Service) HandleAddressRegister(ctx context.Context, event dto.Event) error {
	var payload domain.RegisterAddressEvent
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	if payload.Network == "" {
		payload.Network = s.cfg.Network
	}
	if payload.Asset == "" {
		payload.Asset = s.cfg.Asset
	}

	return s.repository.UpsertWatchAddress(ctx, domain.WatchAddress{
		UserID:  payload.UserID,
		Address: payload.Address,
		Network: payload.Network,
		Asset:   payload.Asset,
	})
}

func (s *Service) Process(ctx context.Context) error {
	latestBlock, err := s.provider.LatestBlock(ctx)
	if err != nil {
		return err
	}

	fromBlock := s.cfg.StartBlock
	cursor, err := s.repository.GetCursor(ctx, s.cfg.Network)
	if err == nil {
		fromBlock = cursor.LastBlock + 1
	} else if !pg.IsNotFound(err) {
		return fmt.Errorf("failed get cursor: %s", err)
	}

	if latestBlock < fromBlock {
		return nil
	}

	toBlock := fromBlock + s.cfg.BlockBatchSize - 1
	if toBlock > latestBlock {
		toBlock = latestBlock
	}

	transfers, err := s.provider.Transfers(ctx, fromBlock, toBlock)
	if err != nil {
		return fmt.Errorf("failed to fetch transfers: %s", err)
	}

	for _, transfer := range transfers {
		if transfer.Network != s.cfg.Network || transfer.Asset != s.cfg.Asset {
			continue
		}

		if err := s.processTransfer(ctx, transfer, latestBlock); err != nil {
			return fmt.Errorf("failed process transfer: %s", err)
		}
	}

	err = s.repository.UpsertCursor(ctx, s.cfg.Network, toBlock, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update cursor: %s", err)
	}

	return nil
}

func (s *Service) processTransfer(ctx context.Context, transfer mocksource.ChainTransfer, latestBlock int64) error {
	watched, err := s.repository.GetWatchAddress(ctx, transfer.Network, transfer.ToAddress)
	if err != nil {
		if pg.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed get address: %s", err)
	}

	confirmations := latestBlock - transfer.BlockNumber + 1
	if confirmations < 0 {
		confirmations = 0
	}

	normalizedAmount, err := normalizeTokenAmount(transfer.Amount, s.cfg.TokenDecimals)
	if err != nil {
		return err
	}

	existing, err := s.repository.GetDeposit(ctx, transfer.Network, transfer.TxHash, transfer.LogIndex)
	if err != nil && !pg.IsNotFound(err) {
		return fmt.Errorf("failed check deposit: %s", err)
	}

	now := time.Now().UTC()
	status := domain.DepositStatusDetected
	var confirmedAt *time.Time
	if confirmations >= s.cfg.ConfirmationsRequired {
		status = domain.DepositStatusConfirmed
		confirmedAt = &now
	}

	if existing == nil {
		if err := s.repository.CreateDeposit(ctx, domain.Deposit{
			UserID:        watched.UserID,
			Address:       transfer.ToAddress,
			Network:       transfer.Network,
			Asset:         transfer.Asset,
			TxHash:        transfer.TxHash,
			LogIndex:      transfer.LogIndex,
			Amount:        normalizedAmount,
			BlockNumber:   transfer.BlockNumber,
			Confirmations: confirmations,
			Status:        status,
			ConfirmedAt:   confirmedAt,
		}); err != nil {
			return err
		}
	} else {
		if err := s.repository.UpdateDepositProgress(
			ctx,
			transfer.Network,
			transfer.TxHash,
			transfer.LogIndex,
			confirmations,
			status,
			confirmedAt,
		); err != nil {
			return err
		}
	}

	if status != domain.DepositStatusConfirmed {
		return nil
	}
	if existing != nil && existing.OutboxSentAt != nil {
		return nil
	}

	event := domain.DepositConfirmedEvent{
		UserID:        watched.UserID,
		Address:       transfer.ToAddress,
		Network:       transfer.Network,
		Asset:         transfer.Asset,
		TxHash:        transfer.TxHash,
		LogIndex:      transfer.LogIndex,
		Amount:        normalizedAmount,
		BlockNumber:   transfer.BlockNumber,
		Confirmations: confirmations,
	}

	if err := s.outbox.Send(ctx, eventDepositConfirmed, event); err != nil {
		return err
	}

	return s.repository.MarkDepositOutboxSent(ctx, transfer.Network, transfer.TxHash, transfer.LogIndex, now)
}

func normalizeTokenAmount(raw string, decimals int32) (string, error) {
	value, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", fmt.Errorf("invalid token amount: %s", raw)
	}

	if decimals < 0 {
		return "", fmt.Errorf("invalid token decimals: %d", decimals)
	}

	denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	rational := new(big.Rat).SetFrac(value, denominator)

	return rational.FloatString(int(decimals)), nil
}
