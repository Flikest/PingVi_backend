package exchange

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
)

func GetByBitInstruments(instrument string) (dto.ByBitExchangeInfo, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://api.bybit.com/v5/market/instruments-info?category=spot&symbol=%s", instrument))
	if err != nil {
		slog.Error("error while fetching bybit instruments", "error", err)
		return dto.ByBitExchangeInfo{}, err
	}
	defer resp.Body.Close()

	var byBitInstruments dto.ByBitExchangeInfo

	err = json.NewDecoder(resp.Body).Decode(&byBitInstruments)
	if err != nil {
		slog.Error("error while decoding bybit instruments", "error", err)
		return dto.ByBitExchangeInfo{}, err
	}

	return byBitInstruments, nil
}

func GetByBitOrderBook(instrument string) (dto.OrderBookData, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://api.bybit.com/v5/market/orderbook?category=spot&symbol=%s", strings.ToUpper(instrument)))
	if err != nil {
		return dto.OrderBookData{}, err
	}
	defer resp.Body.Close()

	var response dto.BybitOrderBookResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return dto.OrderBookData{}, err
	}

	if response.RetCode != 0 {
		return dto.OrderBookData{}, fmt.Errorf("bybit error: %s", response.RetMsg)
	}

	var result dto.OrderBookData

	var maxVolume float64 = 0
	var maxTotal float64 = 0
	var bestBid float64 = 0
	var bestAsk float64 = 0

	for _, bid := range response.Result.Bids {
		price, err := strconv.ParseFloat(bid[0], 64)
		if err != nil {
			return dto.OrderBookData{}, err
		}

		quantity, err := strconv.ParseFloat(bid[1], 64)
		if err != nil {
			return dto.OrderBookData{}, err
		}

		total := price * quantity

		if bestBid == 0 || price > bestBid {
			bestBid = price
		}

		if quantity > maxVolume {
			maxVolume = quantity
		}
		if total > maxTotal {
			maxTotal = total
		}

		result.Bids = append(result.Bids, dto.OrderBookEntry{
			Price:    price,
			Quantity: quantity,
			Total:    total,
		})
	}

	for _, ask := range response.Result.Asks {
		price, err := strconv.ParseFloat(ask[0], 64)
		if err != nil {
			return dto.OrderBookData{}, err
		}

		quantity, err := strconv.ParseFloat(ask[1], 64)
		if err != nil {
			return dto.OrderBookData{}, err
		}

		total := price * quantity

		if bestAsk == 0 || price < bestAsk {
			bestAsk = price
		}

		if quantity > maxVolume {
			maxVolume = quantity
		}
		if total > maxTotal {
			maxTotal = total
		}

		result.Asks = append(result.Asks, dto.OrderBookEntry{
			Price:    price,
			Quantity: quantity,
			Total:    total,
		})
	}

	if bestBid > 0 && bestAsk > 0 {
		result.Spread = bestAsk - bestBid
		result.MidPrice = (bestBid + bestAsk) / 2
	}

	result.MaxVolume = maxVolume
	result.MaxTotal = maxTotal

	return result, nil
}

func GeByBitCandles(instrument, after, interval string) ([]dto.CandleDataset, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	if interval == "" {
		interval = "1"
	}

	resp, err := client.Get(fmt.Sprintf("https://api.bybit.com/v5/market/index-price-kline?symbol=%s&interval=%s&end=%s", strings.ToUpper(instrument), interval, after))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []dto.BybitResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	var result []dto.CandleDataset

	for i, j := range data {
		closeTime := strconv.FormatInt(j.Result.List[i].Timestamp, 10)

		open, err := strconv.ParseFloat(j.Result.List[i].Open, 64)
		if err != nil {
			return nil, err
		}

		high, err := strconv.ParseFloat(j.Result.List[i].High, 64)
		if err != nil {
			return nil, err
		}

		low, err := strconv.ParseFloat(j.Result.List[i].Low, 64)
		if err != nil {
			return nil, err
		}

		close, err := strconv.ParseFloat(j.Result.List[i].Close, 64)
		if err != nil {
			return nil, err
		}

		result = append(result, dto.CandleDataset{
			Time:  closeTime,
			Open:  open,
			High:  high,
			Low:   low,
			Close: close,
		})
	}

	return result, nil
}
