package binance

import (
	"io"
	"net/http"
)

func (b *BinanceConfig) GetCandleHistory() {
	response, err := http.Get()
	if err != nil {
		b.Log.Error("request error: ", "error", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
}
