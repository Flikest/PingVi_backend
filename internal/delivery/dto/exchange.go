package dto

type CandleDataset struct {
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume float64 `json:"volume"`
}
type SearchInstrumentResponse struct {
	ExchangeName   string        `json:"exchange_name"`
	Symbol         string        `json:"symbol"`
	LogoURL        string        `json:"logo_url"`
	InstType       string        `json:"inst_type"`
	OHLCLastCandle CandleDataset `json:"last_close_price"`
}

type OrderBookEntry struct {
	Price    float64
	Quantity float64
	Total    float64
}

type OrderBookData struct {
	Bids      []OrderBookEntry `json:"bids"`
	Asks      []OrderBookEntry `json:"asks"`
	MaxVolume float64          `json:"max_volume"`
	MaxTotal  float64          `json:"max_total"`
	Spread    float64          `json:"spread"`
	MidPrice  float64          `json:"mid_price"`
}
