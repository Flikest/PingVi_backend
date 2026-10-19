package alchemy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/registries"
)

const defaultErc20Decimals = 18

const (
	OpGet         string = "get"
	OpSend        string = "send"
	OpSwap        string = "swap"
	OpP2PExchange string = "p2p exchange"
	OpCreate      string = "create"
	OpRestoration string = "restoration"
)

type OperationRecord struct {
	UserID    string    // user_id — владелец кошелька (UUID)
	Type      string    // type — тип операции (operation_type)
	Symbol    string    // symbol — тикер актива
	Value     string    // value — сумма операции (NUMERIC)
	Remainder string    // remainder — текущий баланс актива (NUMERIC)
	Hash      string    // hash — хеш транзакции
	At        time.Time // at — время операции
}

type TokenBalance struct {
	Symbol       string
	Name         string
	Balance      string
	Network      string
	Decimals     int
	PriceUSD     string
	TokenAddress string
}

type portfolioToken struct {
	Address       string  `json:"address"`
	Network       string  `json:"network"`
	TokenAddress  *string `json:"tokenAddress"` // null = нативная монета
	TokenBalance  string  `json:"tokenBalance"` // hex-строка
	TokenMetadata struct {
		Symbol   *string `json:"symbol"`
		Name     *string `json:"name"`
		Decimals *int    `json:"decimals"`
		Logo     *string `json:"logo"`
	} `json:"tokenMetadata"`
	TokenPrices []struct {
		Currency string `json:"currency"`
		Value    string `json:"value"`
	} `json:"tokenPrices"`
}

type portfolioResponse struct {
	Data struct {
		Tokens  []portfolioToken `json:"tokens"`
		PageKey string           `json:"pageKey"`
	} `json:"data"`
}

func sortByTimeDesc(ops []dto.Operation) {
	for i := 0; i < len(ops); i++ {
		for j := i + 1; j < len(ops); j++ {
			if ops[j].Timestamp.After(ops[i].Timestamp) {
				ops[i], ops[j] = ops[j], ops[i]
			}
		}
	}
}

func balanceKey(network, symbol string) string {
	return network + "|" + symbol
}

func lookupBalance(balances map[string]string, network, symbol string) string {
	if balances == nil || symbol == "" {
		return "0"
	}
	if v, ok := balances[balanceKey(network, symbol)]; ok && v != "" {
		return v
	}
	return "0"
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseRawBalance(raw string) (*big.Int, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, false
	}
	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
		base = 16
	}
	if s == "" {
		return nil, false
	}
	val := new(big.Int)
	if _, ok := val.SetString(s, base); !ok {
		return nil, false
	}
	return val, true
}

func humanizeBalance(raw string, decimals int) string {
	val, ok := parseRawBalance(raw)
	if !ok {
		return "0"
	}
	if decimals <= 0 {
		return val.String()
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.DivMod(val, divisor, remainder)

	intPart := quotient.String()
	fracPart := remainder.String()
	for len(fracPart) < decimals {
		fracPart = "0" + fracPart
	}
	// Убираем хвостовые нули, чтобы не выводить ".000000000000000000"
	fracPart = strings.TrimRight(fracPart, "0")
	if fracPart == "" {
		return intPart
	}
	return intPart + "." + fracPart
}

func rpcRequest(url, method string, params []interface{}) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Result json.RawMessage `json:"result"`
		Error  interface{}     `json:"error"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("RPC error: %v", result.Error)
	}
	return result.Result, nil
}

func GetTokenBalances(apiKey string, evmAddress, solAddress string) ([]TokenBalance, int, error) {
	url := fmt.Sprintf("https://api.g.alchemy.com/data/v1/%s/assets/tokens/by-address", apiKey)

	addresses := []map[string]interface{}{
		{
			"address":  evmAddress,
			"networks": []string{"eth-mainnet", "bnb-mainnet"},
		},
		{
			"address":  solAddress,
			"networks": []string{"solana-mainnet"},
		},
	}

	var result []TokenBalance
	seen := make(map[string]bool) // защита от дублей между страницами
	pageKey := ""
	const maxPages = 200 // предохранитель от бесконечного цикла

	// API отдаёт активы ПОСТРАНИЧНО (pageKey), поэтому обходим все страницы,
	// иначе теряется часть токенов (на 1-й странице только 100 позиций).
	for page := 0; page < maxPages; page++ {
		payload := map[string]interface{}{
			"addresses":           addresses,
			"withMetadata":        true,
			"withPrices":          true,
			"includeNativeTokens": true,
			"includeErc20Tokens":  true,
		}
		if pageKey != "" {
			payload["pageKey"] = pageKey
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			return nil, 0, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, http.StatusTooManyRequests, fmt.Errorf("rate limit exceeded")
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var parsed portfolioResponse
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			return nil, 0, fmt.Errorf("parse error: %w, body: %s", err, string(respBody))
		}

		for _, t := range parsed.Data.Tokens {
			isNative := t.TokenAddress == nil

			symbol := derefString(t.TokenMetadata.Symbol)
			name := derefString(t.TokenMetadata.Name)

			// decimals: null (нет метаданных) отличается от реального 0
			decimals := -1
			if t.TokenMetadata.Decimals != nil {
				decimals = *t.TokenMetadata.Decimals
			}

			// Нативные монеты (tokenAddress == null) API не описывает:
			// при любых параметрах tokenMetadata = null. Поэтому символ,
			// имя и decimals берём из реестра сетей (см. networkRegistry).
			if isNative {
				if native, ok := registries.NetWorkRegistry[t.Network]; ok {
					if symbol == "" {
						symbol = native.Symbol
					}
					if name == "" {
						name = native.Name
					}
					if decimals < 0 {
						decimals = native.Decimals
					}
				}
			}
			// ERC-20/SPL без метаданных — типовая точность.
			if decimals < 0 {
				decimals = defaultErc20Decimals
			}
			// Токен без символа — placeholder, чтобы строка не "ломалась".
			if symbol == "" {
				symbol = "?"
			}

			// Пропускаем пустые балансы (ровно 0)
			if raw, ok := parseRawBalance(t.TokenBalance); !ok || raw.Sign() == 0 {
				continue
			}

			// Цена: ищем именно usd в массиве tokenPrices
			priceUSD := ""
			for _, p := range t.TokenPrices {
				if strings.EqualFold(p.Currency, "usd") {
					priceUSD = p.Value
					break
				}
			}
			if priceUSD == "" && len(t.TokenPrices) > 0 {
				priceUSD = t.TokenPrices[0].Value
			}

			tokenAddr := ""
			if t.TokenAddress != nil {
				tokenAddr = *t.TokenAddress
			}

			// Дедупликация по (сеть, кошелёк, адрес токена)
			key := t.Network + "|" + t.Address + "|" + tokenAddr
			if seen[key] {
				continue
			}
			seen[key] = true

			result = append(result, TokenBalance{
				Symbol:       symbol,
				Name:         name,
				Balance:      humanizeBalance(t.TokenBalance, decimals),
				Network:      t.Network,
				Decimals:     decimals,
				PriceUSD:     priceUSD,
				TokenAddress: tokenAddr,
			})
		}

		pageKey = parsed.Data.PageKey
		if pageKey == "" {
			break
		}
	}

	return result, 0, nil
}

func GetEVMOperations(rpcURL, network, address string, balances map[string]string) ([]dto.Operation, error) {
	var all []dto.Operation

	// --- Входящие (get) ---
	pageKey := ""
	for {
		params := map[string]interface{}{
			"toAddress":    address,
			"category":     []string{"external"},
			"withMetadata": true,
			"maxCount":     "0x3e8",
			"order":        "desc",
		}
		if pageKey != "" {
			params["pageKey"] = pageKey
		}

		raw, err := rpcRequest(rpcURL, "alchemy_getAssetTransfers", []interface{}{params})
		if err != nil {
			return nil, err
		}

		var resp struct {
			Transfers []struct {
				Hash     string  `json:"hash"`
				Value    float64 `json:"value"`
				Asset    string  `json:"asset"`
				Metadata struct {
					BlockTimestamp string `json:"blockTimestamp"`
				} `json:"metadata"`
			} `json:"transfers"`
			PageKey string `json:"pageKey"`
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, err
		}

		for _, t := range resp.Transfers {
			ts, _ := time.Parse(time.RFC3339, t.Metadata.BlockTimestamp)
			all = append(all, dto.Operation{
				Type:      OpGet,
				Asset:     t.Asset,
				Value:     fmt.Sprintf("%f", t.Value),
				Remainder: lookupBalance(balances, network, t.Asset),
				TxHash:    t.Hash,
				Timestamp: ts,
			})
		}

		if resp.PageKey == "" {
			break
		}
		pageKey = resp.PageKey
	}

	// --- Исходящие (send) ---
	pageKey = ""
	for {
		params := map[string]interface{}{
			"fromAddress":  address,
			"category":     []string{"external"},
			"withMetadata": true,
			"maxCount":     "0x3e8",
			"order":        "desc",
		}
		if pageKey != "" {
			params["pageKey"] = pageKey
		}

		raw, err := rpcRequest(rpcURL, "alchemy_getAssetTransfers", []interface{}{params})
		if err != nil {
			return nil, err
		}

		var resp struct {
			Transfers []struct {
				Hash     string  `json:"hash"`
				Value    float64 `json:"value"`
				Asset    string  `json:"asset"`
				Metadata struct {
					BlockTimestamp string `json:"blockTimestamp"`
				} `json:"metadata"`
			} `json:"transfers"`
			PageKey string `json:"pageKey"`
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, err
		}

		for _, t := range resp.Transfers {
			ts, _ := time.Parse(time.RFC3339, t.Metadata.BlockTimestamp)
			all = append(all, dto.Operation{
				Type:      OpSend,
				Asset:     t.Asset,
				Value:     fmt.Sprintf("%f", t.Value),
				Remainder: lookupBalance(balances, network, t.Asset),
				TxHash:    t.Hash,
				Timestamp: ts,
			})
		}

		if resp.PageKey == "" {
			break
		}
		pageKey = resp.PageKey
	}

	sortByTimeDesc(all)
	return all, nil
}

func GetSolanaOperations(rpcURL, network, address string, balances map[string]string) ([]dto.Operation, error) {
	var result []dto.Operation
	var before string

	for {
		config := map[string]interface{}{
			"limit": 1000,
		}
		if before != "" {
			config["before"] = before
		}

		raw, err := rpcRequest(rpcURL, "getSignaturesForAddress", []interface{}{address, config})
		if err != nil {
			return nil, err
		}

		var page []struct {
			Signature string `json:"signature"`
			BlockTime int64  `json:"blockTime"`
		}
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, err
		}

		if len(page) == 0 {
			break
		}

		for _, s := range page {
			if s.BlockTime == 0 {
				continue
			}

			txParams := []interface{}{
				s.Signature,
				map[string]interface{}{
					"encoding":                       "jsonParsed",
					"maxSupportedTransactionVersion": 0,
				},
			}
			txRaw, err := rpcRequest(rpcURL, "getTransaction", txParams)
			if err != nil || txRaw == nil {
				continue
			}

			var tx struct {
				Meta struct {
					PreBalances  []uint64    `json:"preBalances"`
					PostBalances []uint64    `json:"postBalances"`
					Err          interface{} `json:"err"`
				} `json:"meta"`
				Transaction struct {
					Message struct {
						AccountKeys []struct {
							Pubkey string `json:"pubkey"`
						} `json:"accountKeys"`
					} `json:"message"`
				} `json:"transaction"`
			}
			if err := json.Unmarshal(txRaw, &tx); err != nil {
				continue
			}
			if tx.Meta.Err != nil {
				continue
			}

			idx := -1
			for i, k := range tx.Transaction.Message.AccountKeys {
				if k.Pubkey == address {
					idx = i
					break
				}
			}
			if idx == -1 || idx >= len(tx.Meta.PreBalances) || idx >= len(tx.Meta.PostBalances) {
				continue
			}

			diff := int64(tx.Meta.PostBalances[idx]) - int64(tx.Meta.PreBalances[idx])
			if diff == 0 {
				continue
			}

			var opType string
			var absDiff int64
			if diff > 0 {
				opType = OpGet
				absDiff = diff
			} else {
				opType = OpSend
				absDiff = -diff
			}

			result = append(result, dto.Operation{
				Type:      opType,
				Asset:     "SOL",
				Value:     fmt.Sprintf("%f", float64(absDiff)/1e9),
				Remainder: lookupBalance(balances, network, "SOL"),
				TxHash:    s.Signature,
				Timestamp: time.Unix(s.BlockTime, 0),
			})
		}

		before = page[len(page)-1].Signature
	}

	return result, nil
}

func SendTransaction(net, apiKey, txHex string) (int, error) {
	if net == "ethereum" {
		url := fmt.Sprintf("https://eth-mainnet.g.alchemy.com/v2/%s", apiKey)

		payload := strings.NewReader(fmt.Sprintf("{\n  \"jsonrpc\": \"2.0\",\n  \"id\": 1,\n  \"method\": \"eth_sendRawTransaction\",\n  \"params\": [\n    \"%s\"\n  ]\n}", txHex))

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}

		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			return 0, err
		}

		return res.StatusCode, nil
	}
	if net == "arbitrum" {
		url := fmt.Sprintf("https://arb-mainnet.g.alchemy.com/v2/%s", apiKey)

		payload := strings.NewReader(fmt.Sprintf("{\n  \"jsonrpc\": \"2.0\",\n  \"id\": 1,\n  \"method\": \"eth_sendRawTransaction\",\n  \"params\": [\n    \"%s\"\n  ]\n}", txHex))

		req, err := http.NewRequest("POST", url, payload)
		if err != nil {
			return 0, err
		}

		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}

		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			return 0, err
		}

		return res.StatusCode, nil
	}
	if net == "solana" {
		url := "https://solana-mainnet.g.alchemy.com/v2/docs-demo"

		payload := strings.NewReader("{\n  \"jsonrpc\": \"2.0\",\n  \"id\": 1,\n  \"method\": \"sendTransaction\",\n  \"params\": [\n    \"string\",\n    {\n      \"encoding\": \"base58\",\n      \"skipPreflight\": false,\n      \"preflightCommitment\": \"finalized\",\n      \"maxRetries\": 1,\n      \"minContextSlot\": 1\n    }\n  ]\n}")

		req, err := http.NewRequest("POST", url, payload)
		if err != nil {
			return 0, err
		}

		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}

		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			return 0, err
		}

		return res.StatusCode, nil
	} else {
		return 0, errors.New("unknown network")
	}
}
