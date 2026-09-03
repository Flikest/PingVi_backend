package dto

type ByBitExchangeInfo struct {
	Timezone        string           `json:"timezone"`
	ServerTime      int64            `json:"serverTime"`
	RateLimits      []ByBitRateLimit `json:"rateLimits"`
	ExchangeFilters []interface{}    `json:"exchangeFilters"`
	Symbols         []ByBitSymbol    `json:"symbols"`
}

type ByBitRateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
}

type ByBitSymbol struct {
	Symbol                          string        `json:"symbol"`
	Status                          string        `json:"status"`
	BaseAsset                       string        `json:"baseAsset"`
	BaseAssetPrecision              int           `json:"baseAssetPrecision"`
	QuoteAsset                      string        `json:"quoteAsset"`
	QuotePrecision                  int           `json:"quotePrecision"`
	QuoteAssetPrecision             int           `json:"quoteAssetPrecision"`
	BaseCommissionPrecision         int           `json:"baseCommissionPrecision"`
	QuoteCommissionPrecision        int           `json:"quoteCommissionPrecision"`
	OrderTypes                      []string      `json:"orderTypes"`
	IcebergAllowed                  bool          `json:"icebergAllowed"`
	OcoAllowed                      bool          `json:"ocoAllowed"`
	OtoAllowed                      bool          `json:"otoAllowed"`
	OpoAllowed                      bool          `json:"opoAllowed"`
	QuoteOrderQtyMarketAllowed      bool          `json:"quoteOrderQtyMarketAllowed"`
	AllowTrailingStop               bool          `json:"allowTrailingStop"`
	CancelReplaceAllowed            bool          `json:"cancelReplaceAllowed"`
	AmendAllowed                    bool          `json:"amendAllowed"`
	PegInstructionsAllowed          bool          `json:"pegInstructionsAllowed"`
	IsSpotTradingAllowed            bool          `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed          bool          `json:"isMarginTradingAllowed"`
	Filters                         []ByBitFilter `json:"filters"`
	Permissions                     []string      `json:"permissions"`
	PermissionSets                  [][]string    `json:"permissionSets"`
	DefaultSelfTradePreventionMode  string        `json:"defaultSelfTradePreventionMode"`
	AllowedSelfTradePreventionModes []string      `json:"allowedSelfTradePreventionModes"`
}

type ByBitFilter struct {
	FilterType string `json:"filterType"`

	MinPrice string `json:"minPrice,omitempty"`
	MaxPrice string `json:"maxPrice,omitempty"`
	TickSize string `json:"tickSize,omitempty"`

	MinQty   string `json:"minQty,omitempty"`
	MaxQty   string `json:"maxQty,omitempty"`
	StepSize string `json:"stepSize,omitempty"`

	Limit int `json:"limit,omitempty"`

	MinTrailingAboveDelta int `json:"minTrailingAboveDelta,omitempty"`
	MaxTrailingAboveDelta int `json:"maxTrailingAboveDelta,omitempty"`
	MinTrailingBelowDelta int `json:"minTrailingBelowDelta,omitempty"`
	MaxTrailingBelowDelta int `json:"maxTrailingBelowDelta,omitempty"`

	BidMultiplierUp   string `json:"bidMultiplierUp,omitempty"`
	BidMultiplierDown string `json:"bidMultiplierDown,omitempty"`
	AskMultiplierUp   string `json:"askMultiplierUp,omitempty"`
	AskMultiplierDown string `json:"askMultiplierDown,omitempty"`
	AvgPriceMins      int    `json:"avgPriceMins,omitempty"`

	MinNotional      string `json:"minNotional,omitempty"`
	MaxNotional      string `json:"maxNotional,omitempty"`
	ApplyMinToMarket bool   `json:"applyMinToMarket,omitempty"`
	ApplyMaxToMarket bool   `json:"applyMaxToMarket,omitempty"`

	MaxNumOrders int `json:"maxNumOrders,omitempty"`

	MaxNumOrderLists int `json:"maxNumOrderLists,omitempty"`

	MaxNumAlgoOrders int `json:"maxNumAlgoOrders,omitempty"`

	MaxNumOrderAmends int `json:"maxNumOrderAmends,omitempty"`
}

type BybitResponse struct {
	RetCode    int         `json:"retCode"`
	RetMsg     string      `json:"retMsg"`
	Result     BybitResult `json:"result"`
	RetExtInfo interface{} `json:"retExtInfo"`
	Time       int64       `json:"time"`
}

type BybitResult struct {
	Symbol string           `json:"symbol"`
	List   []BybitCandleRow `json:"list"`
}

type BybitCandleRow struct {
	Timestamp int64  `json:"-"`
	Open      string `json:"-"`
	High      string `json:"-"`
	Low       string `json:"-"`
	Close     string `json:"-"`
}

type BybitOrderBookResponse struct {
	RetCode    int                  `json:"retCode"`
	RetMsg     string               `json:"retMsg"`
	Result     BybitOrderBookResult `json:"result"`
	RetExtInfo interface{}          `json:"retExtInfo"`
	Time       int64                `json:"time"`
}

type BybitOrderBookResult struct {
	Symbol    string      `json:"s"`
	Asks      [][2]string `json:"a"`
	Bids      [][2]string `json:"b"`
	Timestamp int64       `json:"ts"`
	UpdateID  int64       `json:"u"`
	Sequence  int64       `json:"seq"`
	CTime     int64       `json:"cts"`
}
