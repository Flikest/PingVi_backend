package mempool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type MemPoolPrices struct {
	Time int `json:"time"`
	USD  int `json:"USD"`
	EUR  int `json:"EUR"`
	GBP  int `json:"GBP"`
	CAD  int `json:"CAD"`
	CHF  int `json:"CHF"`
	AUD  int `json:"AUD"`
	JPY  int `json:"JPY"`
}

type Stats struct {
	FundedTxoCount int   `json:"funded_txo_count"`
	FundedTxoSum   int64 `json:"funded_txo_sum"` // Сумма всех приходов в сатоши
	SpentTxoCount  int   `json:"spent_txo_count"`
	SpentTxoSum    int64 `json:"spent_txo_sum"` // Сумма всех расходов в сатоши
}

type AddressResponse struct {
	Address      string `json:"address"`
	ChainStats   Stats  `json:"chain_stats"`   // Подтвержденные транзакции
	MempoolStats Stats  `json:"mempool_stats"` // Неподтвержденные транзакции
}

func GetBtcBalance(address string) (float64, error) {
	url := fmt.Sprintf("https://mempool.space/api/address/%s", address)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var addrInfo AddressResponse
	if err := json.Unmarshal(body, &addrInfo); err != nil {
		return 0, err
	}

	totalFunded := addrInfo.ChainStats.FundedTxoSum + addrInfo.MempoolStats.FundedTxoSum
	totalSpent := addrInfo.ChainStats.SpentTxoSum + addrInfo.MempoolStats.SpentTxoSum
	balanceSat := totalFunded - totalSpent

	balanceBTC := float64(balanceSat) / 100000000.0

	return balanceBTC, nil
}

func GetBtcPrices() (int, error) {
	resp, err := http.Get("https://mempool.space/api/v1/prices")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var prices MemPoolPrices
	err = json.NewDecoder(resp.Body).Decode(&prices)
	if err != nil {
		return 0, err
	}

	return prices.USD, nil
}

func SendTransaction(txHex string) (int, error) {
	url := "https://mempool.space/api/tx"

	req, err := http.NewRequest("POST", url, strings.NewReader(txHex))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}
