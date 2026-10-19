package dto

import "time"

type CreateWallet struct {
	Eth struct {
		Address string `json:"address"`
	} `json:"eth"`
	Sol struct {
		Address string `json:"address"`
	} `json:"sol"`
	Btc struct {
		Address string `json:"address"`
	} `json:"btc"`
}

type SelectKeys struct {
	Eth string `json:"eth"`
	Sol string `json:"sol"`
	Btc string `jsno:"btc"`
}

type BlockbookTx struct {
	Txid        string `json:"txid"`
	BlockTime   int64  `json:"blockTime"`
	BlockHeight int64  `json:"blockHeight"`
}

type BlockbookResponse struct {
	Page          int           `json:"page"`
	TotalPages    int           `json:"totalPages"`
	Address       string        `json:"address"`
	Balance       string        `json:"balance"`
	TotalReceived string        `json:"totalReceived"`
	TotalSent     string        `json:"totalSent"`
	Txs           int           `json:"txs"`
	Transactions  []BlockbookTx `json:"transactions"`
}

type Operation struct {
	Type      string    // "get" или "send"
	Asset     string    // "ETH", "BNB", "SOL"
	Value     string    // Сумма операции
	Remainder string    // Текущий (актуальный) баланс актива в кошельке
	TxHash    string    // Хеш транзакции
	Timestamp time.Time // Когда произошло
}

type UserCexExchangeKeys struct {
	UserID              string    `json:"string"`
	CexName             string    `json:"cex_name"`
	EncryptedApiKey     string    `json:"encrypted_api_key"`
	EncryptedSecretKey  string    `json:"encrypted_secret_key"`
	EnrcyptedPassphrase string    `json:"encrypted_passphrase,omitempty"`
	EncryptedDek        string    `json:"encrypted_dek"`
	At                  time.Time `json:"at"`
}

type SendTransaction struct {
	Net          string `json:"net"`
	Type         string `json:"type"`
	TXhex        string `json:"tx_hex"`
	ExchangeName string `json:"exchange_name,omitempty"`
	CCY          string `json:"ccy,omitempty"`
	Chain        string `json:"chain,omitempty"`
	ToAddress    string `json:"to_address,omitempty"`
	Amount       string `json:"amount,omitempty"`
	Dest         string `json:"dest,omitempty"`
}

type ConvertCurrency struct {
	Min  string `json:"min"`
	Max  string `json:"max"`
	CCY  string `json:"ccy"`
	Icon string `json:"icon"`
}

type OKXConvertCurrencies struct {
	Code string        `json:"code"`
	Data []OKXCurrency `json:"data"`
	Msg  string        `json:"msg"`
}

type OKXCurrency struct {
	Min string `json:"min"`
	Max string `json:"max"`
	Ccy string `json:"ccy"`
}

type BybitConvertCoinListResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		Coins []struct {
			Coin               string `json:"coin"`
			SingleFromMinLimit string `json:"singleFromMinLimit"`
			SingleFromMaxLimit string `json:"singleFromMaxLimit"`
			DisableFrom        bool   `json:"disableFrom"`
		} `json:"coins"`
	} `json:"result"`
}

type BybitQuoteResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		QuoteTxId    string `json:"quoteTxId"`
		ExchangeRate string `json:"exchangeRate"`
		FromCoin     string `json:"fromCoin"`
		ToCoin       string `json:"toCoin"`
		FromAmount   string `json:"fromAmount"`
		ToAmount     string `json:"toAmount"`
		ExpiredTime  string `json:"expiredTime"`
	} `json:"result"`
}

type OKXEstimateQuoteResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		QuoteId  string `json:"quoteId"`
		CnvtPx   string `json:"cnvtPx"`
		BaseCcy  string `json:"baseCcy"`
		QuoteCcy string `json:"quoteCcy"`
		BaseSz   string `json:"baseSz"`
		QuoteSz  string `json:"quoteSz"`
		TtlMs    string `json:"ttlMs"`
	} `json:"data"`
}

type GetConvertQuote struct {
	ExchangeName string `json:"exchange_name"`
	FromCoin     string `json:"from_coin"`
	ToCoin       string `json:"to_coin"`
	Amount       string `json:"amount"`
}

type ConfirmConvertQuote struct {
	ExchangeName string `json:"exchange_name"`
	QuoteTxId    string `json:"quote_tx_id"`
}

type BybitConfirmQuoteResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		QuoteTxId      string `json:"quoteTxId"`
		ExchangeStatus string `json:"exchangeStatus"` // init / processing / success / failure
	} `json:"result"`
}

type DeleteApiKeys struct {
	ExchangeName string `json:"exchange_name"`
}
