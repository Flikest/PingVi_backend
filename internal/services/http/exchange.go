package servicehttp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/exchange"
	"github.com/Flikest/PingVi_backend/pkg/exchange/formatier"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var exchangeUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ExchangeCall struct {
	Operation  string `json:"operation"`
	Exchange   string `json:"exchange"`
	Instrument string `json:"instrument"`
	Timeframe  string `json:"timeframe"`
}

type BinanceSubRequest struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     string   `json:"id"`
}

type BybitKlineResponse struct {
	Topic string `json:"topic"`
	Type  string `json:"type"`
	Data  []struct {
		Start   string `json:"start"`
		End     string `json:"end"`
		Open    string `json:"open"`
		High    string `json:"high"`
		Low     string `json:"low"`
		Close   string `json:"close"`
		Volume  string `json:"volume"`
		Confirm bool   `json:"confirm"`
	} `json:"data"`
}

type ByBitTransactionLogResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List []struct {
			ID              string `json:"id"`
			Symbol          string `json:"symbol"`
			Category        string `json:"category"`
			Side            string `json:"side"`
			TransactionTime string `json:"transactionTime"`
			Type            string `json:"type"`
			Qty             string `json:"qty"`
			Size            string `json:"size"`
			Currency        string `json:"currency"`
			TradePrice      string `json:"tradePrice"`
			Funding         string `json:"funding"`
			Fee             string `json:"fee"`
			CashFlow        string `json:"cashFlow"`
			Change          string `json:"change"`
			CashBalance     string `json:"cashBalance"`
		} `json:"list"`
		NextPageCursor string `json:"nextPageCursor"`
	} `json:"result"`
	Time int64 `json:"time"`
}

type OKXBillsResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		BillId string `json:"billId"`
		InstId string `json:"instId"`
		Type   string `json:"type"`
		Ccy    string `json:"ccy"`
		BalChg string `json:"balChg"`
		Bal    string `json:"bal"`
		Sz     string `json:"sz"`
		Px     string `json:"px"`
		Ts     string `json:"ts"`
	} `json:"data"`
}

type Transactions struct {
	ID        string `json:"id"`
	Exchange  string `json:"exchange"` // "bybit" / "okx"
	Symbol    string `json:"symbol"`
	Side      string `json:"side"` // "buy" / "sell" / ""
	Type      string `json:"type"` // "trade" / "funding" / "fee" / "transfer"
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	Fee       string `json:"fee"`
	Funding   string `json:"funding"`
	PnL       string `json:"pnl"`
	Balance   string `json:"balance"`
	Timestamp int64  `json:"timestamp"`
}

type BinanceKlineResponse struct {
	Stream string `json:"stream"` // Название потока, например "bnbusdt@kline_5m"
	Data   struct {
		EventType string `json:"e"` // Тип события: "kline"
		EventTime int64  `json:"E"` // Время события
		Symbol    string `json:"s"` // Торговая пара (BNBUSDT)
		Kline     struct {
			Start    string `json:"t"` // Время старта свечи (timestamp ms)
			End      string `json:"T"` // Время окончания свечи (timestamp ms)
			Interval string `json:"i"` // Интервал (например, "5m")
			Open     string `json:"o"` // Цена открытия
			Close    string `json:"c"` // Цена закрытия (текущая цена)
			High     string `json:"h"` // Максимум
			Low      string `json:"l"` // Minimum
			Volume   string `json:"v"` // Объем базового актива
			IsClosed bool   `json:"x"` // true означает, что свеча ЗАКРЫЛАСЬ
		} `json:"k"`
	} `json:"data"`
}

type OKXCandle struct {
	Ts          string // время старта свечи, ms
	Open        string // цена открытия
	High        string // максимум
	Low         string // минимум
	Close       string // цена закрытия (текущая)
	Vol         string // объём в базовой валюте
	VolCcy      string // объём в котируемой валюте
	VolCcyQuote string // объём в quote-валюте (USDT)
	Confirm     bool   // true = свеча закрыта
}

type OKXSubRequest struct {
	Op   string   `json:"op"` // Всегда "subscribe"
	Args []OKXArg `json:"args"`
}

type OKXArg struct {
	Channel string `json:"channel"` // Например, "candle5m"
	InstID  string `json:"instId"`  // Например, "BTC-USDT"
}

type OKXCandleResponse struct {
	Arg  OKXArg      `json:"arg"`
	Data []OKXCandle `json:"data"`
}

type SubscriptionRequest struct {
	Op   string `json:"op"`
	Args []any  `json:"args"`
}

type Quotes struct {
	Instrument string                    `json:"instrument"`
	Data       []dto.CandleStreamDataset `json:"data"`
}

type BalanceResponse struct {
	TotalEquity  string      `json:"totalEquity"`  // общий капитал в USD (одно число)
	TotalPerpUPL string      `json:"totalPerpUPL"` // P&L по перпетуалам
	Coins        []CoinQuote `json:"coins"`
}

type CoinQuote struct {
	ExchangeName  string `json:"exchange_name"`
	Coin          string `json:"coin"`
	Icon          string `json:"icon"`
	WalletBalance string `json:"walletBalance"`
	UsdValue      string `json:"usdValue"`
	UnrealisedPnl string `json:"unrealisedPnl"`
	AvgEntryPrice string `json:"avgEntryPrice,omitempty"`
}

type OKXBalanceResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		TotalEq  string `json:"totalEq"`
		IsoEq    string `json:"isoEq"`
		AdjEq    string `json:"adjEq"`
		AvailEq  string `json:"availEq"`
		OrdFroz  string `json:"ordFroz"`
		Imr      string `json:"imr"`
		Mmr      string `json:"mmr"`
		MgnRatio string `json:"mgnRatio"`
		Upl      string `json:"upl"`
		Details  []struct {
			Ccy       string `json:"ccy"`
			Eq        string `json:"eq"`
			CashBal   string `json:"cashBal"`
			Upl       string `json:"upl"`
			EqUsd     string `json:"eqUsd"`
			AvailBal  string `json:"availBal"`
			FrozenBal string `json:"frozenBal"`
			AvgPx     string `json:"avgPx"`
			OpenAvgPx string `json:"openAvgPx"`
			AccAvgPx  string `json:"accAvgPx"`
		} `json:"details"`
	} `json:"data"`
}

type ByBitWalletRestResponse struct {
	RetCode    int      `json:"retCode"`
	RetMsg     string   `json:"retMsg"`
	RetExtInfo struct{} `json:"retExtInfo"`
	Time       int64    `json:"time"`
	Result     struct {
		List []struct {
			TotalEquity                string `json:"totalEquity"`
			AccountIMRate              string `json:"accountIMRate"`
			AccountIMRateByMp          string `json:"accountIMRateByMp"`
			TotalMarginBalance         string `json:"totalMarginBalance"`
			TotalInitialMargin         string `json:"totalInitialMargin"`
			TotalInitialMarginByMp     string `json:"totalInitialMarginByMp"`
			AccountType                string `json:"accountType"`
			TotalAvailableBalance      string `json:"totalAvailableBalance"`
			AccountMMRate              string `json:"accountMMRate"`
			AccountMMRateByMp          string `json:"accountMMRateByMp"`
			TotalPerpUPL               string `json:"totalPerpUPL"`
			TotalWalletBalance         string `json:"totalWalletBalance"`
			AccountLTV                 string `json:"accountLTV"`
			TotalMaintenanceMargin     string `json:"totalMaintenanceMargin"`
			TotalMaintenanceMarginByMp string `json:"totalMaintenanceMarginByMp"`
			Coin                       []struct {
				AvailableToBorrow   string `json:"availableToBorrow"`
				Bonus               string `json:"bonus"`
				AccruedInterest     string `json:"accruedInterest"`
				AvailableToWithdraw string `json:"availableToWithdraw"`
				TotalOrderIM        string `json:"totalOrderIM"`
				Equity              string `json:"equity"`
				TotalPositionMM     string `json:"totalPositionMM"`
				UsdValue            string `json:"usdValue"`
				SpotHedgingQty      string `json:"spotHedgingQty"`
				UnrealisedPnl       string `json:"unrealisedPnl"`
				CollateralSwitch    bool   `json:"collateralSwitch"`
				BorrowAmount        string `json:"borrowAmount"`
				TotalPositionIM     string `json:"totalPositionIM"`
				WalletBalance       string `json:"walletBalance"`
				CumRealisedPnl      string `json:"cumRealisedPnl"`
				Locked              string `json:"locked"`
				MarginCollateral    bool   `json:"marginCollateral"`
				Coin                string `json:"coin"`
				SpotBorrow          string `json:"spotBorrow"`
			} `json:"coin"`
		} `json:"list"`
	} `json:"result"`
}

type ExchangeClient struct {
	ID   uuid.UUID
	Conn *websocket.Conn
	Send chan any
}

type ExchangeHub struct {
	Online       map[string]*ExchangeClient
	ExchangeMap  map[string]map[string]map[string]map[string]*ExchangeClient // название биржи/торговая пара/таймфрейм/клиенты которые к ней подключенны
	Log          *slog.Logger
	OKXConnect   *websocket.Conn
	ByBitConnect *websocket.Conn
	mu           sync.RWMutex
}

func mergeCoin(m map[string]*CoinQuote, c CoinQuote) {
	if existing, ok := m[c.Coin]; ok {
		existing.WalletBalance = addDecimal(existing.WalletBalance, c.WalletBalance)
		existing.UsdValue = addDecimal(existing.UsdValue, c.UsdValue)
		existing.UnrealisedPnl = addDecimal(existing.UnrealisedPnl, c.UnrealisedPnl)
		if existing.AvgEntryPrice == "" {
			existing.AvgEntryPrice = c.AvgEntryPrice
		}
	} else {
		cc := c
		m[c.Coin] = &cc
	}
}

func addDecimal(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	if errA != nil || errB != nil {
		return a
	}
	return strconv.FormatFloat(fa+fb, 'f', -1, 64)
}

func mapOKXSide(billType string) string {
	switch billType {
	case "1", "3", "4", "204", "206", "207":
		return "buy"
	case "2", "5", "6", "205", "208", "209":
		return "sell"
	}
	return ""
}

func mapOKXType(billType string) string {
	switch billType {
	case "1", "2", "3", "4", "5", "6", "204", "205", "206", "207", "208", "209":
		return "trade"
	case "8":
		return "funding"
	case "22":
		return "fee"
	}
	return "other"
}

func NewExchangeHub(log *slog.Logger) *ExchangeHub {
	return &ExchangeHub{
		ExchangeMap:  make(map[string]map[string]map[string]map[string]*ExchangeClient),
		OKXConnect:   nil,
		ByBitConnect: nil,
		Log:          log,
	}
}

func (eh *ExchangeHub) openByBitConnect(client *ExchangeClient, instrument, timeframe string) {
	instrument = strings.ToUpper(strings.ReplaceAll(instrument, "/", ""))
	if eh.ByBitConnect != nil {
		subReq := SubscriptionRequest{
			Op:   "subscribe",
			Args: []any{"kline." + timeframe + "." + instrument},
		}
		reqBytes, _ := json.Marshal(subReq)
		if err := eh.ByBitConnect.WriteMessage(websocket.TextMessage, reqBytes); err != nil {
			eh.Log.Error("error with ws bybit :", "error", err)

		}
		eh.mu.Lock()
		eh.ExchangeMap["bybit"][instrument][timeframe][client.ID.String()] = client
		eh.mu.Unlock()
		return
	}

	u := "wss://stream.bybit.com/v5/public/spot"
	log.Printf("Подключение к %s", u)

	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	eh.ByBitConnect = c
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Ошибка чтения: %v", err)
				return
			}

			var response BybitKlineResponse
			if err := json.Unmarshal(message, &response); err != nil {
				continue
			}

			go func() {
				var data []dto.CandleStreamDataset

				for i := 0; i < len(response.Data); i++ {
					data = append(data, dto.CandleStreamDataset{
						Time:    response.Data[i].Start,
						Open:    response.Data[i].Open,
						High:    response.Data[i].High,
						Low:     response.Data[i].Low,
						Close:   response.Data[i].Close,
						Volume:  response.Data[i].Volume,
						Confirm: response.Data[i].Confirm,
					})
				}

				eh.broadcast("bybit", instrument, timeframe, Quotes{
					Instrument: response.Topic,
					Data:       data,
				})
			}()
		}
	}()

	subReq := SubscriptionRequest{
		Op:   "subscribe",
		Args: []any{"kline." + timeframe + "." + instrument},
	}

	reqBytes, err := json.Marshal(subReq)
	if err != nil {
		log.Fatalf("Ошибка кодирования запроса подписки: %v", err)
	}

	err = c.WriteMessage(websocket.TextMessage, reqBytes)
	if err != nil {
		log.Fatalf("Ошибка отправки подписки: %v", err)
	}
	log.Println("Запрос на подписку отправлен:", "kline."+timeframe+"."+instrument)

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-done:
			log.Println("Соединение закрыто удаленным сервером.")
			return
		case <-ticker.C:

			pingReq := SubscriptionRequest{Op: "ping"}
			pingBytes, _ := json.Marshal(pingReq)

			err := c.WriteMessage(websocket.TextMessage, pingBytes)
			if err != nil {
				eh.Log.Error("Ошибка отправки ping: %v", "error", err)
				return
			}
		case <-interrupt:
			eh.Log.Info("Получен сигнал прерывания. Закрываем соединение...")

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Ошибка при отправке Close-фрейма: %v", err)
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (eh *ExchangeHub) openOkxConnect(client *ExchangeClient, instrument, timeframe string) {
	instrument = strings.ToUpper(strings.ReplaceAll(instrument, "/", "-"))

	if eh.OKXConnect != nil {
		subReq := OKXSubRequest{
			Op:   "subscribe",
			Args: []OKXArg{{Channel: "candle" + timeframe, InstID: instrument}},
		}
		reqBytes, _ := json.Marshal(subReq)
		if err := eh.OKXConnect.WriteMessage(websocket.TextMessage, reqBytes); err != nil {
			eh.Log.Error("error with ws okx: ")
		}
		eh.mu.Lock()
		eh.ExchangeMap["okx"][instrument][timeframe][client.ID.String()] = client
		eh.mu.Unlock()
		return
	}

	u := "wss://ws.okx.com:8443/ws/v5/business"
	log.Printf("Подключение к OKX WebSocket: %s", u)

	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	eh.OKXConnect = c
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Ошибка чтения: %v", err)
				return
			}
			if string(message) == "pong" {
				continue
			}

			var response OKXCandleResponse
			if err := json.Unmarshal(message, &response); err != nil {
				continue
			}

			if len(response.Data) > 0 {

				go func() {

					var data []dto.CandleStreamDataset

					for i := 0; i < len(response.Data); i++ {
						data = append(data, dto.CandleStreamDataset{
							Time:    response.Data[i].Ts,
							Open:    response.Data[i].Open,
							High:    response.Data[i].High,
							Low:     response.Data[i].Low,
							Close:   response.Data[i].Close,
							Volume:  response.Data[i].Vol,
							Confirm: response.Data[i].Confirm,
						})
					}

					eh.broadcast("okx", response.Arg.InstID, strings.TrimPrefix(response.Arg.Channel, "candle"), Quotes{
						Instrument: response.Arg.InstID,
						Data:       data,
					})
				}()

			} else {
				fmt.Printf("Системное сообщение от OKX: %s\n", string(message))
			}
		}
	}()

	subReq := OKXSubRequest{
		Op: "subscribe",
		Args: []OKXArg{
			{
				Channel: "candle" + timeframe,
				InstID:  instrument,
			},
		},
	}

	reqBytes, err := json.Marshal(subReq)
	if err != nil {
		log.Fatalf("Ошибка кодирования запроса подписки: %v", err)
	}

	err = c.WriteMessage(websocket.TextMessage, reqBytes)
	if err != nil {
		log.Fatalf("Ошибка отправки подписки: %v", err)
	}
	log.Println("Запрос на подписку отправлен для", instrument, "candle"+timeframe)

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-done:
			log.Println("Соединение закрыто сервером OKX.")
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				log.Printf("Ошибка отправки ping: %v", err)
				return
			}
		case <-interrupt:
			log.Println("Получен сигнал прерывания. Отключаемся...")

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Ошибка отправки Close фрейма: %v", err)
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (eh *ExchangeHub) removeClientFromAllInstruments(clientToRemove string) {
	if clientToRemove == "" {
		return
	}

	for _, pairs := range eh.ExchangeMap {
		for _, timeframes := range pairs {
			for _, clients := range timeframes {
				for clientID, client := range clients {
					if client.ID.String() == clientToRemove {
						delete(clients, clientID)
					}
				}
			}
		}
	}
}

func (eh *ExchangeHub) Unsubscribe(exchange string, instrument string, timeframe string, client *ExchangeClient) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	exchangeMap, ok := eh.ExchangeMap[exchange]
	if !ok {
		return
	}

	instrumentMap, ok := exchangeMap[instrument]
	if !ok {
		return
	}

	timeframeMap, ok := instrumentMap[timeframe]
	if !ok {
		return
	}

	clientID := client.ID.String()
	if _, ok := timeframeMap[clientID]; !ok {
		return
	}

	delete(timeframeMap, clientID)

	if len(timeframeMap) == 0 {
		delete(instrumentMap, timeframe)

		if len(instrumentMap) == 0 {
			delete(exchangeMap, instrument)

			if len(exchangeMap) == 0 {
				delete(eh.ExchangeMap, exchange)
				if exchange == "bybit" {
					eh.ByBitConnect.Close()
					eh.ByBitConnect = nil
				}
				eh.OKXConnect.Close()
				eh.OKXConnect = nil
			}
		}
	}

	eh.Log.Debug("client unsubscribed",
		"client_id", client.ID,
		"exchange", exchange,
		"instrument", instrument,
		"timeframe", timeframe,
	)
}

func (eh *ExchangeHub) broadcast(exchange string, instrument string, timeframe string, quotes Quotes) {
	eh.mu.RLock()

	exchangeMap, exchangeExists := eh.ExchangeMap[exchange]
	if !exchangeExists {
		eh.mu.RUnlock()
		return
	}

	clientMap, instrumentExists := exchangeMap[instrument][timeframe]
	if !instrumentExists || len(clientMap) == 0 {
		eh.mu.RUnlock()
		return
	}

	var clients []*ExchangeClient
	for client := range clientMap {
		clients = append(clients, clientMap[client])
	}
	eh.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- quotes:

		default:
			eh.Log.Warn("Client is too slow, disconnecting",
				"client_id", client.ID,
				"exchange", exchange,
				"instrument", instrument,
			)

			go eh.removeClientFromAllInstruments(client.ID.String())
		}
	}
}

func (eh *ServiceWallet) writeToClient(ctx context.Context, exchangeClient *ExchangeClient) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			exchangeClient.Conn.WriteJSON("ping")
		case quotes, ok := <-exchangeClient.Send:
			if !ok {
				return
			}

			exchangeClient.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := exchangeClient.Conn.WriteJSON(quotes); err != nil {
				eh.Log.Error("error writing quotes to client", "error", err)
				break loop
			}
		}
	}
}

func (eh *ServiceWallet) readClient(ctx context.Context, exchangeClient *ExchangeClient) {
	readCtx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		cancel()
		ticker.Stop()
	}()

	go func() {
		<-readCtx.Done()
		exchangeClient.Conn.Close()
	}()

	for {
		var msg ExchangeCall
		if err := exchangeClient.Conn.ReadJSON(&msg); err != nil {
			if readCtx.Err() != nil {
				return
			}
			eh.Log.Error("error with reading message", "error", err)
			return
		}

		switch msg.Operation {

		case "pong":
			ticker = time.NewTicker(10 * time.Second)
		case "subscribe":
			if msg.Exchange == "bybit" {
				go eh.Hub.openByBitConnect(exchangeClient, msg.Instrument, msg.Timeframe)
			} else {
				eh.Hub.openOkxConnect(exchangeClient, msg.Instrument, msg.Timeframe)
			}
		case "unsubscribe":
			eh.Hub.Unsubscribe(msg.Exchange, msg.Instrument, msg.Timeframe, exchangeClient)
		}
	}
}

func (s ServiceWallet) Handshake(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reqCtx := ctx.Request.Context()

	wsConn, err := exchangeUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		s.Log.Error("WebSocket upgrade failed", "error", err)
		return
	}

	client := &ExchangeClient{
		ID:   userID,
		Conn: wsConn,
		Send: make(chan any, 256),
	}

	s.Hub.mu.RLock()
	s.Hub.Online[userID.String()] = client
	s.Hub.mu.RUnlock()

	balance, err := s.getCeFiBalance(reqCtx, userID)
	if err != nil {
		s.Log.Error("error with get user balance: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch balance"})
		return
	}

	for _, coin := range balance.Coins {
		if coin.ExchangeName == "bybit" {
			s.Hub.openByBitConnect(client, coin.Coin, "1")
		} else {
			s.Hub.openOkxConnect(client, coin.Coin, "index-candle1m")
		}
	}

	go s.writeToClient(reqCtx, client)

	s.readClient(reqCtx, client)
}

func (s *ServiceWallet) Searchinstruments(ctx *gin.Context) {
	instrument := ctx.Param("instrument")
	if instrument == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "instrument is required"})
		return
	}

	before, _, _ := strings.Cut(instrument, "/")

	okxSymbol := strings.ReplaceAll(instrument, "/", "-")
	bybitSymbol := strings.ReplaceAll(instrument, "/", "")

	var wg sync.WaitGroup
	var mu sync.Mutex
	var result []dto.SearchInstrumentResponse
	var errs []error

	wg.Add(2)

	go func() {
		defer wg.Done()

		okxInstruments, err := exchange.GetOKXInstruments(okxSymbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("OKX instruments: %w", err))
			mu.Unlock()
			return
		}

		ohlc, err := exchange.GetOKXCandles(okxSymbol, "", "")
		if err != nil {
			s.Log.Error("error with fetching okx candles: ", "error", err)
			mu.Lock()
			errs = append(errs, fmt.Errorf("OKX candles: %w", err))
			mu.Unlock()
			return
		}

		if len(okxInstruments.Data) == 0 || len(ohlc) == 0 {
			mu.Lock()
			errs = append(errs, fmt.Errorf("OKX: no data available"))
			mu.Unlock()
			return
		}

		mu.Lock()
		result = append(result, dto.SearchInstrumentResponse{
			ExchangeName:   "OKX",
			Symbol:         okxInstruments.Data[0].InstId,
			LogoURL:        fmt.Sprintf("https://img.logo.dev/crypto/%s?token=%s", before, os.Getenv("DEV_LOGO_PK")),
			OHLCLastCandle: ohlc[len(ohlc)-1],
		})
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()

		bybitInstruments, err := exchange.GetByBitInstruments(bybitSymbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("ByBit instruments: %w", err))
			mu.Unlock()
			return
		}

		ohlc, err := exchange.GeByBitCandles(bybitSymbol, "", "")
		if err != nil {
			s.Log.Error("error with fetching bybit candles: ", "error", err)
			mu.Lock()
			errs = append(errs, fmt.Errorf("ByBit candles: %w", err))
			mu.Unlock()
			return
		}

		if len(bybitInstruments.Symbols) == 0 || len(ohlc) == 0 {
			mu.Lock()
			errs = append(errs, fmt.Errorf("ByBit: no data available"))
			mu.Unlock()
			return
		}

		mu.Lock()
		result = append(result, dto.SearchInstrumentResponse{
			ExchangeName:   "ByBit",
			Symbol:         bybitInstruments.Symbols[0].Symbol,
			LogoURL:        fmt.Sprintf("https://img.logo.dev/crypto/%s?token=%s", before, os.Getenv("DEV_LOGO_PK")),
			OHLCLastCandle: ohlc[len(ohlc)-1],
		})
		mu.Unlock()
	}()

	wg.Wait()

	if len(errs) > 0 {
		for _, err := range errs {
			s.Log.Error("error in search instruments: ", "error", err)
		}
	}

	if len(result) == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch data from all exchanges"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (s *ServiceWallet) GetCandles(ctx *gin.Context) {
	instrument := ctx.Query("instrument")
	if instrument == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "instrument is required"})
		return
	}

	exchangeName := ctx.Query("exchange")
	if exchangeName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "exchange is required"})
		return
	}

	interval := ctx.Query("interval")
	after := ctx.Query("after")

	interval = formatier.InstrumentFormatier(exchangeName, instrument)

	switch exchangeName {
	case "okx":
		candles, err := exchange.GetOKXCandles(instrument, interval, after)
		if err != nil {
			s.Log.Error("error with fetching okx candles: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch OKX candles"})
			return
		}

		ctx.JSON(http.StatusOK, candles)
	case "bybit":
		candles, err := exchange.GeByBitCandles(instrument, interval, after)
		if err != nil {
			s.Log.Error("error with fetching bybit candles: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ByBit candles"})
			return
		}

		ctx.JSON(http.StatusOK, candles)
	case "binance":
		candles, err := exchange.GetBinanceCandles(instrument, interval, after)
		if err != nil {
			s.Log.Error("error with fetching binance candles: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch Binance candles"})
			return
		}

		ctx.JSON(http.StatusOK, candles)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange"})
		return
	}
}

func (s *ServiceWallet) GetOrderBook(ctx *gin.Context) {
	instrument := ctx.Query("instrument")
	if instrument == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "instrument is required"})
		return
	}

	exchangeName := ctx.Query("exchange")
	if exchangeName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "exchange is required"})
		return
	}

	interval := ctx.Query("interval")
	after := ctx.Query("after")

	interval = formatier.InstrumentFormatier(exchangeName, instrument)

	switch exchangeName {
	case "okx":
		candles, err := exchange.GetOKXCandles(instrument, interval, after)
		if err != nil {
			s.Log.Error("error with fetching okx candles: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch OKX candles"})
			return
		}

		ctx.JSON(http.StatusOK, candles)
	case "bybit":
		candles, err := exchange.GeByBitCandles(instrument, interval, after)
		if err != nil {
			s.Log.Error("error with fetching bybit candles: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ByBit candles"})
			return
		}

		ctx.JSON(http.StatusOK, candles)
	case "binance":
		candles, err := exchange.GetBinanceCandles(instrument, interval, after)
		if err != nil {
			s.Log.Error("error with fetching binance candles: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch Binance candles"})
			return
		}

		ctx.JSON(http.StatusOK, candles)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange"})
		return
	}
}
