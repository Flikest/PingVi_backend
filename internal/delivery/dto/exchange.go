package dto

type CandleDataset struct {
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume float64 `json:"volume"`
}

type CandleStreamDataset struct {
	Time    string `json:"time"`
	Open    string `json:"open"`
	High    string `json:"high"`
	Low     string `json:"low"`
	Close   string `json:"close"`
	Volume  string `json:"volume"`
	Confirm bool   `json:"confirm"`
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

type WithdrawByBitRequest struct {
	Coin        string `json:"coin"`
	Chain       string `json:"chain"`
	Address     string `json:"address"`
	Amount      string `json:"amount"`
	Timestamp   int64  `json:"timestamp"`
	ForceChain  int    `json:"forceChain"`
	AccountType string `json:"accountType"`
}

type WithdrawByBitResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		ID string `json:"id"`
	} `json:"result"`
	Time int64 `json:"time"`
}

type WithdrawOkxRequest struct {
	Amt    string `json:"amt"`
	Dest   string `json:"dest"`
	Ccy    string `json:"ccy"`
	Chain  string `json:"chain"`
	ToAddr string `json:"toAddr"`
}

type WithdrawOkxResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Ccy   string `json:"ccy"`
		Chain string `json:"chain"`
		Amt   string `json:"amt"`
		WdID  string `json:"wdId"`
		To    string `json:"to"`
		Dest  string `json:"dest"`
		TS    string `json:"ts"`
	} `json:"data"`
}
