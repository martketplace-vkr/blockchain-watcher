package mock

import (
	"context"
	"sort"
)

type ChainTransfer struct {
	TxHash      string
	LogIndex    int64
	BlockNumber int64
	FromAddress string
	ToAddress   string
	Amount      string
	Asset       string
	Network     string
}

type Provider struct {
	cfg Config
}

func New(cfg Config) *Provider {
	return &Provider{cfg: cfg}
}

func (p *Provider) LatestBlock(_ context.Context) (int64, error) {
	return p.cfg.LatestBlock, nil
}

func (p *Provider) Transfers(_ context.Context, fromBlock, toBlock int64) ([]ChainTransfer, error) {
	result := make([]ChainTransfer, 0)

	for _, transfer := range p.cfg.Transfers {
		if transfer.BlockNumber < fromBlock || transfer.BlockNumber > toBlock {
			continue
		}

		result = append(result, ChainTransfer{
			TxHash:      transfer.TxHash,
			LogIndex:    transfer.LogIndex,
			BlockNumber: transfer.BlockNumber,
			FromAddress: transfer.FromAddress,
			ToAddress:   transfer.ToAddress,
			Amount:      transfer.Amount,
			Asset:       transfer.Asset,
			Network:     transfer.Network,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].BlockNumber == result[j].BlockNumber {
			return result[i].LogIndex < result[j].LogIndex
		}

		return result[i].BlockNumber < result[j].BlockNumber
	})

	return result, nil
}
