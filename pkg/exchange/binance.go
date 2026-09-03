package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
)

func GetBinanceInstruments(instrument string) (dto.BinanceExchangeInfo, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://api.binance.com/api/v3/exchangeInfo?symbol=%s", strings.ToUpper(instrument)))
	if err != nil {
		return dto.BinanceExchangeInfo{}, err
	}
	defer resp.Body.Close()

	var binanceInstruments dto.BinanceExchangeInfo

	err = json.NewDecoder(resp.Body).Decode(&binanceInstruments)
	if err != nil {
		return dto.BinanceExchangeInfo{}, err
	}

	return binanceInstruments, nil
}

func GetBinanceOrderBook(instrument string) (dto.OrderBookData, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://api.binance.com/api/v3/depth?symbol=%s&limit=100", strings.ToUpper(instrument)))
	if err != nil {
		return dto.OrderBookData{}, err
	}
	defer resp.Body.Close()

	var binanceOrderBook dto.BinanceOrderBook

	err = json.NewDecoder(resp.Body).Decode(&binanceOrderBook)
	if err != nil {
		return dto.OrderBookData{}, err
	}

	var result dto.OrderBookData
	var maxVolume float64 = 0
	var maxTotal float64 = 0
	var bestBid float64 = 0
	var bestAsk float64 = 0

	for _, bid := range binanceOrderBook.Bids {
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

	for _, ask := range binanceOrderBook.Asks {
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

func GetBinanceCandles(instrument, after, interval string) ([]dto.CandleDataset, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&endTime=%s", strings.ToUpper(instrument), interval, after))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []dto.BinanceCandleRaw

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	var result []dto.CandleDataset

	for _, j := range data {
		closeTime := strconv.FormatInt(j.CloseTime, 10)

		open, err := strconv.ParseFloat(j.Open, 64)
		if err != nil {
			return nil, err
		}

		high, err := strconv.ParseFloat(j.High, 64)
		if err != nil {
			return nil, err
		}

		low, err := strconv.ParseFloat(j.Low, 64)
		if err != nil {
			return nil, err
		}

		close, err := strconv.ParseFloat(j.Close, 64)
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
