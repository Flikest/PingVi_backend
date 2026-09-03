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

func GetOKXInstruments(instrument string) (dto.OKXInstrument, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://www.okx.com/api/v5/public/instruments?instType=SPOT&instId=%s", instrument))
	if err != nil {
		return dto.OKXInstrument{}, err
	}
	defer resp.Body.Close()

	var okxInstrument dto.OKXInstrument

	err = json.NewDecoder(resp.Body).Decode(&okxInstrument)
	if err != nil {
		return dto.OKXInstrument{}, err
	}

	return okxInstrument, nil
}

func GetOKXOrderbook(instrument string) (dto.OrderBookData, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://www.okx.com/api/v5/market/books?instId=%s&sz=100", strings.ToUpper(instrument)))
	if err != nil {
		return dto.OrderBookData{}, err
	}
	defer resp.Body.Close()

	var okxResponse dto.OKXOrderBookResponse

	err = json.NewDecoder(resp.Body).Decode(&okxResponse)
	if err != nil {
		return dto.OrderBookData{}, err
	}

	if okxResponse.Code != "0" {
		return dto.OrderBookData{}, fmt.Errorf("okx error: code=%s, msg=%s", okxResponse.Code, okxResponse.Msg)
	}

	if len(okxResponse.Data) == 0 {
		return dto.OrderBookData{}, fmt.Errorf("no orderbook data received")
	}

	orderBook := okxResponse.Data[0]

	var result dto.OrderBookData
	var maxVolume float64 = 0
	var maxTotal float64 = 0
	var bestBid float64 = 0
	var bestAsk float64 = 0

	for _, bid := range orderBook.Bids {
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

	for _, ask := range orderBook.Asks {
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

func GetOKXCandles(instrument, interval, after string) ([]dto.CandleDataset, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("https://www.okx.com/api/v5/market/index-candles?instId=%s&after=%s&bar=%s", strings.ToUpper(instrument), after, interval))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data dto.OKXKlineResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	var result []dto.CandleDataset

	for index, candle := range data.Data {
		closeTime := strconv.FormatInt(candle[index].Timestamp, 10)

		open, err := strconv.ParseFloat(candle[index].Open, 64)
		if err != nil {
			return nil, err
		}

		high, err := strconv.ParseFloat(candle[index].High, 64)
		if err != nil {
			return nil, err
		}

		low, err := strconv.ParseFloat(candle[index].Low, 64)
		if err != nil {
			return nil, err
		}

		close, err := strconv.ParseFloat(candle[index].Close, 64)
		if err != nil {
			return nil, err
		}

		volume, err := strconv.ParseFloat(candle[index].Volume, 64)
		if err != nil {
			return nil, err
		}

		result = append(result, dto.CandleDataset{
			Time:   closeTime,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return result, nil
}
