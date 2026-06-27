package binance

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (b *BinanceConfig) GetCandleHistory(symbol, interval, startTime, endTime, timeZone, limit string) {
	params := url.Values{}
	params.Add("symbol", symbol)
	params.Add("interval", interval)
	params.Add("startTime", startTime)
	params.Add("endTime", endTime)
	params.Add("timeZone", timeZone)
	params.Add("limit", limit)

	response, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/klines/?%s", b.Net, params.Encode()))
	if err != nil {
		b.Log.Error("request error: ", "error", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		b.Log.Error("error read response: ", "error", err)
		return
	}

}
