package dto

type OKXInstrument struct {
	Code string `json:"code"`
	Data []Data `json:"data"`
}

type Data struct {
	Alias                  string        `json:"alias"`
	AuctionEndTime         string        `json:"auctionEndTime"`
	BaseCcy                string        `json:"baseCcy"`
	Category               string        `json:"category"`
	ContTdSwTime           string        `json:"contTdSwTime"`
	CtMult                 string        `json:"ctMult"`
	CtType                 string        `json:"ctType"`
	CtVal                  string        `json:"ctVal"`
	CtValCcy               string        `json:"ctValCcy"`
	ExpTime                string        `json:"expTime"`
	FloatPxLmtPct          string        `json:"floatPxLmtPct"`
	Freq                   string        `json:"freq"`
	FutureSettlement       bool          `json:"futureSettlement"`
	GroupId                string        `json:"groupId"`
	InitPxLmtPct           string        `json:"initPxLmtPct"`
	InstCategory           string        `json:"instCategory"`
	InstFamily             string        `json:"instFamily"`
	InstId                 string        `json:"instId"`
	InstIdCode             int           `json:"instIdCode"`
	InstType               string        `json:"instType"`
	Lever                  string        `json:"lever"`
	ListTime               string        `json:"listTime"`
	LongPosRemainingQuota  string        `json:"longPosRemainingQuota"`
	LotSz                  string        `json:"lotSz"`
	MaxIcebergSz           string        `json:"maxIcebergSz"`
	MaxLmtAmt              string        `json:"maxLmtAmt"`
	MaxLmtSz               string        `json:"maxLmtSz"`
	MaxMktAmt              string        `json:"maxMktAmt"`
	MaxMktSz               string        `json:"maxMktSz"`
	MaxPlatOICoinLmt       string        `json:"maxPlatOICoinLmt"`
	MaxPlatOILmt           string        `json:"maxPlatOILmt"`
	MaxPxLmtPct            string        `json:"maxPxLmtPct"`
	MaxStopSz              string        `json:"maxStopSz"`
	MaxTriggerSz           string        `json:"maxTriggerSz"`
	MaxTwapSz              string        `json:"maxTwapSz"`
	Method                 string        `json:"method"`
	MinSz                  string        `json:"minSz"`
	OpenType               string        `json:"openType"`
	OptType                string        `json:"optType"`
	PosLmtAmt              string        `json:"posLmtAmt"`
	PosLmtPct              string        `json:"posLmtPct"`
	PreMktSwTime           string        `json:"preMktSwTime"`
	QuoteCcy               string        `json:"quoteCcy"`
	RuleType               string        `json:"ruleType"`
	SeriesId               string        `json:"seriesId"`
	SettleCcy              string        `json:"settleCcy"`
	ShortPosRemainingQuota string        `json:"shortPosRemainingQuota"`
	State                  string        `json:"state"`
	Stk                    string        `json:"stk"`
	TickSz                 string        `json:"tickSz"`
	TradeQuoteCcyList      []string      `json:"tradeQuoteCcyList"`
	Uly                    string        `json:"uly"`
	UpcChg                 []interface{} `json:"upcChg"`
}

type OKXKlineResponse struct {
	Code string        `json:"code"`
	Msg  string        `json:"msg"`
	Data [][]OKXCandle `json:"data"`
}

type OKXCandle struct {
	Timestamp int64  `json:"-"`
	Open      string `json:"-"`
	High      string `json:"-"`
	Low       string `json:"-"`
	Close     string `json:"-"`
	Volume    string `json:"-"`
	Confirm   string `json:"-"`
}

type OKXOrderBookResponse struct {
	Code string         `json:"code"`
	Msg  string         `json:"msg"`
	Data []OKXOrderBook `json:"data"`
}

type OKXOrderBook struct {
	Asks  [][4]string `json:"asks"`
	Bids  [][4]string `json:"bids"`
	Ts    string      `json:"ts"`
	SeqID int64       `json:"seqId"`
}
