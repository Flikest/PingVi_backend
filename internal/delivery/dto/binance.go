package dto

type BinanceExchangeInfo struct {
	Timezone        string             `json:"timezone"`
	ServerTime      int64              `json:"serverTime"`
	RateLimits      []BinanceRateLimit `json:"rateLimits"`
	ExchangeFilters []interface{}      `json:"exchangeFilters"`
	Symbols         []BinanceSymbol    `json:"symbols"`
}

type BinanceRateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
}

type BinanceSymbol struct {
	Symbol                          string          `json:"symbol"`
	Status                          string          `json:"status"`
	BaseAsset                       string          `json:"baseAsset"`
	BaseAssetPrecision              int             `json:"baseAssetPrecision"`
	QuoteAsset                      string          `json:"quoteAsset"`
	QuotePrecision                  int             `json:"quotePrecision"`
	QuoteAssetPrecision             int             `json:"quoteAssetPrecision"`
	BaseCommissionPrecision         int             `json:"baseCommissionPrecision"`
	QuoteCommissionPrecision        int             `json:"quoteCommissionPrecision"`
	OrderTypes                      []string        `json:"orderTypes"`
	IcebergAllowed                  bool            `json:"icebergAllowed"`
	OcoAllowed                      bool            `json:"ocoAllowed"`
	OtoAllowed                      bool            `json:"otoAllowed"`
	OpoAllowed                      bool            `json:"opoAllowed"`
	QuoteOrderQtyMarketAllowed      bool            `json:"quoteOrderQtyMarketAllowed"`
	AllowTrailingStop               bool            `json:"allowTrailingStop"`
	CancelReplaceAllowed            bool            `json:"cancelReplaceAllowed"`
	AmendAllowed                    bool            `json:"amendAllowed"`
	PegInstructionsAllowed          bool            `json:"pegInstructionsAllowed"`
	IsSpotTradingAllowed            bool            `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed          bool            `json:"isMarginTradingAllowed"`
	Filters                         []BinanceFilter `json:"filters"`
	Permissions                     []string        `json:"permissions"`
	PermissionSets                  [][]string      `json:"permissionSets"`
	DefaultSelfTradePreventionMode  string          `json:"defaultSelfTradePreventionMode"`
	AllowedSelfTradePreventionModes []string        `json:"allowedSelfTradePreventionModes"`
}

type BinanceFilter struct {
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

type BinanceCandleRaw struct {
	OpenTime       int64  `json:"-"`
	Open           string `json:"-"`
	High           string `json:"-"`
	Low            string `json:"-"`
	Close          string `json:"-"`
	Volume         string `json:"-"`
	CloseTime      int64  `json:"-"`
	QuoteVolume    string `json:"-"`
	TradesCount    int    `json:"-"`
	BuyBaseVolume  string `json:"-"`
	BuyQuoteVolume string `json:"-"`
	Ignore         string `json:"-"`
}

type BinanceOrderBook struct {
	LastUpdateID int64       `json:"lastUpdateId"`
	Bids         [][2]string `json:"bids"`
	Asks         [][2]string `json:"asks"`
}
