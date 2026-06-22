package main

import (
	"context"
	"fmt"
	"time"

	bybit "github.com/bybit-exchange/bybit.go.api"
)

type OHLC struct {
	Time     time.Time `json:"time"`
	Open     float64   `json:"open"`
	High     float64   `json:"high"`
	Low      float64   `json:"low"`
	Close    float64   `json:"close"`
	Volume   float64   `json:"volume"`
	Turnover float64   `json:"Turnover"`
}

func main() {

	client := bybit.NewBybitHttpClient("", "", bybit.WithBaseURL(bybit.MAINNET))

	params := map[string]interface{}{
		"category": "spot",
		"symbol":   "BTCUSDT",
		"interval": "M",
	}

	resp, err := client.NewUtaBybitServiceWithParams(params).GetMarketKline(context.Background())
	if err != nil {
		fmt.Printf("Ошибка при запросе: %v\n", err)
		return
	}

	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		fmt.Println("Ошибка: Result не является мапой")
		return
	}

	candles, _ := resultMap["list"].([]interface{})

	result := []OHLC{}

	for i := range candles {
		result = append(result, OHLC{
			Time:     candles[i][0],
			Open:     i[1],
			High:     i[2],
			Low:      i[3],
			Close:    i[4],
			Volume:   i[5],
			Turnover: i[6],
		})
	}
}
