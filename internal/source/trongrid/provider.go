package trongrid

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	mocksource "github.com/martketplace-vkr/blockchain-watcher/internal/source/mock"
)

const tronAPIKeyHeader = "TRON-PRO-API-KEY"

type Provider struct {
	cfg    Config
	client *http.Client
}

type latestBlockResponse struct {
	BlockHeader struct {
		RawData struct {
			Number int64 `json:"number"`
		} `json:"raw_data"`
	} `json:"block_header"`
}

type contractEventsResponse struct {
	Data []struct {
		BlockNumber   int64             `json:"block_number"`
		TransactionID string            `json:"transaction_id"`
		EventIndex    int64             `json:"event_index"`
		Result        map[string]string `json:"result"`
	} `json:"data"`
	Meta struct {
		Fingerprint string `json:"fingerprint"`
	} `json:"meta"`
}

func New(cfg Config) *Provider {
	return &Provider{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (p *Provider) LatestBlock(ctx context.Context) (int64, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(p.cfg.BaseURL, "/")+"/wallet/getnowblock",
		http.NoBody,
	)
	if err != nil {
		return 0, err
	}
	p.withHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("trongrid latest block request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload latestBlockResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}

	return payload.BlockHeader.RawData.Number, nil
}

func (p *Provider) Transfers(ctx context.Context, fromBlock, toBlock int64) ([]mocksource.ChainTransfer, error) {
	result := make([]mocksource.ChainTransfer, 0)

	for blockNumber := fromBlock; blockNumber <= toBlock; blockNumber++ {
		transfers, err := p.blockTransfers(ctx, blockNumber)
		if err != nil {
			return nil, err
		}

		result = append(result, transfers...)
	}

	return result, nil
}

func (p *Provider) blockTransfers(ctx context.Context, blockNumber int64) ([]mocksource.ChainTransfer, error) {
	result := make([]mocksource.ChainTransfer, 0)
	fingerprint := ""

	for {
		reqURL, err := url.Parse(strings.TrimRight(p.cfg.BaseURL, "/") + "/v1/contracts/" + p.cfg.ContractAddress + "/events")
		if err != nil {
			return nil, err
		}

		query := reqURL.Query()
		query.Set("event_name", "Transfer")
		query.Set("only_confirmed", "true")
		query.Set("block_number", strconv.FormatInt(blockNumber, 10))
		query.Set("limit", strconv.Itoa(p.cfg.PageLimit))
		if fingerprint != "" {
			query.Set("fingerprint", fingerprint)
		}
		reqURL.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
		if err != nil {
			return nil, err
		}
		p.withHeaders(req)

		resp, err := p.client.Do(req)
		if err != nil {
			return nil, err
		}

		var payload contractEventsResponse
		err = decodeResponse(resp, &payload)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		for _, event := range payload.Data {
			transfer := mocksource.ChainTransfer{
				TxHash:      event.TransactionID,
				LogIndex:    event.EventIndex,
				BlockNumber: event.BlockNumber,
				FromAddress: event.Result["from"],
				ToAddress:   event.Result["to"],
				Amount:      event.Result["value"],
			}

			if transfer.TxHash == "" || transfer.ToAddress == "" || transfer.Amount == "" {
				continue
			}

			result = append(result, transfer)
		}

		if payload.Meta.Fingerprint == "" {
			return result, nil
		}

		fingerprint = payload.Meta.Fingerprint
	}
}

func (p *Provider) withHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if p.cfg.APIKey != "" {
		req.Header.Set(tronAPIKeyHeader, p.cfg.APIKey)
	}
}

func decodeResponse(resp *http.Response, target any) error {
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("trongrid request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
